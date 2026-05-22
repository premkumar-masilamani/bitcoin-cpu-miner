package util

import (
	"encoding/hex"
	"fmt"
	"log"

	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcutil"
)

func ConvertBitcoinAddressToP2PKHScript(bitcoinAddress string) (string, error) {
	// Decode the Bitcoin address
	decoded, err := btcutil.DecodeAddress(bitcoinAddress, &chaincfg.MainNetParams)
	if err != nil {
		return "", err
	}

	// Extract the hash160 from the decoded address
	hash160 := decoded.ScriptAddress()

	// Convert hash160 to hexadecimal representation
	hash160Hex := hex.EncodeToString(hash160)

	// Construct the P2PKH script
	script := fmt.Sprintf("OP_DUP OP_HASH160 %s OP_EQUALVERIFY OP_CHECKSIG", hash160Hex)
	log.Printf("Bitcoin Address: %s \nP2PKH Script: %s", bitcoinAddress, script)

	return script, nil
}
