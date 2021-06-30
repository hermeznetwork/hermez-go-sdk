package transaction

import (
	"fmt"
	"math/big"

	"github.com/hermeznetwork/hermez-go-sdk/account"
	"github.com/hermeznetwork/hermez-go-sdk/client"
)

// L2Transfer perform token or ETH transfer within Hermez network (we say L2 or Layer2)
func L2Transfer(hezClient client.HermezClient,
	senderBjjWallet account.BJJWallet,
	receipientAddress string,
	tokenSymbolToTransfer string,
	amount *big.Int,
	feeRangeSelectedID int,
	ethereumChainID int) (apiTxReturn APITx, serverResponse string, err error) {

	// log.Println("[L2Transfer] Parameters")
	// log.Printf("hezClient: %+v\n", hezClient)
	// log.Printf("senderBjjWallet: %+v", senderBjjWallet)
	// log.Println("receipientAddress: ", receipientAddress)
	// log.Println("tokenSymbolToTransfer: ", tokenSymbolToTransfer)
	// log.Println("amount: ", amount.String())
	// log.Println("feeRangeSelectedID: ", feeRangeSelectedID)
	// log.Println("ethereumChainID: ", ethereumChainID)

	err = nil

	senderAccDetails, err := account.GetAccountInfo(hezClient, senderBjjWallet.EthAccount.Address.Hex())
	if err != nil {
		err = fmt.Errorf("[L2Transfer] Error obtaining account details. Account: %s - Error: %s\n", senderBjjWallet.HezEthAddress, err.Error())
		return
	}

	// log.Printf("\n\nSender Account details from Coordinator: %+v\n\n", senderAccDetails)
	// log.Println("BJJ Address in server: ", senderAccDetails.Accounts[0].BJJAddress)
	// log.Println("BJJ Address local: ", bjjWallet.HezBjjAddress)
	// log.Printf("Wallet details %+v\n", bjjWallet)

	recipientAccDetails, err := account.GetAccountInfo(hezClient, receipientAddress)
	if err != nil {
		err = fmt.Errorf("[L2Transfer] Error obtaining account details. Account: %s - Error: %s\n", senderBjjWallet.HezEthAddress, err.Error())
		return
	}

	// log.Printf("\n\nReceipient Account details from Coordinator: %+v\n\n", recipientAccDetails)
	// log.Println("BJJ Address in server: ", recipientAccDetails.Accounts[0].BJJAddress)

	apiTxReturn, err = MarshalTransaction(tokenSymbolToTransfer, senderAccDetails, recipientAccDetails, senderBjjWallet, amount, feeRangeSelectedID, ethereumChainID)
	if err != nil {
		err = fmt.Errorf("[L2Transfer] Error marsheling tx data to prepare to send to coordinator. Error: %s\n", err.Error())
		return
	}

	// log.Printf("\nTX to be submited: %+v\n", apiTxReturn)

	apiTxReturn, serverResponse, err = ExecuteL2Transaction(hezClient, apiTxReturn)
	if err != nil {
		err = fmt.Errorf("[L2Transfer] Error submiting tx transaction pool endpoint. Error: %s\n", err.Error())
		return
	}

	return
}
