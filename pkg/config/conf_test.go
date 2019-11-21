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

func TestConfig_GetQueueConfig(t *testing.T) {
	var tests = []struct {
		conf              Config
		expectedQueueConf Queue
	}{
		{
			conf: Config{
				QueueDurable:    false,
				QueueAutoDelete: true,
				QueueExclusive:  false,
				QueueNoWait:     false,
				QueueArgs:       map[string]interface{}{"test": true, "arguments": "test"},
			},
			expectedQueueConf: Queue{
				Durable:    false,
				AutoDelete: true,
				Exclusive:  false,
				NoWait:     false,
				Args:       map[string]interface{}{"test": true, "arguments": "test"},
			},
		},
		{
			conf: *GetConfig(),
			expectedQueueConf: Queue{
				Durable:    true,
				AutoDelete: false,
				Exclusive:  false,
				NoWait:     false,
				Args:       *new(map[string]interface{}),
			},
		},
	}

	for i, tt := range tests {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			if !reflect.DeepEqual(tt.conf.GetQueueConfig(), tt.expectedQueueConf) {
				t.Error("return value of GetQueueConfig does not match expected value")
			}
		})
	}
}

func TestConfig_GetConsumeConfig(t *testing.T) {
	var tests = []struct {
		conf                Config
		expectedConsumeConf Consume
	}{
		{
			conf: Config{
				Consumer:          "test",
				ConsumerAutoAck:   false,
				ConsumerExclusive: true,
				ConsumerNoLocal:   false,
				ConsumerNoWait:    true,
				ConsumerArgs:      map[string]interface{}{"test": 4},
			},
			expectedConsumeConf: Consume{
				Consumer:  "test",
				AutoAck:   false,
				Exclusive: true,
				NoLocal:   false,
				NoWait:    true,
				Args:      map[string]interface{}{"test": 4},
			},
		},
		{
			conf: *GetConfig(),
			expectedConsumeConf: Consume{
				Consumer:  "",
				AutoAck:   false,
				Exclusive: false,
				NoLocal:   false,
				NoWait:    false,
				Args:      *new(map[string]interface{}),
			},
		},
	}

	for i, tt := range tests {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			if !reflect.DeepEqual(tt.conf.GetConsumeConfig(), tt.expectedConsumeConf) {
				t.Error("return value of GetConsumeConfig does not match expected value")
			}
		})
	}
}

func TestConfig_GetPublishConfig(t *testing.T) {
	var tests = []struct {
		conf              Config
		expectedPubConfig Publish
	}{
		{
			conf: Config{
				PublishMandatory: true,
				PublishImmediate: false,
			},
			expectedPubConfig: Publish{
				Mandatory: true,
				Immediate: false,
			},
		},
		{
			conf: *GetConfig(),
			expectedPubConfig: Publish{
				Mandatory: false,
				Immediate: false,
			},
		},
	}

	for i, tt := range tests {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			if !reflect.DeepEqual(tt.conf.GetPublishConfig(), tt.expectedPubConfig) {
				t.Error("return value of GetPublishConfig does not match expected value")
			}
		})
	}
}

func TestConfig_GetAMQPConfig(t *testing.T) {
	var tests = []struct {
		conf             Config
		expectedAMQPConf AMQP
	}{
		{
			conf: Config{
				AMQPQueueName:     "test",
				QueueDurable:      false,
				QueueAutoDelete:   true,
				QueueExclusive:    false,
				QueueNoWait:       false,
				QueueArgs:         map[string]interface{}{"test": true, "arguments": "test"},
				Consumer:          "test",
				ConsumerAutoAck:   false,
				ConsumerExclusive: true,
				ConsumerNoLocal:   false,
				ConsumerNoWait:    true,
				ConsumerArgs:      map[string]interface{}{"test": 4},
				PublishMandatory:  true,
				PublishImmediate:  false,
			},
			expectedAMQPConf: AMQP{
				QueueName: "test",
				Queue: Queue{
					Durable:    false,
					AutoDelete: true,
					Exclusive:  false,
					NoWait:     false,
					Args:       map[string]interface{}{"test": true, "arguments": "test"},
				},
				Consume: Consume{
					Consumer:  "test",
					AutoAck:   false,
					Exclusive: true,
					NoLocal:   false,
					NoWait:    true,
					Args:      map[string]interface{}{"test": 4},
				},
				Publish: Publish{
					Mandatory: true,
					Immediate: false,
				},
			},
		},
		{
			conf: *GetConfig(),
			expectedAMQPConf: AMQP{
				QueueName: "",
				Queue: Queue{
					Durable:    true,
					AutoDelete: false,
					Exclusive:  false,
					NoWait:     false,
					Args:       *new(map[string]interface{}),
				},
				Consume: Consume{
					Consumer:  "",
					AutoAck:   false,
					Exclusive: false,
					NoLocal:   false,
					NoWait:    false,
					Args:      *new(map[string]interface{}),
				},
				Publish: Publish{
					Mandatory: false,
					Immediate: false,
				},
			},
		},
	}

	for i, tt := range tests {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			if !reflect.DeepEqual(tt.conf.GetAMQPConfig(), tt.expectedAMQPConf) {
				t.Error("return value of GetAMQPConfig does not match expected value")
			}
		})
	}
}

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
				VoluemDriverOpts: map[string]string{"test": "test"},
			},
			expectedVolumeConf: command.VolumeConfig{
				Driver:     "test",
				DriverOpts: map[string]string{"test": "test"},
			},
		},
		{
			conf: *GetConfig(),
			expectedVolumeConf: command.VolumeConfig{
				Driver:     "",
				DriverOpts: *new(map[string]string),
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
		{
			conf: *GetConfig(),
			expectedRestConf: entity.RestConfig{
				Listen: "0.0.0.0:8000",
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
