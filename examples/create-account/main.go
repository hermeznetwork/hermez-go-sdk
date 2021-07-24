package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/ethereum/go-ethereum/common"
	"github.com/hermeznetwork/hermez-go-sdk/account"
	"github.com/hermeznetwork/hermez-go-sdk/client"
	sdkcommon "github.com/hermeznetwork/hermez-go-sdk/common"
	"github.com/hermeznetwork/hermez-go-sdk/node"
	"github.com/hermeznetwork/hermez-go-sdk/util"
)

const (
	ethereumNodeURL = "https://goerli.infura.io/v3/dfc83c12e02149a585a67cd6a6338f9d"
	sourceAccPvtKey = "3030303254656e686f333242697473566f6365506f6465416372656469746172"
	network         = "goerli"
	debug           = false
)

func main() {
	log.Println("Starting Hermez Client...")
	networkDefinition, err := sdkcommon.GetNetworkDefinition(network)
	if err != nil {
		log.Printf("Error getting hermez definition at %s . Error: %s\n", network, err.Error())
		return
	}
	hezClient, err := client.NewHermezClient(ethereumNodeURL, networkDefinition.AuctionContractAddress.Hex(), networkDefinition.ChainID)
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
	bjjWallet, _, err := account.CreateBjjWalletWithAccCreationSignatureFromHexPvtKey(
		sourceAccPvtKey, networkDefinition.ChainID, networkDefinition.RollupContractAddress.Hex())
	if err != nil {
		log.Printf("Error Create a Babyjubjub Wallet from Hexdecimal Private Key. Account: %s - Error: %s\n", bjjWallet.EthAccount.Address.Hex(), err.Error())
		return
	}

	accountCreation := account.AccountCreation{}
	accountCreation.EthereumAddress = bjjWallet.HezEthAddress
	accountCreation.HezBjjAddress = bjjWallet.HezBjjAddress
	accountCreation.Signature = bjjWallet.AccountCreationAuthSignature

	log.Printf("\naccount creation: %+v\n", accountCreation)

	apiTxBody, err := util.MarshallBody(accountCreation)
	if err != nil {
		err = fmt.Errorf("[] Error marshaling HTTP request tx: %+v - Error: %s\n", accountCreation, err.Error())
		log.Println("Error marshaling HTTP request: ", err.Error())
		return
	}

	var URL string
	URL = hezClient.CurrentCoordinatorURL + "/v1/account-creation-authorization"
	request, err := http.NewRequest("POST", URL, apiTxBody)
	if err != nil {
		err = fmt.Errorf("[] Error creating HTTP request. URL: %s - request: %+v - Error: %s\n", URL, apiTxBody, err.Error())
		log.Println("Error creating HTTP request: ", err.Error())
		return
	}
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10.15; rv:88.0) Gecko/20100101 Firefox/88.0")
	request.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8")
	request.Header.Set("Accept-Encoding", "gzip, deflate, br")

	if debug {
		log.Printf("Submitting...\n%+v\n%+v\n", accountCreation, request)
	}

	response, err := hezClient.HttpClient.Do(request)
	if err != nil {
		err = fmt.Errorf("[] Error submitting HTTP request tx. URL: %s - request: %+v - Error: %s\n", URL, apiTxBody, err.Error())
		log.Println("Error submitting HTTP: ", err.Error())
		return
	}
	defer response.Body.Close()

	if response.StatusCode < 200 || response.StatusCode > 299 {
		tempBuf, errResp := io.ReadAll(response.Body)
		if errResp != nil {
			err = fmt.Errorf("[] Error unmarshaling tx: %+v - Error: %s\n", accountCreation, errResp.Error())
			return
		}
		strJSONRequest := string(tempBuf)
		err = fmt.Errorf("[] Error posting account: %+v\nStatusCode:%d \nStatus: %s\nReturned Message: %s\nURL: %s \nRequest: %+v \nResponse: %+v\n",
			accountCreation, response.StatusCode, response.Status, strJSONRequest, URL, request, response)
		log.Println("Error request: ", err.Error())
		return
	}

	b, err := ioutil.ReadAll(response.Body)
	if err != nil {
		err = fmt.Errorf("[] Error reading HTTP return from Coordinator. URL: %s - request: %+v - Error: %s\n", URL, apiTxBody, err.Error())
		log.Println("Error ReadAll: ", err.Error())
		return
	}

	serverResponse := fmt.Sprintln("Transaction ID submmited: ", accountCreation.HezBjjAddress, "Message returned in Hex: %s\n", common.Bytes2Hex(b))
	log.Println("Success!", serverResponse)
}
