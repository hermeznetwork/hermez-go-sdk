package transaction

import (
	hezcommon "github.com/hermeznetwork/hermez-node/common"
)

// APITx is a representation of a transaction API request.
type APITx struct {
	TxID      hezcommon.TxID `json:"id" binding:"required"`
	Type      string         `json:"type"`
	TokenID   uint32         `json:"tokenId"`
	FromIdx   string         `json:"fromAccountIndex" binding:"required"`
	ToIdx     string         `json:"toAccountIndex"`
	ToEthAddr string         `json:"toHezEthereumAddress"`
	ToBJJ     string         `json:"toBjj"`
	Amount    string         `json:"amount" binding:"required"`
	Fee       uint64         `json:"fee"`
	Nonce     uint64         `json:"nonce"`
	Signature string         `json:"signature"`
}
