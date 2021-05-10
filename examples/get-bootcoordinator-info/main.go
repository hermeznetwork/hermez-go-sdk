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
	log.Println("Connected to Hermez Smart Contracts...")
	log.Println("Pulling boot coordinator info...")
	nodeState, err := node.GetBootCoordinatorNodeInfo(hezClient)
	if err != nil {
		log.Printf("Error obtaining boot coordinator info. URL: %s - Error: %s\n", hezClient.BootCoordinatorURL, err.Error())
		return
	}
	log.Printf("\nBoot Coordinator info is: %+v\n\n", nodeState)

	log.Printf("\nBoot Coordinator URL is: %+v\n\n", nodeState.GetBootCoordinatorURL())
}
