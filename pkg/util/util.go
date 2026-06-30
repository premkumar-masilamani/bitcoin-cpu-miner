package util

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"math/big"
	"strconv"
	"time"

	"github.com/btcsuite/btcd/btcjson"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/wire"
	"github.com/btcsuite/btcutil"
)

// GetByteCountForInteger calculates how many bytes are required to represent
// the given 64-bit integer.
func GetByteCountForInteger(number int64) byte {
	var byteCount byte
	for number > 0 {
		number >>= 8
		byteCount++
	}
	return byteCount
}

// ConvertIntegerToLittleEndianByteArray converts a 64-bit integer to a
// little-endian byte slice of the specified byte count.
func ConvertIntegerToLittleEndianByteArray(number int64, byteCount int) []byte {
	littleEndianByteArray := make([]byte, byteCount)
	for i := 0; i < byteCount; i++ {
		littleEndianByteArray[i] = byte(number >> (i * 8))
	}
	return littleEndianByteArray
}

// EncodeBlockHeight encodes the block height as a byte slice according to BIP-34.
// The first byte represents the length of the height value, followed by the
// height encoded as a little-endian integer.
func EncodeBlockHeight(blockHeight int64) []byte {
	byteCount := GetByteCountForInteger(blockHeight)
	littleEndian := ConvertIntegerToLittleEndianByteArray(blockHeight, int(byteCount))
	return append([]byte{byteCount}, littleEndian...)
}

// CreateCoinbaseTx creates a new BIP-34 compliant coinbase transaction.
// It sets a null previous outpoint, puts the block height and custom coinbase message 
// into the scriptSig, and pays the coinbase subsidy + fees to the miner's address.
func CreateCoinbaseTx(blockHeight int64, coinbaseValue int64, minerAddress string, params *chaincfg.Params, coinbaseMsg string) (*wire.MsgTx, error) {
	tx := wire.NewMsgTx(1) // Coinbase Tx version 1

	// A coinbase input spends a null hash and an index of 0xffffffff
	nullHash := chainhash.Hash{}
	prevOut := wire.NewOutPoint(&nullHash, 0xffffffff)

	// Build the coinbase script (BIP-34 requires the block height to be the first item)
	encodedHeight := EncodeBlockHeight(blockHeight)
	scriptSig := append(encodedHeight, []byte(coinbaseMsg)...)

	// Enforce the consensus limit on scriptSig length (between 2 and 100 bytes)
	if len(scriptSig) > 100 {
		scriptSig = scriptSig[:100]
	}

	txIn := wire.NewTxIn(prevOut, scriptSig, nil)
	txIn.Sequence = 0xffffffff
	tx.AddTxIn(txIn)

	// Create output script paying to the miner's address
	pkScript, err := ConvertBitcoinAddressToPayToAddrScript(minerAddress, params)
	if err != nil {
		return nil, fmt.Errorf("failed to generate output script for address: %v", err)
	}

	txOut := wire.NewTxOut(coinbaseValue, pkScript)
	tx.AddTxOut(txOut)

	return tx, nil
}

// CalcMerkleRoot computes the Merkle root of a list of transaction hashes
// using Bitcoin's standard double-SHA256 binary tree algorithm.
func CalcMerkleRoot(hashes []chainhash.Hash) chainhash.Hash {
	if len(hashes) == 0 {
		return chainhash.Hash{}
	}

	// Copy transaction hashes into our working layer
	level := make([]chainhash.Hash, len(hashes))
	copy(level, hashes)

	// Pair up hashes and double-hash them until one remains
	for len(level) > 1 {
		var nextLevel []chainhash.Hash
		for i := 0; i < len(level); i += 2 {
			var concat [64]byte
			if i+1 < len(level) {
				copy(concat[0:32], level[i][:])
				copy(concat[32:64], level[i+1][:])
			} else {
				// Odd count: concatenate the hash with itself
				copy(concat[0:32], level[i][:])
				copy(concat[32:64], level[i][:])
			}
			nextLevel = append(nextLevel, chainhash.DoubleHashH(concat[:]))
		}
		level = nextLevel
	}

	return level[0]
}

// CompactToBig converts a 32-bit compact representation of a difficulty target
// (often referred to as 'nBits' or 'bits') to a 256-bit big integer target threshold.
func CompactToBig(compact uint32) *big.Int {
	mantissa := compact & 0x007fffff
	isNegative := compact&0x00800000 != 0
	exponent := uint(compact >> 24)

	var value *big.Int
	if exponent <= 3 {
		value = big.NewInt(int64(mantissa))
		value.Rsh(value, 8*(3-exponent))
	} else {
		value = big.NewInt(int64(mantissa))
		value.Lsh(value, 8*(exponent-3))
	}

	if isNegative {
		value.Neg(value)
	}

	return value
}

// HashToBig converts a 32-byte block hash (little-endian internal format)
// to a big-endian big.Int for arithmetic comparison against the target.
func HashToBig(hash *chainhash.Hash) *big.Int {
	// A Hash is in little-endian. To treat it as a big-endian number, reverse it.
	buf := *hash
	for i := 0; i < len(buf)/2; i++ {
		buf[i], buf[len(buf)-1-i] = buf[len(buf)-1-i], buf[i]
	}
	return new(big.Int).SetBytes(buf[:])
}

// MineBlock performs Proof of Work to mine a block matching the target difficulty
// defined in the block template. It constructs the coinbase transaction,
// deserializes memory-pool transactions, and runs a hashing loop until a valid nonce
// is found or a timeout occurs.
func MineBlock(blockTemplateResult *btcjson.GetBlockTemplateResult, minerAddress string, params *chaincfg.Params) (*btcutil.Block, error) {
	if blockTemplateResult == nil {
		return nil, fmt.Errorf("block template result cannot be nil")
	}

	// 1. Deserialize memory-pool transactions from the block template
	var templateTxs []*wire.MsgTx
	for _, t := range blockTemplateResult.Transactions {
		txBytes, err := hex.DecodeString(t.Data)
		if err != nil {
			return nil, fmt.Errorf("failed to decode transaction hex: %v", err)
		}

		var msgTx wire.MsgTx
		err = msgTx.Deserialize(bytes.NewReader(txBytes))
		if err != nil {
			return nil, fmt.Errorf("failed to deserialize transaction data: %v", err)
		}
		templateTxs = append(templateTxs, &msgTx)
	}

	// 2. Parse the target difficulty bits
	bitsValue, err := strconv.ParseUint(blockTemplateResult.Bits, 16, 32)
	if err != nil {
		return nil, fmt.Errorf("failed to parse bits hex %q: %v", blockTemplateResult.Bits, err)
	}
	bits := uint32(bitsValue)
	target := CompactToBig(bits)

	// 3. Initialize state variables for mining
	var (
		extraNonce = 0
		nonce      = uint32(0)
		version    = blockTemplateResult.Version
		blockTime  = time.Unix(blockTemplateResult.CurTime, 0)
		startTime  = time.Now()
	)

	prevHash, err := chainhash.NewHashFromStr(blockTemplateResult.PreviousHash)
	if err != nil {
		return nil, fmt.Errorf("invalid previous block hash: %v", err)
	}

	// 4. Hashing Loop
	for {
		// Periodically check for timeout limit
		if time.Since(startTime).Seconds() > float64(MiningDurationInSeconds) {
			return nil, fmt.Errorf("mining timed out after %d seconds", MiningDurationInSeconds)
		}

		// Re-construct coinbase transaction with the current extraNonce
		coinbaseMsg := "smileprem-extra-" + strconv.Itoa(extraNonce)
		coinbaseTx, err := CreateCoinbaseTx(blockTemplateResult.Height, *blockTemplateResult.CoinbaseValue, minerAddress, params, coinbaseMsg)
		if err != nil {
			return nil, fmt.Errorf("failed to create coinbase transaction: %v", err)
		}

		// Bundle all transactions
		blockTxs := make([]*wire.MsgTx, 0, 1+len(templateTxs))
		blockTxs = append(blockTxs, coinbaseTx)
		blockTxs = append(blockTxs, templateTxs...)

		// Calculate Merkle Root
		txHashes := make([]chainhash.Hash, len(blockTxs))
		for i, tx := range blockTxs {
			txHashes[i] = tx.TxHash()
		}
		merkleRoot := CalcMerkleRoot(txHashes)

		// Set up the block header template
		header := wire.BlockHeader{
			Version:    version,
			PrevBlock:  *prevHash,
			MerkleRoot: merkleRoot,
			Timestamp:  blockTime,
			Bits:       bits,
			Nonce:      nonce,
		}

		// Inner loop scanning nonces
		for {
			blockHash := header.BlockHash()
			hashVal := HashToBig(&blockHash)

			// If block hash satisfies target, we found a block!
			if hashVal.Cmp(target) <= 0 {
				msgBlock := wire.MsgBlock{
					Header:       header,
					Transactions: blockTxs,
				}
				return btcutil.NewBlock(&msgBlock), nil
			}

			nonce++
			header.Nonce = nonce

			// If nonce wrapped back to 0, break to vary extraNonce / time
			if nonce == 0 {
				break
			}
		}

		// Vary extraNonce and update block timestamp when nonce range is exhausted
		extraNonce++
		nowUnix := time.Now().Unix()
		if nowUnix > blockTemplateResult.CurTime {
			blockTime = time.Unix(nowUnix, 0)
		} else {
			blockTime = blockTime.Add(time.Second)
		}
	}
}
