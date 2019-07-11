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

//Package docker provides a quick naive interface to Docker calls over ssh
package docker

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/whiteblock/genesis/db"
	"github.com/whiteblock/genesis/protocols/helpers"
	"github.com/whiteblock/genesis/protocols/services"
	"github.com/whiteblock/genesis/ssh"
	"github.com/whiteblock/genesis/testnet"
	"github.com/whiteblock/genesis/util"
	"strings"
)

var conf = util.GetConfig()

// KillNode kills a single node by index on a server
func KillNode(client ssh.Client, node int) error {
	_, err := client.Run(fmt.Sprintf("docker rm -f %s%d", conf.NodePrefix, node))
	return err
}

//Kill kills a node and all of its sidecars
func Kill(client ssh.Client, node int) error {
	_, err := client.Run(fmt.Sprintf("docker rm -f $(docker ps -aq -f name=\"%s%d\")", conf.NodePrefix, node))
	return err
}

// KillAll kills all nodes on a server
func KillAll(client ssh.Client) error {
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
func NetworkDestroy(client ssh.Client, node int) error {
	_, err := client.Run(fmt.Sprintf("docker network rm %s%d", conf.NodeNetworkPrefix, node))
	return err
}

// NetworkDestroyAll removes all whiteblock networks on a node
func NetworkDestroyAll(client ssh.Client) error {
	_, err := client.Run(fmt.Sprintf(
		"for net in $(docker network ls | grep %s | awk '{print $1}'); do docker network rm $net; done", conf.NodeNetworkPrefix))
	return err
}

// Login is an abstraction of docker login
func Login(client ssh.Client, username string, password string) error {
	user := strings.Replace(username, "\"", "\\\"", -1) //Escape the quotes
	pass := strings.Replace(password, "\"", "\\\"", -1) //Escape the quotes
	_, err := client.Run(fmt.Sprintf("docker login -u \"%s\" -p \"%s\"", user, pass))
	return err
}

// Logout is an abstraction of docker logout
func Logout(client ssh.Client) error {
	_, err := client.Run("docker logout")
	return err
}

// Pull pulls an image on all the given servers
func Pull(clients []ssh.Client, image string) error {
	for _, client := range clients {
		_, err := client.Run("docker pull " + image)
		if err != nil {
			return util.LogError(err)
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

	if c.GetResources().Volumes != nil && conf.EnableDockerVolumes {
		for _, volume := range c.GetResources().Volumes {
			command += fmt.Sprintf(" -v %s", volume)
		}
	}

	if conf.EnablePortForwarding {
		ports := c.GetPorts()
		for _, port := range ports {
			command += fmt.Sprintf(" -p %s", port)
		}
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
		return "", util.LogError(err)
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
		return util.LogError(err)
	}
	_, err = tn.Clients[serverID].Run(command)
	if err != nil {
		return util.LogError(err)
	}
	return nil
}

func serviceDockerRunCmd(network string, ip string, name string, env map[string]string, volumes []string, ports []string, image string, cmd string) string {
	envFlags := ""
	for k, v := range env {
		envFlags += fmt.Sprintf("-e \"%s=%s\" ", k, v)
	}
	envFlags += fmt.Sprintf("-e \"BIND_ADDR=%s\"", ip)
	ipFlag := ""
	if len(ip) > 0 {
		ipFlag = fmt.Sprintf("--ip %s", ip)
	}
	volumestr := ""
	if conf.EnableDockerVolumes {
		for _, vol := range volumes {
			volumestr += fmt.Sprintf("-v %s ", vol)
		}
	}

	portstr := ""
	if conf.EnablePortForwarding {
		for _, port := range ports {
			portstr += fmt.Sprintf("-p %s ", port)
		}
	}

	return fmt.Sprintf("docker run -itd --network %s %s --hostname %s --name %s %s %s %s %s %s",
		network,
		ipFlag,
		name,
		name,
		envFlags,
		volumestr,
		portstr,
		image,
		cmd)
}

// StopServices stops all services and remove the service network from a server
func StopServices(tn *testnet.TestNet) error {
	return helpers.AllServerExecCon(tn, func(client ssh.Client, _ *db.Server) error {
		_, err := client.Run(fmt.Sprintf("docker rm -f $(docker ps -aq -f name=%s)", conf.ServicePrefix))
		if err != nil {
			log.WithFields(log.Fields{"error": err}).Info("no service containers to remove")
		}

		_, err = client.Run("docker network rm " + conf.ServiceNetworkName)
		if err != nil {
			log.WithFields(log.Fields{"error": err}).Info("no service network to remove")
		}

		return nil
	})
}

// StartServices creates the service network and starts all the services on a server
func StartServices(tn *testnet.TestNet, servs []services.Service) error {
	gateway, subnet, err := util.GetServiceNetwork()
	if err != nil {
		return util.LogError(err)
	}
	client := tn.GetFlatClients()[0] //TODO make this nice
	_, err = client.KeepTryRun(dockerNetworkCreateCmd(subnet, gateway, -1, conf.ServiceNetworkName))
	if err != nil {
		return util.LogError(err)
	}
	ips, err := services.GetServiceIps(servs)
	if err != nil {
		return util.LogError(err)
	}

	for i, service := range servs {
		net := conf.ServiceNetworkName
		ip := ips[service.GetName()]
		if len(service.GetNetwork()) != 0 {
			net = service.GetNetwork()
			ip = ""
		}
		err = service.Prepare(client, tn)
		if err != nil {
			return util.LogError(err)
		}
		_, err = client.KeepTryRun(serviceDockerRunCmd(net, ip,
			fmt.Sprintf("%s%d", conf.ServicePrefix, i),
			service.GetEnv(),
			service.GetVolumes(),
			service.GetPorts(),
			service.GetImage(),
			service.GetCommand()))
		if err != nil {
			return util.LogError(err)
		}
		tn.BuildState.IncrementDeployProgress()
	}
	return nil
}
