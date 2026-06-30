package util

import (
	"encoding/hex"
	"testing"

	"github.com/btcsuite/btcd/chaincfg"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConvertBitcoinAddressToPayToAddrScript_Valid(t *testing.T) {
	tests := []struct {
		name           string
		bitcoinAddress string
		params         *chaincfg.Params
		expectedHex    string
	}{
		{
			name:           "MainNet Address 1GCRgM...",
			bitcoinAddress: "1GCRgM2L6tzjwfm7okZNL16K1J9wus85We",
			params:         &chaincfg.MainNetParams,
			expectedHex:    "76a914a6b31013949f07e6e244e3f563aa336dd4c5840288ac",
		},
		{
			name:           "MainNet Address 1C7zdT...",
			bitcoinAddress: "1C7zdTfnkzmr13HfA2vNm5SJYRK6nEKyq8",
			params:         &chaincfg.MainNetParams,
			expectedHex:    "76a91479fbfc3f34e7745860d76137da68f362380c606c88ac",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actualBytes, err := ConvertBitcoinAddressToPayToAddrScript(tt.bitcoinAddress, tt.params)
			require.NoError(t, err)
			assert.Equal(t, tt.expectedHex, hex.EncodeToString(actualBytes))
		})
	}
}

func TestConvertBitcoinAddressToPayToAddrScript_Invalid(t *testing.T) {
	bitcoinAddress := "invalid bitcoin address"
	_, err := ConvertBitcoinAddressToPayToAddrScript(bitcoinAddress, &chaincfg.MainNetParams)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "decoded address is of unknown format")
}
