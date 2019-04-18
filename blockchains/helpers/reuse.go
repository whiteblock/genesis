package helpers

import (
	ssh "../../ssh"
	state "../../state"
	"fmt"
	"log"
)

func ScpAndDeferRemoval(client *ssh.Client, buildState *state.BuildState, src string, dst string) {
	buildState.Defer(func() { client.Run(fmt.Sprintf("rm -rf %s", dst)) })
	err := client.Scp(src, dst)
	if err != nil {
		log.Println(err)
		buildState.ReportError(err)
		return
	}
}
