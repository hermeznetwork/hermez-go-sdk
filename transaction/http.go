package transaction

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/ethereum/go-ethereum/common"
	"github.com/hermeznetwork/hermez-go-sdk/client"
	"github.com/hermeznetwork/hermez-go-sdk/util"

	hezCommon "github.com/hermeznetwork/hermez-node/common"
)

// ExecuteL2Transaction submits L2 transaction to the current coordinator endpoint
func ExecuteL2Transaction(hezClient client.HermezClient, apiTx APITx) (apiTxReturn APITx, serverResponse string, err error) {
	apiTxBody, err := util.MarshallBody(apiTx)
	if err != nil {
		err = fmt.Errorf("[ExecuteL2Transaction] Error marshaling HTTP request tx: %+v - Error: %s\n", apiTx, err.Error())
		return
	}

	var URL string
	URL = hezClient.CurrentCoordinatorURL + "/v1/transactions-pool"
	request, err := http.NewRequest("POST", URL, apiTxBody)
	if err != nil {
		err = fmt.Errorf("[ExecuteL2Transaction] Error creating HTTP request. URL: %s - request: %+v - Error: %s\n", URL, apiTxBody, err.Error())
		return
	}
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10.15; rv:88.0) Gecko/20100101 Firefox/88.0")
	request.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8")
	request.Header.Set("Accept-Encoding", "gzip, deflate, br")

	response, err := hezClient.HttpClient.Do(request)
	if err != nil {
		err = fmt.Errorf("[ExecuteL2Transaction] Error submitting HTTP request tx. URL: %s - request: %+v - Error: %s\n", URL, apiTxBody, err.Error())
		return
	}
	defer response.Body.Close()

	if response.StatusCode < 200 || response.StatusCode > 299 {
		tempBuf, errResp := io.ReadAll(response.Body)
		if errResp != nil {
			err = fmt.Errorf("[ExecuteL2Transaction] Error unmarshaling tx: %+v - Error: %s\n", apiTx, errResp.Error())
			return
		}
		strJSONRequest := string(tempBuf)
		err = fmt.Errorf("[ExecuteL2Transaction] Error posting TX: %+v\nStatusCode:%d \nStatus: %s\nReturned Message: %s\nURL: %s \nRequest: %+v \nResponse: %+v\n",
			apiTx, response.StatusCode, response.Status, strJSONRequest, URL, request, response)
		return
	}

	b, err := ioutil.ReadAll(response.Body)
	if err != nil {
		err = fmt.Errorf("[ExecuteL2Transaction] Error reading HTTP return from Coordinator. URL: %s - request: %+v - Error: %s\n", URL, apiTxBody, err.Error())
		return
	}

	serverResponse = fmt.Sprintln("Transaction ID submmited: ", apiTx.TxID.String(), "Message returned in Hex: %s\n", common.Bytes2Hex(b))
	apiTxReturn = apiTx

	return
}

// SendAtomicTxsGroup submits Atomic transaction to the current coordinator endpoint
func SendAtomicTxsGroup(hezClient client.HermezClient, atomicTxs hezCommon.AtomicGroup) (serverResponse string, err error) {
	apiTxBody, err := util.MarshallBody(atomicTxs)
	if err != nil {
		err = fmt.Errorf("[SendAtomicTxsGroup] Error marshaling HTTP request tx: %+v - Error: %s\n", atomicTxs, err.Error())
		return
	}
	log.Println("Bytes sent: ", apiTxBody)

	var URL string
	URL = hezClient.CurrentCoordinatorURL + "/v1/atomic-pool"
	request, err := http.NewRequest("POST", URL, apiTxBody)
	if err != nil {
		err = fmt.Errorf("[SendAtomicTxsGroup] Error creating HTTP request. URL: %s - request: %+v - Error: %s\n", URL, apiTxBody, err.Error())
		return
	}
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10.15; rv:88.0) Gecko/20100101 Firefox/88.0")
	request.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8")
	request.Header.Set("Accept-Encoding", "gzip, deflate, br")

	response, err := hezClient.HttpClient.Do(request)
	if err != nil {
		err = fmt.Errorf("[SendAtomicTxsGroup] Error submitting HTTP request tx. URL: %s - request: %+v - Error: %s\n", URL, apiTxBody, err.Error())
		return
	}
	defer response.Body.Close()

	if response.StatusCode < 200 || response.StatusCode > 299 {
		tempBuf, errResp := io.ReadAll(response.Body)
		if errResp != nil {
			err = fmt.Errorf("[SendAtomicTxsGroup] Error unmarshaling tx: %+v - Error: %s\n", atomicTxs, errResp.Error())
			return
		}
		strJSONRequest := string(tempBuf)
		err = fmt.Errorf("[SendAtomicTxsGroup] Error posting TX: %+v\nStatusCode:%d \nStatus: %s\nReturned Message: %s\nURL: %s \nRequest: %+v \nResponse: %+v\n",
			atomicTxs, response.StatusCode, response.Status, strJSONRequest, URL, request, response)
		return
	}

	b, err := ioutil.ReadAll(response.Body)
	if err != nil {
		err = fmt.Errorf("[SendAtomicTxsGroup] Error reading HTTP return from Coordinator. URL: %s - request: %+v - Error: %s\n", URL, apiTxBody, err.Error())
		return
	}

	serverResponse = fmt.Sprintln("Transaction ID submmited: ", atomicTxs.ID.String(), "Message returned in Hex: %s\n", common.Bytes2Hex(b))
	return
}

// GetTransactionsInPool connects to a hermez node and pull all transactions in the pool
func GetTransactionsInPool(hezClient client.HermezClient) (transactions TransactionsAPIResponse, err error) {
	if len(hezClient.BootCoordinatorURL) < 10 {
		err = fmt.Errorf("[Transaction][GetTransactionsInPool] Boot Coordinator is not set : %s", hezClient.BootCoordinatorURL)
		return
	}
	url := "/v1/transactions-pool?limit=1000"
	req, err := hezClient.BootCoordinatorClient.New().Get(url).Request()
	if err != nil {
		log.Printf("[Transaction][GetTransactionsInPool] Error pulling transactions info from request: %s\n", err.Error())
		return
	}
	var failureBody interface{}
	res, err := hezClient.BootCoordinatorClient.Do(req, &transactions, &failureBody)
	if err != nil {
		log.Printf("[Transaction][GetTransactionsInPool] Error pulling transactions info from hermez node: %s - Error: %s\n", hezClient.BootCoordinatorURL, err.Error())
		return
	}
	if res.StatusCode != http.StatusOK {
		log.Printf("[Transaction][GetTransactionsInPool] HTTP Error pulling transactions info from hermez node: %+v - Error: %d\n", failureBody, res.StatusCode)
		err = fmt.Errorf("[Transaction][GetTransactionsInPool] HTTP Error pulling transactions info from hermez node: %+v - Error: %d\n", failureBody, res.StatusCode))
		return
	}
	return
}

func GetTransactionsPool(hezClient client.HermezClient, txID hezCommon.TxID) (transaction hezCommon.PoolL2Tx, err error) {
	if len(hezClient.BootCoordinatorURL) < 10 {
		err = fmt.Errorf("[Transaction][GetTransactionsInPool] Boot Coordinator is not set : %s", hezClient.BootCoordinatorURL)
		return
	}
	url := "/v1/transactions-pool/" + txID.String()
	req, err := hezClient.BootCoordinatorClient.New().Get(url).Request()
	if err != nil {
		log.Printf("[Transaction][GetTransactionsInPool] Error pulling transactions info from request: %s\n", err.Error())
		return
	}
	var failureBody interface{}
	res, err := hezClient.BootCoordinatorClient.Do(req, &transaction, &failureBody)
	if err != nil {
		log.Printf("[Transaction][GetTransactionsInPool] Error pulling transactions info from hermez node: %s - Error: %s\n", hezClient.BootCoordinatorURL, err.Error())
		return
	}
	if res.StatusCode != http.StatusOK {
		log.Printf("[Transaction][GetTransactionsInPool] HTTP Error pulling transactions info from hermez node: %+v - Error: %d\n", failureBody, res.StatusCode)
		return
	}
	return
}
