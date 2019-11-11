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
	"github.com/docker/docker/api/types/network"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"github.com/whiteblock/genesis/pkg/entity"
)

// Config groups all of the global configuration parameters into
// a single struct
type Config struct {
	QueueDurable         bool                                 `mapstructure:"queueDurable"`
	QueueAutoDelete      bool                                 `mapstructure:"queueAutoDelete"`
	QueueExclusive       bool                                 `mapstructure:"queueExclusive"`
	QueueNoWait          bool                                 `mapstructure:"queueNoWait"`
	QueueArgs            bool                                 `mapstructure:"queueArgs"`
	Consumer             string                               `mapstructure:"consumer"`
	ConsumerAutoAck      bool                                 `mapstructure:"consumerAutoAck"`
	ConsumerExclusive    bool                                 `mapstructure:"consumerExclusive"`
	ConsumerNoLocal      bool                                 `mapstructure:"consumerNoLocal"`
	PublishMandatory     bool                                 `mapstructure:"publishMandatory"`
	PublishImmediate     bool                                 `mapstructure:"publishImmediate"`
	AMQPQueueName        string                               `mapstructure:"amqpQueueName"`
	AMQPQueue            entity.QueueConfig                   `mapstructure:"amqpQueue"`
	AMQPConsume          entity.ConsumeConfig                 `mapstructure:"amqpConsume"`
	AMQPPublish          entity.PublishConfig                 `mapstructure:"amqpPublish"`
	NetworkEndpoints     map[string]*network.EndpointSettings `mapstructure:"networkEndpoints"`
	BoundCPUs            []int                                `mapstructure:"boundCPUs"`
	ContainerDetach      bool                                 `mapstructure:"containerDetach"`
	ContainerEntryPoint  string                               `mapstructure:"containerEntryPoint"`
	ContainerEnvironment map[string]string                    `mapstructure:"containerEnvironment"`
	ContainerLabels      map[string]string                    `mapstructure:"containerLabels"`
	ContainerName        string                               `mapstructure:"containerName"`
	ContainerNetwork     string                               `mapstructure:"containerNetwork"`
	ContainerNetworkConf entity.NetworkConfig                 `mapstructure:"containerNetworkConf"`
	ContainerPorts       map[int]int                          `mapstructure:"containerPorts"`
	ContainerVolumes     map[string]entity.Volume             `mapstructure:"containerVolume"`
	ContainerResources   entity.Resources                     `mapstructure:"containerResources"`
	ContainerImage       string                               `mapstructure:"containerImage"`
	ContainerArgs        []string                             `mapstructure:"containerArgs"`
	DockerCACertPath     string                               `mapstructure:"dockerCACertPath"`
	DockerCertPath       string                               `mapstructure:"dockerCertPath"`
	DockerKeyPath        string                               `mapstructure:"dockerKeyPath"`
	FilePath             string                               `mapstructure:"filePath"`
	FileData             string                               `mapstructure:"fileData"`
	NetconfLimit         int                                  `mapstructure:"netconfLimit"`
	NetconfLoss          float64                              `mapstructure:"netconfLoss"`
	NetconfDelay         int                                  `mapstructure:"netconfDelay"`
	NetconfRate				string	`mapstructure:"netconfRate"`
	NetconfDuplication	float64	`mapstructure:"netconfDuplication"`
	NetconfCorrupt	float64	`mapstructure:"netconfCorrupt"`
	NetconfReorder	float64	`mapstructure:"netconfReorder"`
	NetworkName	string	`mapstructure:"networkName"`
	ResourcesCpus	string	`mapstructure:"resourcesCpus"`
	ResourcesMemory	string	`mapstructure:"resourcesMemory"`
	VolumeDriver	string	`mapstructure:"volumeDriver"`
	VoluemDriverOpts	map[string]string	`mapstructure:"volumeDriverOpts"`
	VolumeName		string	`mapstructure:"volumeName"`
	VolumeLabels	map[string]string	`mapstructure:"volumeLabels"`
	//todo should we set a log level?
}

//NodesPerCluster represents the maximum number of nodes allowed in a cluster
var NodesPerCluster uint32

var conf = new(Config)

func setViperEnvBindings() {
	// todo?
}
func setViperDefaults() {
	// todo?
}

// GCPFormatter enables the ability to use genesis logging with Stackdriver
type GCPFormatter struct { //todo: does this stay the same?
	JSON           *log.JSONFormatter
	ConstantFields log.Fields
}

// Format takes in the entry and processes it into the appropiate log entry
func (gf GCPFormatter) Format(entry *log.Entry) ([]byte, error) {
	for k, v := range gf.ConstantFields {
		entry.Data[k] = v
	}
	return gf.JSON.Format(entry)
}

func init() {
	setViperDefaults()
	setViperEnvBindings()
	viper.AddConfigPath("/etc/whiteblock/")          // path to look for the config file in
	viper.AddConfigPath("$HOME/.config/whiteblock/") // call multiple times to add many search paths
	viper.SetConfigName("genesis")
	viper.SetConfigType("yaml")
	err := viper.ReadInConfig()

	if err != nil {
		log.WithFields(log.Fields{"error": err}).Warn("could not find the config file")
	}
	err = viper.Unmarshal(&conf)
	if err != nil {
		log.Fatalf("unable to decode into struct, %v", err)
	}

	//todo what else?
}

// GetConfig gets a pointer to the global config object.
// Do not modify conf object
func GetConfig() *Config {
	return conf
}
