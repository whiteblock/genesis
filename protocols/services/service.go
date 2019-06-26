package services

import (
	"fmt"
	"github.com/whiteblock/genesis/ssh"
	"github.com/whiteblock/genesis/testnet"
	"github.com/whiteblock/genesis/util"
	"net"
)

var conf = util.GetConfig()

// Service represents a service
type Service interface {
	// Prepare prepares the service
	Prepare(client ssh.Client, tn *testnet.TestNet) error

	// GetName gets the name of service
	GetName() string

	// GetImage gets the image of service
	GetImage() string

	// GetEnv gets the environment variables of service
	GetEnv() map[string]string

	// GetNetwork gets the network of the service
	GetNetwork() string

	// GetPorts gets the ports published by the service
	GetPorts() []string

	// GetVolumes gets the volumes mounted on the service container
	GetVolumes() []string

	// GetCommand gets the command to run for the service with Docker.
	GetCommand() string
}

// SimpleService represents a service for a blockchain.
// All env variables will be passed to the container.
type SimpleService struct {
	Name    string            `json:"name"`
	Image   string            `json:"image"`
	Env     map[string]string `json:"env"`
	Network string            `json:"network"`
	Ports   []string          `json:"ports"`
	Volumes []string          `json:"volumes"`
}

// Prepare just returns nil. Simple service has no prepare step
func (s SimpleService) Prepare(client ssh.Client, tn *testnet.TestNet) error {
	return nil
}

// GetName gets the name of service
func (s SimpleService) GetName() string {
	return s.Name
}

// GetImage gets the image of service
func (s SimpleService) GetImage() string {
	return s.Image
}

// GetEnv gets the environment variables of service
func (s SimpleService) GetEnv() map[string]string {
	return s.Env
}

// GetNetwork gets the network of the service
func (s SimpleService) GetNetwork() string {
	return s.Network
}

// GetPorts gets the ports published by the service
func (s SimpleService) GetPorts() []string {
	return s.Ports
}

// GetVolumes gets the volumes mounted on the service container
func (s SimpleService) GetVolumes() []string {
	return s.Volumes
}

// GetCommand gets the command to run for the service with Docker.
func (s SimpleService) GetCommand() string {
	return ""
}

// GetServiceIps creates a map of the service names to their ip addresses. Useful
// for determining the ip address of a service.
func GetServiceIps(services []Service) (map[string]string, error) {
	out := make(map[string]string)
	ip, ipnet, err := net.ParseCIDR(conf.ServiceNetwork)
	ip = ip.Mask(ipnet.Mask)
	if err != nil {
		return nil, util.LogError(err)
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
