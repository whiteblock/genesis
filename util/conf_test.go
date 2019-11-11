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

package util

import (
	"fmt"
	"reflect"
	"strconv"
	"testing"

	"github.com/whiteblock/genesis/pkg/entity"
)

func TestConfig_GetQueueConfig(t *testing.T) {
	var tests = []struct {
		conf 				Config
		expectedQueueConf 	entity.QueueConfig
	}{
		{
			conf: Config{
				QueueDurable: false,
				QueueAutoDelete: true,
				QueueExclusive: false,
				QueueNoWait: false,
				QueueArgs: map[string]interface{}{"test": true, "arguments": "test"},
			},
			expectedQueueConf: entity.QueueConfig{
				Durable: false,
				AutoDelete: true,
				Exclusive: false,
				NoWait: false,
				Args: map[string]interface{}{"test": true, "arguments": "test"},
			},
		},
		{
			conf: *GetConfig(),
			expectedQueueConf: entity.QueueConfig{
				Durable: true,
				AutoDelete: false,
				Exclusive: false,
				NoWait: false,
				Args: *new(map[string]interface{}),
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
		expectedConsumeConf entity.ConsumeConfig
	}{
		{
			conf: Config{
				Consumer: "test",
				ConsumerAutoAck: false,
				ConsumerExclusive: true,
				ConsumerNoLocal: false,
				ConsumerNoWait: true,
				ConsumerArgs: map[string]interface{}{"test": 4},
			},
			expectedConsumeConf: entity.ConsumeConfig{
				Consumer: "test",
				AutoAck: false,
				Exclusive: true,
				NoLocal: false,
				NoWait: true,
				Args: map[string]interface{}{"test": 4},
			},
		},
		{
			conf: *GetConfig(),
			expectedConsumeConf: entity.ConsumeConfig{
				Consumer: "",
				AutoAck: false,
				Exclusive: false,
				NoLocal: false,
				NoWait: false,
				Args: *new(map[string]interface{}),
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
		expectedPubConfig entity.PublishConfig
	}{
		{
			conf: Config{
				PublishMandatory: true,
				PublishImmediate: false,
			},
			expectedPubConfig: entity.PublishConfig{
				Mandatory: true,
				Immediate: false,
			},
		},
		{
			conf: *GetConfig(),
			expectedPubConfig: entity.PublishConfig{
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
		conf              Config
		expectedAMQPConf entity.AMQPConfig
	}{
		{
			conf: Config{
				AMQPQueueName: "test",
				QueueDurable: false,
				QueueAutoDelete: true,
				QueueExclusive: false,
				QueueNoWait: false,
				QueueArgs: map[string]interface{}{"test": true, "arguments": "test"},
				Consumer: "test",
				ConsumerAutoAck: false,
				ConsumerExclusive: true,
				ConsumerNoLocal: false,
				ConsumerNoWait: true,
				ConsumerArgs: map[string]interface{}{"test": 4},
				PublishMandatory: true,
				PublishImmediate: false,
			},
			expectedAMQPConf: entity.AMQPConfig{
				QueueName: "test",
				Queue: entity.QueueConfig{
					Durable: false,
					AutoDelete: true,
					Exclusive: false,
					NoWait: false,
					Args: map[string]interface{}{"test": true, "arguments": "test"},
				},
				Consume: entity.ConsumeConfig{
					Consumer: "test",
					AutoAck: false,
					Exclusive: true,
					NoLocal: false,
					NoWait: true,
					Args: map[string]interface{}{"test": 4},
				},
				Publish: entity.PublishConfig{
					Mandatory: true,
					Immediate: false,
				},
			},
		},
		{
			conf: *GetConfig(),
			expectedAMQPConf: entity.AMQPConfig{
				QueueName: "",
				Queue: entity.QueueConfig{
					Durable: true,
					AutoDelete: false,
					Exclusive: false,
					NoWait: false,
					Args: *new(map[string]interface{}),
				},
				Consume: entity.ConsumeConfig{
					Consumer: "",
					AutoAck: false,
					Exclusive: false,
					NoLocal: false,
					NoWait: false,
					Args: *new(map[string]interface{}),
				},
				Publish: entity.PublishConfig{
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
		conf              Config
		expectedDockerConf entity.DockerConfig
	}{
		{
			conf: Config{
				DockerCACertPath: "test",
				DockerCertPath:"test",
				DockerKeyPath: "test",
			},
			expectedDockerConf: entity.DockerConfig{
				CACertPath: "test",
				CertPath: "test",
				KeyPath: "test",
			},
		},
		{
			conf: *GetConfig(),
			expectedDockerConf: entity.DockerConfig{
				CACertPath: "",
				CertPath: "",
				KeyPath: "",
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
		expectedVolumeConf entity.VolumeConfig
	}{
		{
			conf: Config{
				VolumeDriver: "test",
				VoluemDriverOpts: map[string]string{"test": "test"},
			},
			expectedVolumeConf: entity.VolumeConfig{
				Driver: "test",
				DriverOpts: map[string]string{"test": "test"},
			},
		},
		{
			conf: *GetConfig(),
			expectedVolumeConf: entity.VolumeConfig{
				Driver: "",
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
				Listen: "0.0.0.0:8000", //todo is this meant to happen?
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
				fmt.Println(tt.conf.GetRestConfig())
				fmt.Println(tt.expectedRestConf)
				t.Error("return value of GetRestConfig does not match expected value")
			}
		})
	}
}

func TestGetConfig(t *testing.T) {
	var tests = []struct {
		expectedConf	*Config
	}{
		{
			expectedConf: new(Config),
		},
		{
			expectedConf: &Config{
				QueueDurable: true,
				QueueAutoDelete: false,
				QueueExclusive: false,
				QueueNoWait: false,
				QueueArgs: *new(map[string]interface{}),
				Consumer: "",
				ConsumerAutoAck: false,
				ConsumerExclusive: false,
				ConsumerNoLocal: false,
				ConsumerNoWait: false,
				ConsumerArgs: *new(map[string]interface{}),
				PublishMandatory: false,
				PublishImmediate: false,
				AMQPQueueName: "",
				DockerCACertPath: "",
				DockerCertPath: "",
				DockerKeyPath: "",
				VolumeDriver: "",
				VoluemDriverOpts: *new(map[string]string),
				Verbosity: "INFO",
				Listen: "0.0.0.0:8000",
			},
		},
	}

	for i, tt := range tests {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			if !reflect.DeepEqual(GetConfig(), tt.expectedConf) {
				fmt.Println(GetConfig())
				fmt.Println()
				fmt.Println(tt.expectedConf)
				t.Error("return value of GetConfig does not match expected value")
			}
		})
	}
}