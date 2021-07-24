package token

import "time"

type TokensAPIResponse struct {
	Tokens []Token `json:"tokens"`
}

type Token struct {
	ItemID           int       `json:"itemId"`
	ID               int       `json:"id"`
	EthereumBlockNum int       `json:"ethereumBlockNum"`
	EthereumAddress  string    `json:"ethereumAddress"`
	Name             string    `json:"name"`
	Symbol           string    `json:"symbol"`
	Decimals         int       `json:"decimals"`
	USD              float64   `json:"USD"`
	FiatUpdate       time.Time `json:"fiatUpdate"`
}
