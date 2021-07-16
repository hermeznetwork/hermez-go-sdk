package main

/*
This integration test will perform atomic transactions.

Config requirements:
- Two EthPrivKeys that have HEZ accounts with funds on Goerli
- HermezNodeURL operating on Goerli, should win the auction too, oterwise the test will run forever
*/

import (
	"math/big"
	"sync"
	"time"

	integration "github.com/hermeznetwork/hermez-go-sdk/test"
	"github.com/hermeznetwork/hermez-go-sdk/transaction"
	"github.com/hermeznetwork/hermez-node/common"
	"github.com/hermeznetwork/hermez-node/log"
)

func main() {
	// Load integration testing
	it, err := integration.NewIntegrationTest()
	if err != nil {
		log.Error(err)
		panic("Failed initializing integration testing framework")
	}
	if len(it.Wallets) < 2 {
		panic("To run this test at least two wallets with HEZ deposited must be provided.")
	}
	// // Test happy path
	// if err := caseHappyPath(it); err != nil {
	// 	log.Error(err)
	// 	panic("Failed testing case: happy path")
	// }
	// Test reject and follow
	if err := caseRejectAndFollow(it); err != nil {
		log.Error(err)
		panic("Failed testing case: reject and follow")
	}
}

// caseHappyPath send a valid (nonce, balance, ...) atomic group of two linked txs
func caseHappyPath(it integration.IntegrationTest) error {
	log.Info("Start atomic test case: happy path")
	log.Info("Sending atomic group of two linked txs")
	// Define txs
	tx1 := transaction.AtomicTxItem{
		SenderBjjWallet:       it.Wallets[0],
		RecipientAddress:      it.Wallets[1].HezEthAddress,
		TokenSymbolToTransfer: "HEZ",
		Amount:                big.NewInt(100000000000000000),
		FeeRangeSelectedID:    126,
		RqOffSet:              1, //+1
	}

	tx2 := transaction.AtomicTxItem{
		SenderBjjWallet:       it.Wallets[1],
		RecipientAddress:      it.Wallets[0].HezEthAddress,
		TokenSymbolToTransfer: "HEZ",
		Amount:                big.NewInt(100000000000000000),
		FeeRangeSelectedID:    126,
		RqOffSet:              7, //-1
	}
	txs := make([]transaction.AtomicTxItem, 2)
	txs[0] = tx1
	txs[1] = tx2

	// create PoolL2Txs
	atomicGroup := common.AtomicGroup{}
	fullTxs, err := transaction.CreateFullTxs(it.Client, txs)
	if err != nil {
		return err
	}
	atomicGroup.Txs = fullTxs

	// set AtomicGroupID
	atomicGroup = transaction.SetAtomicGroupID(atomicGroup)

	// Sign the txs
	for i := range txs {
		var txHash *big.Int
		txHash, err = atomicGroup.Txs[i].HashToSign(uint16(5))
		if err != nil {
			return err
		}
		signature := txs[i].SenderBjjWallet.PrivateKey.SignPoseidon(txHash)
		atomicGroup.Txs[i].Signature = signature.Compress()
	}

	// Post
	var serverResponse string
	serverResponse, err = transaction.SendAtomicTxsGroup(it.Client, atomicGroup)
	if err != nil {
		return err
	}
	log.Info("Txs sent successfuly: ", serverResponse)
	log.Info("You can manually check the status of the txs at: " + it.Client.CurrentCoordinatorURL + "/v1/atomic-pool/" + atomicGroup.ID.String())
	// Wait until txs are forged
	log.Info("Entering a wait loop until txs are forged")
	const timeBetweenChecks = 30 * time.Second
	if err := integration.WaitUntilTxsAreForged(atomicGroup.Txs, timeBetweenChecks, it.Client); err != nil {
		return err
	}
	log.Info("Atomic group has been forged")
	return nil
}

// caseRejectAndFollow this case is intended to test that atomic groups are rejected safely
// without side effects on the StateDB. To achieve that, the following txs will be sent:
func caseRejectAndFollow(it integration.IntegrationTest) error {
	// 1. Send atomic group 1 with fee I and nonce X
	// 2. Send atomic group 2 with fee J and nonce Y
	// 3. Send atomic group 3 with fee K and nonce Z
	// Force the txselector to reject groups by having (the lower the nonce the lower the fee):
	// I < J < K
	// X < Y < Z
	// Each atomic group should be forged in different consecutive batches (due to current txselector implementation)
	// TODO: check the batch num in which each batch is forged: should be sequentialy one per batch
	log.Info("Start atomic test case: reject and follow")
	log.Info("Sending 3 atomic groups of two linked txs that should be forged in 3 consecutive batches")
	// Define txs
	tx1 := transaction.AtomicTxItem{
		SenderBjjWallet:       it.Wallets[0],
		RecipientAddress:      it.Wallets[1].HezEthAddress,
		TokenSymbolToTransfer: "HEZ",
		Amount:                big.NewInt(100000000000000000),
		FeeRangeSelectedID:    126,
		RqOffSet:              1, //+1
	}

	tx2 := transaction.AtomicTxItem{
		SenderBjjWallet:       it.Wallets[1],
		RecipientAddress:      it.Wallets[0].HezEthAddress,
		TokenSymbolToTransfer: "HEZ",
		Amount:                big.NewInt(100000000000000000),
		FeeRangeSelectedID:    126,
		RqOffSet:              7, //-1
	}
	txs := make([]transaction.AtomicTxItem, 2)
	txs[0] = tx1
	txs[1] = tx2

	// create PoolL2Txs
	atomicGroup1 := common.AtomicGroup{}
	atomicGroup2 := common.AtomicGroup{}
	atomicGroup3 := common.AtomicGroup{}
	fullTxs, err := transaction.CreateFullTxs(it.Client, txs)
	if err != nil {
		return err
	}
	atomicGroup1.Txs = make([]common.PoolL2Tx, 2)
	copy(atomicGroup1.Txs, fullTxs)
	atomicGroup2.Txs = make([]common.PoolL2Tx, 2)
	copy(atomicGroup2.Txs, fullTxs)
	atomicGroup3.Txs = make([]common.PoolL2Tx, 2)
	copy(atomicGroup3.Txs, fullTxs)
	// Manually modify fee and nonce for atomicGroup2 & atomicGroup3
	for i := 0; i < len(fullTxs); i++ {
		atomicGroup2.Txs[i].Fee += 1
		atomicGroup2.Txs[i].Nonce += 1
		atomicGroup2.Txs[i].RqFee += 1
		atomicGroup2.Txs[i].RqNonce += 1
		atomicGroup3.Txs[i].Fee += 2
		atomicGroup3.Txs[i].Nonce += 2
		atomicGroup3.Txs[i].RqFee += 2
		atomicGroup3.Txs[i].RqNonce += 2
	}

	// Sign the txs for atomic group 1
	for i := range txs {
		var txHash *big.Int
		txHash, err = atomicGroup1.Txs[i].HashToSign(uint16(5))
		if err != nil {
			return err
		}
		signature := txs[i].SenderBjjWallet.PrivateKey.SignPoseidon(txHash)
		atomicGroup1.Txs[i].Signature = signature.Compress()
	}
	// Sign the txs for atomic group 2 and reset txID
	for i := range txs {
		atomicGroup2.Txs[i].TxID = common.EmptyTxID
		if _, err := common.NewPoolL2Tx(&atomicGroup2.Txs[i]); err != nil {
			return err
		}
		var txHash *big.Int
		txHash, err = atomicGroup2.Txs[i].HashToSign(uint16(5))
		if err != nil {
			return err
		}
		signature := txs[i].SenderBjjWallet.PrivateKey.SignPoseidon(txHash)
		atomicGroup2.Txs[i].Signature = signature.Compress()
	}
	// Sign the txs for atomic group 3 and reset txID
	for i := range txs {
		atomicGroup3.Txs[i].TxID = common.EmptyTxID
		if _, err := common.NewPoolL2Tx(&atomicGroup3.Txs[i]); err != nil {
			return err
		}
		var txHash *big.Int
		txHash, err = atomicGroup3.Txs[i].HashToSign(uint16(5))
		if err != nil {
			return err
		}
		signature := txs[i].SenderBjjWallet.PrivateKey.SignPoseidon(txHash)
		atomicGroup3.Txs[i].Signature = signature.Compress()
	}
	// set AtomicGroupID
	atomicGroup1 = transaction.SetAtomicGroupID(atomicGroup1)
	atomicGroup2 = transaction.SetAtomicGroupID(atomicGroup2)
	atomicGroup3 = transaction.SetAtomicGroupID(atomicGroup3)

	// Post
	wg := new(sync.WaitGroup)
	wg.Add(3)
	const timeBetweenChecks = 30 * time.Second
	waitForAtomicGroupToBeForged := func(ag common.AtomicGroup, agNum int) {
		if err := integration.WaitUntilTxsAreForged(ag.Txs, timeBetweenChecks, it.Client); err != nil {
			panic(err)
		}
		log.Info("Atomic group ", agNum, "has been forged, atomic group id: ", ag.ID.String())
		wg.Done()
	}
	for i, ag := range []common.AtomicGroup{atomicGroup1, atomicGroup2, atomicGroup3} {
		var serverResponse string
		serverResponse, err = transaction.SendAtomicTxsGroup(it.Client, ag)
		if err != nil {
			return err
		}
		log.Info("Atomic group ", i+1, " sent successfuly: ", serverResponse)
		log.Info("You can manually check the status of the txs at: " + it.Client.CurrentCoordinatorURL + "/v1/atomic-pool/" + ag.ID.String())
		go waitForAtomicGroupToBeForged(ag, i+1)
	}
	// Wait until txs are forged
	log.Info("Entering a wait loop until txs are forged")
	wg.Wait()
	log.Info("All transactions forged successfuly")

	return nil
}
