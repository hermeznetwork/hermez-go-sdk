package account

import (
	"fmt"
	"log"
	"net/http"
	"regexp"
	"strings"

	"github.com/hermeznetwork/hermez-go-sdk/client"
)

// PostNewBJJAccount post new BJJ Account to Hermez Network to allow L2 - Layer 2 transactions
// func PostNewBJJAccount(newBjj BJJWallet, hezClient client.HermezClient) (err error) {
// 	log.Println("[Account][PostNewBJJAccount] Posting new account ", newBjj.HezBjjAddress, " to ", hezClient.ActualCoordinatorURL, " coordinator... ")

// 	URL := hezClient.ActualCoordinatorURL + "/v1/account-creation-authorization"
// 	request, err := http.NewRequest("POST", URL, apiTxBody)
// 	if err != nil {
// 		err = fmt.Errorf("[Account][PostNewBJJAccount] Error creating HTTP request. URL: %s - request: %+v - Error: %s\n", URL, apiTxBody, err.Error())
// 		return
// 	}
// 	request.Header.Set("Content-Type", "application/json")
// 	request.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10.15; rv:88.0) Gecko/20100101 Firefox/88.0")
// 	request.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8")
// 	request.Header.Set("Accept-Encoding", "gzip, deflate, br")

// 	response, err := hezClient.HttpClient.Do(request)
// 	if err != nil {
// 		err = fmt.Errorf("[Account][PostNewBJJAccount] Error submitting HTTP request new account creation. URL: %s - request: %+v - Error: %s\n", URL, apiTxBody, err.Error())
// 		return
// 	}
// 	defer response.Body.Close()

// 	if response.StatusCode < 200 || response.StatusCode > 299 {
// 		tempBuf, errResp := io.ReadAll(response.Body)
// 		if errResp != nil {
// 			err = fmt.Errorf("[ExecuteL2Transaction] Error unmarshaling tx: %+v - Error: %s\n", apiTx, errResp.Error())
// 			return
// 		}
// 		strJSONRequest := string(tempBuf)
// 		err = fmt.Errorf("[ExecuteL2Transaction] Error posting TX: %+v\nStatusCode:%d \nStatus: %s\nReturned Message: %s\nURL: %s \nRequest: %+v \nResponse: %+v\n",
// 			apiTx, response.StatusCode, response.Status, strJSONRequest, URL, request, response)
// 		return
// 	}

// 	b, err := ioutil.ReadAll(response.Body)
// 	if err != nil {
// 		err = fmt.Errorf("[ExecuteL2Transaction] Error reading HTTP return from Coordinator. URL: %s - request: %+v - Error: %s\n", URL, apiTxBody, err.Error())
// 		return
// 	}
// 	return
// }

// GetAccountInfo connects to a hermez node and pull account data
func GetAccountInfo(hezClient client.HermezClient, account string) (hezAccount AccountAPIResponse, err error) {
	log.Println("[Account][GetAccountInfo] Pulling account info ", account, " from a coordinator...")
	if len(account) < 5 {
		err = fmt.Errorf("[Account][GetAccountInfo] Invalid account to query: %s", account)
		return
	}
	if len(hezClient.BootCoordinatorURL) < 10 {
		err = fmt.Errorf("[Account][GetAccountInfo] Boot Coordinator is not set : %s", hezClient.BootCoordinatorURL)
		return
	}
	account = formatHezAccountAddress(account)
	url := "/v1/accounts?" + account
	// log.Printf("[Account][GetAccountInfo] URL %s", url)
	req, err := hezClient.BootCoordinatorClient.New().Get(url).Request()
	if err != nil {
		log.Printf("[Account][GetAccountInfo] Error creating pulling account info request: %s\n", err.Error())
		return
	}
	// log.Printf("[Account][GetAccountInfo] req %+v\n", req)
	var failureBody interface{}
	res, err := hezClient.BootCoordinatorClient.Do(req, &hezAccount, &failureBody)
	if err != nil {
		log.Printf("[Account][GetAccountInfo] Error pulling account info from hermez node: %s - Error: %s\n", hezClient.BootCoordinatorURL, err.Error())
		return
	}
	if res.StatusCode != http.StatusOK {
		log.Printf("[Account][GetAccountInfo] Error pulling account info from hermez node: %+v - Error: %d\n", failureBody, res.StatusCode)
		return
	}
	// log.Printf("[Account][GetAccountInfo] res \n\n%+v\n\n", res)
	return
}

func formatHezAccountAddress(account string) (hezAccountString string) {
	if strings.HasPrefix(account, "0x") {
		hezAccountString = "hezEthereumAddress=hez:" + account
		return
	}
	matchIdx, _ := regexp.MatchString("[A-Za-z]{3}:\\d+", account)
	if matchIdx {
		hezAccountString = "hez:" + account
		return
	}
	if len(account) > 20 {
		hezAccountString = "BJJ=hez:" + account
		return
	}
	return
}
