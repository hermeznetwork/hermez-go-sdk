package transaction

import (
	"fmt"
	"math/big"
	"strconv"
	"strings"

	"github.com/hermeznetwork/hermez-go-sdk/account"
	"github.com/hermeznetwork/hermez-go-sdk/client"

	ethCommon "github.com/ethereum/go-ethereum/common"
	hezCommon "github.com/hermeznetwork/hermez-node/common"
)

type AtomicTxItem struct {
	SenderBjjWallet       account.BJJWallet
	ReceiverAddress       string
	TokenSymbolToTransfer string
	Amount                *big.Int
	FeeRangeSelectedID    int
	RqOffSet              int
}

func getAccountDetails(hezClient client.HermezClient, address string,
	tokenToTransfer string) (idx hezCommon.Idx, nonce hezCommon.Nonce, tokenId hezCommon.TokenID, err error) {
	var accDetails account.AccountAPIResponse
	accDetails, err = account.GetAccountInfo(hezClient, address)
	if err != nil {
		err = fmt.Errorf("error obtaining account details. Account: %s - Error: %s\n", address, err.Error())
		return
	}
	for _, innerAccount := range accDetails.Accounts {
		if strings.ToUpper(innerAccount.Token.Symbol) == tokenToTransfer {
			tokenId = hezCommon.TokenID(innerAccount.Token.ID)
			nonce = hezCommon.Nonce(innerAccount.Nonce)
			tempAccountsIdx := strings.Split(innerAccount.AccountIndex, ":")
			if len(tempAccountsIdx) == 3 {
				var tempAccIdx int
				tempAccIdx, err = strconv.Atoi(tempAccountsIdx[2])
				if err != nil {
					err = fmt.Errorf("error getting account index. Account: %+v - Error: %s\n", innerAccount, err.Error())
					return
				}
				idx = hezCommon.Idx(tempAccIdx)
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
func CreateFullTxs(hezClient client.HermezClient, txs []AtomicTxItem) (fullTxs []hezCommon.PoolL2Tx, err error) {
	// configure transactions and do basic validations
	for currentAtomicTxId := range txs {
		localTx := hezCommon.PoolL2Tx{}
		localTx.ToEthAddr = ethCommon.HexToAddress(txs[currentAtomicTxId].ReceiverAddress)
		localTx.ToBJJ = hezCommon.EmptyBJJComp
		localTx.Amount = txs[currentAtomicTxId].Amount
		localTx.Fee = hezCommon.FeeSelector(uint8(txs[currentAtomicTxId].FeeRangeSelectedID))
		localTx.TokenSymbol = txs[currentAtomicTxId].TokenSymbolToTransfer

		// SenderAccount
		var idx hezCommon.Idx
		var nonce hezCommon.Nonce
		var tokenId hezCommon.TokenID
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

		// Receiver Account
		idx, _, _, err = getAccountDetails(hezClient, txs[currentAtomicTxId].ReceiverAddress, txs[currentAtomicTxId].TokenSymbolToTransfer)
		if err != nil {
			err = fmt.Errorf("[AtomicTransfer] Error obtaining receipient account details. Account: %s - Error: %s\n", txs[currentAtomicTxId].SenderBjjWallet.EthAccount.Address.Hex(), err.Error())
			return
		}
		localTx.ToIdx = idx
		if err != nil {
			return
		}

		hezCommon.NewPoolL2Tx(&localTx)
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
			case 0:
				err = fmt.Errorf("[CreateFullTxs]: Invalid rqOffSet")
				return
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
func SetAtomicGroupID(atomicGroup hezCommon.AtomicGroup) hezCommon.AtomicGroup {
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
func AtomicTransfer(hezClient client.HermezClient, txs []AtomicTxItem) (serverResponse string, atomicGroupID hezCommon.AtomicGroupID, err error) {
	atomicGroup := hezCommon.AtomicGroup{}

	// create PoolL2Txs
	atomicGroup.Txs, err = CreateFullTxs(hezClient, txs)
	if err != nil {
		err = fmt.Errorf("[AtomicTransfer] Error generating PoolL2Tx. Error: %s\n", err.Error())
		return
	}

	// set AtomicGroupID
	atomicGroup = SetAtomicGroupID(atomicGroup)
	atomicGroupID = atomicGroup.ID

	// Sign the txs
	for i := range txs {
		var txHash *big.Int
		txHash, err = atomicGroup.Txs[i].HashToSign(uint16(hezClient.EthereumChainID))
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

// AtomicTransferJSON receives an array of AtomicTxItems in JSON format, create PoolL2Txs, atomic group, sign and post
func AtomicTransferJSON(hezClient client.HermezClient, txsJSON []string) (serverResponse string, atomicGroupID hezCommon.AtomicGroupID, err error) {
	atomicGroup := hezCommon.AtomicGroup{}

	for _, currentJSON := range txsJSON {
		localTx := hezCommon.PoolL2Tx{}
		err = localTx.UnmarshalJSON([]byte(currentJSON))
		if err != nil {
			err = fmt.Errorf("[AtomicTransferJSON] Error Unmarshalling JSON %s - Error: %s\n", currentJSON, err.Error())
			return
		}
		hezCommon.NewPoolL2Tx(&localTx)
		atomicGroup.Txs = append(atomicGroup.Txs, localTx)
	}

	// Populate RqID and set the RqFields
	for i := range atomicGroup.Txs {
		position := 0
		if atomicGroup.Txs[i].RqOffset > 0 && atomicGroup.Txs[i].RqOffset < 4 {
			position = i + int(atomicGroup.Txs[i].RqOffset)
		} else {
			switch atomicGroup.Txs[i].RqOffset {
			case 4:
				position = i - 4
			case 5:
				position = i - 3
			case 6:
				position = i - 2
			case 7:
				position = i - 1
			case 0:
				err = fmt.Errorf("[CreateFullTxs]: Invalid rqOffSet")
				return
			}
		}
		atomicGroup.Txs[i].RqFromIdx = atomicGroup.Txs[position].FromIdx
		atomicGroup.Txs[i].RqToIdx = atomicGroup.Txs[position].ToIdx
		atomicGroup.Txs[i].RqToEthAddr = atomicGroup.Txs[position].ToEthAddr
		atomicGroup.Txs[i].RqToBJJ = atomicGroup.Txs[position].ToBJJ
		atomicGroup.Txs[i].RqNonce = atomicGroup.Txs[position].Nonce
		atomicGroup.Txs[i].RqFee = atomicGroup.Txs[position].Fee
		atomicGroup.Txs[i].RqAmount = atomicGroup.Txs[position].Amount
		atomicGroup.Txs[i].RqTokenID = atomicGroup.Txs[position].TokenID
	}

	// set AtomicGroupID
	atomicGroup = SetAtomicGroupID(atomicGroup)
	atomicGroupID = atomicGroup.ID

	// Post
	serverResponse, err = SendAtomicTxsGroup(hezClient, atomicGroup)
	if err != nil {
		err = fmt.Errorf("[AtomicTransfer] Error sending transactions. Error: %s\n", err.Error())
		return
	}

	return
}
