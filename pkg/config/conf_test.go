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

package config

import (
	"strconv"
	"testing"

	"github.com/whiteblock/genesis/pkg/command"
	"github.com/whiteblock/genesis/pkg/entity"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestConfig_GetDockerConfig(t *testing.T) {
	var tests = []struct {
		conf               Config
		expectedDockerConf entity.DockerConfig
	}{
		{
			conf: Config{
				DockerCACertPath: "test",
				DockerCertPath:   "test",
				DockerKeyPath:    "test",
				LocalMode:        false,
			},
			expectedDockerConf: entity.DockerConfig{
				CACertPath: "test",
				CertPath:   "test",
				KeyPath:    "test",
				LocalMode:  false,
			},
		},
	}

	for i, tt := range tests {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			assert.Equal(t, tt.expectedDockerConf, tt.conf.GetDockerConfig())
		})
	}
}

func TestConfig_GetVolumeConfig(t *testing.T) {
	var tests = []struct {
		conf               Config
		expectedVolumeConf command.VolumeConfig
	}{
		{
			conf: Config{
				VolumeDriver:     "test",
				VolumeDriverOpts: map[string]string{"test": "test"},
			},
			expectedVolumeConf: command.VolumeConfig{
				Driver:     "test",
				DriverOpts: map[string]string{"test": "test"},
			},
		},
		{
			conf: Config{VolumeDriver: "", VolumeDriverOpts: nil},
			expectedVolumeConf: command.VolumeConfig{
				Driver:     "",
				DriverOpts: nil,
			},
		},
	}

	for i, tt := range tests {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			assert.Equal(t, tt.expectedVolumeConf, tt.conf.GetVolumeConfig())
		})
	}
}

func TestConfig_GetRestConfig(t *testing.T) {
	var tests = []struct {
		conf             Config
		expectedRestConf entity.RestConfig
	}{
		{
			conf: Config{
				Listen: "129.9.9.0:3000",
			},
			expectedRestConf: entity.RestConfig{
				Listen: "129.9.9.0:3000",
			},
		},
	}

	for i, tt := range tests {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			assert.Equal(t, tt.expectedRestConf, tt.conf.GetRestConfig())
		})
	}
}

func TestConfig_GetLogger_Success(t *testing.T) {
	conf := Config{
		Verbosity: "INFO",
	}

	logger, err := conf.GetLogger()
	assert.NotNil(t, logger)
	assert.NoError(t, err)
	assert.True(t, logger.IsLevelEnabled(logrus.InfoLevel))
	assert.False(t, logger.IsLevelEnabled(logrus.DebugLevel))
}

func TestConfig_GetLogger_Failure(t *testing.T) {
	conf := Config{
		Verbosity: "3434ds",
	}

	logger, err := conf.GetLogger()
	assert.Nil(t, logger)
	assert.Error(t, err)
}

func TestConfig_CompletionAMQP(t *testing.T) {
	conf := Config{
		CompletionQueueName: "compl",
	}
	res, _ := conf.CompletionAMQP()
	assert.Equal(t, conf.CompletionQueueName, res.QueueName)
}

func TestConfig_CommandAMQP(t *testing.T) {
	conf := Config{
		CommandQueueName: "comm",
	}
	res, _ := conf.CommandAMQP()
	assert.Equal(t, conf.CommandQueueName, res.QueueName)
}
