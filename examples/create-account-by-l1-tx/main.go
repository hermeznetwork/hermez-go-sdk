package main

import (
	"context"
	"crypto/ecdsa"
	"log"
	"math/big"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/hermeznetwork/hermez-go-sdk/account"
	"github.com/hermeznetwork/hermez-go-sdk/client"
	sdkcommon "github.com/hermeznetwork/hermez-go-sdk/common"
	"github.com/hermeznetwork/hermez-go-sdk/node"
	hezcommon "github.com/hermeznetwork/hermez-node/common"
	hermez "github.com/hermeznetwork/hermez-node/eth/contracts/hermez"
)

const (
	ethereumNodeURL = "https://geth.marcelonode.xyz"
	sourceAccPvtKey = ""
	network         = "goerli"
	debug           = false
)

func main() {
	networkDefinition, err := sdkcommon.GetNetworkDefinition(network)
	if err != nil {
		log.Printf("Error getting hermez definition at %s . Error: %s\n", network, err.Error())
		return
	}
	log.Println("Starting Hermez Client...")
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

	hermezSC, err := hermez.NewHermez(networkDefinition.RollupContractAddress, hezClient.EthClient)

	privateKey, err := crypto.HexToECDSA(sourceAccPvtKey)
	if err != nil {
		log.Fatalf("get private key error: %s", err.Error())
	}
	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		log.Fatalf("cannot assert type: publicKey is not of type *ecdsa.PublicKey")
	}
	fromAddress := crypto.PubkeyToAddress(*publicKeyECDSA)

	gasPrice, err := hezClient.EthClient.SuggestGasPrice(context.Background())
	if err != nil {
		log.Fatalf("try get gas price suggestion error: %s", err.Error())
	}

	nonce, err := hezClient.EthClient.PendingNonceAt(context.Background(), fromAddress)
	if err != nil {
		log.Fatalf("try get pending nonce error: %s", err.Error())
	}

	/*
		pending, err := hezClient.EthClient.PendingBalanceAt(context.Background(), fromAddress)
		if err != nil {
			log.Fatalf("balanceAt error: %s ", err.Error())
		}
		fmt.Println("pending balance", pending)
	*/

	tokenAddress := common.HexToAddress("0x55a1Db90A5753e6Ff50FD018d7E648d58A081486")
	gasLimit, err := hezClient.EthClient.EstimateGas(context.Background(), ethereum.CallMsg{
		To: &tokenAddress,
	})
	if err != nil {
		log.Fatalf("try estimateGas error: %s", err.Error())
	}

	// opts from contract
	auth := bind.NewKeyedTransactor(privateKey)

	auth.Nonce = big.NewInt(int64(nonce))
	auth.Value = big.NewInt(0)
	auth.GasLimit = uint64(gasLimit)
	auth.GasPrice = gasPrice

	// baby pub key
	var babyPubKey *big.Int
	pkCompB := hezcommon.SwapEndianness(bjjWallet.PublicKey[:])
	babyPubKey = new(big.Int).SetBytes(pkCompB)
	// from idx from Mario's account : 264
	fromIdxBig := big.NewInt(0)
	// loadAmoundF
	// depositAmount, _ := new(big.Int).SetString("1000000000000000000000", 10)
	loadAmountF, err := hezcommon.NewFloat40(big.NewInt(0))
	// amountF
	amountF, err := hezcommon.NewFloat40(big.NewInt(0))
	// tokenID: HEZ Token = 1 to Goerli
	tokenID := uint32(0)
	// toIdx
	toIdxBig := big.NewInt(0)
	// permit -> var permit []byte
	var permit []byte
	permit = []byte("0x")

	log.Println("try send l1 tx to", bjjWallet.PublicKey)
	txs, err := hermezSC.AddL1Transaction(auth, babyPubKey, fromIdxBig, big.NewInt(int64(loadAmountF)), big.NewInt(int64(amountF)), tokenID, toIdxBig, permit)
	if err != nil {
		log.Printf("Error to create L1 tx to %s (private key) -> Error: %s \n", bjjWallet.PublicKey, err.Error())
		return
	}
	log.Printf("%+v\n", txs)
}
