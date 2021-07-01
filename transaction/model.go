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

// AtomicTx is a representation of a member of an atomic group
type AtomicTx struct {
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
	RqID      hezcommon.TxID `json:"requestId" binding:"required"`
	RqOffSet  int            `json:"requestOffset" binding:"required"`
}

// AtomicGroup is a group of transactions that should run all or fail all
type AtomicGroup struct {
	AtomicGroupId hezcommon.TxID `json:"atomicGroupId" binding:"required"`
	Txs           []AtomicTx     `json:"transactions" binding:"required"`
}

// AddAtomicItem Add a tx to txs
func (ag *AtomicGroup) AddAtomicItem(tx AtomicTx) {
	ag.Txs = append(ag.Txs, tx)
}
