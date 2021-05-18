package account

import (
	"time"

	"github.com/iden3/go-iden3-crypto/babyjub"
)

type AccountAPIResponse struct {
	Accounts     []Account `json:"accounts"`
	PendingItems int       `json:"pendingItems"`
}

type Account struct {
	AccountIndex       string `json:"accountIndex"`
	Balance            string `json:"balance"`
	BJJAddress         string `json:"bjj"`
	HezEthereumAddress string `json:"hezEthereumAddress"`
	ItemID             int    `json:"itemId"`
	Nonce              int    `json:"nonce"`
	Token              Token  `json:"token"`
}

type Token struct {
	USD              float64   `json:"USD"`
	Decimals         int       `json:"decimals"`
	EthereumAddress  string    `json:"ethereumAddress"`
	EthereumBlockNum int       `json:"ethereumBlockNum"`
	FiatUpdate       time.Time `json:"fiatUpdate"`
	ID               int       `json:"id"`
	ItemID           int       `json:"itemId"`
	Name             string    `json:"name"`
	Symbol           string    `json:"symbol"`
}

// BJJWallet BJJ Wallet
type BJJWallet struct {
	PrivateKey    babyjub.PrivateKey
	PublicKey     babyjub.PublicKeyComp
	HezBjjAddress string
	HezEthAddress string
}
