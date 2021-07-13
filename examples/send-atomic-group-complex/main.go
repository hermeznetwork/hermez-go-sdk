package main

import (
	"fmt"
	"log"
	"math/big"

	"github.com/hermeznetwork/hermez-go-sdk/account"
	"github.com/hermeznetwork/hermez-go-sdk/client"
	"github.com/hermeznetwork/hermez-go-sdk/node"
	"github.com/hermeznetwork/hermez-go-sdk/transaction"
	hezCommon "github.com/hermeznetwork/hermez-node/common"
)

const (
	ethereumNodeURL           = "https://goerli.infura.io/v3/171aba97e221493db75f0c9900902580"
	sourceAccPvtKey1          = ""
	sourceAccPvtKey2          = ""
	auctionContractAddressHex = "0x1D5c3Dd2003118743D596D7DB7EA07de6C90fB20"
)

func main() {
	var debug bool
	debug = true

	// Client initialization and
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
		RecipientAddress:      "0xf495CC4be6896733e8fe5141a62D261110CEb9F3",
		TokenSymbolToTransfer: "HEZ",
		Amount:                big.NewInt(100000000000000000),
		FeeRangeSelectedID:    126,
		RqOffSet:              1, //+1
	}

	tx2 := transaction.AtomicTxItem{
		SenderBjjWallet:       bjjWallet2,
		RecipientAddress:      "0x137455AFCD2bEc304E39C988b8BCc66a56FDF32a",
		TokenSymbolToTransfer: "HEZ",
		Amount:                big.NewInt(100000000000000000),
		FeeRangeSelectedID:    126,
		RqOffSet:              7, //-1
	}

	txs := make([]transaction.AtomicTxItem, 2)
	txs[0] = tx1
	txs[1] = tx2

	// start the txs generation to send
	atomicGroup := hezCommon.AtomicGroup{}

	// create PoolL2Txs
	atomicGroup.Txs, err = transaction.CreateFullTxs(hezClient, txs)
	if err != nil {
		err = fmt.Errorf("[AtomicTransfer] Error generating PoolL2Tx. Error: %s\n", err.Error())
		log.Printf(err.Error())
		return
	}

	// set AtomicGroupID
	atomicGroup = transaction.SetAtomicGroupID(atomicGroup)

	// Sign the txs
	for currentAtomicTxId := range txs {
		var txHash *big.Int
		txHash, err = atomicGroup.Txs[currentAtomicTxId].HashToSign(uint16(5))
		if err != nil {
			err = fmt.Errorf("[AtomicTransfer] Error generating currentAtomicTxItem hash. TX: %+v - Error: %s\n", atomicGroup.Txs[currentAtomicTxId], err.Error())
			log.Printf(err.Error())
			return
		}
		signature := txs[currentAtomicTxId].SenderBjjWallet.PrivateKey.SignPoseidon(txHash)
		atomicGroup.Txs[currentAtomicTxId].Signature = signature.Compress()
	}

	// Post
	var serverResponse string
	serverResponse, err = transaction.SendAtomicTxsGroup(hezClient, atomicGroup)
	if err != nil {
		log.Printf(err.Error())
	} else {
		log.Printf(serverResponse)
	}
}
