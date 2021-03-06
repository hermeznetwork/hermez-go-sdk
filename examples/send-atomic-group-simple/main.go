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
	ethereumNodeURL           = ""
	sourceAccPvtKey1          = ""
	sourceAccPvtKey2          = ""
	auctionContractAddressHex = "0x1D5c3Dd2003118743D596D7DB7EA07de6C90fB20"
)

func main() {
	var debug bool
	debug = false

	// Client initialization and
	log.Println("Starting Hermez Client...")
	hezClient, err := client.NewHermezClient(ethereumNodeURL, auctionContractAddressHex, 5)
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

	log.Println("Generating BJJ wallet 1...")
	bjjWallet1, _, err := account.CreateBjjWalletFromHexPvtKey(sourceAccPvtKey1)
	if err != nil {
		log.Printf("Error Create a Babyjubjub Wallet from Hexdecimal Private Key. Account: %s - Error: %s\n", bjjWallet1.EthAccount.Address.Hex(), err.Error())
		return
	}
	log.Println("Generating BJJ wallet 2...")
	bjjWallet2, _, err := account.CreateBjjWalletFromHexPvtKey(sourceAccPvtKey2)
	if err != nil {
		log.Printf("Error Create a Babyjubjub Wallet from Hexdecimal Private Key. Account: %s - Error: %s\n", bjjWallet2.EthAccount.Address.Hex(), err.Error())
		return
	}

	tx1 := transaction.AtomicTxItem{
		SenderBjjWallet:       bjjWallet1,
		ReceiverAddress:       "0xf495CC4be6896733e8fe5141a62D261110CEb9F3",
		TokenSymbolToTransfer: "HEZ",
		Amount:                big.NewInt(100000000000000000),
		FeeRangeSelectedID:    126,
		RqOffSet:              1, //+1
	}

	tx2 := transaction.AtomicTxItem{
		SenderBjjWallet:       bjjWallet2,
		ReceiverAddress:       "0x137455AFCD2bEc304E39C988b8BCc66a56FDF32a",
		TokenSymbolToTransfer: "HEZ",
		Amount:                big.NewInt(100000000000000000),
		FeeRangeSelectedID:    126,
		RqOffSet:              7, //-1
	}

	txs := make([]transaction.AtomicTxItem, 2)
	txs[0] = tx1
	txs[1] = tx2

	server, atomicGroupID, err := transaction.AtomicTransfer(hezClient, txs)
	log.Printf("Atomic group ID: %s", atomicGroupID)
	if err != nil {
		log.Printf(err.Error())
	} else {
		log.Printf(server)
	}
}
