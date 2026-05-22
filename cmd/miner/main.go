package main

import (
	"flag"
	"log"
	"os"

	"github.com/smileprem/go-bitcoin-miner/pkg/config"
	"github.com/smileprem/go-bitcoin-miner/pkg/miner"
)

func main() {

	// Check the command line arguments
	configPtr := flag.String("config", "./cmd/config/local.config.yaml", "config file to be passed to the blockchain miner")
	flag.Parse()
	if _, err := os.Stat(*configPtr); err != nil {
		log.Fatalf("Unable to read config file. Please check the file path and permissions. Config File: %v Error: %v", *configPtr, err)
	}

	// Fetch the bitcoin configurations
	appConfig, err := config.LoadConfig(*configPtr)
	if err != nil {
		log.Fatalf("Unable to load config file. Config File: %s Error: %v", *configPtr, err)
	}

	// Initialize MinerService
	minerService, err := miner.NewMinerService(appConfig)
	if err != nil {
		log.Print(err)
		return
	}
	defer minerService.RPCClient.Shutdown()

	// Mine Block
	err = minerService.MineBlock()
	if err != nil {
		log.Print(err)
		return
	}
	log.Printf("Whohoo... successfully mined a block in bitcoin :)")
}
