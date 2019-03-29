package eos

import(
	"encoding/json"
	"io/ioutil"
	"strings"
	"fmt"
	util "../../util"
	db "../../db"
)

type KeyMaster struct {
	PrivateKeys		[]string
	PublicKeys		[]string
	index			int
}

func NewKeyMaster() (*KeyMaster,error) {
	out := new(KeyMaster)
	dat, err := ioutil.ReadFile("./resources/eos/privatekeys.json")
	if err != nil {
		return nil,err
	}
	err = json.Unmarshal(dat,&out.PrivateKeys)
	if err != nil {
		return nil,err
	}
	dat, err = ioutil.ReadFile("./resources/eos/publickeys.json")
	if err != nil {
		return nil,err
	}
	err = json.Unmarshal(dat,&out.PublicKeys)
	if err != nil {
		return nil,err
	}
	out.index = 0
	return out,nil
}
func (this *KeyMaster) GenerateKeyPair(client *util.SshClient) (util.KeyPair,error) {
    data,err := client.DockerExec(0,"cleos create key --to-console | awk '{print $3}'")
    if err != nil {
        return util.KeyPair{},err
    }
    keyPair := strings.Split(data, "\n")
    if(len(data) < 10){
        return util.KeyPair{},fmt.Errorf("Unexpected create key output %s\n",keyPair)
        panic(1)
    }
    return util.KeyPair{PrivateKey: keyPair[0], PublicKey: keyPair[1]},nil
}


func (this *KeyMaster) GetKeyPair(client *util.SshClient) (util.KeyPair,error) {
	if this.index >= len(this.PrivateKeys) || this.index >= len(this.PublicKeys) {
		return this.GenerateKeyPair(client)
	}

	out := util.KeyPair{PrivateKey: this.PrivateKeys[this.index], PublicKey: this.PublicKeys[this.index]}
	this.index++;
	return out,nil
}


func (this *KeyMaster) GetMappedKeyPairs(args []string,client *util.SshClient) (map[string]util.KeyPair,error) {
	keyPairs := make(map[string]util.KeyPair)

	for _, arg := range args{
		keyPair,err := this.GetKeyPair(client)
		if err != nil {
			return nil,err
		}
		keyPairs[arg] = keyPair
	}
	return keyPairs,nil
}

func (this *KeyMaster) GetServerKeyPairs(servers []db.Server,clients []*util.SshClient) (map[string]util.KeyPair,error){
	ips := []string{}
	for _, server := range servers {
		for _, ip := range server.Ips {
			ips = append(ips,ip)
		}
	}
	return this.GetMappedKeyPairs(ips,clients[0])
}

