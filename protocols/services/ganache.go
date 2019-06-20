package services

import (
	"github.com/whiteblock/genesis/ssh"
	"github.com/whiteblock/genesis/testnet"
	"strconv"
)

// GanacheService represents the Ganache service
type GanacheService struct {
	SimpleService
}

// Prepare prepares the ganache service
func (p GanacheService) Prepare(client ssh.Client, tn *testnet.TestNet) error {
	return nil
}

func (p GanacheService) GetCommand() string {
	return conf.GanacheCLIOptions
}

// RegisterGanache exposes a Ganache service on the testnet.
func RegisterGanache() Service {
	return GanacheService{
		SimpleService{
			Name:    "ganache",
			Image:   "trufflesuite/ganache-cli",
			Env:     map[string]string{},
			Ports:   []string{strconv.Itoa(conf.GanacheRPCPort) + ":8545"},
			Volumes: []string{},
		},
	}
}
