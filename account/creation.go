package account

import (
	"crypto/ecdsa"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"log"

	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/crypto"
	hezcommon "github.com/hermeznetwork/hermez-node/common"
	"github.com/iden3/go-iden3-crypto/babyjub"
)

// CreateBjjWalletFromHexPvtKey Create a Babyjubjub Wallet from Hexdecimal Private Key
func CreateBjjWalletFromHexPvtKey(hexPvtKey string) (bjjWallet BJJWallet, ethAccount accounts.Account, err error) {
	ecdsaPvtKey, err := crypto.HexToECDSA(hexPvtKey)
	if err != nil {
		log.Printf("[CreateBjjWalletFromHexPvtKey] Error when creating private key: %s\n", err.Error())
		return
	}
	ecdsaPubKey := ecdsaPvtKey.Public().(*ecdsa.PublicKey)
	ethAccount.Address = crypto.PubkeyToAddress(*ecdsaPubKey)

	// BJJ account
	const hermezWalletMsg = "Hermez Network account access.\n\nSign this message if you are in a trusted application only."

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
	bjjWallet.HezEthAddress = "hez:" + ethAccount.Address.String()

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
