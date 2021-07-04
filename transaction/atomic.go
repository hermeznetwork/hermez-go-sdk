package transaction

import (
	"fmt"
	"math/big"
	"strconv"
	"strings"

	"github.com/hermeznetwork/hermez-go-sdk/account"
	"github.com/hermeznetwork/hermez-go-sdk/client"
	"github.com/hermeznetwork/hermez-node/api"

	ethCommon "github.com/ethereum/go-ethereum/common"
	hezcommon "github.com/hermeznetwork/hermez-node/common"
)

type AtomicTxItem struct {
	SenderBjjWallet       account.BJJWallet
	RecipientAddress      string
	TokenSymbolToTransfer string
	Amount                *big.Int
	FeeRangeSelectedID    int
	RqOffSet              int
}

func getAccountDetails(hezClient client.HermezClient, address string,
	tokenToTransfer string) (idx hezcommon.Idx, nonce hezcommon.Nonce, tokenId hezcommon.TokenID, err error) {
	var accDetails account.AccountAPIResponse
	accDetails, err = account.GetAccountInfo(hezClient, address)
	if err != nil {
		err = fmt.Errorf("error obtaining account details. Account: %s - Error: %s\n", address, err.Error())
		return
	}
	for _, innerAccount := range accDetails.Accounts {
		if strings.ToUpper(innerAccount.Token.Symbol) == tokenToTransfer {
			tokenId = hezcommon.TokenID(innerAccount.Token.ID)
			nonce = hezcommon.Nonce(innerAccount.Nonce)
			tempAccountsIdx := strings.Split(innerAccount.AccountIndex, ":")
			if len(tempAccountsIdx) == 3 {
				var tempAccIdx int
				tempAccIdx, err = strconv.Atoi(tempAccountsIdx[2])
				if err != nil {
					err = fmt.Errorf("error getting account index. Account: %+v - Error: %s\n", innerAccount, err.Error())
					return
				}
				idx = hezcommon.Idx(tempAccIdx)
			}
		}
	}
	if len(string(idx)) < 1 {
		err = fmt.Errorf("there is no account to this user %s for this Token %s", address, tokenToTransfer)
		return
	}
	return
}

// AtomicTransfer perform token or ETH transfers in a pool of transactions
func AtomicTransfer(hezClient client.HermezClient, ethereumChainID int,
	txs []AtomicTxItem) (serverResponse string, err error) {
	atomicGroup := api.AtomicGroup{}

	// configure transactions and do basic validations
	for _, currentAtomicTxItem := range txs {
		localTx := hezcommon.PoolL2Tx{}
		localTx.ToEthAddr = ethCommon.HexToAddress(currentAtomicTxItem.RecipientAddress)
		localTx.ToBJJ = hezcommon.EmptyBJJComp
		localTx.Amount = currentAtomicTxItem.Amount
		localTx.Type = hezcommon.TxTypeTransfer
		localTx.Fee = hezcommon.FeeSelector(uint8(currentAtomicTxItem.FeeRangeSelectedID))
		localTx.TokenSymbol = currentAtomicTxItem.TokenSymbolToTransfer

		// SenderAccount
		var idx hezcommon.Idx
		var nonce hezcommon.Nonce
		var tokenId hezcommon.TokenID
		idx, nonce, tokenId, err = getAccountDetails(hezClient, currentAtomicTxItem.SenderBjjWallet.EthAccount.Address.Hex(), currentAtomicTxItem.TokenSymbolToTransfer)
		if err != nil {
			err = fmt.Errorf("[AtomicTransfer] Error obtaining sender account details. Account: %s - Error: %s\n", currentAtomicTxItem.SenderBjjWallet.EthAccount.Address.Hex(), err.Error())
			return
		}
		localTx.TokenID = tokenId
		localTx.Nonce = nonce
		localTx.FromIdx = idx

		// Recipient Account
		idx, _, _, err = getAccountDetails(hezClient, currentAtomicTxItem.RecipientAddress, currentAtomicTxItem.TokenSymbolToTransfer)
		if err != nil {
			err = fmt.Errorf("[AtomicTransfer] Error obtaining receipient account details. Account: %s - Error: %s\n", currentAtomicTxItem.SenderBjjWallet.EthAccount.Address.Hex(), err.Error())
			return
		}
		localTx.ToIdx = idx
		err = localTx.SetID()
		if err != nil {
			return
		}

		atomicGroup.Txs = append(atomicGroup.Txs, localTx)
	}

	// Populate RqID and set the RqFields
	for current, currentAtomicTxItem := range txs {
		position := 0
		if currentAtomicTxItem.RqOffSet > 0 && currentAtomicTxItem.RqOffSet < 4 {
			position = current + currentAtomicTxItem.RqOffSet
		} else {
			switch currentAtomicTxItem.RqOffSet {
			case 4:
				position = current - 4
			case 5:
				position = current - 3
			case 6:
				position = current - 2
			case 7:
				position = current - 1
			}
		}
		atomicGroup.Txs[current].RqFromIdx = atomicGroup.Txs[position].FromIdx
		atomicGroup.Txs[current].RqToIdx = atomicGroup.Txs[position].ToIdx
		atomicGroup.Txs[current].RqToEthAddr = atomicGroup.Txs[position].ToEthAddr
		atomicGroup.Txs[current].RqToBJJ = atomicGroup.Txs[position].ToBJJ
		atomicGroup.Txs[current].RqNonce = atomicGroup.Txs[position].Nonce
		atomicGroup.Txs[current].RqFee = atomicGroup.Txs[position].Fee
		atomicGroup.Txs[current].RqAmount = atomicGroup.Txs[position].Amount
		atomicGroup.Txs[current].RqOffset = uint8(currentAtomicTxItem.RqOffSet)
	}

	// Generate atomic group id
	atomicGroup.SetAtomicGroupID()

	// Sign the txs
	for current, currentAtomicTxItem := range txs {
		var txHash *big.Int
		txHash, err = atomicGroup.Txs[current].HashToSign(uint16(ethereumChainID))
		if err != nil {
			err = fmt.Errorf("[AtomicTransfer] Error generating currentAtomicTxItem hash. TX: %+v - Error: %s\n", atomicGroup.Txs[current], err.Error())
			return
		}
		signedTx := currentAtomicTxItem.SenderBjjWallet.PrivateKey.SignPoseidon(txHash)
		atomicGroup.Txs[current].Signature = signedTx.Compress()
	}

	// Post
	serverResponse, err = ExecuteAtomicTransaction(hezClient, atomicGroup)

	return
}
