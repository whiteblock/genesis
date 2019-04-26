package deploy

import (
	"../blockchains/helpers"
	"../db"
	"../ssh"
	"../testnet"
	"../util"
	"fmt"
	"log"
	"strings"
)

/**Quick naive interface to Docker calls over ssh*/

/*
   Kill a single node by index on a server
*/
func DockerKill(client *ssh.Client, node int) error {
	_, err := client.Run(fmt.Sprintf("docker rm -f %s%d", conf.NodePrefix, node))
	return err
}

/*
   Kill all nodes on a server
*/
func DockerKillAll(client *ssh.Client) error {
	_, err := client.Run(fmt.Sprintf("docker rm -f $(docker ps -aq -f name=\"%s\")", conf.NodePrefix))
	return err
}

/*
   Create the command to a docker network for a node
*/
func dockerNetworkCreateCmd(subnet string, gateway string, network int, name string) string {
	return fmt.Sprintf("docker network create --subnet %s --gateway %s -o \"com.docker.network.bridge.name=%s%d\" %s",
		subnet,
		gateway,
		conf.BridgePrefix,
		network,
		name)
}

/*
   Create a docker network for a node
*/
func DockerNetworkCreate(tn *testnet.TestNet, serverID int, subnetID int, node int) error {
	command := dockerNetworkCreateCmd(
		util.GetNetworkAddress(subnetID, node),
		util.GetGateway(subnetID, node),
		node,
		fmt.Sprintf("%s%d", conf.NodeNetworkPrefix, node))

	_, err := tn.Clients[serverID].KeepTryRun(command)

	return err
}

func DockerNetworkDestroy(client *ssh.Client, node int) error {
	_, err := client.Run(fmt.Sprintf("docker network rm %s%d", conf.NodeNetworkPrefix, node))
	return err
}

/*
   Remove all whiteblock networks on a node
*/
func DockerNetworkDestroyAll(client *ssh.Client) error {
	_, err := client.Run(fmt.Sprintf(
		"for net in $(docker network ls | grep %s | awk '{print $1}'); do docker network rm $net; done", conf.NodeNetworkPrefix))
	return err
}

func DockerLogin(client *ssh.Client, username string, password string) error {
	user := strings.Replace(username, "\"", "\\\"", -1) //Escape the quotes
	pass := strings.Replace(password, "\"", "\\\"", -1) //Escape the quotes
	_, err := client.Run(fmt.Sprintf("docker login -u \"%s\" -p \"%s\"", user, pass))
	return err
}

func DockerLogout(client *ssh.Client) error {
	_, err := client.Run("docker logout")
	return err
}

/*
   Pull an image on all the given servers
*/
func DockerPull(clients []*ssh.Client, image string) error {
	for _, client := range clients {
		_, err := client.Run("docker pull " + image)
		if err != nil {
			log.Println(err)
			return err
		}
	}
	return nil
}

/*
   Makes a docker run command to start a node
*/
func dockerRunCmd(subnetID int, resources util.Resources, node int, image string, env map[string]string) (string, error) {
	command := "docker run -itd --entrypoint /bin/sh "
	command += fmt.Sprintf("--network %s%d", conf.NodeNetworkPrefix, node)

	if !resources.NoCPULimits() {
		command += fmt.Sprintf(" --cpus %s", resources.Cpus)
	}

	if !resources.NoMemoryLimits() {
		mem, err := resources.GetMemory()
		if err != nil {
			return "", fmt.Errorf("invalid value for memory")
		}
		command += fmt.Sprintf(" --memory %d", mem)
	}
	for key, value := range env {
		command += fmt.Sprintf(" -e \"%s=%s\"", key, value)
	}
	command += fmt.Sprintf(" --ip %s", util.GetNodeIP(subnetID, node))
	command += fmt.Sprintf(" --hostname %s%d", conf.NodePrefix, node)
	command += fmt.Sprintf(" --name %s%d", conf.NodePrefix, node)
	command += " " + image
	return command, nil
}

/*
   Starts a node
*/
func DockerRun(tn *testnet.TestNet, serverID int, subnetID int, resources util.Resources, node int, image string, env map[string]string) error {
	command, err := dockerRunCmd(subnetID, resources, node, image, env)
	if err != nil {
		log.Println(err)
		return err
	}
	_, err = tn.Clients[serverID].Run(command)
	if err != nil {
		log.Println(err)
		return err
	}
	return nil
}

func serviceDockerRunCmd(network string, ip string, name string, env map[string]string, image string) string {
	envFlags := ""
	for k, v := range env {
		envFlags += fmt.Sprintf("-e \"%s=%s\" ", k, v)
	}
	envFlags += fmt.Sprintf("-e \"BIND_ADDR=%s\"", ip)
	ipFlag := ""
	if len(ip) > 0 {
		ipFlag = fmt.Sprintf("--ip %s", ip)
	}
	return fmt.Sprintf("docker run -itd --network %s %s --hostname %s --name %s %s %s",
		network,
		ipFlag,
		name,
		name,
		envFlags,
		image)
}

/*
   Stop all services and remove the service network from a server
*/
func DockerStopServices(tn *testnet.TestNet) error {
	return helpers.AllServerExecCon(tn, func(client *ssh.Client, _ *db.Server) error {
		_, err := client.Run(fmt.Sprintf("docker rm -f $(docker ps -aq -f name=%s)", conf.ServicePrefix))
		client.Run("docker network rm " + conf.ServiceNetworkName)
		if err != nil {
			log.Println(err)
		}
		return nil
	})
}

/*
   Creates the service network and starts all the services on a server
*/
func DockerStartServices(tn *testnet.TestNet, services []util.Service) error {
	gateway, subnet, err := util.GetServiceNetwork()
	if err != nil {
		log.Println(err)
		return err
	}
	client := tn.GetFlatClients()[0] //TODO make this nice
	_, err = client.KeepTryRun(dockerNetworkCreateCmd(subnet, gateway, -1, conf.ServiceNetworkName))
	if err != nil {
		log.Println(err)
		return err
	}
	ips, err := util.GetServiceIps(services)
	if err != nil {
		log.Println(err)
		return err
	}

	for i, service := range services {
		net := conf.ServiceNetworkName
		ip := ips[service.Name]
		if len(service.Network) != 0 {
			net = service.Network
			ip = ""
		}
		_, err = client.KeepTryRun(serviceDockerRunCmd(net, ip,
			fmt.Sprintf("%s%d", conf.ServicePrefix, i),
			service.Env,
			service.Image))
		if err != nil {
			log.Println(err)
			return err
		}
		tn.BuildState.IncrementDeployProgress()
	}
	return nil
}
