/*
	Copyright 2019 whiteblock Inc.
	This file is a part of the genesis.

	Genesis is free software: you can redistribute it and/or modify
    it under the terms of the GNU General Public License as published by
    the Free Software Foundation, either version 3 of the License, or
    (at your option) any later version.

    Genesis is distributed in the hope that it will be useful,
    but WITHOUT ANY WARRANTY; without even the implied warranty of
    MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
    GNU General Public License for more details.

    You should have received a copy of the GNU General Public License
    along with this program.  If not, see <https://www.gnu.org/licenses/>.
*/

//Package ethereum provides functions to assist with Ethereum related functionality
package ethereum

import (
	"crypto/ecdsa"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/whiteblock/genesis/util"
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

// UnmarshalJSON handles the conversion from json to account
func (acc *Account) UnmarshalJSON(data []byte) error {
	var tmp map[string]string
	err := json.Unmarshal(data, &tmp)
	if err != nil {
		return util.LogError(err)
	}
	pk, ok := tmp["privateKey"]
	if !ok {
		return fmt.Errorf("missing field \"privateKey\"")
	}
	newAcc, err := CreateAccountFromHex(pk)
	if err != nil {
		return util.LogError(err)
	}
	acc.PrivateKey = newAcc.PrivateKey
	acc.PublicKey = newAcc.PublicKey
	acc.Address = newAcc.Address
	return nil
}

// NewAccount creates an account from a SECP256K1 ECDSA private key
func NewAccount(privKey *ecdsa.PrivateKey) *Account {
	pubKey := privKey.Public().(*ecdsa.PublicKey)
	addr := crypto.PubkeyToAddress(*pubKey)
	return &Account{PrivateKey: privKey, PublicKey: pubKey, Address: addr}
}

// GenerateEthereumAddress generates a new, random Ethereum account
func GenerateEthereumAddress() (*Account, error) {
	privKey, err := crypto.GenerateKey()
	if err != nil {
		return nil, util.LogError(err)
	}
	return NewAccount(privKey), nil
}

// CreateAccountFromHex creates an account from a hex encoded private key
func CreateAccountFromHex(hexPK string) (*Account, error) {
	privKey, err := crypto.HexToECDSA(hexPK)
	if err != nil {
		return nil, util.LogError(err)
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
			return nil, util.LogError(err)
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
