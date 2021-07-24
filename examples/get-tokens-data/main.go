package main

import (
	"log"

	"github.com/hermeznetwork/hermez-go-sdk/client"
	sdkcommon "github.com/hermeznetwork/hermez-go-sdk/common"
	"github.com/hermeznetwork/hermez-go-sdk/node"
	"github.com/hermeznetwork/hermez-go-sdk/token"
)

const (
	ethereumNodeURL = "https://mainnet.infura.io/v3/"
	network         = "mainnet"
	debug           = false
)

func main() {
	log.Println("Starting Hermez Client...")
	networkDefinition, err := sdkcommon.GetNetworkDefinition(network)
	if err != nil {
		log.Printf("Error getting hermez definition at %s . Error: %s\n", network, err.Error())
		return
	}
	hezClient, err := client.NewHermezClient(ethereumNodeURL, networkDefinition.AuctionContractAddress.Hex(), networkDefinition.ChainID)
	if err != nil {
		log.Printf("Error during Hermez client initialization: %s\n", err.Error())
		return
	}
	log.Println("Connected to Hermez Smart Contracts...")
	log.Println("Pulling account info from a coordinator...")

	bootCoordNodeState, err := node.GetBootCoordinatorNodeInfo(hezClient)
	if err != nil {
		log.Printf("Error obtaining boot coordinator info. URL: %s - Error: %s\n", hezClient.BootCoordinatorURL, err.Error())
		return
	}
	log.Println("Setting current client ...")
	hezClient.SetCurrentCoordinator(bootCoordNodeState.Network.NextForgers[0].Coordinator.URL)
	log.Println("Current client is set.")

	apiResponseTokens, err := token.GetTokens(hezClient)
	if err != nil {
		log.Printf("Error obtaining tokens info. URL: %s - Error: %s\n", hezClient.BootCoordinatorURL, err.Error())
		return
	}

	for _, token := range apiResponseTokens.Tokens {
		log.Printf("%+v\n", token)
	}
}
