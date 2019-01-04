package deploy

import (
    "fmt"
    db "../db"
    util "../util"
)

func setupHPSwitch(server db.Server,nodes int){
    fmt.Printf("Setting up the HP Switch...")
    util.BashExec("tmux kill-session -t hpconf")
    util.BashExec("tmux new -s hpconf -d")
    util.BashExec(fmt.Sprintf("tmux send-keys -t hpconf 'ssh admin@%s' C-m",server.Switches[0].Addr))
    util.BashExec("tmux send-keys -t hpconf ' ' C-m")
    util.BashExec("tmux send-keys -t hpconf 'config' C-m")

    
    gws := util.GetGateways(server.Id, nodes)
    vlan := 101
    for i,gw := range gws {
        util.BashExec(
            fmt.Sprintf("tmux send-keys -t hpconf 'vlan %d' C-m",vlan+i))

        util.BashExec("tmux send-keys -t hpconf 'no ip address' C-m")

        util.BashExec(
            fmt.Sprintf("tmux send-keys -t hpconf 'name VLAN_%d' C-m",vlan+i))
        util.BashExec("tmux send-keys -t hpconf 'tagged 25-26' C-m")
        util.BashExec(
            fmt.Sprintf("tmux send-keys -t hpconf 'ip address %s/%d' C-m",gw,util.GetSubnet()))


    }   
    util.BashExec("tmux send-keys -t hpconf 'wr me' C-m")
    fmt.Printf("done\n")
}