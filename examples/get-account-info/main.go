package main

import (
	"log"

	"github.com/hermeznetwork/hermez-go-sdk/account"
	"github.com/hermeznetwork/hermez-go-sdk/client"
)

const (
	ethereumNodeURL           = ""
	auctionContractAddressHex = "0x1D5c3Dd2003118743D596D7DB7EA07de6C90fB20"
)

func main() {
	log.Println("Starting Hermez Client...")
	hezClient, err := client.NewHermezClient(ethereumNodeURL, auctionContractAddressHex)
	if err != nil {
		log.Printf("Error during Hermez client initialization: %s\n", err.Error())
		return
	}
	log.Println("Connected to Hermez Smart Contracts...")
	log.Println("Pulling account info from a coordinator...")

	testAccountString := "ETH:21499"
	accountDetails, err := account.GetAccountInfo(hezClient, testAccountString)
	if err != nil {
		log.Printf("Error obtaining account details. Account: %s - Error: %s\n", testAccountString, err.Error())
		return
	}
	log.Printf("\n\nAccount info is: %+v\n\n", accountDetails)
}
