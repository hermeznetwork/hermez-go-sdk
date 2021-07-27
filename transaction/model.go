package transaction

import (
	"time"

	"github.com/hermeznetwork/hermez-go-sdk/token"
	"github.com/hermeznetwork/hermez-node/api/apitypes"
	hezcommon "github.com/hermeznetwork/hermez-node/common"
	"github.com/iden3/go-iden3-crypto/babyjub"
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

type PoolTxAPI struct {
	ItemID               uint64                  `json:"itemId"`
	TxID                 hezcommon.TxID          `json:"id"`
	FromIdx              apitypes.HezIdx         `json:"fromAccountIndex"`
	EffectiveFromEthAddr apitypes.HezEthAddr     `json:"fromHezEthereumAddress"`
	EffectiveFromBJJ     apitypes.HezBJJ         `json:"fromBJJ"`
	ToIdx                apitypes.HezIdx         `json:"toAccountIndex"`
	EffectiveToEthAddr   apitypes.HezEthAddr     `json:"toHezEthereumAddress"`
	EffectiveToBJJ       apitypes.HezBJJ         `json:"toBJJ"`
	Amount               apitypes.BigIntStr      `json:"amount"`
	Fee                  hezcommon.FeeSelector   `json:"fee"`
	Nonce                hezcommon.Nonce         `json:"nonce"`
	State                hezcommon.PoolL2TxState `json:"state"`
	MaxNumBatch          uint32                  `json:"maxNumBatch"`
	Info                 string                  `json:"info"`
	ErrorCode            int                     `json:"errorCode"`
	ErrorType            string                  `json:"errorType"`
	Signature            babyjub.SignatureComp   `json:"signature"`
	RqFromIdx            apitypes.HezIdx         `json:"requestFromAccountIndex"`
	RqToIdx              apitypes.HezIdx         `json:"requestToAccountIndex"`
	RqToEthAddr          apitypes.HezEthAddr     `json:"requestToHezEthereumAddress"`
	RqToBJJ              apitypes.HezBJJ         `json:"requestToBJJ"`
	RqTokenID            hezcommon.TokenID       `json:"requestTokenId"`
	RqAmount             apitypes.BigIntStr      `json:"requestAmount"`
	RqFee                hezcommon.FeeSelector   `json:"requestFee"`
	RqNonce              hezcommon.Nonce         `json:"requestNonce"`
	Type                 hezcommon.TxType        `json:"type"`
	Timestamp            time.Time               `json:"timestamp"`
	TotalItems           uint64                  `json:"total_items"`
	Token                token.Token             `json:"token"`
}

type TransactionsAPIResponse struct {
	Transactions []PoolTxAPI `json:"transactions"`
}

type TxReceiverMetadata struct {
	ToEthAddr   string `json:"to_eth_addr"`
	FeeSelector uint   `json:"fee_selector"`
	Amount      string `json:"amount"`
}
