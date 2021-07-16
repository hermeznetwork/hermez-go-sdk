package account

import (
	"fmt"
	"log"
	"net/http"
	"regexp"
	"strings"

	"github.com/hermeznetwork/hermez-go-sdk/client"
)

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
	// account = formatHezAccountAddress(account)
	url := "/v1/accounts?BJJ=" + account
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
