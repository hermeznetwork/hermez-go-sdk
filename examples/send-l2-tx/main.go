package main

import (
	"log"
	"math/big"

	"github.com/hermeznetwork/hermez-go-sdk/account"
	"github.com/hermeznetwork/hermez-go-sdk/client"
	"github.com/hermeznetwork/hermez-go-sdk/node"
	"github.com/hermeznetwork/hermez-go-sdk/transaction"
)

const (
	ethereumNodeURL           = "https://goerli.infura.io/v3/171aba97e221493db75f0c9900902580"
	sourceAccPvtKey           = ""
	auctionContractAddressHex = "0x1D5c3Dd2003118743D596D7DB7EA07de6C90fB20"
)

func main() {
	var debug bool
	debug = false

	log.Println("Starting Hermez Client...")
	hezClient, err := client.NewHermezClient(ethereumNodeURL, auctionContractAddressHex)
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

	log.Printf("Pulling current coordinator (%s) info...\n", hezClient.CurrentCoordinatorURL)
	currentCoordNodeState, err := node.GetCurrentCoordinatorNodeInfo(hezClient)
	if err != nil {
		log.Printf("Error obtaining boot coordinator info. URL: %s - Error: %s\n", hezClient.BootCoordinatorURL, err.Error())
		return
	}

	if debug {
		log.Printf("Current coordinator info is: %+v\n", currentCoordNodeState)
	}

	if len(currentCoordNodeState.Network.NextForgers) > 0 {
		log.Printf("Current coordinator URL is: %s\n", currentCoordNodeState.Network.NextForgers[0].Coordinator.URL)
	}

	if debug {
		log.Printf("Boot coordinator URL is: %+v\n", currentCoordNodeState.Auction.BootCoordinatorURL)
	}

	log.Println("Generating BJJ wallet...")
	bjjWallet, _, err := account.CreateBjjWalletFromHexPvtKey(sourceAccPvtKey)
	if err != nil {
		log.Printf("Error Create a Babyjubjub Wallet from Hexdecimal Private Key. Account: %s - Error: %s\n", bjjWallet.EthAccount.Address.Hex(), err.Error())
		return
	}

	apiTx, response, err := transaction.L2Transfer(hezClient,
		bjjWallet,
		"0x263C3Ab7E4832eDF623fBdD66ACee71c028Ff591",
		"HEZ",
		big.NewInt(983000000000000000),
		126,
		5)

	log.Println("Transaction ID: ", apiTx.TxID.String())
	log.Printf("Transaction submitted: %s\n", response)
}
