package rchain

import(
	"encoding/json"
	"io/ioutil"
	"errors"
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
	dat, err := ioutil.ReadFile("./resources/rchain/privatekeys.json")
	if err != nil {
		return nil,err
	}
	err = json.Unmarshal(dat,&out.PrivateKeys)
	if err != nil {
		return nil,err
	}
	dat, err = ioutil.ReadFile("./resources/rchain/publickeys.json")
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

func (this *KeyMaster) GetKeyPair() (util.KeyPair,error) {
	if this.index >= len(this.PrivateKeys) || this.index >= len(this.PublicKeys) {
		return util.KeyPair{},errors.New("No more keys left")
	}

	out := util.KeyPair{PrivateKey: this.PrivateKeys[this.index], PublicKey: this.PublicKeys[this.index]}
	this.index++;
	return out,nil
}


func (this *KeyMaster) GetMappedKeyPairs(args []string) (map[string]util.KeyPair,error) {
	keyPairs := make(map[string]util.KeyPair)

	for _, arg := range args{
		keyPair,err := this.GetKeyPair()
		if err != nil {
			return nil,err
		}
		keyPairs[arg] = keyPair
	}
	return keyPairs,nil
}

func (this *KeyMaster) GetServerKeyPairs(servers []db.Server) (map[string]util.KeyPair,error){
	ips := []string{}
	for _, server := range servers {
		for _, ip := range server.Ips {
			ips = append(ips,ip)
		}
	}
	return this.GetMappedKeyPairs(ips)
}
