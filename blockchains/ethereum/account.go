//Package ethereum provides functions to assist with Ethereum related functionality
package ethereum

import (
	"crypto/ecdsa"
	"encoding/hex"
	"encoding/json"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"log"
	"strings"
)

// Account represents an ethereum account
type Account struct {
	PrivateKey *ecdsa.PrivateKey
	PublicKey  *ecdsa.PublicKey
	Address    common.Address
}

// HexPrivateKey gets the private key in hex format
func (acc Account) HexPrivateKey() string {
	return hex.EncodeToString(crypto.FromECDSA(acc.PrivateKey))
}

// HexPublicKey gets the public key in hex format
func (acc Account) HexPublicKey() string {
	return hex.EncodeToString(crypto.FromECDSAPub(acc.PublicKey))
}

// HexAddress gets the address in hex format
func (acc Account) HexAddress() string {
	return strings.ToLower(acc.Address.Hex())
}

// MarshalJSON handles the marshaling of Acount into JSON, so that
// the fields are exposed in their hex encodings
func (acc Account) MarshalJSON() ([]byte, error) {
	return json.Marshal(&struct {
		PrivateKey string `json:"privateKey"`
		PublicKey  string `json:"publicKey"`
		Address    string `json:"address"`
	}{
		PrivateKey: acc.HexPrivateKey(),
		PublicKey:  acc.HexPublicKey(),
		Address:    acc.HexAddress(),
	})
}

// GenerateEthereumAddress generates a new, random Ethereum account
func GenerateEthereumAddress() (*Account, error) {
	privKey, err := crypto.GenerateKey()
	if err != nil {
		log.Println(err)
		return nil, err
	}
	pubKey := privKey.Public().(*ecdsa.PublicKey)
	addr := crypto.PubkeyToAddress(*pubKey)
	return &Account{PrivateKey: privKey, PublicKey: pubKey, Address: addr}, nil
}

// GenerateAccounts is a convience function to generate an arbitrary number of accounts
// using GenerateEthereumAddress
func GenerateAccounts(accounts int) ([]*Account, error) {
	out := []*Account{}
	for i := 0; i < accounts; i++ {
		acc, err := GenerateEthereumAddress()
		if err != nil {
			log.Println(err)
			return nil, err
		}
		out = append(out, acc)
	}
	return out, nil
}
