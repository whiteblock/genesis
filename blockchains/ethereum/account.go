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
	return hex.EncodeToString(crypto.FromECDSAPub(acc.PublicKey))[2:]
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

func NewAccount(privKey *ecdsa.PrivateKey) *Account {
	pubKey := privKey.Public().(*ecdsa.PublicKey)
	addr := crypto.PubkeyToAddress(*pubKey)
	return &Account{PrivateKey: privKey, PublicKey: pubKey, Address: addr}
}

// GenerateEthereumAddress generates a new, random Ethereum account
func GenerateEthereumAddress() (*Account, error) {
	privKey, err := crypto.GenerateKey()
	if err != nil {
		log.Println(err)
		return nil, err
	}
	return NewAccount(privKey), nil
}

func CreateAccountFromHex(hexPK string) (*Account, error) {
	privKey, err := crypto.HexToECDSA(hexPK)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	return NewAccount(privKey), nil
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

//ExtractAddresses turns an array of accounts into an array of addresses
func ExtractAddresses(accs []*Account) []string {
	out := make([]string, len(accs))
	for i := range accs {
		out[i] = accs[i].HexAddress()
	}
	return out
}

//ExtractAddressesNoPrefix turns an array of accounts into an array of addresses without the 0x prefix
func ExtractAddressesNoPrefix(accs []*Account) []string {
	out := make([]string, len(accs))
	for i := range accs {
		out[i] = accs[i].HexAddress()[2:]
	}
	return out
}
