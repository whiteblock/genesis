package helpers

import (
	db "../../db"
	util "../../util"
	"encoding/json"
	"fmt"
	"log"
)

/*
	Static resource key manager
	Uses keys stored in the blockchains resource directory, so that
	keys can remain consistent among builds and also to save
	time on builds where a large number of keys are needed.
*/
type KeyMaster struct {
	PrivateKeys []string
	PublicKeys  []string
	index       int
	generator   func(client *util.SshClient) (util.KeyPair, error)
}

func NewKeyMaster(details *db.DeploymentDetails, blockchain string) (*KeyMaster, error) {
	out := new(KeyMaster)
	dat, err := util.GetBlockchainConfig(blockchain, "privatekeys.json", details.Files)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	err = json.Unmarshal(dat, &out.PrivateKeys)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	dat, err = util.GetBlockchainConfig(blockchain, "publickeys.json", details.Files)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	err = json.Unmarshal(dat, &out.PublicKeys)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	out.index = 0
	return out, nil
}

func (this *KeyMaster) AddGenerator(gen func(client *util.SshClient) (util.KeyPair, error)) {
	this.generator = gen
}

func (this *KeyMaster) GenerateKeyPair(client *util.SshClient) (util.KeyPair, error) {
	if this.generator != nil {
		return this.generator(client)
	}
	return util.KeyPair{}, fmt.Errorf("No generator provided")
}

func (this *KeyMaster) GetKeyPair(client *util.SshClient) (util.KeyPair, error) {
	if this.index >= len(this.PrivateKeys) || this.index >= len(this.PublicKeys) {
		return this.GenerateKeyPair(client)
	}

	out := util.KeyPair{PrivateKey: this.PrivateKeys[this.index], PublicKey: this.PublicKeys[this.index]}
	this.index++
	return out, nil
}

func (this *KeyMaster) GetMappedKeyPairs(args []string, client *util.SshClient) (map[string]util.KeyPair, error) {
	keyPairs := make(map[string]util.KeyPair)

	for _, arg := range args {
		keyPair, err := this.GetKeyPair(client)
		if err != nil {
			log.Println(err)
			return nil, err
		}
		keyPairs[arg] = keyPair
	}
	return keyPairs, nil
}

func (this *KeyMaster) GetServerKeyPairs(servers []db.Server, clients []*util.SshClient) (map[string]util.KeyPair, error) {
	ips := []string{}
	for _, server := range servers {
		for _, ip := range server.Ips {
			ips = append(ips, ip)
		}
	}
	return this.GetMappedKeyPairs(ips, clients[0])
}