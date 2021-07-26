package token

import (
	"fmt"
	"log"
	"net/http"

	"github.com/hermeznetwork/hermez-go-sdk/client"
)

// GetTokens connects to a hermez node and pull all tokens supported in that specific Hermez network instance
func GetTokens(hezClient client.HermezClient) (tokens TokensAPIResponse, err error) {
	if len(hezClient.BootCoordinatorURL) < 10 {
		err = fmt.Errorf("[Token][GetTokens] Boot Coordinator is not set : %s", hezClient.BootCoordinatorURL)
		return
	}
	url := "/v1/tokens?limit=100"
	req, err := hezClient.BootCoordinatorClient.New().Get(url).Request()
	if err != nil {
		log.Printf("[Token][GetTokens] Error pulling tokens info from request: %s\n", err.Error())
		return
	}
	var failureBody interface{}
	res, err := hezClient.BootCoordinatorClient.Do(req, &tokens, &failureBody)
	if err != nil {
		log.Printf("[Token][GetTokens] Error pulling tokens info from hermez node: %s - Error: %s\n", hezClient.BootCoordinatorURL, err.Error())
		return
	}
	if res.StatusCode != http.StatusOK {
		log.Printf("[Token][GetTokens] HTTP Error pulling tokens info from hermez node: %+v - Error: %d\n", failureBody, res.StatusCode)
		return
	}
	return
}
