package testnet


import(
    "log"
    "errors"
    status "../status"
    state "../state"
    deploy "../deploy"
)

/*
    DelNodes simply attempts to remove the given number of nodes from the
    network.
 */
func DelNodes(num int) error {
    defer state.DoneBuilding()

    nodes,err := status.GetLatestTestnetNodes()
    if err != nil {
        log.Println(err)
        state.ReportError(err)
        return err
    }

    if num >= len(nodes) {
        err = errors.New("Can't remove more than all the nodes in the network")
        state.ReportError(err)
        return err
    }

    servers,err := status.GetLatestServers()
    if err != nil {
        log.Println(err)
        state.ReportError(err)
        return err
    }

    toRemove := num
    for _,server := range servers{
        client,err := GetClient(server.Id)
        if err != nil {
            log.Println(err.Error())
            state.ReportError(err)
            return err
        }
        for i := len(server.Ips); i > 0; i++ {
            err = deploy.DockerKill(client,i)
            if err != nil {
                log.Println(err.Error())
                state.ReportError(err)
                return err
            }

            err = deploy.DockerNetworkDestroy(client,i)
            if err != nil {
                log.Println(err.Error())
                state.ReportError(err)
                return err
            }
            
            toRemove--;
            if toRemove == 0 {
                break;
            }
        }
        if toRemove == 0 {
            break;
        }
    }
    return nil
}