package helpers

import (
	"fmt"
	"github.com/whiteblock/genesis/ssh"
	"github.com/whiteblock/genesis/testnet"
	"github.com/whiteblock/genesis/util"
	"log"
	"net"
)


type Service interface {
// Prepare the service
Prepare(client *ssh.Client, tn *testnet.TestNet) error

// name of service
GetName() string

// image of service
GetImage() string

// environment variables of service
GetEnv() map[string]string

// the network of the service
GetNetwork() string

// the ports published by the service
GetPorts() []string

// the volumes mounted on the service container
GetVolumes() []string

}

// Service represents a service for a blockchain.
// All env variables will be passed to the container.
type SimpleService struct {
Name    string            `json:"name"`
Image   string            `json:"image"`
Env     map[string]string `json:"env"`
Network string            `json:"network"`
Ports   []string          `json:"ports"`
Volumes []string          `json:"volumes"`
}

// Simple service has no prepare step
func (s SimpleService) Prepare(client *ssh.Client, tn *testnet.TestNet) error {
  return nil
}

func (s SimpleService) GetName() string {
return s.Name
}

func (s SimpleService) GetImage() string {
return s.Image
}

func (s SimpleService) GetEnv() map[string]string {
return s.Env
}

func (s SimpleService) GetNetwork() string {
return s.Network
}

func (s SimpleService) GetPorts() []string {
return s.Ports
}

func (s SimpleService) GetVolumes() []string {
return s.Volumes
}

// GetServiceIps creates a map of the service names to their ip addresses. Useful
// for determining the ip address of a service.
func GetServiceIps(services []Service) (map[string]string, error) {
	out := make(map[string]string)
	ip, ipnet, err := net.ParseCIDR(conf.ServiceNetwork)
	ip = ip.Mask(ipnet.Mask)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	util.Inc(ip) //skip first ip

	for _, service := range services {
		util.Inc(ip)
		if !ipnet.Contains(ip) {
			return nil, fmt.Errorf("CIDR range too small")
		}
		out[service.GetName()] = ip.String()
	}
	return out, nil
}