package main

import (
	"fmt"
	"log"
	"math/big"
	"strings"

	"github.com/jeffprestes/hermez-go-sdk/account"
	"github.com/jeffprestes/hermez-go-sdk/client"
	"github.com/jeffprestes/hermez-go-sdk/node"
	"github.com/jeffprestes/hermez-go-sdk/transaction"
)

// SignatureConstantBytes contains the SignatureConstant in byte array
// format, which is equivalent to 3322668559 as uint32 in byte array in
// big endian representation.
var SignatureConstantBytes = []byte{198, 11, 230, 15}

func main() {
	var debug bool
	debug = false

	log.Println("Starting Hermez Client...")
	hezClient, err := client.NewHermezClient()
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
	bjjWallet, ethAccount, err := account.CreateBjjWalletFromHexPvtKey("e7064c29eb71fa44e6a14e78f5fcba3c1625b6382d107ef21096555074a98cd9")
	if err != nil {
		log.Printf("Error Create a Babyjubjub Wallet from Hexdecimal Private Key. Account: %s - Error: %s\n", bjjWallet.HezEthAddress, err.Error())
		return
	}

	senderAccDetails, err := account.GetAccountInfo(hezClient, ethAccount.Address.Hex())
	if err != nil {
		log.Printf("Error obtaining account details. Account: %s - Error: %s\n", bjjWallet.HezEthAddress, err.Error())
		return
	}

	if strings.ToUpper(bjjWallet.HezBjjAddress) != strings.ToUpper(senderAccDetails.Accounts[0].BJJAddress) {
		err = fmt.Errorf("Local BJJ address and remote BJJ account are different: %s %s", bjjWallet.HezBjjAddress, senderAccDetails.Accounts[0].BJJAddress)
		log.Println(err)
		return
	}

	if debug {
		log.Printf("\n\nAccount details from Coordinator: %+v\n\n", senderAccDetails)
		log.Println("BJJ Address in server: ", senderAccDetails.Accounts[0].BJJAddress)
		log.Println("BJJ Address local: ", bjjWallet.HezBjjAddress)
		log.Printf("Wallet details %+v\n", bjjWallet)
	}

	recipientAccDetails, err := account.GetAccountInfo(hezClient, "0x263C3Ab7E4832eDF623fBdD66ACee71c028Ff591")
	if err != nil {
		log.Printf("Error obtaining account details. Account: %s - Error: %s\n", bjjWallet.HezEthAddress, err.Error())
		return
	}

	// Which token do you want to transfer?
	itemToTransfer := "HEZ"
	amount := big.NewInt(982000000000000000)
	feeSelector := 126
	chainID := 5 // goerli

	apiTx, err := transaction.MarshalTransaction(itemToTransfer, senderAccDetails, recipientAccDetails, bjjWallet, amount, feeSelector, chainID)
	if err != nil {
		log.Printf("Error marsheling tx data to prepare to send to coordinator. Error: %s\n", err.Error())
		return
	}

	apiTx, response, err := transaction.ExecuteL2Transaction(hezClient, apiTx)
	if err != nil {
		log.Printf("Error submiting tx transaction pool endpoint. Error: %s\n", err.Error())
		return
	}

	log.Println("Transaction ID: ", apiTx.TxID.String())
	log.Printf("Transaction submitted: %s\n", response)
}
