package main

import (
	"log"

	"github.com/jeffprestes/hermez-go-sdk/account"
	"github.com/jeffprestes/hermez-go-sdk/client"
)

const (
	ethereumNodeURL           = "http://geth.marcelonode.xyz:8545"
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
	// testAccountString := "6LJgHxQWiqB8hxYShBLTcBpAKuVyoJ8Bpol2EXDNuwM9"
	// testAccountString := "0x263c3ab7e4832edf623fbdd66acee71c028ff591"
	testAccountString := "ETH:21499"
	accountDetails, err := account.GetAccountInfo(hezClient, testAccountString)
	if err != nil {
		log.Printf("Error obtaining account details. Account: %s - Error: %s\n", testAccountString, err.Error())
		return
	}
	log.Printf("\n\nAccount info is: %+v\n\n", accountDetails)
}
