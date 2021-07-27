package main

import (
	"encoding/json"
	"log"
	"math/big"

	"github.com/hermeznetwork/hermez-go-sdk/account"
	"github.com/hermeznetwork/hermez-go-sdk/client"
	"github.com/hermeznetwork/hermez-go-sdk/node"
	"github.com/hermeznetwork/hermez-go-sdk/transaction"
	hezCommon "github.com/hermeznetwork/hermez-node/common"
)

const (
	ethereumNodeURL           = ""
	chainID                   = 5 // set this accordingly to the ethereum node
	auctionContractAddressHex = "0x1D5c3Dd2003118743D596D7DB7EA07de6C90fB20"

	sourceAccPvtKey = ""

	txsReceiverMetadataJson = `
	[
		{ "to_eth_addr": "0xb48cA794d49EeC406A5dD2c547717e37b5952a83", "fee_selector": 126, "amount": "1100000000000000000" },
		{ "to_eth_addr": "0x263C3Ab7E4832eDF623fBdD66ACee71c028Ff591", "fee_selector": 126, "amount": "1200000000000000000" },
		{ "to_eth_addr": "0xb8eD2B0a6e17649c9cE891895D3D9297Ab448f03", "fee_selector": 126, "amount": "1300000000000000000" },
		{ "to_eth_addr": "0x4E857Ac4A07cAD0B50CD006158f5E5A521A880CE", "fee_selector": 126, "amount": "1400000000000000000" }
	]`
)

var (
	hezToken = hezCommon.Token{
		TokenID: 1,
		Symbol:  "HEZ",
	}
)

func main() {
	var debug bool = false

	log.Println("Starting Hermez Client...")
	hezClient, err := client.NewHermezClient(ethereumNodeURL, auctionContractAddressHex, chainID)
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

	log.Println("Getting sender account info...")
	payerAccInfo, err := account.GetAccountInfo(hezClient, bjjWallet.EthAccount.Address.Hex())
	if err != nil {
		log.Printf("Error getting sender account info. Account: %s - Error: %s\n", bjjWallet.EthAccount.Address.Hex(), err.Error())
		return
	}

	log.Println("Getting sender idx and nonce...")
	var idx uint64
	var nonce int
	for _, acc := range payerAccInfo.Accounts {
		if acc.Token.Symbol == hezToken.Symbol {
			var strHezIdx hezCommon.StrHezIdx
			if err = strHezIdx.UnmarshalText([]byte(acc.AccountIndex)); err != nil {
				log.Printf("Error parsing account idx. Account Index: %s - Error: %s\n", acc.AccountIndex, err.Error())
				return
			}
			idx = uint64(strHezIdx.Idx)
			nonce = acc.Nonce
			break
		}
	}
	log.Printf("Nonce is: %+v\n", nonce)

	var txsMd []transaction.TxReceiverMetadata
	if err := json.Unmarshal([]byte(txsReceiverMetadataJson), &txsMd); err != nil {
		log.Printf("Error parsing txs receiver meta data. Error: %s\n", err.Error())
		return
	}

	for _, txMd := range txsMd {

		amount, ok := big.NewInt(0).SetString(txMd.Amount, 10)
		if !ok {
			log.Printf("Error parsing tx receiver metadata amount: %s - Error: %s\n", txMd.Amount, err.Error())
			return
		}

		apiTx, err := transaction.NewSignedAPITxToEthAddr(chainID, bjjWallet, idx, txMd.ToEthAddr, amount, hezCommon.FeeSelector(txMd.FeeSelector), hezToken, nonce)
		if err != nil {
			log.Printf("Error creating tx to Eth Address: %s - Error: %s\n", txMd.ToEthAddr, err.Error())
			return
		}

		apiTx, response, err := transaction.ExecuteL2Transaction(hezClient, apiTx)
		if err != nil {
			log.Printf("Error executing tx to Eth Address: %s - Error: %s\n", txMd.ToEthAddr, err.Error())
			return
		}

		log.Println("Transaction ID: ", apiTx.TxID.String())
		log.Printf("Transaction submitted: %s\n", response)

		nonce++
	}
}
