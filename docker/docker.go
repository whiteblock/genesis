//Package docker provides a quick naive interface to Docker calls over ssh
package docker

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

var conf *util.Config

func init() {
	conf = util.GetConfig()
}

// DockerKill kills a single node by index on a server
func Kill(client *ssh.Client, node int) error {
	_, err := client.Run(fmt.Sprintf("docker rm -f %s%d", conf.NodePrefix, node))
	return err
}

// DockerKillAll kills all nodes on a server
func KillAll(client *ssh.Client) error {
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

// NetworkCreate creates a docker network for a node
func NetworkCreate(tn *testnet.TestNet, serverID int, subnetID int, node int) error {
	command := dockerNetworkCreateCmd(
		util.GetNetworkAddress(subnetID, node),
		util.GetGateway(subnetID, node),
		node,
		fmt.Sprintf("%s%d", conf.NodeNetworkPrefix, node))

	_, err := tn.Clients[serverID].KeepTryRun(command)

	return err
}

// NetworkDestroy tears down a single docker network
func NetworkDestroy(client *ssh.Client, node int) error {
	_, err := client.Run(fmt.Sprintf("docker network rm %s%d", conf.NodeNetworkPrefix, node))
	return err
}

// NetworkDestroyAll removes all whiteblock networks on a node
func NetworkDestroyAll(client *ssh.Client) error {
	_, err := client.Run(fmt.Sprintf(
		"for net in $(docker network ls | grep %s | awk '{print $1}'); do docker network rm $net; done", conf.NodeNetworkPrefix))
	return err
}

// Login is an abstraction of docker login
func Login(client *ssh.Client, username string, password string) error {
	user := strings.Replace(username, "\"", "\\\"", -1) //Escape the quotes
	pass := strings.Replace(password, "\"", "\\\"", -1) //Escape the quotes
	_, err := client.Run(fmt.Sprintf("docker login -u \"%s\" -p \"%s\"", user, pass))
	return err
}

// Logout is an abstraction of docker logout
func Logout(client *ssh.Client) error {
	_, err := client.Run("docker logout")
	return err
}

// Pull pulls an image on all the given servers
func Pull(clients []*ssh.Client, image string) error {
	for _, client := range clients {
		_, err := client.Run("docker pull " + image)
		if err != nil {
			log.Println(err)
			return err
		}
	}
	return nil
}

// dockerRunCmd makes a docker run command to start a node
func dockerRunCmd(c Container) (string, error) {
	command := "docker run -itd --entrypoint /bin/sh "
	command += fmt.Sprintf("--network %s", c.GetNetworkName())

	if !c.GetResources().NoCPULimits() {
		command += fmt.Sprintf(" --cpus %s", c.GetResources().Cpus)
	}

	if !c.GetResources().NoMemoryLimits() {
		mem, err := c.GetResources().GetMemory()
		if err != nil {
			return "", fmt.Errorf("invalid value for memory")
		}
		command += fmt.Sprintf(" --memory %d", mem)
	}
	for key, value := range c.GetEnvironment() {
		command += fmt.Sprintf(" -e \"%s=%s\"", key, value)
	}
	ip, err := c.GetIP()
	if err != nil {
		log.Println(err)
		return "", err
	}
	command += fmt.Sprintf(" --ip %s", ip)
	command += fmt.Sprintf(" --hostname %s", c.GetName())
	command += fmt.Sprintf(" --name %s", c.GetName())
	command += " " + c.GetImage()
	return command, nil
}

// Run starts a node
func Run(tn *testnet.TestNet, serverID int, container Container) error {
	command, err := dockerRunCmd(container)
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

// DockerStopServices stops all services and remove the service network from a server
func StopServices(tn *testnet.TestNet) error {
	return helpers.AllServerExecCon(tn, func(client *ssh.Client, _ *db.Server) error {
		_, err := client.Run(fmt.Sprintf("docker rm -f $(docker ps -aq -f name=%s)", conf.ServicePrefix))
		client.Run("docker network rm " + conf.ServiceNetworkName)
		if err != nil {
			log.Println(err)
		}
		return nil
	})
}

// DockerStartServices creates the service network and starts all the services on a server
func StartServices(tn *testnet.TestNet, services []util.Service) error {
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
