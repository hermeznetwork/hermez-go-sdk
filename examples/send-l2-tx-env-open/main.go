package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"math/big"
	"net/http"
	"os"

	"github.com/ethereum/go-ethereum/common"
	"github.com/hermeznetwork/hermez-go-sdk/account"
	"github.com/hermeznetwork/hermez-go-sdk/client"
	"github.com/hermeznetwork/hermez-go-sdk/node"
	"github.com/hermeznetwork/hermez-go-sdk/transaction"
	"github.com/hermeznetwork/hermez-go-sdk/util"
)

func main() {
	var debug bool
	debug = true

	log.Println("Starting Hermez Client...")
	hezClient, err := client.NewHermezClientFromEnv()
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

	log.Println("Getting pvt key from env...")
	sourceAccPvtKey := os.Getenv("PVT_KEY")
	if len(sourceAccPvtKey) < 64 {
		log.Fatalln("Invalid Private key.")
	}

	log.Println("Generating BJJ wallet...")
	bjjWallet, _, err := account.CreateBjjWalletFromHexPvtKey(sourceAccPvtKey)
	if err != nil {
		log.Printf("Error Create a Babyjubjub Wallet from Hexdecimal Private Key. Account: %s - Error: %s\n", bjjWallet.EthAccount.Address.Hex(), err.Error())
		return
	}

	// apiTx, response, err := transaction.L2Transfer(hezClient,
	// 	bjjWallet,
	// 	"0x263C3Ab7E4832eDF623fBdD66ACee71c028Ff591",
	// 	"HEZ",
	// 	big.NewInt(983000000000000000),
	// 	126)

	senderBjjWallet := bjjWallet
	// receiverAddress := "0x263C3Ab7E4832eDF623fBdD66ACee71c028Ff591"
	receiverAddress := "0x9302d35786643f680B617B03BcdE983DB6A7319d"
	tokenSymbolToTransfer := "HEZ"
	amount := big.NewInt(953000000000000000)
	feeRangeSelectedID := 126
	ethereumChainID := hezClient.EthereumChainID

	if debug {
		log.Println("[L2Transfer] Parameters")
		log.Printf("hezClient: %+v\n", hezClient)
		log.Printf("senderBjjWallet: %+v", bjjWallet)
		log.Println("receiverAddress: ", receiverAddress)
		log.Println("tokenSymbolToTransfer: ", tokenSymbolToTransfer)
		log.Println("amount: ", amount.String())
		log.Println("feeRangeSelectedID: ", feeRangeSelectedID)
		log.Println("ethereumChainID: ", ethereumChainID)
	}

	err = nil

	senderAccDetails, err := account.GetAccountInfo(hezClient, senderBjjWallet.EthAccount.Address.Hex())
	if err != nil {
		err = fmt.Errorf("[L2Transfer] Error obtaining account details. Account: %s - Error: %s\n", senderBjjWallet.HezEthAddress, err.Error())
		return
	}

	if debug {
		log.Printf("\n\nSender Account details from Coordinator: %+v\n\n", senderAccDetails)
		log.Println("BJJ Address in server: ", senderAccDetails.Accounts[0].BJJAddress)
		log.Println("BJJ Address local: ", bjjWallet.HezBjjAddress)
		log.Printf("Wallet details %+v\n", bjjWallet)
	}

	receiverAccDetails, err := account.GetAccountInfo(hezClient, receiverAddress)
	if err != nil {
		err = fmt.Errorf("[L2Transfer] Error obtaining account details. Account: %s - Error: %s\n", receiverAddress, err.Error())
		return
	}

	if debug {
		log.Printf("\n\nReceipient Account details from Coordinator: %+v\n\n", receiverAccDetails)
		log.Println("BJJ Address in server: ", receiverAccDetails.Accounts[0].BJJAddress)
	}

	apiTxReturn, err := transaction.MarshalTransaction(tokenSymbolToTransfer, senderAccDetails, receiverAccDetails, senderBjjWallet, amount, feeRangeSelectedID, hezClient.EthereumChainID)
	if err != nil {
		err = fmt.Errorf("[L2Transfer] Error marsheling tx data to prepare to send to coordinator. Error: %s\n", err.Error())
		return
	}

	if debug {
		log.Printf("\nTX to be submited: %+v\n", apiTxReturn)
	}

	apiTx := apiTxReturn
	apiTxBody, err := util.MarshallBody(apiTx)
	if err != nil {
		err = fmt.Errorf("[ExecuteL2Transaction] Error marshaling HTTP request tx: %+v - Error: %s\n", apiTx, err.Error())
		return
	}

	var submitTransaction bool
	submitTransaction = false

	if submitTransaction {
		var URL string
		URL = "https://www.jeffpresteshermez.xyz"
		URL += "/v1/transactions-pool"
		request, err := http.NewRequest("POST", URL, apiTxBody)
		if err != nil {
			err = fmt.Errorf("[ExecuteL2Transaction] Error creating HTTP request. URL: %s - request: %+v - Error: %s\n", URL, apiTxBody, err.Error())
			return
		}
		request.Header.Set("Content-Type", "application/json")
		request.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10.15; rv:88.0) Gecko/20100101 Firefox/88.0")
		request.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8")
		request.Header.Set("Accept-Encoding", "gzip, deflate, br")

		response, err := hezClient.HttpClient.Do(request)
		if err != nil {
			log.Fatalf("[ExecuteL2Transaction] Error submitting HTTP request tx. URL: %s - request: %+v - Error: %s\n", URL, apiTxBody, err.Error())
			return
		}
		defer response.Body.Close()

		if response.StatusCode < 200 || response.StatusCode > 299 {
			tempBuf, errResp := io.ReadAll(response.Body)
			if errResp != nil {
				log.Fatalf("[ExecuteL2Transaction] Error unmarshaling tx: %+v - Error: %s\n", apiTx, errResp.Error())
				return
			}
			strJSONRequest := string(tempBuf)
			log.Fatalf("[ExecuteL2Transaction] Error posting TX: %+v\nStatusCode:%d \nStatus: %s\nReturned Message: %s\nURL: %s \nRequest: %+v \nResponse: %+v\n",
				apiTx, response.StatusCode, response.Status, strJSONRequest, URL, request, response)
			return
		}

		b, err := ioutil.ReadAll(response.Body)
		if err != nil {
			log.Fatalf("[ExecuteL2Transaction] Error reading HTTP return from Coordinator. URL: %s - request: %+v - Error: %s\n", URL, apiTxBody, err.Error())
			return
		}

		serverResponse := fmt.Sprintln("Transaction ID submmited: ", apiTx.TxID.String(), "Message returned in Hex: %s\n", common.Bytes2Hex(b))
		log.Println("Transaction ID: ", apiTx.TxID.String())
		log.Println(serverResponse)
	} else {
		log.Printf("%+v", apiTx)
		log.Println("Transaction ID: ", apiTx.TxID.String())
	}
}
