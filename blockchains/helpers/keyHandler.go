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

package helpers

import (
	"encoding/json"
	"fmt"
	"github.com/whiteblock/genesis/db"
	"github.com/whiteblock/genesis/ssh"
	"github.com/whiteblock/genesis/util"
)

// KeyMaster is a static resource key manager
// Uses keys stored in the blockchains resource directory, so that
// keys can remain consistent among builds and also to save
// time on builds where a large number of keys are needed.
// Note: This is not thread safe and may need external synchronization.
type KeyMaster struct {
	//PrivateKeys contains the static pool of private keys
	PrivateKeys []string
	//PublicKeys contains the static pool of private keys.
	PublicKeys []string
	//index is the current index in the pool.
	index int
	//generator is the function which can dynamically generate keys in case the static pool runs out
	generator func(client *ssh.Client) (util.KeyPair, error)
}

// NewKeyMaster creates a new KeyMaster using the provided deployment details and blockchain.
// Currently details is not used, but in the future, it should be used to allow the user to provide
// their own static keys to be used in the pool.
func NewKeyMaster(details *db.DeploymentDetails, blockchain string) (*KeyMaster, error) {
	out := new(KeyMaster)
	dat, err := GetStaticBlockchainConfig(blockchain, "privatekeys.json")
	if err != nil {
		return nil, util.LogError(err)
	}
	err = json.Unmarshal(dat, &out.PrivateKeys)
	if err != nil {
		return nil, util.LogError(err)
	}
	dat, err = GetStaticBlockchainConfig(blockchain, "publickeys.json")
	if err != nil {
		return nil, util.LogError(err)
	}
	err = json.Unmarshal(dat, &out.PublicKeys)
	if err != nil {
		return nil, util.LogError(err)
	}
	out.index = 0
	return out, nil
}

// AddGenerator sets the backup key generator function for km KeyMaster
func (km *KeyMaster) AddGenerator(gen func(client *ssh.Client) (util.KeyPair, error)) {
	km.generator = gen
}

// GenerateKeyPair generates a new key pair if a generator function has been provided
func (km *KeyMaster) GenerateKeyPair(client *ssh.Client) (util.KeyPair, error) {
	if km.generator != nil {
		return km.generator(client)
	}
	return util.KeyPair{}, fmt.Errorf("no generator provided")
}

// GetKeyPair fetches a key pair, will use up keys in the static pool until it runs out,
// if it runs out, it will use the given generator to create new keys
func (km *KeyMaster) GetKeyPair(client *ssh.Client) (util.KeyPair, error) {
	if km.index >= len(km.PrivateKeys) || km.index >= len(km.PublicKeys) {
		return km.GenerateKeyPair(client)
	}

	out := util.KeyPair{PrivateKey: km.PrivateKeys[km.index], PublicKey: km.PublicKeys[km.index]}
	km.index++
	return out, nil
}

// GetMappedKeyPairs returns key pairs mapped to arbitrary string values.
// Useful for named key pairs
func (km *KeyMaster) GetMappedKeyPairs(args []string, client *ssh.Client) (map[string]util.KeyPair, error) {
	keyPairs := make(map[string]util.KeyPair)

	for _, arg := range args {
		keyPair, err := km.GetKeyPair(client)
		if err != nil {
			return nil, util.LogError(err)
		}
		keyPairs[arg] = keyPair
	}
	return keyPairs, nil
}

//GetServerKeyPairs is DEPRECATED, but maps the ip addresses of nodes to their own key pair
func (km *KeyMaster) GetServerKeyPairs(servers []db.Server, clients []*ssh.Client) (map[string]util.KeyPair, error) {
	ips := []string{}
	for _, server := range servers {
		for _, ip := range server.Ips {
			ips = append(ips, ip)
		}
	}
	return km.GetMappedKeyPairs(ips, clients[0])
}
