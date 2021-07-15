package main

import (
	"github.com/hermeznetwork/hermez-go-sdk/transaction"
	"log"

	"github.com/hermeznetwork/hermez-go-sdk/client"
	"github.com/hermeznetwork/hermez-go-sdk/node"
)

const (
	ethereumNodeURL           = ""
	auctionContractAddressHex = "0x1D5c3Dd2003118743D596D7DB7EA07de6C90fB20"
)

func main() {
	// Client initialization
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

	if len(currentCoordNodeState.Network.NextForgers) > 0 {
		log.Printf("Current coordinator URL is: %s\n", currentCoordNodeState.Network.NextForgers[0].Coordinator.URL)
	}

	txs := make([]string, 2)
	txs[0] = `{	"type": "Transfer",
				"tokenId": 1,
				"fromAccountIndex": "hez:HEZ:24265",
				"toAccountIndex": "hez:HEZ:24268",
				"toHezEthereumAddress": "hez:0xf495CC4be6896733e8fe5141a62D261110CEb9F3",
				"toBjj": null,
				"amount": "100000000000000000",
				"fee": 126,
				"nonce": 2,
				"maxNumBatch": 0,
				"signature": "35840ade5cdd7bab78bba78a1a10997a5bfe6a18c787540f040ea8f33884969ae22088a40b6d596450056b0a14e0f6e3330d605d1a04459f3147d73127564f05",
				"requestOffset": 1 }`
	txs[1] = `{	"type": "Transfer",
				"tokenId": 1,
				"fromAccountIndex": "hez:HEZ:24268",
				"toAccountIndex": "hez:HEZ:24269",
				"toHezEthereumAddress": "hez:0x137455AFCD2bEc304E39C988b8BCc66a56FDF32a",
				"toBjj": null,
				"amount": "100000000000000000",
				"fee": 126,
				"nonce": 0,
				"maxNumBatch": 0,
				"signature": "03272ada607829d29c35b27e7cb1889ba65e02961950ac45bf9e49f2639cf91a980fda0f6c9553389162af70d022d1447cf1daee90651745609de870ba6eba05",
				"requestOffset": 7 }`

	server, err := transaction.AtomicTransferJSON(hezClient, 5, txs)
	if err != nil {
		log.Printf(err.Error())
	} else {
		log.Printf(server)
	}
}
