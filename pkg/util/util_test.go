package util

import (
	"math/big"
	"testing"
	"time"

	"github.com/btcsuite/btcd/btcjson"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetByteCountForInteger(t *testing.T) {
	tests := []struct {
		number   int64
		expected byte
	}{
		{0, 0},
		{1, 1},
		{255, 1},
		{256, 2},
		{65535, 2},
		{65536, 3},
		{1234567, 3},
		{0x7fffffffffffffff, 8},
	}

	for _, tt := range tests {
		assert.Equal(t, tt.expected, GetByteCountForInteger(tt.number))
	}
}

func TestConvertIntegerToLittleEndianByteArray(t *testing.T) {
	tests := []struct {
		number    int64
		byteCount int
		expected  []byte
	}{
		{0x0a0b0c0d, 4, []byte{0x0d, 0x0c, 0x0b, 0x0a}},
		{0x12345678, 4, []byte{0x78, 0x56, 0x34, 0x12}},
		{0x10000000, 4, []byte{0x00, 0x00, 0x00, 0x10}},
	}

	for _, tt := range tests {
		assert.Equal(t, tt.expected, ConvertIntegerToLittleEndianByteArray(tt.number, tt.byteCount))
	}
}

func TestEncodeBlockHeight(t *testing.T) {
	tests := []struct {
		height   int64
		expected []byte
	}{
		// 123456 (0x01e240) -> 3 bytes -> [3, 0x40, 0xe2, 0x01]
		{0x01e240, []byte{0x03, 0x40, 0xe2, 0x01}},
		// 717900 (0x0af44c) -> 3 bytes -> [3, 0x4c, 0xf4, 0x0a]
		{0x0af44c, []byte{0x03, 0x4c, 0xf4, 0x0a}},
	}

	for _, tt := range tests {
		assert.Equal(t, tt.expected, EncodeBlockHeight(tt.height))
	}
}

func TestCreateCoinbaseTx(t *testing.T) {
	minerAddress := "1GCRgM2L6tzjwfm7okZNL16K1J9wus85We"
	params := &chaincfg.MainNetParams
	blockHeight := int64(735169)
	coinbaseValue := int64(625000000) // 6.25 BTC

	tx, err := CreateCoinbaseTx(blockHeight, coinbaseValue, minerAddress, params, "smileprem-extra-42")
	require.NoError(t, err)
	require.NotNil(t, tx)

	// Verify transaction input
	assert.Len(t, tx.TxIn, 1)
	assert.Equal(t, chainhash.Hash{}, tx.TxIn[0].PreviousOutPoint.Hash)
	assert.Equal(t, uint32(0xffffffff), tx.TxIn[0].PreviousOutPoint.Index)

	// Verify BIP-34 height in scriptSig
	encodedHeight := EncodeBlockHeight(blockHeight)
	assert.Equal(t, encodedHeight, tx.TxIn[0].SignatureScript[:len(encodedHeight)])

	// Verify output
	assert.Len(t, tx.TxOut, 1)
	assert.Equal(t, coinbaseValue, tx.TxOut[0].Value)

	// Expected public key script for 1GCRgM...
	expectedPkScript, err := ConvertBitcoinAddressToPayToAddrScript(minerAddress, params)
	require.NoError(t, err)
	assert.Equal(t, expectedPkScript, tx.TxOut[0].PkScript)

	// Test with invalid address
	_, err = CreateCoinbaseTx(blockHeight, coinbaseValue, "invalid_address", params, "msg")
	assert.Error(t, err)

	// Test scriptSig length enforcement (> 100 bytes)
	longMsg := "this is a very very very very very very very very very very very very very very very very very very very long message"
	txLong, err := CreateCoinbaseTx(blockHeight, coinbaseValue, minerAddress, params, longMsg)
	require.NoError(t, err)
	assert.Equal(t, 100, len(txLong.TxIn[0].SignatureScript))
}

func TestCalcMerkleRoot(t *testing.T) {
	// Empty case
	assert.Equal(t, chainhash.Hash{}, CalcMerkleRoot(nil))

	// Single hash case
	h1 := chainhash.DoubleHashH([]byte("tx1"))
	assert.Equal(t, h1, CalcMerkleRoot([]chainhash.Hash{h1}))

	// Even count case (2 hashes)
	h2 := chainhash.DoubleHashH([]byte("tx2"))
	expectedEvenConcat := append(h1[:], h2[:]...)
	expectedEvenRoot := chainhash.DoubleHashH(expectedEvenConcat)
	assert.Equal(t, expectedEvenRoot, CalcMerkleRoot([]chainhash.Hash{h1, h2}))

	// Odd count case (3 hashes)
	h3 := chainhash.DoubleHashH([]byte("tx3"))
	// Layer 1: pair 1 (h1+h2) -> h12, pair 2 (h3+h3) -> h33
	h12 := chainhash.DoubleHashH(append(h1[:], h2[:]...))
	h33 := chainhash.DoubleHashH(append(h3[:], h3[:]...))
	// Layer 2: pair (h12+h33) -> h1233 (root)
	expectedOddRoot := chainhash.DoubleHashH(append(h12[:], h33[:]...))
	assert.Equal(t, expectedOddRoot, CalcMerkleRoot([]chainhash.Hash{h1, h2, h3}))
}

func TestCompactToBig(t *testing.T) {
	// Test standard positive values
	// Target = coefficient * 256^(exponent - 3)
	// For compact 0x1b0404cb: exponent = 0x1b, mantissa = 0x0404cb
	// Target = 0x0404cb * 2^(8*(0x1b-3))
	compactVal := uint32(0x1b0404cb)
	expectedTarget := CompactToBig(compactVal)
	assert.NotNil(t, expectedTarget)
	assert.True(t, expectedTarget.Sign() > 0)

	// Test negative compact values (sign bit 0x00800000 set)
	// e.g. 0x1b8404cb
	negCompactVal := uint32(0x1b8404cb)
	negTarget := CompactToBig(negCompactVal)
	assert.NotNil(t, negTarget)
	assert.True(t, negTarget.Sign() < 0)

	// Test exponent <= 3
	smallCompactVal := uint32(0x02000001)
	smallTarget := CompactToBig(smallCompactVal)
	assert.NotNil(t, smallTarget)
}

func TestHashToBig(t *testing.T) {
	hashBytes := [32]byte{
		0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08,
		0x09, 0x0a, 0x0b, 0x0c, 0x0d, 0x0e, 0x0f, 0x10,
		0x11, 0x12, 0x13, 0x14, 0x15, 0x16, 0x17, 0x18,
		0x19, 0x1a, 0x1b, 0x1c, 0x1d, 0x1e, 0x1f, 0x20,
	}
	hash := chainhash.Hash(hashBytes)
	bigInt := HashToBig(&hash)

	// Treat hash as little-endian, so it is reversed in big-endian representation
	reversedBytes := make([]byte, 32)
	for i := 0; i < 32; i++ {
		reversedBytes[i] = hashBytes[31-i]
	}
	expectedBig := new(big.Int).SetBytes(reversedBytes)
	assert.Equal(t, expectedBig, bigInt)
}

func TestMineBlock_Success(t *testing.T) {
	// Prepare a simple block template with an extremely easy target: "207fffff"
	// This ensures mining finishes successfully in 1 step.
	val := int64(625000000)
	template := &btcjson.GetBlockTemplateResult{
		Version:           1,
		PreviousHash:      "0000000000000000000000000000000000000000000000000000000000000000",
		Transactions: []btcjson.GetBlockTemplateResultTx{
			{
				// Raw Tx hex for a simple dummy transaction
				Data: "01000000010000000000000000000000000000000000000000000000000000000000000000ffffffff0401010101ffffffff0100e1f505000000001976a91479fbfc3f34e7745860d76137da68f362380c606c88ac00000000",
			},
		},
		CoinbaseValue: &val,
		Bits:          "207fffff",
		Height:        100,
		CurTime:       time.Now().Unix(),
	}

	minerAddress := "1GCRgM2L6tzjwfm7okZNL16K1J9wus85We"
	params := &chaincfg.MainNetParams

	block, err := MineBlock(template, minerAddress, params)
	require.NoError(t, err)
	require.NotNil(t, block)

	// Block should contain 2 transactions: coinbase + the template tx
	assert.Len(t, block.MsgBlock().Transactions, 2)
	assert.Equal(t, int32(1), block.MsgBlock().Header.Version)
}

func TestMineBlock_NilTemplate(t *testing.T) {
	minerAddress := "1GCRgM2L6tzjwfm7okZNL16K1J9wus85We"
	_, err := MineBlock(nil, minerAddress, &chaincfg.MainNetParams)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "block template result cannot be nil")
}

func TestMineBlock_InvalidTxHex(t *testing.T) {
	val := int64(100)
	template := &btcjson.GetBlockTemplateResult{
		Transactions: []btcjson.GetBlockTemplateResultTx{
			{
				Data: "invalid_hex",
			},
		},
		CoinbaseValue: &val,
	}
	_, err := MineBlock(template, "1GCRgM2L6tzjwfm7okZNL16K1J9wus85We", &chaincfg.MainNetParams)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to decode transaction hex")
}

func TestMineBlock_InvalidTxDeserialize(t *testing.T) {
	val := int64(100)
	template := &btcjson.GetBlockTemplateResult{
		Transactions: []btcjson.GetBlockTemplateResultTx{
			{
				Data: "aabbcc", // Valid hex, but not a valid transaction
			},
		},
		CoinbaseValue: &val,
	}
	_, err := MineBlock(template, "1GCRgM2L6tzjwfm7okZNL16K1J9wus85We", &chaincfg.MainNetParams)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to deserialize transaction data")
}

func TestMineBlock_InvalidBitsHex(t *testing.T) {
	val := int64(100)
	template := &btcjson.GetBlockTemplateResult{
		Transactions:  []btcjson.GetBlockTemplateResultTx{},
		CoinbaseValue: &val,
		Bits:          "invalid_bits",
	}
	_, err := MineBlock(template, "1GCRgM2L6tzjwfm7okZNL16K1J9wus85We", &chaincfg.MainNetParams)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to parse bits hex")
}

func TestMineBlock_InvalidPreviousHash(t *testing.T) {
	val := int64(100)
	template := &btcjson.GetBlockTemplateResult{
		Transactions:  []btcjson.GetBlockTemplateResultTx{},
		CoinbaseValue: &val,
		Bits:          "207fffff",
		PreviousHash:  "invalid_hash_string",
	}
	_, err := MineBlock(template, "1GCRgM2L6tzjwfm7okZNL16K1J9wus85We", &chaincfg.MainNetParams)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid previous block hash")
}

func TestMineBlock_InvalidMinerAddress(t *testing.T) {
	val := int64(100)
	template := &btcjson.GetBlockTemplateResult{
		Transactions:      []btcjson.GetBlockTemplateResultTx{},
		CoinbaseValue:     &val,
		Bits:              "207fffff",
		PreviousHash:  "0000000000000000000000000000000000000000000000000000000000000000",
		CurTime:           time.Now().Unix(),
	}
	_, err := MineBlock(template, "invalid_miner_address", &chaincfg.MainNetParams)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to create coinbase transaction")
}

func TestMineBlock_Timeout(t *testing.T) {
	// Temporarily override MiningDurationInSeconds to 0 to simulate timeout
	oldDuration := MiningDurationInSeconds
	MiningDurationInSeconds = -1
	defer func() { MiningDurationInSeconds = oldDuration }()

	val := int64(100)
	template := &btcjson.GetBlockTemplateResult{
		Transactions:      []btcjson.GetBlockTemplateResultTx{},
		CoinbaseValue:     &val,
		Bits:              "207fffff",
		PreviousHash:  "0000000000000000000000000000000000000000000000000000000000000000",
		CurTime:           time.Now().Unix(),
	}

	_, err := MineBlock(template, "1GCRgM2L6tzjwfm7okZNL16K1J9wus85We", &chaincfg.MainNetParams)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "mining timed out")
}
