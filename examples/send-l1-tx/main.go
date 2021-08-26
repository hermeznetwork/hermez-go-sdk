package main

import (
	"log"

	"github.com/hermeznetwork/hermez-go-sdk/client"
	"github.com/hermeznetwork/hermez-node/eth"
)

func main() {
	var debug bool
	debug = true

	log.Println("Starting Hermez Client...")
	hezClient, err := client.NewHermezClientFromEnv()
	if err != nil {
		log.Printf("Error during Hermez client initialization: %s\n", err.Error())
		return
	}
	log.Println("Connected to Hermez Smart Contracts...")

	rollupAddress, err := hezClient.AuctionContract.HermezRollup(nil)
	if err != nil {
		log.Printf("Error getting rollup smart contract address: %s\n", err.Error())
		return
	}
	log.Println("Connected to Hermez Smart Contracts...")

	eth.NewRollupClient(hezClient.EthClient, rollupAddress)
}
