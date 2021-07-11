package node

import (
	"fmt"
	"log"
	"net/http"

	"github.com/hermeznetwork/hermez-go-sdk/client"
	"github.com/hermeznetwork/hermez-node/db/historydb"
)

func GetCurrentCoordinatorNodeInfo(hezClient client.HermezClient) (nodeState historydb.StateAPI, err error) {
	if len(hezClient.CurrentCoordinatorURL) < 10 {
		err = fmt.Errorf("[Node][GetCurrentCoordinatorNodeInfo] Current Coordinator is not set : %s", hezClient.CurrentCoordinatorURL)
		return
	}
	url := "/v1/state"
	// log.Printf("[Node][GetCurrentCoordinatorNodeInfo] URL %s", url)
	req, err := hezClient.CurrentCoordinatorClient.New().Get(url).Request()
	if err != nil {
		log.Printf("[Node][GetCurrentCoordinatorNodeInfo] Error boot coordinator info request: %s\n", err.Error())
		return
	}
	var failureBody interface{}
	res, err := hezClient.CurrentCoordinatorClient.Do(req, &nodeState, &failureBody)
	if err != nil {
		log.Printf("[Node][GetCurrentCoordinatorNodeInfo] Error pulling current coordinator info: %s - Error: %s\n", hezClient.CurrentCoordinatorURL, err.Error())
		return
	}
	if res.StatusCode != http.StatusOK {
		log.Printf("[Node][GetCurrentCoordinatorNodeInfo] Error pulling current coordinator info from hermez node: %+v - Error: %d\n", failureBody, res.StatusCode)
		return
	}
	return
}
