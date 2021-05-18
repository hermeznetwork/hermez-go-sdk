package main

import (
	"crypto/ecdsa"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"io"
	"io/ioutil"
	"log"
	"math/big"
	"net/http"
	"strconv"
	"strings"

	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	hezcommon "github.com/hermeznetwork/hermez-node/common"
	"github.com/iden3/go-iden3-crypto/babyjub"
	"github.com/jeffprestes/hermez-go-sdk/account"
	"github.com/jeffprestes/hermez-go-sdk/client"
	"github.com/jeffprestes/hermez-go-sdk/node"
	"github.com/jeffprestes/hermez-go-sdk/util"
)

// SignatureConstantBytes contains the SignatureConstant in byte array
// format, which is equivalent to 3322668559 as uint32 in byte array in
// big endian representation.
var SignatureConstantBytes = []byte{198, 11, 230, 15}

// BJJWallet BJJ Wallet
type BJJWallet struct {
	PrivateKey    babyjub.PrivateKey
	PublicKey     babyjub.PublicKeyComp
	HezBjjAddress string
	HezEthAddress string
}

func main() {
	log.Println("Starting Hermez Client...")
	hezClient, err := client.NewHermezClient()
	if err != nil {
		log.Printf("Error during Hermez client initialization: %s\n", err.Error())
		return
	}
	log.Println("Connected to Hermez Smart Contracts...")
	log.Println("Pulling account info from a coordinator...")

	log.Println("Pulling boot coordinator info...")
	bootCoordNodeState, err := node.GetBootCoordinatorNodeInfo(hezClient)
	if err != nil {
		log.Printf("Error obtaining boot coordinator info. URL: %s - Error: %s\n", hezClient.BootCoordinatorURL, err.Error())
		return
	}
	log.Println("Setting actual client ...")
	hezClient.SetActualCoordinator(bootCoordNodeState.Network.NextForgers[0].Coordinator.URL)
	log.Println("Actual client is set.")
	log.Printf("Pulling actual coordinator (%s) info...\n", hezClient.ActualCoordinatorURL)
	actualCoordNodeState, err := node.GetActualCoordinatorNodeInfo(hezClient)
	if err != nil {
		log.Printf("Error obtaining boot coordinator info. URL: %s - Error: %s\n", hezClient.BootCoordinatorURL, err.Error())
		return
	}
	log.Printf("Actual coordinator info is: %+v\n", actualCoordNodeState)
	if len(actualCoordNodeState.Network.NextForgers) > 0 {
		log.Printf("Actual coordinator URL is: %s\n", actualCoordNodeState.Network.NextForgers[0].Coordinator.URL)
	}
	log.Printf("Boot coordinator URL is: %+v\n", actualCoordNodeState.Auction.BootCoordinatorURL)

	log.Println("Generating the private key...")
	// ecdsaPvtKey, err := crypto.HexToECDSA("53616f5061756c6f42617263656c6f6e614d616e696c6143616e617269617339")
	ecdsaPvtKey, err := crypto.HexToECDSA("e7064c29eb71fa44e6a14e78f5fcba3c1625b6382d107ef21096555074a98cd9")
	if err != nil {
		log.Printf("Error when creating private key: %s\n", err.Error())
		return
	}
	ecdsaPubKey := ecdsaPvtKey.Public().(*ecdsa.PublicKey)
	ethAddress := crypto.PubkeyToAddress(*ecdsaPubKey)

	// BJJ account
	const hermezWalletMsg = "Hermez Network account access.\n\nSign this message if you are in a trusted application only."

	hermezWalletMsgHash := accounts.TextHash([]byte(hermezWalletMsg))
	hermezWalletMsgSigned, err := crypto.Sign(hermezWalletMsgHash, ecdsaPvtKey)
	if err != nil {
		log.Printf("Error signing key msg to generate BJJ private key. Account: %s - Error: %s\n", ethAddress.Hex(), err.Error())
		return
	}

	// Change the last item value of the byte array from 1 to 28
	hermezWalletMsgSigned[len(hermezWalletMsgSigned)-1] += 27

	hermezWalletMsgEncoded := "0x" + hex.EncodeToString(hermezWalletMsgSigned)
	// log.Println("hermezWalletMsgEncoded: ", hermezWalletMsgEncoded)
	hermezWalletMsgSignedHash := crypto.Keccak256([]byte(hermezWalletMsgEncoded))

	var bjjPvtKey babyjub.PrivateKey
	// Copy to the private key
	copy(bjjPvtKey[:], hermezWalletMsgSignedHash[:])

	bjjAddress, err := FromBJJPubKeyCompToHezBJJAddress(bjjPvtKey.Public().Compress())
	if err != nil {
		log.Printf("Error generating BJJ address from BJJ public key. Account: %+v - Error: %s\n", bjjPvtKey.Public().Compress(), err.Error())
		return
	}

	accountDetails, err := account.GetAccountInfo(hezClient, ethAddress.Hex())
	if err != nil {
		log.Printf("Error obtaining account details. Account: %s - Error: %s\n", ethAddress.Hex(), err.Error())
		return
	}
	log.Printf("\n\nAccount info is: %+v\n\n", accountDetails)

	log.Println("BJJ Address in server: ", accountDetails.Accounts[0].BJJAddress)
	log.Println("BJJ Address local: ", bjjAddress)

	decodedBjjPubKey, err := hex.DecodeString(bjjPvtKey.Public().Compress().String())
	if err != nil {
		log.Printf("Error decoding BJJ public key. Account: %+v - Error: %s\n", bjjPvtKey, err.Error())
		return
	}

	var bjjPubKeyCompressed babyjub.PublicKeyComp
	temp := hezcommon.SwapEndianness(decodedBjjPubKey)
	copy(bjjPubKeyCompressed[:], temp[:])

	var bjjWallet BJJWallet
	bjjWallet.PrivateKey = bjjPvtKey
	bjjWallet.PublicKey = bjjPubKeyCompressed
	bjjWallet.HezBjjAddress = bjjAddress
	bjjWallet.HezEthAddress = "hez:" + ethAddress.String()

	log.Printf("Wallet details %+v\n", bjjWallet)

	// Which token do you want to transfer?
	itemToTransfer := "HEZ"

	var token hezcommon.Token
	var nonce hezcommon.Nonce
	var fromIdx hezcommon.Idx

	// Get from account Token and nonce details
	for _, account := range accountDetails.Accounts {
		if strings.ToUpper(account.Token.Symbol) == itemToTransfer {
			token.TokenID = hezcommon.TokenID(account.Token.ID)
			token.Symbol = account.Token.Symbol
			nonce = hezcommon.Nonce(account.Nonce)
			tempAccountsIdx := strings.Split(account.AccountIndex, ":")
			if len(tempAccountsIdx) == 3 {
				tempAccIdx, err := strconv.Atoi(tempAccountsIdx[2])
				if err != nil {
					log.Printf("Error getting account index. Account: %+v - Error: %s\n", account, err.Error())
					return
				}
				fromIdx = hezcommon.Idx(tempAccIdx)
			}
		}
	}

	// If there is no account created to this specific token stop the code
	if len(fromIdx.String()) < 1 {
		log.Println("There is no account to this user ", bjjAddress, " for this Token ", itemToTransfer)
		return
	}

	toIdx := hezcommon.Idx(uint64(21541))
	// toIdx :=
	amount := big.NewInt(889000000000000000)
	// fee := hezcommon.FeeSelector(100)
	fee := hezcommon.FeeSelector(126) // 10.2%

	tx := new(hezcommon.PoolL2Tx)
	tx.FromIdx = fromIdx
	tx.ToEthAddr = common.HexToAddress("0x263c3ab7e4832edf623fbdd66acee71c028ff591")
	tx.ToBJJ = hezcommon.EmptyBJJComp
	tx.ToIdx = toIdx
	tx.Amount = amount
	tx.Fee = fee
	tx.TokenID = token.TokenID
	tx.Nonce = nonce
	tx.Type = hezcommon.TxTypeTransfer

	tx, err = hezcommon.NewPoolL2Tx(tx)
	if err != nil {
		log.Printf("Error posting tx to hermez node. TX: %+v - Error: %s\n", tx, err.Error())
		return
	}

	txHash, err := tx.HashToSign(uint16(5))
	if err != nil {
		log.Printf("Error generating tx hash. TX: %+v - Error: %s\n", tx, err.Error())
		return
	}

	signedTx := bjjPvtKey.SignPoseidon(txHash)
	tx.Signature = signedTx.Compress()

	apiTx := NewHermezAPITxRequest(tx, token)

	apiTxBody, err := util.MarshallBody(apiTx)
	if err != nil {
		log.Printf("Error marshaling HTTP request tx: %+v - Error: %s\n", tx, err.Error())
		return
	}

	var URL string
	URL = hezClient.ActualCoordinatorURL + "/v1/transactions-pool"
	// URL = "https://vps30599.publiccloud.com.br/v1/transactions-pool"
	request, err := http.NewRequest("POST", URL, apiTxBody)
	if err != nil {
		log.Printf("Error creating HTTP request. URL: %s - request: %+v - Error: %s\n", URL, apiTxBody, err.Error())
		return
	}
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10.15; rv:88.0) Gecko/20100101 Firefox/88.0")
	request.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8")
	request.Header.Set("Accept-Encoding", "gzip, deflate, br")

	log.Printf("Submiting this TX: %s\n Full Tx: %+v\nRequest details: %+v\n\n", apiTx.TxID, apiTx, request)

	response, err := hezClient.HttpClient.Do(request)
	if err != nil {
		log.Printf("Error submitting HTTP request tx. URL: %s - request: %+v - Error: %s\n", URL, apiTxBody, err.Error())
		return
	}
	defer response.Body.Close()

	if response.StatusCode < 200 || response.StatusCode > 299 {
		tempBuf, err := io.ReadAll(response.Body)
		if err != nil {
			log.Printf("Error unmarshaling tx: %+v - Error: %s\n", tx, err.Error())
			return
		}
		strJSONRequest := string(tempBuf)
		log.Printf("Error posting TX. \nStatusCode:%d \nStatus: %s\nReturned Message: %s\nURL: %s \nRequest: %+v \nResponse: %+v\n", response.StatusCode, response.Status, strJSONRequest, URL, request, response)
		return
	}

	b, err := ioutil.ReadAll(response.Body)
	if err != nil {
		log.Printf("Error reading HTTP return from Coordinator. URL: %s - request: %+v - Error: %s\n", URL, apiTxBody, err.Error())
		return
	}
	if b == nil || len(b) == 0 {
		log.Printf("Error no HTTP return from Coordinator. URL: %s - request: %+v \n", URL, apiTxBody)
		return
	}

	log.Printf("Transaction submitted: %s\n", string(b))
}

// FromBJJPubKeyCompToHezBJJAddress creates a Hermez BJJ account from a BJJ PubKey Compressed
// Calling this method with a nil bjj causes panic
func FromBJJPubKeyCompToHezBJJAddress(pkComp babyjub.PublicKeyComp) (string, error) {
	if len(pkComp.String()) < 10 {
		return "", errors.New("Invalid BJJ PubKey Compressed")
	}
	sum := pkComp[0]
	for i := 1; i < len(pkComp); i++ {
		sum += pkComp[i]
	}
	bjjSum := append(pkComp[:], sum)
	return "hez:" + base64.RawURLEncoding.EncodeToString(bjjSum), nil
}

// NewHermezAPITxRequest convert L2 tx to Hermez API request model
func NewHermezAPITxRequest(poolTx *hezcommon.PoolL2Tx, token hezcommon.Token) APITx {
	toIdx := "hez:ETH:0"
	if poolTx.ToIdx > 0 {
		toIdx = idxToHez(poolTx.ToIdx, token.Symbol)
	}
	toEth := ""
	if poolTx.ToEthAddr != hezcommon.EmptyAddr {
		toEth = ethAddrToHez(poolTx.ToEthAddr)
	}
	toBJJ := bjjToString(poolTx.ToBJJ)
	if poolTx.ToBJJ != hezcommon.EmptyBJJComp {
		toBJJ = bjjToString(poolTx.ToBJJ)
	}
	return APITx{
		TxID:      poolTx.TxID,
		Type:      string(poolTx.Type),
		TokenID:   uint32(poolTx.TokenID),
		FromIdx:   idxToHez(poolTx.FromIdx, token.Symbol),
		ToIdx:     toIdx,
		ToEthAddr: toEth,
		ToBJJ:     toBJJ,
		Amount:    poolTx.Amount.String(),
		Fee:       uint64(poolTx.Fee),
		Nonce:     uint64(poolTx.Nonce),
		Signature: poolTx.Signature.String(),
	}
}

// idxToHez convert idx to hez idx
func idxToHez(idx hezcommon.Idx, tokenSymbol string) string {
	return "hez:" + tokenSymbol + ":" + strconv.Itoa(int(idx))
}

// idxToHez convert eth address to hez address
func ethAddrToHez(addr common.Address) string {
	return "hez:" + addr.String()
}

// bjjToString convert the BJJ public key to string
func bjjToString(bjj babyjub.PublicKeyComp) string {
	pkComp := [32]byte(bjj)
	sum := pkComp[0]
	for i := 1; i < len(pkComp); i++ {
		sum += pkComp[i]
	}
	bjjSum := append(pkComp[:], sum)
	return "hez:" + base64.RawURLEncoding.EncodeToString(bjjSum)
}

// APITx is a representation of a transaction API request.
type APITx struct {
	TxID      hezcommon.TxID `json:"id" binding:"required"`
	Type      string         `json:"type"`
	TokenID   uint32         `json:"tokenId"`
	FromIdx   string         `json:"fromAccountIndex" binding:"required"`
	ToIdx     string         `json:"toAccountIndex"`
	ToEthAddr string         `json:"toHezEthereumAddress"`
	ToBJJ     string         `json:"toBjj"`
	Amount    string         `json:"amount" binding:"required"`
	Fee       uint64         `json:"fee"`
	Nonce     uint64         `json:"nonce"`
	Signature string         `json:"signature"`
}
