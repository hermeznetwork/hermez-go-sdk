package main

import (
	"log"

	"github.com/jeffprestes/hermez-go-sdk/client"
	"github.com/jeffprestes/hermez-go-sdk/node"
)

func main() {
	log.Println("Starting Hermez Client...")
	hezClient, err := client.NewHermezClient()
	if err != nil {
		log.Printf("Error during Hermez client initialization: %s\n", err.Error())
		return
	}
	log.Println("Connected to Hermez Smart Contracts.")
	log.Println("Pulling boot coordinator info...")
	bootCoordNodeState, err := node.GetBootCoordinatorNodeInfo(hezClient)
	if err != nil {
		log.Printf("Error obtaining boot coordinator info. URL: %s - Error: %s\n", hezClient.BootCoordinatorURL, err.Error())
		return
	}
	log.Println("Setting actual client ...")
	hezClient.SetActualCoordinator(bootCoordNodeState.Network.NextForgers[0].Coordinator.URL)
	log.Println("Actual client is set.")
	log.Printf("Pulling actual coordinator (%s) info...\n", hezClient.ActualCoordinatorURL)
	actualCoordNodeState, err := node.GetActualCoordinatorNodeInfo(hezClient)
	if err != nil {
		log.Printf("Error obtaining boot coordinator info. URL: %s - Error: %s\n", hezClient.BootCoordinatorURL, err.Error())
		return
	}
	log.Printf("Actual coordinator info is: %+v\n", actualCoordNodeState)
	if len(actualCoordNodeState.Network.NextForgers) > 0 {
		log.Printf("Actual coordinator URL is: %s\n", actualCoordNodeState.Network.NextForgers[0].Coordinator.URL)
	}
	log.Printf("Boot coordinator URL is: %+v\n", actualCoordNodeState.Auction.BootCoordinatorURL)
}
