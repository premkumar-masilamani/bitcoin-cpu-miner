package util

import (
	"testing"
)

func TestValidConvertBitcoinAddressToP2PKHScript(t *testing.T) {
	bitcoinAddress := "1GCRgM2L6tzjwfm7okZNL16K1J9wus85We"
	// bitcoinAddress := "1C7zdTfnkzmr13HfA2vNm5SJYRK6nEKyq8"
	expectedP2PKHScript := "OP_DUP OP_HASH160 79fbfc3f34e7745860d76137da68f362380c606c OP_EQUALVERIFY OP_CHECKSIG"
	actualP2PKHScript, err := ConvertBitcoinAddressToP2PKHScript(bitcoinAddress)
	if err != nil {
		t.Error(err)
	}
	if actualP2PKHScript != expectedP2PKHScript {
		t.Errorf("Mismatch for Bitcoin Address.\nExpected: %s\nActual: %s", expectedP2PKHScript, actualP2PKHScript)
	}
}

func TestInValidConvertBitcoinAddressToP2PKHScript(t *testing.T) {
	bitcoinAddress := "invalid bitcoin address"
	expectedError := "decoded address is of unknown format"
	_, err := ConvertBitcoinAddressToP2PKHScript(bitcoinAddress)
	if err == nil || err.Error() != expectedError {
		t.Error(err)
	}
}
