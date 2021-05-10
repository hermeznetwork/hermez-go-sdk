package node

import (
	"fmt"
	"log"
	"net/http"

	"github.com/jeffprestes/hermez-go-sdk/client"
)

func GetActualCoordinatorNodeInfo(hezClient client.HermezClient) (nodeState NodeState, err error) {
	if len(hezClient.ActualCoordinatorURL) < 10 {
		err = fmt.Errorf("[Node][GetActualCoordinatorNodeInfo] Actual Coordinator is not set : %s", hezClient.ActualCoordinatorURL)
		return
	}
	url := "/v1/state"
	// log.Printf("[Node][GetActualCoordinatorNodeInfo] URL %s", url)
	req, err := hezClient.ActualCoordinatorClient.New().Get(url).Request()
	if err != nil {
		log.Printf("[Node][GetActualCoordinatorNodeInfo] Error boot coordinator info request: %s\n", err.Error())
		return
	}
	var failureBody interface{}
	res, err := hezClient.ActualCoordinatorClient.Do(req, &nodeState, &failureBody)
	if err != nil {
		log.Printf("[Node][GetActualCoordinatorNodeInfo] Error pulling actual coordinator info: %s - Error: %s\n", hezClient.ActualCoordinatorURL, err.Error())
		return
	}
	if res.StatusCode != http.StatusOK {
		log.Printf("[Node][GetActualCoordinatorNodeInfo] Error pulling actual coordinator info from hermez node: %+v - Error: %d\n", failureBody, res.StatusCode)
		return
	}
	return
}
