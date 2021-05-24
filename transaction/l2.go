package transaction

import (
	"fmt"
	"math/big"

	"github.com/jeffprestes/hermez-go-sdk/account"
	"github.com/jeffprestes/hermez-go-sdk/client"
)

// L2Transfer perform token or ETH transfer within Hermez network (we say L2 or Layer2)
func L2Transfer(hezClient client.HermezClient,
	bjjWallet account.BJJWallet,
	receipientAddress string,
	tokenSymbolToTransfer string,
	amount *big.Int,
	feeRangeSelectedID int,
	ethereumChainID int) (apiTxReturn APITx, serverResponse string, err error) {

	senderAccDetails, err := account.GetAccountInfo(hezClient, bjjWallet.HezBjjAddress)
	if err != nil {
		err = fmt.Errorf("[L2Transfer] Error obtaining account details. Account: %s - Error: %s\n", bjjWallet.HezEthAddress, err.Error())
		return
	}

	// log.Printf("\n\nAccount details from Coordinator: %+v\n\n", senderAccDetails)
	// log.Println("BJJ Address in server: ", senderAccDetails.Accounts[0].BJJAddress)
	// log.Println("BJJ Address local: ", bjjWallet.HezBjjAddress)
	// log.Printf("Wallet details %+v\n", bjjWallet)

	recipientAccDetails, err := account.GetAccountInfo(hezClient, receipientAddress)
	if err != nil {
		err = fmt.Errorf("[L2Transfer] Error obtaining account details. Account: %s - Error: %s\n", bjjWallet.HezEthAddress, err.Error())
		return
	}

	apiTx, err := MarshalTransaction(tokenSymbolToTransfer, senderAccDetails, recipientAccDetails, bjjWallet, amount, feeRangeSelectedID, ethereumChainID)
	if err != nil {
		err = fmt.Errorf("[L2Transfer] Error marsheling tx data to prepare to send to coordinator. Error: %s\n", err.Error())
		return
	}

	apiTx, serverResponse, err = ExecuteL2Transaction(hezClient, apiTx)
	if err != nil {
		err = fmt.Errorf("[L2Transfer] Error submiting tx transaction pool endpoint. Error: %s\n", err.Error())
		return
	}

	return
}
