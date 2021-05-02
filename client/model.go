package client

import (
	"github.com/dghubble/sling"
	"github.com/ethereum/go-ethereum/ethclient"
	HermezAuctionProtocol "github.com/hermeznetwork/hermez-node/eth/contracts/auction"
)

// HermezClient connect to Ethereum node and Hermez Coordinator and Smart Contracts
type HermezClient struct {
	EthClient               *ethclient.Client
	AuctionContract         *HermezAuctionProtocol.HermezAuctionProtocol
	BootCoordinatorURL      string
	BootCoordinatorClient   *sling.Sling
	ActualCoordinatorURL    string
	ActualCoordinatorClient *sling.Sling
}

// SetActualCoordinator updates coordinator definitions based on actual coordinator URL
func (hezClient *HermezClient) SetActualCoordinator(URL string) {
	hezClient.ActualCoordinatorURL = URL
	hezClient.ActualCoordinatorClient = sling.New().Base(hezClient.ActualCoordinatorURL).Client(newHttpClient())
}
