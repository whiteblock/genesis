package deploy

import (
	"../blockchains/helpers"
	"../db"
	"../ssh"
	"../testnet"
	"../util"
	"fmt"
	"log"
)

func handleExtraPublicKeys(tn *testnet.TestNet) error {
	postBuild, ok := util.ExtractStringMap(tn.LDD.Extras, "postbuild")
	if !ok {
		return nil
	}
	fmt.Printf("%#v\n", tn.LDD.Extras["postbuild"])
	SSHDetails, ok := util.ExtractStringMap(postBuild, "ssh")
	if !ok || SSHDetails == nil {
		return nil
	}
	iPubKeys, ok := SSHDetails["pubKeys"]
	if !ok || iPubKeys == nil {
		return nil
	}
	pubKeys := iPubKeys.([]interface{})

	tn.BuildState.Async(func() {
		helpers.AllNewNodeExecCon(tn, func(client *ssh.Client, _ *db.Server, node ssh.Node) error {
			for i := range pubKeys {
				_, err := client.DockerExec(node, fmt.Sprintf(`bash -c 'echo "%v" >> /root/.ssh/authorized_keys'`, pubKeys[i]))
				if err != nil {
					log.Println(err)
					return err
				}
			}
			return nil
		})
	})
	return nil
}

func handlePostBuild(tn *testnet.TestNet) error {
	return handleExtraPublicKeys(tn)
	//return nil
}
