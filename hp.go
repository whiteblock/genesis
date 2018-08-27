package main

import (
	"fmt"
)

func setupHPSwitch(server Server,nodes int){
	fmt.Printf("Setting up the HP Switch...")
	bashExec("tmux kill-session -t hpconf")
	bashExec("tmux new -s hpconf -d")
	bashExec(fmt.Sprintf("tmux send-keys -t hpconf 'ssh admin@%s' C-m",server.switches[0].addr))
	bashExec("tmux send-keys -t hpconf ' ' C-m")
	bashExec("tmux send-keys -t hpconf 'config' C-m")

	
	gws := getGateways(server.id, nodes)
	vlan := 101
	for i,gw := range gws {
		bashExec(
			fmt.Sprintf("tmux send-keys -t hpconf 'vlan %d' C-m",vlan+i))

		bashExec("tmux send-keys -t hpconf 'no ip address' C-m")

		bashExec(
			fmt.Sprintf("tmux send-keys -t hpconf 'name VLAN_%d' C-m",vlan+i))
		bashExec("tmux send-keys -t hpconf 'tagged 25-26' C-m")
		bashExec(
			fmt.Sprintf("tmux send-keys -t hpconf 'ip address %s/%d' C-m",gw,getSubnet()))


	}	
	bashExec("tmux send-keys -t hpconf 'wr me' C-m")
	fmt.Printf("done\n")
}