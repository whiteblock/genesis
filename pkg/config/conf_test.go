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
	"github.com/whiteblock/genesis/pkg/command"
	"reflect"
	"strconv"
	"testing"

	"github.com/whiteblock/genesis/pkg/entity"
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
			if !reflect.DeepEqual(tt.conf.GetDockerConfig(), tt.expectedDockerConf) {
				t.Error("return value of GetDockerConfig does not match expected value")
			}
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
			if !reflect.DeepEqual(tt.conf.GetVolumeConfig(), tt.expectedVolumeConf) {
				t.Error("return value of GetVolumeConfig does not match expected value")
			}
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
				Listen: "129.9.9.0:3000", //todo is this meant to happen?
			},
		},
	}

	for i, tt := range tests {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			if !reflect.DeepEqual(tt.conf.GetRestConfig(), tt.expectedRestConf) {
				t.Error("return value of GetRestConfig does not match expected value")
			}
		})
	}
}
