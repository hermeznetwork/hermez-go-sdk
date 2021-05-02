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

func NewHermezClient() (hezClient HermezClient, err error) {
	ethClient, err := getCustomEthereumClient("http://191.252.179.69:8545")
	if err != nil {
		log.Printf("Error during ETH client initialization: %s\n", err.Error())
		return
	}
	hezClient.EthClient = ethClient
	auctionContractAddress := common.HexToAddress("0x1D5c3Dd2003118743D596D7DB7EA07de6C90fB20")
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

	hezClient.BootCoordinatorURL = bootCoordURL
	hezClient.BootCoordinatorClient = sling.New().Base(bootCoordURL).Client(newHttpClient())
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

func newHttpClient() *http.Client {
	tr := &http.Transport{
		MaxIdleConns:       defaultMaxIdleConns,
		IdleConnTimeout:    defaultIdleConnTimeout,
		DisableCompression: true,
	}
	httpClient := new(http.Client)
	httpClient.Timeout = defaultTimeoutCall
	httpClient.Transport = tr
	return httpClient
}
