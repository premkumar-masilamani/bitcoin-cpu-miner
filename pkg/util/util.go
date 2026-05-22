package util

import (
	"encoding/hex"
	"log"
	"strconv"
	"time"

	"github.com/btcsuite/btcd/btcjson"
	"github.com/btcsuite/btcutil"
)

/*
Right shift the binary number by 8 bits in a loop until the value becomes zero.

Ex. Let's find out how many bytes are needed to represent 12345. By looking at the
representation, we can easily figure out that it requires 2 bytes.

Decimal: 12345
Binary:	00110000 00111001

Let's do that programmatically.
Right shift by 8 bits: 00000000 00110000. The value is greater than zero.
Right shift by 8 bits: 00000000 00000000. The value is NOT greater than zero.

We right shifted 8 bits, two times. So, the bytes required to represent 12345 is 2 bytes.
*/
func GetByteCountForInteger(number int64) byte {
	var byteCount byte
	for number > 0 {
		rightShiftedBy8Bits := number >> 8
		number = rightShiftedBy8Bits
		byteCount++
	}
	return byteCount
}

/*
https://learnmeabitcoin.com/technical/little-endian
*/
func ConvertIntegerToLittleEndianByteArray(number int64, byteCount int) []byte {
	littleEndianByteArray := make([]byte, byteCount)
	for i := 0; i < byteCount; i++ {
		littleEndianByteArray[i] = byte(number >> (i * 8))
	}
	return littleEndianByteArray
}

/*
https://developer.bitcoin.org/reference/transactions.html#compactsize-unsigned-integers
*/
func ConvertIntegerToCompactUnsignedInteger(number int64) []byte {
	if number <= 0xfc {
		return ConvertIntegerToLittleEndianByteArray(number, 1)
	} else if number <= 0xffff {
		return append([]byte{0xfd}, ConvertIntegerToLittleEndianByteArray(number, 2)...)
	} else if number <= 0xffffffff {
		return append([]byte{0xfe}, ConvertIntegerToLittleEndianByteArray(number, 4)...)
	}
	return append([]byte{0xff}, ConvertIntegerToLittleEndianByteArray(number, 8)...)
}

/*
Encode the block height to be used in coinbase transaction as per BIP 0034
https://github.com/bitcoin/bips/blob/master/bip-0034.mediawiki
First Byte: No. of bytes required to represent the block height (0x03)
Subsequent Bytes: Little Endian representation of block height
*/
func EncodeBlockHeight(blockHeight int64) []byte {

	// First Byte: Byte Count
	byteCount := GetByteCountForInteger(blockHeight)

	// Subsequent Bytes: Little Endian Integer Representation
	littleEndianByteArray := ConvertIntegerToLittleEndianByteArray(blockHeight, int(byteCount))

	// Append them together
	encodedBlockHeight := make([]byte, byteCount+1)
	encodedBlockHeight = append(encodedBlockHeight, byteCount)
	encodedBlockHeight = append(encodedBlockHeight, littleEndianByteArray...)

	// Appending a slice to another slice increases the capacity by
	// 8 bytes. We need only the bytes with values. So, trimming it.
	return encodedBlockHeight[byteCount+1:]
}

/*
Coinbase Transaction is the first transaction in any block, without a transaction input.
Bitcoins are created out of nothing. The bitcoin value produced by the coinbase transaction
is the Miner's Reward Fee + Fees from all the transactions included in the current block.

https://developer.bitcoin.org/reference/transactions.html#coinbase-input-the-input-of-the-first-transaction-in-a-block
*/
func MakeCoinbaseTransaction(coinbaseScript string, coinbaseAddress string, coinbaseValue int64) string {

	// OP_DUP OP_HASH160 <len to push> <pubkey> OP_EQUALVERIFY OP_CHECKSIG
	publicKeyScript := "76" + "a9" + "14" + "a6b31013949f07e6e244e3f563aa336dd4c58402" + "88" + "ac"
	// TODO: Fix the Hash160 conversion
	log.Printf("btcutil.Hash160([]byte(coinbaseAddress)): %s", hex.EncodeToString(btcutil.Hash160([]byte(coinbaseAddress))))

	// Version
	coinbaseTxData := "01000000"
	// Number of inputs
	coinbaseTxData += "01"
	// Previous outpoint TXID
	coinbaseTxData += "0000000000000000000000000000000000000000000000000000000000000000"
	// Previous outpoint index
	coinbaseTxData += "ffffffff"
	// Bytes in coinbase
	coinbaseTxData += hex.EncodeToString(ConvertIntegerToCompactUnsignedInteger(int64(len(coinbaseScript))))
	//// input[0] script
	coinbaseTxData += hex.EncodeToString([]byte(coinbaseScript))
	//// input[0] seq num
	coinbaseTxData += "ffffffff"
	//// out-counter
	coinbaseTxData += "01"
	//// output[0] value
	coinbaseTxData += hex.EncodeToString(ConvertIntegerToLittleEndianByteArray(coinbaseValue, 8))
	// output[0] script len
	coinbaseTxData += "19" //TODO: hex.EncodeToString(ConvertIntegerToCompactUnsignedInteger(int64(len(publicKeyScript))))
	// output[0] script
	coinbaseTxData += publicKeyScript
	// lock-time
	coinbaseTxData += "00000000"

	log.Printf("coinbaseTxData: %s", coinbaseTxData)
	log.Printf("hex.EncodeToString(ConvertIntegerToCompactUnsignedInteger(int64(len(coinbaseScript)))): %s", hex.EncodeToString(ConvertIntegerToCompactUnsignedInteger(int64(len(coinbaseScript)))))
	log.Printf("hex.EncodeToString(ConvertIntegerToCompactUnsignedInteger(int64(len(publicKeyScript)))): %s", hex.EncodeToString(ConvertIntegerToCompactUnsignedInteger(int64(len(publicKeyScript)))))
	return coinbaseTxData
}

/*
Single threaded implementation to mine a block
*/
func MineBlock(blockTemplateResult *btcjson.GetBlockTemplateResult, coinbaseAddress string) (*btcutil.Block, error) {

	// Any string encoded in hex to be included in the coinbase transaction
	coinbaseMessage := hex.EncodeToString([]byte("smileprem"))
	log.Printf("coinbaseMessage: %s", coinbaseMessage)

	// Block Height encoded in ascii hex
	encodedBlockHeight := hex.EncodeToString(EncodeBlockHeight(blockTemplateResult.Height))
	log.Printf("encodedBlockHeight: %s", encodedBlockHeight)

	// Bitcoin address to which the block rewards and fees should be sent
	log.Printf("coinbaseAddress: %s", coinbaseAddress)

	targetHash := blockTemplateResult.Target
	log.Printf("targetHash: %s", targetHash)

	startTime := time.Now()
	log.Printf("startTime: %s", startTime)

	nonce := 0
	log.Printf("nonce: %d", nonce)

	for nonce < 0xffffffff {
		coinbaseScript := encodedBlockHeight + coinbaseMessage + strconv.Itoa(nonce)

		coinbaseTransaction := &btcjson.GetBlockTemplateResultTx{
			Data:    MakeCoinbaseTransaction(coinbaseScript, coinbaseAddress, *blockTemplateResult.CoinbaseValue),
			Hash:    "",
			TxID:    "",
			Depends: nil,
			Fee:     0,
			SigOps:  0,
			Weight:  0,
		}
		blockTemplateResult.CoinbaseTxn = coinbaseTransaction
		nonce++
		//https: //medium.com/coinmonks/how-to-create-a-raw-bitcoin-transaction-step-by-step-239b888e87f2
	}

	block := &btcutil.Block{}
	return block, nil
}
