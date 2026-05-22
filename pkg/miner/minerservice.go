package miner

import (
	"errors"
	"fmt"
	"log"

	"github.com/btcsuite/btcd/btcjson"
	"github.com/btcsuite/btcd/rpcclient"
	"github.com/smileprem/go-bitcoin-miner/pkg/config"
	"github.com/smileprem/go-bitcoin-miner/pkg/util"
)

type MinerService struct {
	Config    *config.Config
	RPCClient *rpcclient.Client
}

func NewMinerService(appConfig *config.Config) (*MinerService, error) {
	// Configure the local bitcoin core RPC server using HTTP POST mode.
	rpcClient, err := rpcclient.New(
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
	blockChainInfo, err := rpcClient.GetBlockChainInfo()
	if err != nil {
		return nil, errors.New("Unable to fetch block chain info. Error: " + err.Error())
	}
	log.Println("Connected to bitcoin local RPC server using the config file ", *appConfig)

	// Check if local bitcoin node is synchronized with the network
	if blockChainInfo.Blocks < blockChainInfo.Headers {
		return nil, errors.New(
			fmt.Sprintf(
				"Blocks: %d, Headers: %d. Local bitcoin node is behind by %d blocks.\n",
				blockChainInfo.Blocks,
				blockChainInfo.Headers,
				blockChainInfo.Headers-blockChainInfo.Blocks,
			))
	}

	return &MinerService{
		Config:    appConfig,
		RPCClient: rpcClient,
	}, nil
}

func (ms MinerService) MineBlock() error {
	// Get the block of transactions from memory pool
	blockTemplateResult, err := ms.RPCClient.GetBlockTemplate(
		&btcjson.TemplateRequest{
			Rules: []string{"segwit"},
		})
	if err != nil {
		return errors.New("Unable to get block template. Error: " + err.Error())
	}
	log.Printf("blockTemplateResult: %+v", blockTemplateResult)

	// Mine the block of transactions from memory pool
	block, err := util.MineBlock(blockTemplateResult, ms.Config.Bitcoin.Address)
	if err != nil {
		return errors.New("Error mining the block. Error: " + err.Error())
	}
	log.Printf("Bitcoin Address: %s, Mined block: %+v", ms.Config.Bitcoin.Address, block)

	// Submit the successfully mined block
	err = ms.RPCClient.SubmitBlock(block, &btcjson.SubmitBlockOptions{})
	if err != nil {
		return errors.New("Error submitting the block. Error: " + err.Error())
	}
	log.Printf("Bitcoin Address: %s, Submitted block: %+v", ms.Config.Bitcoin.Address, block)

	return nil
}
