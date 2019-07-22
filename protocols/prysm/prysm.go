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

//Package prysm handles prysm specific functionality
package prysm

import (
	"crypto/rand"
	"fmt"
	"github.com/libp2p/go-libp2p-crypto"
	peer "github.com/libp2p/go-libp2p-peer"
	log "github.com/sirupsen/logrus"
	"github.com/whiteblock/genesis/db"
	"github.com/whiteblock/genesis/protocols/helpers"
	"github.com/whiteblock/genesis/protocols/registrar"
	"github.com/whiteblock/genesis/ssh"
	"github.com/whiteblock/genesis/testnet"
	"github.com/whiteblock/genesis/util"
	"reflect"
)

var conf *util.Config

const (
	blockchain    = "prysm"
	p2pPort       = 3000
	numValidators = 8
)

func init() {
	conf = util.GetConfig()

	registrar.RegisterBuild(blockchain, build)
	registrar.RegisterAddNodes(blockchain, add)
	registrar.RegisterServices(blockchain, GetServices)
	registrar.RegisterDefaults(blockchain, helpers.DefaultGetDefaultsFn(blockchain))
	registrar.RegisterParams(blockchain, helpers.DefaultGetParamsFn(blockchain))
}

// build builds out a fresh new prysm test network
func build(tn *testnet.TestNet) error {
	_, err := newConf(tn.LDD.Params)
	if err != nil {
		return util.LogError(err)
	}
	tn.BuildState.SetBuildSteps(0 + (tn.LDD.Nodes * 4))

	nodeKeyPairs := map[string]crypto.PrivKey{}
	for _, node := range tn.Nodes {
		prvKey, _, _ := crypto.GenerateKeyPairWithReader(crypto.RSA, 2048, rand.Reader)
		nodeKeyPairs[node.ID] = prvKey
	}

	prysmIPList := []string{}

	tn.BuildState.SetBuildStage("Starting prysm")
	err = helpers.AllNodeExecCon(tn, func(client ssh.Client, _ *db.Server, node ssh.Node) error {

		prysmIPList = append(prysmIPList, node.GetIP())

		peers := ""

		for _, peerNode := range tn.Nodes {
			if node.GetID() == peerNode.GetID() {
				continue
			}
			peers += fmt.Sprintf(" --peer=/ip4/%s/tcp/%d/p2p/%s:%d", peerNode.IP, p2pPort, idString(nodeKeyPairs[peerNode.GetID()]), p2pPort)
			tn.BuildState.IncrementBuildProgress()
		}

		marshaled, err := crypto.MarshalPrivateKey(nodeKeyPairs[node.GetID()])
		if err != nil {
			log.WithError(err).Error("Could not marshal key")
			return err
		}
		keyStr := crypto.ConfigEncodeKey(marshaled)

		err = helpers.SingleCp(client, tn.BuildState, node, []byte(keyStr), "/etc/identity.key")
		if err != nil {
			log.WithError(err).Error("Could not marshal key")
			return err
		}
		defer tn.BuildState.IncrementBuildProgress()
		var prometheusInstrumentationPort string
		obj := tn.CombinedDetails.Params["prometheusInstrumentationPort"]
		if obj != nil && reflect.TypeOf(obj).Kind() == reflect.String {
			prometheusInstrumentationPort = obj.(string)
		}
		if prometheusInstrumentationPort == "" {
			prometheusInstrumentationPort = "8008"
		}

		/*
			var contract string
			obj = tn.CombinedDetails.Params["contract"]
			if obj != nil && reflect.TypeOf(obj).Kind() == reflect.String {
				contract = obj.(string)
			}
		*/

		var validatorsPassword string
		obj = tn.CombinedDetails.Params["validatorsPassword"]
		if obj != nil && reflect.TypeOf(obj).Kind() == reflect.String {
			validatorsPassword = obj.(string)
		}

		var logFolder string
		obj = tn.CombinedDetails.Params["logFolder"]
		if obj != nil && reflect.TypeOf(obj).Kind() == reflect.String {
			logFolder = obj.(string)
		} else {
			logFolder = ""
		}

		_, err = client.DockerExecd(node,
			fmt.Sprintf("/prysm/bazel-bin/beacon-chain/linux_amd64_stripped/beacon-chain "+
				"--monitoring-port=%s --no-discovery %s --log-file %s/beacon-chain%d.log "+
				" --p2p-priv-key /etc/identity.key --clear-db --hobbits --p2p-port %d --p2p-host-ip %s"+
				" --verbosity trace",
				prometheusInstrumentationPort, peers, logFolder,
				node.GetAbsoluteNumber(), p2pPort, node.GetIP()))
		if err != nil {
			return util.LogError(err)
		}

		for i := 1; i <= numValidators; i++ {
			_, err = client.DockerExecd(node,
				fmt.Sprintf("/prysm/bazel-bin/validator/linux_amd64_pure_stripped/validator "+
					"accounts create --password %s --keystore-path %s/key%d-%d",
					validatorsPassword, logFolder, node.GetRelativeNumber(), i))
			if err != nil {
				return util.LogError(err)
			}
		}

		for i := 1; i <= numValidators; i++ {
			_, err = client.DockerExecd(node,
				fmt.Sprintf("bash -c \"/prysm/bazel-bin/validator/linux_amd64_pure_stripped/validator"+
					" --password %s --keystore-path %s/key%d-%d --monitoring-port 10%d%d 2>&1 | tee /output.log\"",
					validatorsPassword, logFolder, node.GetRelativeNumber(), i, node.GetRelativeNumber(), i))
			if err != nil {
				return util.LogError(err)
			}
		}

		return err
	})
	tn.BuildState.Set("IPList", prysmIPList)
	tn.BuildState.Set("p2pPort", p2pPort)
	return util.LogError(err)
}

// add handles adding nodes to the testnet
func add(tn *testnet.TestNet) error {
	return nil
}

func idString(k crypto.PrivKey) string {
	pid, err := peer.IDFromPrivateKey(k)
	if err != nil {
		panic(err)
	}
	return pid.Pretty()
}
