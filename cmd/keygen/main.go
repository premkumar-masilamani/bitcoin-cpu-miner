package main

import (
	"fmt"
	"log"

	"github.com/btcsuite/btcd/btcec"
	"github.com/btcsuite/btcutil"
	"github.com/btcsuite/btcd/chaincfg"
)

func main() {
	// Generate a new private key using the secp256k1 curve
	privKey, err := btcec.NewPrivateKey(btcec.S256())
	if err != nil {
		log.Fatalf("Failed to generate private key: %v", err)
	}

	// Supported environments
	environments := []struct {
		name   string
		params *chaincfg.Params
	}{
		{"MainNet", &chaincfg.MainNetParams},
		{"TestNet3", &chaincfg.TestNet3Params},
		{"RegTest", &chaincfg.RegressionNetParams},
		{"SimNet", &chaincfg.SimNetParams},
	}

	fmt.Println("================================================================================")
	fmt.Println("             Generated Offline Bitcoin Keypairs & Addresses                     ")
	fmt.Println("================================================================================")

	for _, env := range environments {
		// Encode private key in Wallet Import Format (WIF)
		wif, err := btcutil.NewWIF(privKey, env.params, true)
		if err != nil {
			log.Fatalf("Failed to generate WIF for %s: %v", env.name, err)
		}

		// Derive compressed P2PKH address
		pubKeyBytes := privKey.PubKey().SerializeCompressed()
		address, err := btcutil.NewAddressPubKey(pubKeyBytes, env.params)
		if err != nil {
			log.Fatalf("Failed to generate address for %s: %v", env.name, err)
		}

		fmt.Printf("Environment:   %s\n", env.name)
		fmt.Printf("WIF Private:   %s\n", wif.String())
		fmt.Printf("P2PKH Address: %s\n", address.EncodeAddress())
		fmt.Println("--------------------------------------------------------------------------------")
	}
}
