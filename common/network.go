package common

import (
	"errors"
	"sync"

	"github.com/ethereum/go-ethereum/common"
)

// EthereumNetworks list of ethereum networks where Hermez Network connects and his definitions in each ethereum network
var EthereumNetworks sync.Map

type HermezDefinitionEthereumNetwork struct {
	ChainID                int
	RollupContractAddress  common.Address
	AuctionContractAddress common.Address
}

func init() {
	mainnet := HermezDefinitionEthereumNetwork{}
	mainnet.ChainID = 1
	mainnet.RollupContractAddress = common.HexToAddress("0xA68D85dF56E733A06443306A095646317B5Fa633")
	mainnet.AuctionContractAddress = common.HexToAddress("0x15468b45eD46C8383F5c0b1b6Cf2EcF403C2AeC2")
	EthereumNetworks.Store("mainnet", mainnet)

	rinkeby := HermezDefinitionEthereumNetwork{}
	rinkeby.ChainID = 4
	rinkeby.RollupContractAddress = common.HexToAddress("0x0a8a6D65Ad9046c2a57a5Ca8Bab2ae9c3345316d")
	rinkeby.AuctionContractAddress = common.HexToAddress("0x15468b45eD46C8383F5c0b1b6Cf2EcF403C2AeC2")
	EthereumNetworks.Store("rinkeby", rinkeby)

	goerli := HermezDefinitionEthereumNetwork{}
	goerli.ChainID = 5
	goerli.RollupContractAddress = common.HexToAddress("0xe6E56C74630F8eE824039308794639D5a02BF9E5")
	goerli.AuctionContractAddress = common.HexToAddress("0x748964F22eFd023eB78A246A7AC2506e84CC4545")
	EthereumNetworks.Store("goerli", goerli)
}

// GetNetworkDefinition return Hermez network definitions (contract addresses) per Ethereum network
func GetNetworkDefinition(networkName string) (network HermezDefinitionEthereumNetwork, err error) {
	err = nil
	tmp, ok := EthereumNetworks.Load(networkName)
	if !ok {
		err = errors.New("hermez definition to this network " + networkName + " not found")
		return
	}
	network = tmp.(HermezDefinitionEthereumNetwork)
	return
}
