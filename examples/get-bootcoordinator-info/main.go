package main

import (
	"log"

	"github.com/hermeznetwork/hermez-go-sdk/client"
	"github.com/hermeznetwork/hermez-go-sdk/node"
)

const (
	nodeURL                   = ""
	auctionContractAddressHex = "0x1D5c3Dd2003118743D596D7DB7EA07de6C90fB20"
)

func main() {
	log.Println("Starting Hermez Client...")
	hezClient, err := client.NewHermezClient(nodeURL, auctionContractAddressHex, 5)
	if err != nil {
		log.Printf("Error during Hermez client initialization: %s\n", err.Error())
		return
	}
	log.Println("Connected to Hermez Smart Contracts...")
	log.Println("Pulling boot coordinator info...")
	nodeState, err := node.GetBootCoordinatorNodeInfo(hezClient)
	if err != nil {
		log.Printf("Error obtaining boot coordinator info. URL: %s - Error: %s\n", hezClient.BootCoordinatorURL, err.Error())
		return
	}
	log.Printf("\nBoot Coordinator info is: %+v\n\n", nodeState)

	log.Printf("\nBoot Coordinator URL is: %+v\n\n", nodeState.Auction.BootCoordinatorURL)
}
