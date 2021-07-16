package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"strings"
	"time"

	"github.com/hermeznetwork/hermez-go-sdk/node"
	integration "github.com/hermeznetwork/hermez-go-sdk/test"
	"github.com/hermeznetwork/hermez-go-sdk/transaction"
	"github.com/hermeznetwork/hermez-node/common"
	"github.com/hermeznetwork/hermez-node/log"
)

/*
This integration test will perform transactions that include MaxNumBatch != 0.

Config requirements:
- One EthPrivKeys that have HEZ accounts with funds on Goerli
- HermezNodeURL operating on Goerli, should win the auction too, oterwise the test will run forever
- The coordinator should forge the tx within the next 30 batches from the moment the test is executed
*/

func main() {
	// Load integration testing
	it, err := integration.NewIntegrationTest()
	if err != nil {
		log.Error(err)
		panic("Failed initializing integration testing framework")
	}
	if len(it.Wallets) == 0 {
		panic("To run this test at least one wallet with HEZ deposited must be provided.")
	}
	// Test fail and success
	if err := caseFailAndSuccess(it); err != nil {
		log.Error(err)
		panic("Failed testing case: fail and success")
	}
}

// caseFailAndSuccess sends valid and invalid txs with MaxNumBatch != 0
func caseFailAndSuccess(it integration.IntegrationTest) error {
	/*
	 TODO:
	 -
	*/
	// Get current batch num
	var currentBatchNum common.BatchNum
	if state, err := node.GetBootCoordinatorNodeInfo(it.Client); err != nil {
		return err
	} else {
		currentBatchNum = state.Network.LastBatch.BatchNum
	}
	// Get account details
	idx, nonce, tokenId, err := transaction.GetAccountDetails(it.Client, it.Wallets[0].HezBjjAddress, "HEZ")
	if err != nil {
		return err
	}
	// Generate invalid tx because of MaxNumBatch
	invalidTx := common.PoolL2Tx{
		FromIdx:     idx,
		ToIdx:       idx,
		Nonce:       nonce,
		Amount:      big.NewInt(100000000000000000),
		TokenID:     tokenId,
		TokenSymbol: "HEZ",
		Fee:         170,
		MaxNumBatch: uint32(currentBatchNum - 1),
	}
	if _, err := common.NewPoolL2Tx(&invalidTx); err != nil {
		return err
	}
	if hash, err := invalidTx.HashToSign(5); err != nil {
		return err
	} else {
		invalidTx.Signature = it.Wallets[0].PrivateKey.SignPoseidon(hash).Compress()
	}
	// Generate valid tx with MaxNumBatch != 0
	validTx := common.PoolL2Tx{
		FromIdx:     idx,
		ToIdx:       idx,
		Nonce:       nonce,
		Amount:      big.NewInt(100000000000000000),
		TokenID:     tokenId,
		TokenSymbol: "HEZ",
		Fee:         126,
		MaxNumBatch: uint32(currentBatchNum + 30),
	}
	if _, err := common.NewPoolL2Tx(&validTx); err != nil {
		return err
	}
	if hash, err := validTx.HashToSign(5); err != nil {
		return err
	} else {
		validTx.Signature = it.Wallets[0].PrivateKey.SignPoseidon(hash).Compress()
	}
	// POST txs
	for i, tx := range []common.PoolL2Tx{invalidTx, validTx} {
		apiTxBody := new(bytes.Buffer)
		if err := json.NewEncoder(apiTxBody).Encode(tx); err != nil {
			return err
		}

		request, err := http.NewRequest(
			"POST",
			it.Client.CurrentCoordinatorURL+"/v1/transactions-pool",
			apiTxBody,
		)
		if err != nil {
			return err
		}
		response, err := it.Client.HttpClient.Do(request)
		if err != nil {
			return err
		}

		if response.StatusCode != 200 {
			tempBuf, err := io.ReadAll(response.Body)
			if err != nil {
				response.Body.Close()
				return err
			}
			response.Body.Close()
			return errors.New(string(tempBuf))
		} else {
			checkMessage := fmt.Sprintf("You can manually check the status of the tx at %s/v1/transactions-pool/%s", it.Client.CurrentCoordinatorURL, tx.TxID.String())
			if i == 0 {
				log.Info("Invalid tx sent succesfully. ", checkMessage)
			} else {
				log.Info("Valid tx sent succesfully. ", checkMessage)
			}
		}
	}

	// Wait until valid tx is forged
	log.Info("Entering a wait loop until txs are forged")
	const timeBetweenChecks = 30 * time.Second
	if err := integration.WaitUntilTxsAreForged([]common.PoolL2Tx{validTx}, timeBetweenChecks, it.Client); err != nil {
		return err
	}
	log.Info("Valid tx is forged")

	// Check Info of invalid tx
	if tx, err := transaction.GetPoolL2Tx(it.Client, invalidTx.TxID); err != nil {
		return err
	} else if !strings.Contains(tx.Info, "MaxNumBatch exceeded") {
		return fmt.Errorf("Unexpected rejection message, expected: MaxNumBatch exceeded. Actual %s", tx.Info)
	}
	return nil
}
