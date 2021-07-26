package transaction

import (
	"encoding/base64"
	"errors"
	"fmt"
	"log"
	"math/big"
	"strconv"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	ethCommon "github.com/ethereum/go-ethereum/common"
	"github.com/hermeznetwork/hermez-go-sdk/account"
	hezCommon "github.com/hermeznetwork/hermez-node/common"
	"github.com/iden3/go-iden3-crypto/babyjub"
)

// NewHermezAPITxRequest convert L2 tx to Hermez API request model
func NewHermezAPITxRequest(poolTx *hezCommon.PoolL2Tx, token hezCommon.Token) APITx {
	toIdx := "hez:ETH:0"
	if poolTx.ToIdx > 0 {
		toIdx = IdxToHez(poolTx.ToIdx, token.Symbol)
	}
	toEth := ""
	if poolTx.ToEthAddr != hezCommon.EmptyAddr {
		toEth = ethAddrToHez(poolTx.ToEthAddr)
	}
	toBJJ := BjjToString(poolTx.ToBJJ)
	if poolTx.ToBJJ != hezCommon.EmptyBJJComp {
		toBJJ = BjjToString(poolTx.ToBJJ)
	}
	return APITx{
		TxID:      poolTx.TxID,
		Type:      string(poolTx.Type),
		TokenID:   uint32(poolTx.TokenID),
		FromIdx:   IdxToHez(poolTx.FromIdx, token.Symbol),
		ToIdx:     toIdx,
		ToEthAddr: toEth,
		ToBJJ:     toBJJ,
		Amount:    poolTx.Amount.String(),
		Fee:       uint64(poolTx.Fee),
		Nonce:     uint64(poolTx.Nonce),
		Signature: poolTx.Signature.String(),
	}
}

func NewSignedAPITxFromEthAddr(chainID int, fromBjjWallet account.BJJWallet, fromIdx int64, toEthAddress string, amount *big.Int, feeSelector hezCommon.FeeSelector, token hezCommon.Token, nonce int) (APITx, error) {

	f40Amount, err := AmountToFloat40(amount)
	if err != nil {
		return APITx{}, err
	}

	tx := &hezCommon.PoolL2Tx{
		FromIdx:   hezCommon.Idx(fromIdx),
		ToEthAddr: ethCommon.HexToAddress(toEthAddress),
		ToBJJ:     hezCommon.EmptyBJJComp,
		ToIdx:     0,
		Amount:    f40Amount,
		Fee:       hezCommon.FeeSelector(uint8(feeSelector)),
		TokenID:   hezCommon.TokenID(token.TokenID),
		Nonce:     hezCommon.Nonce(nonce),
		Type:      hezCommon.TxTypeTransferToEthAddr,
	}

	return SignAPITx(chainID, fromBjjWallet, token, tx)
}

func SignAPITx(chainID int, fromBjjWallet account.BJJWallet, token hezCommon.Token, tx *hezCommon.PoolL2Tx) (APITx, error) {
	tx, err := hezCommon.NewPoolL2Tx(tx)
	if err != nil {
		return APITx{}, err
	}

	txHash, err := tx.HashToSign(uint16(chainID))
	if err != nil {
		return APITx{}, err
	}

	signedTx := fromBjjWallet.PrivateKey.SignPoseidon(txHash)
	tx.Signature = signedTx.Compress()

	t := hezCommon.Token{
		TokenID: hezCommon.TokenID(token.TokenID),
		Symbol:  token.Symbol,
	}

	apiTxRequest := NewHermezAPITxRequest(tx, t)

	return apiTxRequest, nil
}

// IdxToHez convert idx to hez idx
func IdxToHez(idx hezCommon.Idx, tokenSymbol string) string {
	// log.Printf("idx %+v\ntoken: %s\n", idx, tokenSymbol)
	return "hez:" + tokenSymbol + ":" + strconv.Itoa(int(idx))
}

func HezToIdx(hezIdx string) (int64, error) {
	hezIdxSlice := strings.Split(hezIdx, ":")

	if len(hezIdxSlice) == 3 {
		if idx, err := strconv.ParseInt(hezIdxSlice[2], 10, 64); err == nil {
			return idx, nil
		}
	}

	return 0, errors.New("invalid hezIdx")
}

func AmountToFloat40(amount *big.Int) (*big.Int, error) {

	amountfloat40, err := hezCommon.NewFloat40Floor(amount)
	if err != nil {
		return nil, err
	}

	f40Amount, err := amountfloat40.BigInt()
	if err != nil {
		return nil, err
	}

	return f40Amount, nil
}

// EthAddrToHez convert eth address to hez address
func ethAddrToHez(addr common.Address) string {
	return "hez:" + addr.String()
}

// BjjToString convert the BJJ public key to string
func BjjToString(bjj babyjub.PublicKeyComp) string {
	pkComp := [32]byte(bjj)
	sum := pkComp[0]
	for i := 1; i < len(pkComp); i++ {
		sum += pkComp[i]
	}
	bjjSum := append(pkComp[:], sum)
	return "hez:" + base64.RawURLEncoding.EncodeToString(bjjSum)
}

// MarshalTransaction marshal transaction information into a Hermez transaction API request
func MarshalTransaction(itemToTransfer string,
	senderAcctDetails account.AccountAPIResponse,
	receiverAcctDetails account.AccountAPIResponse,
	senderBjjWallet account.BJJWallet,
	amount *big.Int,
	feeSelector int,
	ethereumChainID int) (apiTxRequest APITx, err error) {

	var token hezCommon.Token
	var nonce hezCommon.Nonce
	var fromIdx, toIdx hezCommon.Idx

	// Get from innerAccount Token and nonce details from sender innerAccount
	for _, innerAccount := range senderAcctDetails.Accounts {
		if strings.ToUpper(innerAccount.Token.Symbol) == itemToTransfer {
			token.TokenID = hezCommon.TokenID(innerAccount.Token.ID)
			token.Symbol = innerAccount.Token.Symbol
			nonce = hezCommon.Nonce(innerAccount.Nonce)
			tempAccountsIdx := strings.Split(innerAccount.AccountIndex, ":")
			if len(tempAccountsIdx) == 3 {
				tempAccIdx, errAtoi := strconv.Atoi(tempAccountsIdx[2])
				if errAtoi != nil {
					err = fmt.Errorf("[MarshalTransaction] Error getting sender Account index. Account: %+v - Error: %s\n", innerAccount, err.Error())
					return
				}
				fromIdx = hezCommon.Idx(tempAccIdx)
			}
		}
	}

	// Get from innerAccount Token and nonce details from receiver innerAccount
	for _, innerAccount := range receiverAcctDetails.Accounts {
		if strings.ToUpper(innerAccount.Token.Symbol) == itemToTransfer {
			tempAccountsIdx := strings.Split(innerAccount.AccountIndex, ":")
			if len(tempAccountsIdx) == 3 {
				tempAccIdx, errAtoi := strconv.Atoi(tempAccountsIdx[2])
				if errAtoi != nil {
					log.Printf("[MarshalTransaction] Error getting receipient Account index. Account: %+v - Error: %s\n", innerAccount, err.Error())
					return
				}
				toIdx = hezCommon.Idx(tempAccIdx)
			}
		}
	}

	// If there is no innerAccount created to this specific token stop the code
	if len(fromIdx.String()) < 1 {
		err = fmt.Errorf("[MarshalTransaction] There is no sender Account to this user %s for this Token %s", senderBjjWallet.HezBjjAddress, itemToTransfer)
		log.Println(err.Error())
		return
	}

	// If there is no innerAccount created to this specific token stop the code
	if len(toIdx.String()) < 1 {
		err = fmt.Errorf("[MarshalTransaction] There is no receipient Account to this user %+v for this Token %s", receiverAcctDetails, itemToTransfer)
		log.Println(err.Error())
		return
	}

	// fee := hezcommon.FeeSelector(100)
	fee := hezCommon.FeeSelector(uint8(feeSelector)) // 10.2%

	tx := new(hezCommon.PoolL2Tx)
	tx.FromIdx = fromIdx
	tx.ToEthAddr = hezCommon.EmptyAddr
	tx.ToBJJ = hezCommon.EmptyBJJComp
	tx.ToIdx = toIdx
	tx.Amount = amount
	tx.Fee = fee
	tx.TokenID = token.TokenID
	tx.Nonce = nonce
	tx.Type = hezCommon.TxTypeTransfer

	// log.Println("")
	// log.Println("[MarshalTransaction] hezcommon.PoolL2Tx")
	// log.Printf("%+v\n\n", tx)

	tx, err = hezCommon.NewPoolL2Tx(tx)
	if err != nil {
		err = fmt.Errorf("[MarshalTransaction] Error creating L2 TX Pool object. TX: %+v - Error: %s\n", tx, err.Error())
		return
	}

	// log.Println("[MarshalTransaction] after hezcommon.NewPoolL2Tx")
	// log.Printf("%+v\n\n", tx)

	txHash, err := tx.HashToSign(uint16(ethereumChainID))
	if err != nil {
		err = fmt.Errorf("[MarshalTransaction] Error generating tx hash. TX: %+v - Error: %s\n", tx, err.Error())
		return
	}

	signedTx := senderBjjWallet.PrivateKey.SignPoseidon(txHash)
	tx.Signature = signedTx.Compress()

	// log.Println("[MarshalTransaction] tx signed")
	// log.Printf("%+v\n\n", tx)

	apiTxRequest = NewHermezAPITxRequest(tx, token)

	// log.Println("[MarshalTransaction] apiTxRequest")
	// log.Printf("%+v\n\n", apiTxRequest)
	return
}
