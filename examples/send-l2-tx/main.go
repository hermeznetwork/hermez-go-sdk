package main

import (
	"log"
	"math/big"

	"github.com/jeffprestes/hermez-go-sdk/account"
	"github.com/jeffprestes/hermez-go-sdk/client"
	"github.com/jeffprestes/hermez-go-sdk/node"
	"github.com/jeffprestes/hermez-go-sdk/transaction"
)

const (
	nodeURL         = "http://marcelonode.xyz:8545"
	sourceAccPvtKey = ""
)

func main() {
	var debug bool
	debug = false

	log.Println("Starting Hermez Client...")
	hezClient, err := client.NewHermezClient(nodeURL)
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
	log.Println("Setting actual client ...")
	hezClient.SetActualCoordinator(bootCoordNodeState.Network.NextForgers[0].Coordinator.URL)
	log.Println("Actual client is set.")

	log.Printf("Pulling actual coordinator (%s) info...\n", hezClient.ActualCoordinatorURL)
	actualCoordNodeState, err := node.GetActualCoordinatorNodeInfo(hezClient)
	if err != nil {
		log.Printf("Error obtaining boot coordinator info. URL: %s - Error: %s\n", hezClient.BootCoordinatorURL, err.Error())
		return
	}

	if debug {
		log.Printf("Actual coordinator info is: %+v\n", actualCoordNodeState)
	}

	if len(actualCoordNodeState.Network.NextForgers) > 0 {
		log.Printf("Actual coordinator URL is: %s\n", actualCoordNodeState.Network.NextForgers[0].Coordinator.URL)
	}

	if debug {
		log.Printf("Boot coordinator URL is: %+v\n", actualCoordNodeState.Auction.BootCoordinatorURL)
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
		big.NewInt(982000000000000000),
		126,
		5)

	log.Println("Transaction ID: ", apiTx.TxID.String())
	log.Printf("Transaction submitted: %s\n", response)
}
