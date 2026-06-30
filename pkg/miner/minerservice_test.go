package miner

import (
	"errors"
	"testing"
	"time"

	"github.com/btcsuite/btcd/btcjson"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/rpcclient"
	"github.com/btcsuite/btcutil"
	"github.com/smileprem/go-bitcoin-miner/pkg/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// MockBitcoinRPCClient implements BitcoinRPCClient interface for unit testing
type MockBitcoinRPCClient struct {
	GetBlockChainInfoFunc func() (*btcjson.GetBlockChainInfoResult, error)
	GetBlockTemplateFunc  func(request *btcjson.TemplateRequest) (*btcjson.GetBlockTemplateResult, error)
	SubmitBlockFunc       func(block *btcutil.Block, options *btcjson.SubmitBlockOptions) error
	ShutdownFunc          func()
}

func (m *MockBitcoinRPCClient) GetBlockChainInfo() (*btcjson.GetBlockChainInfoResult, error) {
	if m.GetBlockChainInfoFunc != nil {
		return m.GetBlockChainInfoFunc()
	}
	return &btcjson.GetBlockChainInfoResult{
		Chain:   "regtest",
		Blocks:  100,
		Headers: 100,
	}, nil
}

func (m *MockBitcoinRPCClient) GetBlockTemplate(request *btcjson.TemplateRequest) (*btcjson.GetBlockTemplateResult, error) {
	if m.GetBlockTemplateFunc != nil {
		return m.GetBlockTemplateFunc(request)
	}
	val := int64(625000000)
	return &btcjson.GetBlockTemplateResult{
		Version:           1,
		PreviousHash:      "0000000000000000000000000000000000000000000000000000000000000000",
		Transactions:      []btcjson.GetBlockTemplateResultTx{},
		CoinbaseValue:     &val,
		Bits:              "207fffff", // very easy target
		Height:            101,
		CurTime:           time.Now().Unix(),
	}, nil
}

func (m *MockBitcoinRPCClient) SubmitBlock(block *btcutil.Block, options *btcjson.SubmitBlockOptions) error {
	if m.SubmitBlockFunc != nil {
		return m.SubmitBlockFunc(block, options)
	}
	return nil
}

func (m *MockBitcoinRPCClient) Shutdown() {
	if m.ShutdownFunc != nil {
		m.ShutdownFunc()
	}
}

func TestNewMinerService_Success(t *testing.T) {
	// Mock rpcClientNew
	oldRpcClientNew := rpcClientNew
	defer func() { rpcClientNew = oldRpcClientNew }()

	mockClient := &MockBitcoinRPCClient{
		GetBlockChainInfoFunc: func() (*btcjson.GetBlockChainInfoResult, error) {
			return &btcjson.GetBlockChainInfoResult{
				Chain:   "regtest",
				Blocks:  100,
				Headers: 100,
			}, nil
		},
	}

	rpcClientNew = func(config *rpcclient.ConnConfig, ntfnHandlers *rpcclient.NotificationHandlers) (BitcoinRPCClient, error) {
		return mockClient, nil
	}

	cfg := &config.Config{
		Bitcoin: config.Bitcoin{
			Host:     "localhost",
			Port:     "18443",
			Username: "user",
			Password: "pwd",
		},
	}

	ms, err := NewMinerService(cfg)
	require.NoError(t, err)
	require.NotNil(t, ms)
	assert.Equal(t, &chaincfg.RegressionNetParams, ms.ChainParams)
}

func TestNewMinerService_ClientCreationError(t *testing.T) {
	oldRpcClientNew := rpcClientNew
	defer func() { rpcClientNew = oldRpcClientNew }()

	rpcClientNew = func(config *rpcclient.ConnConfig, ntfnHandlers *rpcclient.NotificationHandlers) (BitcoinRPCClient, error) {
		return nil, errors.New("client creation error")
	}

	cfg := &config.Config{}
	_, err := NewMinerService(cfg)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "client creation error")
}

func TestNewMinerService_GetBlockChainInfoError(t *testing.T) {
	oldRpcClientNew := rpcClientNew
	defer func() { rpcClientNew = oldRpcClientNew }()

	shutdownCalled := false
	mockClient := &MockBitcoinRPCClient{
		GetBlockChainInfoFunc: func() (*btcjson.GetBlockChainInfoResult, error) {
			return nil, errors.New("info error")
		},
		ShutdownFunc: func() {
			shutdownCalled = true
		},
	}

	rpcClientNew = func(config *rpcclient.ConnConfig, ntfnHandlers *rpcclient.NotificationHandlers) (BitcoinRPCClient, error) {
		return mockClient, nil
	}

	cfg := &config.Config{}
	_, err := NewMinerService(cfg)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Unable to fetch block chain info")
	assert.True(t, shutdownCalled)
}

func TestNewMinerService_BlocksBehindHeadersError(t *testing.T) {
	oldRpcClientNew := rpcClientNew
	defer func() { rpcClientNew = oldRpcClientNew }()

	shutdownCalled := false
	mockClient := &MockBitcoinRPCClient{
		GetBlockChainInfoFunc: func() (*btcjson.GetBlockChainInfoResult, error) {
			return &btcjson.GetBlockChainInfoResult{
				Chain:   "regtest",
				Blocks:  50,
				Headers: 100, // behind by 50 blocks
			}, nil
		},
		ShutdownFunc: func() {
			shutdownCalled = true
		},
	}

	rpcClientNew = func(config *rpcclient.ConnConfig, ntfnHandlers *rpcclient.NotificationHandlers) (BitcoinRPCClient, error) {
		return mockClient, nil
	}

	cfg := &config.Config{}
	_, err := NewMinerService(cfg)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Local bitcoin node is behind by 50 blocks")
	assert.True(t, shutdownCalled)
}

func TestNewMinerService_UnknownChainParamsError(t *testing.T) {
	oldRpcClientNew := rpcClientNew
	defer func() { rpcClientNew = oldRpcClientNew }()

	shutdownCalled := false
	mockClient := &MockBitcoinRPCClient{
		GetBlockChainInfoFunc: func() (*btcjson.GetBlockChainInfoResult, error) {
			return &btcjson.GetBlockChainInfoResult{
				Chain:   "unknown_net",
				Blocks:  100,
				Headers: 100,
			}, nil
		},
		ShutdownFunc: func() {
			shutdownCalled = true
		},
	}

	rpcClientNew = func(config *rpcclient.ConnConfig, ntfnHandlers *rpcclient.NotificationHandlers) (BitcoinRPCClient, error) {
		return mockClient, nil
	}

	cfg := &config.Config{}
	_, err := NewMinerService(cfg)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unknown blockchain network name")
	assert.True(t, shutdownCalled)
}

func TestGetChainParams(t *testing.T) {
	tests := []struct {
		name      string
		chainName string
		expected  *chaincfg.Params
		hasError  bool
	}{
		{"MainNet", "main", &chaincfg.MainNetParams, false},
		{"TestNet", "test", &chaincfg.TestNet3Params, false},
		{"TestNet3", "testnet", &chaincfg.TestNet3Params, false},
		{"TestNet3_Alternate", "testnet3", &chaincfg.TestNet3Params, false},
		{"RegTest", "regtest", &chaincfg.RegressionNetParams, false},
		{"SimNet", "simnet", &chaincfg.SimNetParams, false},
		{"Unknown", "invalid", nil, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			params, err := getChainParams(tt.chainName)
			if tt.hasError {
				assert.Error(t, err)
				assert.Nil(t, params)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, params)
			}
		})
	}
}

func TestMineBlock_Success(t *testing.T) {
	mockClient := &MockBitcoinRPCClient{}
	cfg := &config.Config{
		Bitcoin: config.Bitcoin{
			Address: "1GCRgM2L6tzjwfm7okZNL16K1J9wus85We",
		},
	}

	ms := &MinerService{
		Config:      cfg,
		RPCClient:   mockClient,
		ChainParams: &chaincfg.MainNetParams,
	}

	err := ms.MineBlock()
	assert.NoError(t, err)
}

func TestMineBlock_GetBlockTemplateError(t *testing.T) {
	mockClient := &MockBitcoinRPCClient{
		GetBlockTemplateFunc: func(request *btcjson.TemplateRequest) (*btcjson.GetBlockTemplateResult, error) {
			return nil, errors.New("rpc template error")
		},
	}
	cfg := &config.Config{}
	ms := &MinerService{
		Config:      cfg,
		RPCClient:   mockClient,
		ChainParams: &chaincfg.MainNetParams,
	}

	err := ms.MineBlock()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Unable to get block template")
}

func TestMineBlock_MineBlockError(t *testing.T) {
	mockClient := &MockBitcoinRPCClient{
		GetBlockTemplateFunc: func(request *btcjson.TemplateRequest) (*btcjson.GetBlockTemplateResult, error) {
			// invalid bits hex triggers error in MineBlock
			val := int64(100)
			return &btcjson.GetBlockTemplateResult{
				Bits:          "invalid_bits",
				CoinbaseValue: &val,
			}, nil
		},
	}
	cfg := &config.Config{}
	ms := &MinerService{
		Config:      cfg,
		RPCClient:   mockClient,
		ChainParams: &chaincfg.MainNetParams,
	}

	err := ms.MineBlock()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Error mining the block")
}

func TestMineBlock_SubmitBlockError(t *testing.T) {
	mockClient := &MockBitcoinRPCClient{
		SubmitBlockFunc: func(block *btcutil.Block, options *btcjson.SubmitBlockOptions) error {
			return errors.New("rpc submit error")
		},
	}
	cfg := &config.Config{
		Bitcoin: config.Bitcoin{
			Address: "1GCRgM2L6tzjwfm7okZNL16K1J9wus85We",
		},
	}

	ms := &MinerService{
		Config:      cfg,
		RPCClient:   mockClient,
		ChainParams: &chaincfg.MainNetParams,
	}

	err := ms.MineBlock()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Error submitting the block")
}
