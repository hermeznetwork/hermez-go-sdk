package main

import (
	"io"
	"io/ioutil"
	"log"
	"math/big"
	"net/http"

	"github.com/jeffprestes/hermez-go-sdk/account"
	"github.com/jeffprestes/hermez-go-sdk/client"
	"github.com/jeffprestes/hermez-go-sdk/node"
	"github.com/jeffprestes/hermez-go-sdk/transaction"
	"github.com/jeffprestes/hermez-go-sdk/util"
)

// SignatureConstantBytes contains the SignatureConstant in byte array
// format, which is equivalent to 3322668559 as uint32 in byte array in
// big endian representation.
var SignatureConstantBytes = []byte{198, 11, 230, 15}

func main() {
	var debug bool
	debug = true

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
	bjjWallet, _, err := account.CreateBjjWalletFromHexPvtKey("e7064c29eb71fa44e6a14e78f5fcba3c1625b6382d107ef21096555074a98cd9")
	if err != nil {
		log.Printf("Error Create a Babyjubjub Wallet from Hexdecimal Private Key. Account: %s - Error: %s\n", bjjWallet.HezEthAddress, err.Error())
		return
	}

	senderAccDetails, err := account.GetAccountInfo(hezClient, bjjWallet.HezEthAddress)
	if err != nil {
		log.Printf("Error obtaining account details. Account: %s - Error: %s\n", bjjWallet.HezEthAddress, err.Error())
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
	amount := big.NewInt(889000000000000000)
	feeSelector := 126
	chainID := 5 // goerli

	apiTx, err := transaction.MarshalTransaction(itemToTransfer, senderAccDetails, recipientAccDetails, bjjWallet, amount, feeSelector, chainID)
	if err != nil {
		log.Printf("Error marsheling tx data to prepare to send to coordinator. Error: %s\n", err.Error())
		return
	}

	apiTxBody, err := util.MarshallBody(apiTx)
	if err != nil {
		log.Printf("Error marshaling HTTP request tx: %+v - Error: %s\n", tx, err.Error())
		return
	}

	var URL string
	URL = hezClient.ActualCoordinatorURL + "/v1/transactions-pool"
	request, err := http.NewRequest("POST", URL, apiTxBody)
	if err != nil {
		log.Printf("Error creating HTTP request. URL: %s - request: %+v - Error: %s\n", URL, apiTxBody, err.Error())
		return
	}
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10.15; rv:88.0) Gecko/20100101 Firefox/88.0")
	request.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8")
	request.Header.Set("Accept-Encoding", "gzip, deflate, br")

	log.Printf("Submiting this TX: %s\n Full Tx: %+v\nRequest details: %+v\n\n", apiTx.TxID, apiTx, request)

	response, err := hezClient.HttpClient.Do(request)
	if err != nil {
		log.Printf("Error submitting HTTP request tx. URL: %s - request: %+v - Error: %s\n", URL, apiTxBody, err.Error())
		return
	}
	defer response.Body.Close()

	if response.StatusCode < 200 || response.StatusCode > 299 {
		tempBuf, err := io.ReadAll(response.Body)
		if err != nil {
			log.Printf("Error unmarshaling tx: %+v - Error: %s\n", tx, err.Error())
			return
		}
		strJSONRequest := string(tempBuf)
		log.Printf("Error posting TX. \nStatusCode:%d \nStatus: %s\nReturned Message: %s\nURL: %s \nRequest: %+v \nResponse: %+v\n", response.StatusCode, response.Status, strJSONRequest, URL, request, response)
		return
	}

	b, err := ioutil.ReadAll(response.Body)
	if err != nil {
		log.Printf("Error reading HTTP return from Coordinator. URL: %s - request: %+v - Error: %s\n", URL, apiTxBody, err.Error())
		return
	}
	if b == nil || len(b) == 0 {
		log.Printf("Error no HTTP return from Coordinator. URL: %s - request: %+v \n", URL, apiTxBody)
		return
	}

	log.Printf("Transaction submitted: %s\n", string(b))
}
