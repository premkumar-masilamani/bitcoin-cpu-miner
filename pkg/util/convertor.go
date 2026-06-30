package util

import (
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcutil"
)

// ConvertBitcoinAddressToPayToAddrScript decodes a bitcoin address string
// using the provided network parameters and constructs a standard pay-to-address
// output script (scriptPubKey) in raw bytes.
// Supports P2PKH, P2SH, and SegWit address types.
func ConvertBitcoinAddressToPayToAddrScript(bitcoinAddress string, params *chaincfg.Params) ([]byte, error) {
	// Decode the Bitcoin address string to btcutil.Address
	address, err := btcutil.DecodeAddress(bitcoinAddress, params)
	if err != nil {
		return nil, err
	}

	// Generate the pay-to-address script bytes
	script, err := txscript.PayToAddrScript(address)
	if err != nil {
		return nil, err
	}

	return script, nil
}
