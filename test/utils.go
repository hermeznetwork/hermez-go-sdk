package integration

import (
	"fmt"
	"io/ioutil"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/hermeznetwork/hermez-go-sdk/account"
	"github.com/hermeznetwork/hermez-go-sdk/client"
	"github.com/hermeznetwork/hermez-go-sdk/transaction"
	"github.com/hermeznetwork/hermez-node/common"
	"github.com/hermeznetwork/hermez-node/log"
)

const (
	configPath = "./config.toml"
)

type config struct {
	AuctionAddress string
	EthNodeURL     string
	HezNodeURL     string
	EthPrivKeys    []string
}

type IntegrationTest struct {
	Client  client.HermezClient
	Wallets []account.BJJWallet
}

func NewIntegrationTest() (IntegrationTest, error) {
	// Load config file
	it := IntegrationTest{}
	conf, err := loadConfig()
	if err != nil {
		return it, err
	}
	// Create SDK client
	hezClient, err := client.NewHermezClient(conf.EthNodeURL, conf.AuctionAddress)
	if err != nil {
		return it, err
	}
	hezClient.BootCoordinatorURL = conf.HezNodeURL
	hezClient.CurrentCoordinatorURL = conf.HezNodeURL
	// Create wallets
	wallets := []account.BJJWallet{}
	for _, privK := range conf.EthPrivKeys {
		w, _, err := account.CreateBjjWalletFromHexPvtKey(privK)
		if err != nil {
			return it, err
		}
		wallets = append(wallets, w)
	}

	return IntegrationTest{
		Client:  hezClient,
		Wallets: wallets,
	}, nil
}

func loadConfig() (config, error) {
	f, err := ioutil.ReadFile(configPath)
	if err != nil {
		return config{}, err
	}
	cfgToml := string(f)
	c := config{}
	_, err = toml.Decode(cfgToml, &c)
	return c, err
}

func WaitUntilTxsAreForged(txs []common.PoolL2Tx, timeBetweenChecks time.Duration, hezClient client.HermezClient) error {
	for {
		// IsForged?
		forgedTxs := 0
		for _, tx := range txs {
			_, err := transaction.GetL2Tx(hezClient, tx.TxID)
			if err == nil {
				forgedTxs += 1
				continue
			} else if err.Error() == "tx not forged" {
				log.Info("txs not forged yet, waiting before retry")
				time.Sleep(timeBetweenChecks)
				break
			} else {
				return fmt.Errorf("Unexpected error: %e", err)
			}
		}
		if forgedTxs == 0 {
			// 0 txs forged, keep trying
			continue
		} else if forgedTxs == len(txs) {
			// All txs forged, done
			return nil
		} else {
			return fmt.Errorf("SOME txs forged (%d/%d)", forgedTxs, len(txs))
		}
	}
}
