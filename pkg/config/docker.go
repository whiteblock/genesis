/*
	Copyright 2019 Whiteblock Inc.
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

package config

import (
	"github.com/spf13/viper"
)

// Docker represents the configuration needed to communicate with docker daemons
type Docker struct {
	// CACertPath is the filepath to the CA Certificate
	CACertPath string `mapstructure:"dockerCACertPath"`
	// CertPath is the filepath to the Certificate for TLS
	CertPath string `mapstructure:"dockerCertPath"`
	// KeyPath is the filepath to the private key for TLS
	KeyPath string `mapstructure:"dockerKeyPath"`
	// LocalMode causes the TLS parameters to be ignored and Genesis
	// to assume that the docker daemon is on the local machine
	LocalMode bool `mapstructure:"localMode"`

	LogDriver string `mapstructure:"dockerLogDriver"`

	LogLabels string `mapstructure:"dockerLogLabels"`

	// SwarmPort is the docker swarm port
	SwarmPort int `mapstructure:"dockerSwarmPort"`
	// DaemonPort is the port docker daemon is listening on
	DaemonPort string `mapstructure:"dockerDaemonPort"`

	GlusterImage string `mapstructure:"dockerGlusterImage"`

	GlusterDriver string `mapstructure:"dockerGlusterDriver"`
}

// NewDocker creates a new docker configuration from viper
func NewDocker(v *viper.Viper) (out Docker, err error) {
	return out, v.Unmarshal(&out)
}

func setDockerBindings(v *viper.Viper) error {
	err := v.BindEnv("dockerCACertPath", "DOCKER_CACERT_PATH")
	if err != nil {
		return err
	}

	err = v.BindEnv("dockerCertPath", "DOCKER_CERT_PATH")
	if err != nil {
		return err
	}

	err = v.BindEnv("dockerLogDriver", "DOCKER_LOG_DRIVER")
	if err != nil {
		return err
	}

	err = v.BindEnv("dockerLogLabels", "DOCKER_LOG_DRIVER")
	if err != nil {
		return err
	}

	err = v.BindEnv("dockerDaemonPort", "DOCKER_DAEMON_PORT")
	if err != nil {
		return err
	}

	err = v.BindEnv("dockerSwarmPort", "DOCKER_SWARM_PORT")
	if err != nil {
		return err
	}

	err = v.BindEnv("dockerGlusterImage", "DOCKER_GLUSTER_IMAGE")
	if err != nil {
		return err
	}

	err = v.BindEnv("dockerGlusterDriver", "DOCKER_GLUSTER_DRIVER")
	if err != nil {
		return err
	}

	return v.BindEnv("dockerKeyPath", "DOCKER_KEY_PATH")
}

func setDockerDefaults(v *viper.Viper) {
	v.SetDefault("dockerLogDriver", "journald")
	v.SetDefault("dockerLogLabels", "org,name,testRun,test,phase")
	v.SetDefault("dockerSwarmPort", 2477)
	v.SetDefault("dockerDaemonPort", "2376")
	v.SetDefault("dockerGlusterImage", "gcr.io/infra-dev-249211/gluster")
	v.SetDefault("dockerGlusterDriver", "glusterfs")
}
