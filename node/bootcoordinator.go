package node

import (
	"fmt"
	"log"
	"net/http"

	"github.com/hermeznetwork/hermez-node/db/historydb"
	"github.com/jeffprestes/hermez-go-sdk/client"
)

func GetBootCoordinatorNodeInfo(hezClient client.HermezClient) (nodeState historydb.StateAPI, err error) {
	if len(hezClient.BootCoordinatorURL) < 10 {
		err = fmt.Errorf("[Node][GetBootCoordinatorNodeInfo] Boot Coordinator is not set : %s", hezClient.BootCoordinatorURL)
		return
	}
	url := "/v1/state"
	// log.Printf("[Node][GetBootCoordinatorNodeInfo] URL %s", url)
	req, err := hezClient.BootCoordinatorClient.New().Get(url).Request()
	if err != nil {
		log.Printf("[Node][GetBootCoordinatorNodeInfo] Error boot coordinator info request: %s\n", err.Error())
		return
	}
	var failureBody interface{}
	res, err := hezClient.BootCoordinatorClient.Do(req, &nodeState, &failureBody)
	if err != nil {
		log.Printf("[Node][GetBootCoordinatorNodeInfo] Error pulling boot coordinator info: %s - Error: %s\n", hezClient.BootCoordinatorURL, err.Error())
		return
	}
	if res.StatusCode != http.StatusOK {
		log.Printf("[Node][GetBootCoordinatorNodeInfo] Error pulling boot coordinator info from hermez node: %+v - Error: %d\n", failureBody, res.StatusCode)
		return
	}
	return
}
