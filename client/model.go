package client

import (
	"net/http"

	"github.com/dghubble/sling"
	"github.com/ethereum/go-ethereum/ethclient"
	HermezAuctionProtocol "github.com/hermeznetwork/hermez-node/eth/contracts/auction"
)

// HermezClient connect to Ethereum node and Hermez Coordinator and Smart Contracts
type HermezClient struct {
	EthClient                *ethclient.Client
	AuctionContract          *HermezAuctionProtocol.Auction
	HttpClient               http.Client
	BootCoordinatorURL       string
	BootCoordinatorClient    *sling.Sling
	CurrentCoordinatorURL    string
	CurrentCoordinatorClient *sling.Sling
	EthereumChainID          int
}

// SetCurrentCoordinator updates coordinator definitions based on current coordinator URL
func (hezClient *HermezClient) SetCurrentCoordinator(URL string) {
	hezClient.CurrentCoordinatorURL = URL
	httpClient := NewHttpClient()
	hezClient.CurrentCoordinatorClient = sling.New().Base(hezClient.CurrentCoordinatorURL).Client(&httpClient)
}
