package miner

import (
	"errors"
	"fmt"
	"log"

	"github.com/btcsuite/btcd/btcjson"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/rpcclient"
	"github.com/btcsuite/btcutil"
	"github.com/smileprem/go-bitcoin-miner/pkg/config"
	"github.com/smileprem/go-bitcoin-miner/pkg/util"
)

// BitcoinRPCClient defines the interface for interacting with the Bitcoin full node.
// This decouples MinerService from the concrete rpcclient.Client implementation,
// enabling 100% unit test coverage.
type BitcoinRPCClient interface {
	GetBlockChainInfo() (*btcjson.GetBlockChainInfoResult, error)
	GetBlockTemplate(request *btcjson.TemplateRequest) (*btcjson.GetBlockTemplateResult, error)
	SubmitBlock(block *btcutil.Block, options *btcjson.SubmitBlockOptions) error
	Shutdown()
}

// rpcClientNew is a package-level variable wrapping the concrete rpcclient.New constructor.
// Overriding this in unit tests allows mock injection.
var rpcClientNew = func(config *rpcclient.ConnConfig, ntfnHandlers *rpcclient.NotificationHandlers) (BitcoinRPCClient, error) {
	c, err := rpcclient.New(config, ntfnHandlers)
	if err != nil {
		return nil, err
	}
	return c, nil
}

// MinerService coordinates the mining process. It communicates with the local
// Bitcoin node via JSON-RPC, retrieves work, orchestrates the PoW solver,
// and submits solved blocks.
type MinerService struct {
	Config      *config.Config
	RPCClient   BitcoinRPCClient
	ChainParams *chaincfg.Params
}

// NewMinerService initializes a new MinerService, establishes an RPC connection
// to the Bitcoin node, verifies connectivity and sync status, and dynamically
// detects the network parameters (MainNet, TestNet, RegTest, SimNet).
func NewMinerService(appConfig *config.Config) (*MinerService, error) {
	// Configure the local bitcoin core RPC server using HTTP POST mode.
	client, err := rpcClientNew(
		&rpcclient.ConnConfig{
			Host:         fmt.Sprintf("%s:%s", appConfig.Bitcoin.Host, appConfig.Bitcoin.Port),
			User:         appConfig.Bitcoin.Username,
			Pass:         appConfig.Bitcoin.Password,
			HTTPPostMode: true, // Bitcoin core only supports HTTP POST mode
			DisableTLS:   true, // Bitcoin core does not provide TLS by default
		}, nil)
	if err != nil {
		return nil, err
	}

	// Check if local bitcoin rpc server is running
	blockChainInfo, err := client.GetBlockChainInfo()
	if err != nil {
		client.Shutdown()
		return nil, errors.New("Unable to fetch block chain info. Error: " + err.Error())
	}
	log.Println("Connected to bitcoin local RPC server")

	// Check if local bitcoin node is synchronized with the network
	if blockChainInfo.Blocks < blockChainInfo.Headers {
		client.Shutdown()
		return nil, fmt.Errorf(
			"Blocks: %d, Headers: %d. Local bitcoin node is behind by %d blocks.",
			blockChainInfo.Blocks,
			blockChainInfo.Headers,
			blockChainInfo.Headers-blockChainInfo.Blocks,
		)
	}

	// Resolve the network params dynamically based on blockchain info chain name
	params, err := getChainParams(blockChainInfo.Chain)
	if err != nil {
		client.Shutdown()
		return nil, err
	}
	log.Printf("Detected Bitcoin network environment: %s", blockChainInfo.Chain)

	return &MinerService{
		Config:      appConfig,
		RPCClient:   client,
		ChainParams: params,
	}, nil
}

// MineBlock orchestrates the block template retrieval, solves the block Proof of Work
// locally, and submits the completed block back to the Bitcoin full node.
func (ms *MinerService) MineBlock() error {
	// Get the block of transactions from memory pool
	blockTemplateResult, err := ms.RPCClient.GetBlockTemplate(
		&btcjson.TemplateRequest{
			Rules: []string{"segwit"},
		})
	if err != nil {
		return errors.New("Unable to get block template. Error: " + err.Error())
	}
	log.Printf("Received block template for height: %d", blockTemplateResult.Height)

	// Mine the block of transactions
	block, err := util.MineBlock(blockTemplateResult, ms.Config.Bitcoin.Address, ms.ChainParams)
	if err != nil {
		return errors.New("Error mining the block. Error: " + err.Error())
	}
	log.Printf("Successfully mined block: %s", block.Hash())

	// Submit the successfully mined block
	err = ms.RPCClient.SubmitBlock(block, &btcjson.SubmitBlockOptions{})
	if err != nil {
		return errors.New("Error submitting the block. Error: " + err.Error())
	}
	log.Printf("Submitted mined block %s to node successfully", block.Hash())

	return nil
}

// getChainParams maps network names returned by getblockchaininfo to chaincfg Params
func getChainParams(chainName string) (*chaincfg.Params, error) {
	switch chainName {
	case "main":
		return &chaincfg.MainNetParams, nil
	case "test", "testnet", "testnet3":
		return &chaincfg.TestNet3Params, nil
	case "regtest":
		return &chaincfg.RegressionNetParams, nil
	case "simnet":
		return &chaincfg.SimNetParams, nil
	default:
		return nil, fmt.Errorf("unknown blockchain network name: %q", chainName)
	}
}
