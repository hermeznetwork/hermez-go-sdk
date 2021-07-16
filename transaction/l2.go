package transaction

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/big"
	"net/http"

	"github.com/hermeznetwork/hermez-go-sdk/account"
	"github.com/hermeznetwork/hermez-go-sdk/client"
	hezCommon "github.com/hermeznetwork/hermez-node/common"
	"github.com/hermeznetwork/hermez-node/log"
)

// L2Transfer perform token or ETH transfer within Hermez network (we say L2 or Layer2)
func L2Transfer(hezClient client.HermezClient,
	senderBjjWallet account.BJJWallet,
	recipientAddress string,
	tokenSymbolToTransfer string,
	amount *big.Int,
	feeRangeSelectedID int,
	ethereumChainID int) (apiTxReturn APITx, serverResponse string, err error) {

	// log.Println("[L2Transfer] Parameters")
	// log.Printf("hezClient: %+v\n", hezClient)
	// log.Printf("senderBjjWallet: %+v", senderBjjWallet)
	// log.Println("recipientAddress: ", recipientAddress)
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

	recipientAccDetails, err := account.GetAccountInfo(hezClient, recipientAddress)
	if err != nil {
		err = fmt.Errorf("[L2Transfer] Error obtaining account details. Account: %s - Error: %s\n", recipientAddress, err.Error())
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

func GetL2Tx(hezClient client.HermezClient, ID hezCommon.TxID) (hezCommon.PoolL2Tx, error) {
	URL := hezClient.BootCoordinatorURL + "/v1/transactions-history/" + ID.String()
	request, err := http.NewRequest("GET", URL, nil)
	if err != nil {
		err = fmt.Errorf("[GetL2Tx] Error creating HTTP request. URL: %s - Error: %s\n", URL, err.Error())
		return hezCommon.PoolL2Tx{}, err
	}
	response, err := hezClient.HttpClient.Do(request)
	if err != nil {
		return hezCommon.PoolL2Tx{}, fmt.Errorf("[GetL2Tx] Error submitting HTTP request tx. URL: %s - request: %+v - Error: %s\n", URL, ID, err.Error())
	}
	defer response.Body.Close()

	if response.StatusCode == 200 {
		tempBuf, errResp := io.ReadAll(response.Body)
		if errResp != nil {
			err = fmt.Errorf("[GetL2Tx] Error unmarshaling tx: %s - Error: %s\n", ID, errResp.Error())
			return hezCommon.PoolL2Tx{}, err
		}
		tx := hezCommon.PoolL2Tx{}
		err = json.Unmarshal(tempBuf, &tx)
		log.Info(string(tempBuf))
		if err != nil {
			err = fmt.Errorf("[GetL2Tx] Error unmarshaling tx: %s - Error: %s\n", ID, errResp.Error())
			return hezCommon.PoolL2Tx{}, err
		}
		return tx, nil
	} else if response.StatusCode == 404 {
		tempBuf, errResp := io.ReadAll(response.Body)
		if errResp != nil {
			err = fmt.Errorf("[GetL2Tx] Error unmarshaling tx: %s - Error: %s\n", ID, errResp.Error())
			return hezCommon.PoolL2Tx{}, err
		}
		log.Info(string(tempBuf))
		return hezCommon.PoolL2Tx{}, errors.New("tx not forged")
	} else {
		return hezCommon.PoolL2Tx{}, errors.New("unexpected error")
	}
}
