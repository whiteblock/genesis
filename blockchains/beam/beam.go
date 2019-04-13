package beam

import (
	db "../../db"
	state "../../state"
	util "../../util"
	"fmt"
	"log"
	"regexp"
	"strings"
)

var conf *util.Config

func init() {
	conf = util.GetConfig()
}

const port int = 10000

func Build(details db.DeploymentDetails, servers []db.Server, clients []*util.SshClient,
	buildState *state.BuildState) ([]string, error) {

	beamConf, err := NewConf(details.Params)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	buildState.SetBuildSteps(0 + (details.Nodes * 4))

	buildState.SetBuildStage("Setting up the wallets")
	/**Set up wallets**/
	ownerKeys := []string{}
	secretMinerKeys := []string{}
	// walletIDs := []string{}
	for i, server := range servers {
		for localId, _ := range server.Ips {
			clients[i].DockerExec(localId, "beam-wallet --command init --pass password") //ign err

			res1, _ := clients[i].DockerExec(localId, "beam-wallet --command export_owner_key --pass password") //ign err

			buildState.IncrementBuildProgress()

			re := regexp.MustCompile(`(?m)^Owner([A-z|0-9|\s|\:|\/|\+|\=])*$`)
			ownKLine := re.FindAllString(res1, -1)[0]
			ownerKeys = append(ownerKeys, strings.Split(ownKLine, " ")[3])

			res2, _ := clients[i].DockerExec(localId, "beam-wallet --command export_miner_key --subkey=1 --pass password") //ign err

			re = regexp.MustCompile(`(?m)^Secret([A-z|0-9|\s|\:|\/|\+|\=])*$`)
			secMLine := re.FindAllString(res2, -1)[0]
			secretMinerKeys = append(secretMinerKeys, strings.Split(secMLine, " ")[3])
			buildState.IncrementBuildProgress()
		}
	}

	ips := []string{}
	for _, server := range servers {
		for _, ip := range server.Ips {
			ips = append(ips, ip)
		}
	}
	buildState.SetBuildStage("Creating node configuration files")
	/**Create node config files**/
	node := 0
	for i, server := range servers {
		for range server.Ips {
			beam_node_config, err := makeNodeConfig(beamConf, ownerKeys[node], secretMinerKeys[node])

			beam_wallet_config := []string{
				"# Emission.Value0=800000000",
				"# Emission.Drop0=525600",
				"# Emission.Drop1=2102400",
				"Maturity.Coinbase=1",
				"# Maturity.Std=0",
				"# MaxBodySize=0x100000",
				"DA.Target_s=1",
				"# DA.MaxAhead_s=900",
				"# DA.WindowWork=120",
				"# DA.WindowMedian0=25",
				"# DA.WindowMedian1=7",
				"DA.Difficulty0=100",
				"# AllowPublicUtxos=0",
				"# FakePoW=0",
			}
			for _, ip := range append(ips[:node], ips[node+1:]...) {
				beam_node_config += fmt.Sprintf("peer=%s:%d\n", ip, port)
			}
			err = buildState.Write("beam-node.cfg", beam_node_config)
			if err != nil {
				log.Println(err)
				return nil, err
			}
			err = buildState.Write("beam-wallet.cfg", util.CombineConfig(beam_wallet_config))
			if err != nil {
				log.Println(err)
				return nil, err
			}

			err = clients[i].Scp("beam-node.cfg", "/home/appo/beam-node.cfg")
			if err != nil {
				log.Println(err)
				return nil, err
			}
			defer clients[i].Run("rm -f /home/appo/beam-node.cfg")

			err = clients[i].Scp("beam-wallet.cfg", "/home/appo/beam-wallet.cfg")
			if err != nil {
				log.Println(err)
				return nil, err
			}
			defer clients[i].Run("rm -f /home/appo/beam-wallet.cfg")

			err = clients[i].DockerCp(node, "/home/appo/beam-node.cfg", "/beam/")
			if err != nil {
				log.Println(err)
				return nil, err
			}

			err = clients[i].DockerCp(node, "/home/appo/beam-wallet.cfg", "/beam/")
			if err != nil {
				log.Println(err)
				return nil, err
			}

			// fmt.Println(config)
			node++
			buildState.IncrementBuildProgress()
		}
	}

	totNodes := 0
	buildState.SetBuildStage("Starting beam")
	for i, server := range servers {
		for localId, ip := range server.Ips {
			if totNodes >= int(beamConf.Validators) {
				_, err := clients[i].DockerExecd(localId, "beam-node --mining_threads 1")
				if err != nil {
					log.Println(err)
					return nil, err
				}
			} else {
				_, err := clients[i].DockerExecd(localId, "beam-node")
				if err != nil {
					log.Println(err)
					return nil, err
				}

			}
			err = clients[i].DockerExecdLog(localId, fmt.Sprintf("beam-wallet --command listen -n %s:%d --pass password", ip, port))
			if err != nil {
				log.Println(err)
				return nil, err
			}
			buildState.IncrementBuildProgress()
			totNodes++
		}
	}
	return nil, nil
}

func Add(details db.DeploymentDetails, servers []db.Server, clients []*util.SshClient,
	newNodes map[int][]string, buildState *state.BuildState) ([]string, error) {
	return nil, nil
}
