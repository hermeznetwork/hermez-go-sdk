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
	if len(strconv.FormatUint(uint64(idx), 10)) < 1 {
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
	for currentAtomicTxId := range txs {
		localTx := hezcommon.PoolL2Tx{}
		localTx.ToEthAddr = ethCommon.HexToAddress(txs[currentAtomicTxId].RecipientAddress)
		localTx.ToBJJ = hezcommon.EmptyBJJComp
		localTx.Amount = txs[currentAtomicTxId].Amount
		localTx.Type = hezcommon.TxTypeTransfer
		localTx.Fee = hezcommon.FeeSelector(uint8(txs[currentAtomicTxId].FeeRangeSelectedID))
		localTx.TokenSymbol = txs[currentAtomicTxId].TokenSymbolToTransfer

		// SenderAccount
		var idx hezcommon.Idx
		var nonce hezcommon.Nonce
		var tokenId hezcommon.TokenID
		idx, nonce, tokenId, err = getAccountDetails(hezClient, txs[currentAtomicTxId].SenderBjjWallet.EthAccount.Address.Hex(), txs[currentAtomicTxId].TokenSymbolToTransfer)
		if err != nil {
			err = fmt.Errorf("[AtomicTransfer] Error obtaining sender account details. Account: %s - Error: %s\n", txs[currentAtomicTxId].SenderBjjWallet.EthAccount.Address.Hex(), err.Error())
			return
		}
		localTx.TokenID = tokenId
		localTx.Nonce = nonce
		localTx.FromIdx = idx

		// Recipient Account
		idx, _, _, err = getAccountDetails(hezClient, txs[currentAtomicTxId].RecipientAddress, txs[currentAtomicTxId].TokenSymbolToTransfer)
		if err != nil {
			err = fmt.Errorf("[AtomicTransfer] Error obtaining receipient account details. Account: %s - Error: %s\n", txs[currentAtomicTxId].SenderBjjWallet.EthAccount.Address.Hex(), err.Error())
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
	for currentAtomicTxId := range txs {
		position := 0
		if txs[currentAtomicTxId].RqOffSet > 0 && txs[currentAtomicTxId].RqOffSet < 4 {
			position = currentAtomicTxId + txs[currentAtomicTxId].RqOffSet
		} else {
			switch txs[currentAtomicTxId].RqOffSet {
			case 4:
				position = currentAtomicTxId - 4
			case 5:
				position = currentAtomicTxId - 3
			case 6:
				position = currentAtomicTxId - 2
			case 7:
				position = currentAtomicTxId - 1
			}
		}
		atomicGroup.Txs[currentAtomicTxId].RqFromIdx = atomicGroup.Txs[position].FromIdx
		atomicGroup.Txs[currentAtomicTxId].RqToIdx = atomicGroup.Txs[position].ToIdx
		atomicGroup.Txs[currentAtomicTxId].RqToEthAddr = atomicGroup.Txs[position].ToEthAddr
		atomicGroup.Txs[currentAtomicTxId].RqToBJJ = atomicGroup.Txs[position].ToBJJ
		atomicGroup.Txs[currentAtomicTxId].RqNonce = atomicGroup.Txs[position].Nonce
		atomicGroup.Txs[currentAtomicTxId].RqFee = atomicGroup.Txs[position].Fee
		atomicGroup.Txs[currentAtomicTxId].RqAmount = atomicGroup.Txs[position].Amount
		atomicGroup.Txs[currentAtomicTxId].RqTokenSymbol = atomicGroup.Txs[position].TokenSymbol
		atomicGroup.Txs[currentAtomicTxId].RqOffset = uint8(txs[currentAtomicTxId].RqOffSet)
	}

	// Generate atomic group id
	atomicGroup.SetAtomicGroupID()

	// Sign the txs
	for currentAtomicTxId := range txs {
		atomicGroup.Txs[currentAtomicTxId].AtomicGroupID = atomicGroup.ID
		var txHash *big.Int
		txHash, err = atomicGroup.Txs[currentAtomicTxId].HashToSign(uint16(ethereumChainID))
		if err != nil {
			err = fmt.Errorf("[AtomicTransfer] Error generating currentAtomicTxItem hash. TX: %+v - Error: %s\n", atomicGroup.Txs[currentAtomicTxId], err.Error())
			return
		}
		signedTx := txs[currentAtomicTxId].SenderBjjWallet.PrivateKey.SignPoseidon(txHash)
		atomicGroup.Txs[currentAtomicTxId].Signature = signedTx.Compress()
	}

	// Post
	serverResponse, err = ExecuteAtomicTransaction(hezClient, atomicGroup)

	return
}
