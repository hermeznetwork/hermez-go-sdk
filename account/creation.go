package account

import (
	"crypto/ecdsa"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"log"

	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/crypto"
	hezcommon "github.com/hermeznetwork/hermez-node/common"
	"github.com/iden3/go-iden3-crypto/babyjub"
	hdwallet "github.com/miguelmota/go-ethereum-hdwallet"
)

const (
	ethDerivationPath = "m/44'/60'/0'/0/%d"
	hermezWalletMsg   = "Hermez Network account access.\n\nSign this message if you are in a trusted application only."
)

// CreateBJJWalletFromSignedMsg creates BJJWallet from signed hermez standard message to generate account
func CreateBJJWalletFromSignedMsg(signedMsg []byte) (bjjWallet BJJWallet, ethAccount accounts.Account, err error) {
	// Change the last item value of the byte array from 1 to 28
	signedMsg[len(signedMsg)-1] += 27

	hermezWalletMsgEncoded := "0x" + hex.EncodeToString(signedMsg)
	// log.Println("hermezWalletMsgEncoded: ", hermezWalletMsgEncoded)
	hermezWalletMsgSignedHash := crypto.Keccak256([]byte(hermezWalletMsgEncoded))

	var bjjPvtKey babyjub.PrivateKey
	// Copy to the private key
	copy(bjjPvtKey[:], hermezWalletMsgSignedHash[:])

	bjjAddress, err := FromBJJPubKeyCompToHezBJJAddress(bjjPvtKey.Public().Compress())
	if err != nil {
		err = fmt.Errorf("[CreateBJJWalletFromSignedMsg] Error generating BJJ address from BJJ public key. Account: %+v - Error: %s\n", bjjPvtKey.Public().Compress(), err.Error())
		log.Println(err.Error())
		return
	}

	decodedBjjPubKey, err := hex.DecodeString(bjjPvtKey.Public().Compress().String())
	if err != nil {
		err = fmt.Errorf("[CreateBJJWalletFromSignedMsg] Error decoding BJJ public key. Account: %+v - Error: %s\n", bjjPvtKey, err.Error())
		log.Println(err.Error())
		return
	}

	var bjjPubKeyCompressed babyjub.PublicKeyComp
	temp := hezcommon.SwapEndianness(decodedBjjPubKey)
	copy(bjjPubKeyCompressed[:], temp[:])

	bjjWallet.PrivateKey = bjjPvtKey
	bjjWallet.PublicKey = bjjPubKeyCompressed
	bjjWallet.HezBjjAddress = bjjAddress
	bjjWallet.EthAccount = ethAccount
	bjjWallet.HezEthAddress = "hez:" + ethAccount.Address.Hex()

	return
}

// CreateBjjWalletFromMnemonic Create a Babyjubjub Wallet from Mnemonic
func CreateBjjWalletFromMnemonic(mnemonic string) (bjjWallet BJJWallet, ethAccount accounts.Account, err error) {
	ethWallet, err := hdwallet.NewFromMnemonic(mnemonic)
	if err != nil {
		err = fmt.Errorf("[CreateBjjWalletFromMnemonic] Error creating ethereum account from mnemonic - Error: %s\n", err.Error())
		log.Println(err.Error())
		return
	}

	// Generate ETH account
	path := hdwallet.MustParseDerivationPath(fmt.Sprintf(ethDerivationPath, 0))
	ethAccount, err = ethWallet.Derive(path, true)
	if err != nil {
		err = fmt.Errorf("[CreateBjjWalletFromMnemonic] Error deriving the account from mnemonic Error: %s\n", err.Error())
		log.Println(err.Error())
		return
	}

	hermezWalletMsgSigned, err := ethWallet.SignText(ethAccount, []byte(hermezWalletMsg))
	if err != nil {
		err = fmt.Errorf("[CreateBjjWalletFromMnemonic] Error signing key msg to generate BJJ private key. Account: %s - Error: %s\n", ethAccount.Address.Hex(), err.Error())
		log.Println(err.Error())
		return
	}

	bjjWallet, ethAccount, err = CreateBJJWalletFromSignedMsg(hermezWalletMsgSigned)
	if err != nil {
		err = fmt.Errorf("[CreateBjjWalletFromMnemonic] Error generating BJJ Wallet. Account: %s - Error: %s\n", ethAccount.Address.Hex(), err.Error())
		log.Println(err.Error())
		return
	}

	return
}

// CreateBjjWalletFromHexPvtKey Create a Babyjubjub Wallet from Hexdecimal Private Key
func CreateBjjWalletFromHexPvtKey(hexPvtKey string) (bjjWallet BJJWallet, ethAccount accounts.Account, err error) {
	ecdsaPvtKey, err := crypto.HexToECDSA(hexPvtKey)
	if err != nil {
		log.Printf("[CreateBjjWalletFromHexPvtKey] Error when creating private key: %s\n", err.Error())
		return
	}
	ecdsaPubKey := ecdsaPvtKey.Public().(*ecdsa.PublicKey)
	ethAccount.Address = crypto.PubkeyToAddress(*ecdsaPubKey)

	hermezWalletMsgHash := accounts.TextHash([]byte(hermezWalletMsg))
	hermezWalletMsgSigned, err := crypto.Sign(hermezWalletMsgHash, ecdsaPvtKey)
	if err != nil {
		log.Printf("[CreateBjjWalletFromHexPvtKey] Error signing key msg to generate BJJ private key. Account: %s - Error: %s\n", ethAccount.Address.Hex(), err.Error())
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
		log.Printf("[CreateBjjWalletFromHexPvtKey] Error generating BJJ address from BJJ public key. Account: %+v - Error: %s\n", bjjPvtKey.Public().Compress(), err.Error())
		return
	}

	decodedBjjPubKey, err := hex.DecodeString(bjjPvtKey.Public().Compress().String())
	if err != nil {
		log.Printf("[CreateBjjWalletFromHexPvtKey] Error decoding BJJ public key. Account: %+v - Error: %s\n", bjjPvtKey, err.Error())
		return
	}

	var bjjPubKeyCompressed babyjub.PublicKeyComp
	temp := hezcommon.SwapEndianness(decodedBjjPubKey)
	copy(bjjPubKeyCompressed[:], temp[:])

	bjjWallet.PrivateKey = bjjPvtKey
	bjjWallet.PublicKey = bjjPubKeyCompressed
	bjjWallet.HezBjjAddress = bjjAddress
	bjjWallet.EthAccount = ethAccount
	bjjWallet.HezEthAddress = "hez:" + ethAccount.Address.Hex()

	return
}

// FromBJJPubKeyCompToHezBJJAddress creates a Hermez BJJ account from a BJJ PubKey Compressed
func FromBJJPubKeyCompToHezBJJAddress(pkComp babyjub.PublicKeyComp) (string, error) {
	if len(pkComp.String()) < 10 {
		return "", errors.New("[FromBJJPubKeyCompToHezBJJAddress] Invalid BJJ PubKey Compressed")
	}
	sum := pkComp[0]
	for i := 1; i < len(pkComp); i++ {
		sum += pkComp[i]
	}
	bjjSum := append(pkComp[:], sum)
	return "hez:" + base64.RawURLEncoding.EncodeToString(bjjSum), nil
}
