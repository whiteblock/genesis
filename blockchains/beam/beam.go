package beam

import (
	"fmt"
	"log"
	"regexp"
	"strings"

	db "../../db"
	util "../../util"
)

var conf *util.Config

func init() {
	conf = util.GetConfig()
}

const port int = 8100

func Build(data map[string]interface{}, nodes int, servers []db.Server, clients []*util.SshClient) ([]string, error) {

	beamConf, err := NewConf(data)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	/**Set up wallets**/
	ownerKeys := []string{}
	secretMinerKeys := []string{}
	for i, server := range servers {
		for localId, _ := range server.Ips {
			_, err := clients[i].DockerExec(localId, "beam-wallet --command init --pass password")
			if err != nil {
				log.Println(err)
				// return nil, err
			}

			res1, err := clients[i].DockerExec(localId, "beam-wallet --command export_owner_key --pass password")
			if err != nil {
				log.Println(err)
				// return nil, err
			}
			re := regexp.MustCompile(`(?m)^Owner([A-z|0-9|\s|\:|\/|\+|\=])*$`)
			ownKLine := re.FindAllString(res1, -1)[0]
			ownerKeys = append(ownerKeys, strings.Split(ownKLine, " ")[3])

			res2, err := clients[i].DockerExec(localId, "beam-wallet --command export_miner_key --subkey=1 --pass password")
			if err != nil {
				log.Println(err)
				// return nil, err
			}
			re = regexp.MustCompile(`(?m)^Secret([A-z|0-9|\s|\:|\/|\+|\=])*$`)
			secMLine := re.FindAllString(res2, -1)[0]
			secretMinerKeys = append(secretMinerKeys, strings.Split(secMLine, " ")[3])
		}
	}

	ips := []string{}
	for _, server := range servers {
		for _, ip := range server.Ips {
			ips = append(ips, ip)
		}
	}

	/**Create node config files**/
	node := 0
	for i, server := range servers {
		for range server.Ips {
			config := []string{
				"# port=10000",
				"# stratum_port=0",
				"# stratum_secrets_path=.",
				"# wallet_seed=some_secret_string",
				"# wallet_phrase=",
				"# log_level=verbose",
				"# file_log_level=verbose",
				"# storage=node.db",
				"# history_dir=",
				"# temp_dir=",
				"# treasury_path=treasury.mw",
				"# mining_threads=1",
				"miner_type=cpu",
				"# verification_threads=-1",
				"# import=0",
				"# resync=0",
				"# crash=0",

				fmt.Sprintf("key_owner=%s", ownerKeys[node]),
				fmt.Sprintf("key_mine=%s", secretMinerKeys[node]),
				"pass=password",

				"# EmissionValue0=800000000",
				"# EmissionDrop0=525600",
				"# EmissionDrop1=2102400",
				"# MaturityCoinbase=240",
				"# MaturityStd=0",
				"# MaxBodySize=0x100000",
				"# DesiredRate_s=60",
				"# DifficultyReviewWindow=1440",
				"# TimestampAheadThreshold_s=7200",
				"# WindowForMedian=25",
				"# AllowPublicUtxos=0",
				"# FakePoW=0",
			}
			for _, ip := range append(ips[:node], ips[node+1:]...) {
				config = append(config, fmt.Sprintf("peer=%s:%d", ip, port))
			}
			err := util.Write("./beam-node.cfg", util.CombineConfig(config))
			if err != nil {
				log.Println(err)
				return nil, err
			}

			defer util.Rm("./beam-node.cfg")

			err = clients[i].Scp("./beam-node.cfg", "/home/appo/beam-node.cfg")
			if err != nil {
				log.Println(err)
				return nil, err
			}

			defer clients[i].Run("rm -f /home/appo/beam-node.cfg")

			_, err = clients[i].Run(fmt.Sprintf("docker cp /home/appo/beam-node.cfg %s%d:/", conf.NodePrefix, node))
			if err != nil {
				log.Println(err)
				return nil, err
			}

			node++
		}
	}

	totNodes := 0

	for i, server := range servers {
		for localId, _ := range server.Ips {
			if totNodes < int(beamConf.Miners) {
				_, err := clients[i].DockerExecd(localId, "beam-node")
				if err != nil {
					log.Println(err)
					return nil, err
				}
			} else if totNodes >= int(beamConf.Miners) {
				_, err := clients[i].DockerExecd(localId, "beam-node --mining_threads 1")
				if err != nil {
					log.Println(err)
					return nil, err
				}
			}
			totNodes++
		}
	}
	return nil, nil
}
