/**
 * Copyright 2019 Whiteblock Inc. All rights reserved.
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 */

package config

import (
	"github.com/spf13/viper"
)

// Docker represents the configuration needed to communicate with docker daemons
type Docker struct {
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
	err := v.BindEnv("dockerLogDriver", "DOCKER_LOG_DRIVER")
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

	return nil
}

func setDockerDefaults(v *viper.Viper) {
	v.SetDefault("dockerLogDriver", "journald")
	v.SetDefault("dockerLogLabels", "org,name,testRun,test,phase")
	v.SetDefault("dockerSwarmPort", 2477)
	v.SetDefault("dockerDaemonPort", "2376")
	v.SetDefault("dockerGlusterImage", "gcr.io/whiteblock/gluster:latest")
	v.SetDefault("dockerGlusterDriver", "glusterfs")
}
