package client

import (
	"log"
	"net/http"
	"time"

	"github.com/dghubble/sling"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"

	HermezAuctionProtocol "github.com/hermeznetwork/hermez-node/eth/contracts/auction"
)

const (
	defaultMaxIdleConns    = 10
	defaultIdleConnTimeout = 2 * time.Second
	defaultTimeoutCall     = 2 * time.Minute
)

func NewHermezClient(nodeURL string, auctionContractAddressHex string, ethereumChainID int) (hezClient HermezClient, err error) {
	ethClient, err := getCustomEthereumClient(nodeURL)
	if err != nil {
		log.Printf("Error during ETH client initialization: %s\n", err.Error())
		return
	}
	hezClient.EthClient = ethClient
	auctionContractAddress := common.HexToAddress(auctionContractAddressHex)
	auctionContract, err := HermezAuctionProtocol.NewHermezAuctionProtocol(auctionContractAddress, hezClient.EthClient)
	if err != nil {
		log.Printf("Error during Auction smart contract wrapper initialization: %s\n", err.Error())
		return
	}
	hezClient.AuctionContract = auctionContract
	bootCoordURL, err := hezClient.AuctionContract.BootCoordinatorURL(nil)
	if err != nil {
		log.Printf("Error during boot coordinator url query: %s\n", err.Error())
		return
	}

	hezClient.EthereumChainID = ethereumChainID
	hezClient.HttpClient = NewHttpClient()
	bootCoordHttpClient := NewHttpClient()
	hezClient.BootCoordinatorURL = bootCoordURL
	hezClient.BootCoordinatorClient = sling.New().Base(bootCoordURL).Client(&bootCoordHttpClient)
	return
}

/*
getCustomEthereumClient connects and return a client to user defined Ethereum network
*/
func getCustomEthereumClient(URL string) (client *ethclient.Client, err error) {
	err = nil
	client, err = ethclient.Dial(URL)
	if err != nil {
		log.Printf("There was a failure connecting to %s: %+v", URL, err)
		return
	}
	return
}

// NewHttpClient generates new HTTP Client
func NewHttpClient() http.Client {
	tr := &http.Transport{
		MaxIdleConns:       defaultMaxIdleConns,
		IdleConnTimeout:    defaultIdleConnTimeout,
		DisableCompression: true,
	}
	httpClient := new(http.Client)
	httpClient.Timeout = defaultTimeoutCall
	httpClient.Transport = tr
	return *httpClient
}
