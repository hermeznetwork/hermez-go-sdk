package transaction

import (
	"fmt"
	hezcommon "github.com/hermeznetwork/hermez-node/common"
	"log"
	"math/big"
	"strconv"
	"strings"

	"github.com/hermeznetwork/hermez-go-sdk/account"
	"github.com/hermeznetwork/hermez-go-sdk/client"
)

type AtomicTxItem struct {
	SenderBjjWallet       account.BJJWallet
	RecipientAddress      string
	TokenSymbolToTransfer string
	Amount                *big.Int
	FeeRangeSelectedID    int
	RqOffSet              int
	EthereumChainID       int
}

// AtomicTransfer perform token or ETH transfers in a pool of transactions
func AtomicTransfer(hezClient client.HermezClient,
	txs []AtomicTxItem) (serverResponse string, err error) {
	atomicGroup := new(AtomicGroup)

	// configure transactions and do basic validations
	for i, currentAtomicTxItem := range txs {
		atx := new(AtomicTx)
		atx.Type = string(hezcommon.TxTypeTransfer)
		atx.Amount = currentAtomicTxItem.Amount.String()
		atx.Fee = uint64(hezcommon.FeeSelector(uint8(currentAtomicTxItem.FeeRangeSelectedID)))
		atx.ToBJJ = hezcommon.EmptyBJJComp.String()
		atx.ToEthAddr = hezcommon.EmptyAddr.String()
		atx.RqOffSet = currentAtomicTxItem.RqOffSet

		localTx := new(hezcommon.PoolL2Tx)
		localTx.ToEthAddr = hezcommon.EmptyAddr
		localTx.ToBJJ = hezcommon.EmptyBJJComp
		localTx.Amount = currentAtomicTxItem.Amount
		localTx.Type = hezcommon.TxTypeTransfer
		localTx.Fee = hezcommon.FeeSelector(uint8(currentAtomicTxItem.FeeRangeSelectedID))

		// SenderAccount
		senderAccDetails, err2 := account.GetAccountInfo(hezClient, currentAtomicTxItem.SenderBjjWallet.EthAccount.Address.Hex())
		if err2 != nil {
			err = fmt.Errorf("[AtomicTransfer] Error obtaining account details. Account: %s - Error: %s\n", currentAtomicTxItem.SenderBjjWallet.HezEthAddress, err.Error())
			return
		}
		for _, innerAccount := range senderAccDetails.Accounts {
			if strings.ToUpper(innerAccount.Token.Symbol) == currentAtomicTxItem.TokenSymbolToTransfer {
				localTx.TokenID = hezcommon.TokenID(innerAccount.Token.ID)
				atx.TokenID = uint32(hezcommon.TokenID(innerAccount.Token.ID))
				localTx.Nonce = hezcommon.Nonce(innerAccount.Nonce)
				atx.Nonce = uint64(hezcommon.Nonce(innerAccount.Nonce))
				tempAccountsIdx := strings.Split(innerAccount.AccountIndex, ":")
				if len(tempAccountsIdx) == 3 {
					tempAccIdx, errAtoi := strconv.Atoi(tempAccountsIdx[2])
					if errAtoi != nil {
						err = fmt.Errorf("[AtomicTransfer] Error getting sender Account index. Account: %+v - Error: %s\n", innerAccount, err.Error())
						return
					}
					localTx.FromIdx = hezcommon.Idx(tempAccIdx)
					atx.FromIdx = string(hezcommon.Idx(tempAccIdx))
				}
			}
		}
		if len(atx.FromIdx) < 1 {
			err = fmt.Errorf("[AtomicTransfer] There is no sender Account to this user %s for this Token %s", currentAtomicTxItem.SenderBjjWallet.HezBjjAddress, currentAtomicTxItem.TokenSymbolToTransfer)
			log.Println(err.Error())
			return
		}

		// Recipient Account
		recipientAccDetails, err3 := account.GetAccountInfo(hezClient, currentAtomicTxItem.RecipientAddress)
		if err3 != nil {
			err = fmt.Errorf("[AtomicTransfer] Error obtaining account details. Account: %s - Error: %s\n", currentAtomicTxItem.RecipientAddress, err.Error())
			return
		}
		for _, innerAccount := range recipientAccDetails.Accounts {
			if strings.ToUpper(innerAccount.Token.Symbol) == currentAtomicTxItem.TokenSymbolToTransfer {
				tempAccountsIdx := strings.Split(innerAccount.AccountIndex, ":")
				if len(tempAccountsIdx) == 3 {
					tempAccIdx, errAtoi := strconv.Atoi(tempAccountsIdx[2])
					if errAtoi != nil {
						log.Printf("[AtomicTransfer] Error getting receipient Account index. Account: %+v - Error: %s\n", innerAccount, err.Error())
						return
					}
					localTx.ToIdx = hezcommon.Idx(tempAccIdx)
					atx.ToIdx = string(hezcommon.Idx(tempAccIdx))
				}
			}
		}
		if len(atx.ToIdx) < 1 {
			err = fmt.Errorf("[AtomicTransfer] There is no receipient Account to this user %+v for this Token %s", recipientAccDetails, currentAtomicTxItem.TokenSymbolToTransfer)
			log.Println(err.Error())
			return
		}

		// Signature
		htx, err4 := hezcommon.NewPoolL2Tx(localTx)
		if err4 != nil {
			err = fmt.Errorf("[AtomicTransfer] Error creating L2 TX Pool object. TX: %+v - Error: %s\n", currentAtomicTxItem, err.Error())
			return
		}
		txHash, err5 := htx.HashToSign(uint16(currentAtomicTxItem.EthereumChainID))
		if err5 != nil {
			err = fmt.Errorf("[AtomicTransfer] Error generating currentAtomicTxItem hash. TX: %+v - Error: %s\n", currentAtomicTxItem, err.Error())
			return
		}
		signedTx := currentAtomicTxItem.SenderBjjWallet.PrivateKey.SignPoseidon(txHash)
		atx.Signature = signedTx.Compress().String()

		// Idx
		atx.TxID = htx.TxID
		if i == 0 {
			atomicGroup.AtomicGroupId = atx.TxID
		}

		// Populate array
		atomicGroup.AddAtomicItem(AtomicTx{
			TxID:      atx.TxID,
			Type:      atx.Type,
			TokenID:   atx.TokenID,
			FromIdx:   atx.FromIdx,
			ToIdx:     atx.ToIdx,
			ToEthAddr: atx.ToEthAddr,
			ToBJJ:     atx.ToBJJ,
			Amount:    atx.Amount,
			Fee:       atx.Fee,
			Nonce:     atx.Nonce,
			Signature: atx.Signature,
			RqID:      atx.RqID,
			RqOffSet:  atx.RqOffSet,
		})
	}

	// Populate RqID
	for current, tx := range atomicGroup.Txs {
		if tx.RqOffSet > 0 && tx.RqOffSet < 4 {
			atomicGroup.Txs[current].RqID = atomicGroup.Txs[current+tx.RqOffSet].TxID
		} else {
			switch tx.RqOffSet {
			case 4:
				atomicGroup.Txs[current].RqID = atomicGroup.Txs[current-4].TxID
			case 5:
				atomicGroup.Txs[current].RqID = atomicGroup.Txs[current-3].TxID
			case 6:
				atomicGroup.Txs[current].RqID = atomicGroup.Txs[current-2].TxID
			case 7:
				atomicGroup.Txs[current].RqID = atomicGroup.Txs[current-1].TxID
			}
		}
	}

	// Post
	serverResponse, err = ExecuteAtomicTransaction(hezClient, atomicGroup)

	return
}
