package transaction

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"

	"github.com/ethereum/go-ethereum/common"
	"github.com/hermeznetwork/hermez-go-sdk/client"
	"github.com/hermeznetwork/hermez-go-sdk/util"
)

// ExecuteL2Transaction submits L2 transaction to the actual coordinator endpoint
func ExecuteL2Transaction(hezClient client.HermezClient, apiTx APITx) (apiTxReturn APITx, serverResponse string, err error) {
	apiTxBody, err := util.MarshallBody(apiTx)
	if err != nil {
		err = fmt.Errorf("[ExecuteL2Transaction] Error marshaling HTTP request tx: %+v - Error: %s\n", apiTx, err.Error())
		return
	}

	var URL string
	URL = hezClient.ActualCoordinatorURL + "/v1/transactions-pool"
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
