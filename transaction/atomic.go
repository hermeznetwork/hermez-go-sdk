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

// CreateFullTxs turn the basic information in a PoolL2Tx, set metadata and fields based on the current state. Also
// links the txs setting the Rq* fields.
func CreateFullTxs(hezClient client.HermezClient, txs []AtomicTxItem) (fullTxs []hezcommon.PoolL2Tx, err error) {
	// configure transactions and do basic validations
	for currentAtomicTxId := range txs {
		localTx := hezcommon.PoolL2Tx{}
		hezcommon.NewPoolL2Tx(&localTx)

		localTx.ToEthAddr = ethCommon.HexToAddress(txs[currentAtomicTxId].RecipientAddress)
		localTx.ToBJJ = hezcommon.EmptyBJJComp
		localTx.Amount = txs[currentAtomicTxId].Amount
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
		if nonce == 0 {
			localTx.Nonce = 0
		} else {
			localTx.Nonce = nonce + 1
		}

		localTx.FromIdx = idx

		// Recipient Account
		idx, _, _, err = getAccountDetails(hezClient, txs[currentAtomicTxId].RecipientAddress, txs[currentAtomicTxId].TokenSymbolToTransfer)
		if err != nil {
			err = fmt.Errorf("[AtomicTransfer] Error obtaining receipient account details. Account: %s - Error: %s\n", txs[currentAtomicTxId].SenderBjjWallet.EthAccount.Address.Hex(), err.Error())
			return
		}
		localTx.ToIdx = idx
		if err != nil {
			return
		}

		fullTxs = append(fullTxs, localTx)
	}

	// Populate RqID and set the RqFields
	for i := range txs {
		position := 0
		if txs[i].RqOffSet > 0 && txs[i].RqOffSet < 4 {
			position = i + txs[i].RqOffSet
		} else {
			switch txs[i].RqOffSet {
			case 4:
				position = i - 4
			case 5:
				position = i - 3
			case 6:
				position = i - 2
			case 7:
				position = i - 1
			}
		}
		fullTxs[i].RqFromIdx = fullTxs[position].FromIdx
		fullTxs[i].RqToIdx = fullTxs[position].ToIdx
		fullTxs[i].RqToEthAddr = fullTxs[position].ToEthAddr
		fullTxs[i].RqToBJJ = fullTxs[position].ToBJJ
		fullTxs[i].RqNonce = fullTxs[position].Nonce
		fullTxs[i].RqFee = fullTxs[position].Fee
		fullTxs[i].RqAmount = fullTxs[position].Amount
		fullTxs[i].RqTokenID = fullTxs[position].TokenID
		fullTxs[i].RqOffset = uint8(txs[i].RqOffSet)
	}
	return
}

// SetAtomicGroupID defines the AtomicGroup ID and propagate to txs
func SetAtomicGroupID(atomicGroup api.AtomicGroup) api.AtomicGroup {
	// Generate atomic group id
	atomicGroup.SetAtomicGroupID()

	for currentAtomicTxId := range atomicGroup.Txs {
		atomicGroup.Txs[currentAtomicTxId].AtomicGroupID = atomicGroup.ID
	}
	return atomicGroup
}

// AtomicTransfer creates PoolL2Txs using basic information provided in the AtomicTxItems, set metadata and fields based
// on the current state. Also links the txs setting the Rq* fields and sign txs. After performs token or ETH transfers
// in a pool of transactions.
func AtomicTransfer(hezClient client.HermezClient, ethereumChainID int,
	txs []AtomicTxItem) (serverResponse string, err error) {
	atomicGroup := api.AtomicGroup{}

	// create PoolL2Txs
	atomicGroup.Txs, err = CreateFullTxs(hezClient, txs)
	if err != nil {
		err = fmt.Errorf("[AtomicTransfer] Error generating PoolL2Tx. Error: %s\n", err.Error())
		return
	}

	// set AtomicGroupID
	atomicGroup = SetAtomicGroupID(atomicGroup)

	// Sign the txs
	for i := range txs {
		var txHash *big.Int
		txHash, err = atomicGroup.Txs[i].HashToSign(uint16(ethereumChainID))
		if err != nil {
			err = fmt.Errorf("[AtomicTransfer] Error generating currentAtomicTxItem hash. TX: %+v - Error: %s\n", atomicGroup.Txs[i], err.Error())
			return
		}
		signedTx := txs[i].SenderBjjWallet.PrivateKey.SignPoseidon(txHash)
		atomicGroup.Txs[i].Signature = signedTx.Compress()
	}

	// Post
	serverResponse, err = SendAtomicTxsGroup(hezClient, atomicGroup)
	if err != nil {
		err = fmt.Errorf("[AtomicTransfer] Error sending transactions. Error: %s\n", err.Error())
		return
	}

	return
}
