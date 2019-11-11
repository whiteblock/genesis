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
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"github.com/streadway/amqp"
	"github.com/whiteblock/genesis/pkg/entity"
)

// Config groups all of the global configuration parameters into
// a single struct
type Config struct {
	QueueDurable      bool                                 `mapstructure:"queueDurable"`
	QueueAutoDelete   bool                                 `mapstructure:"queueAutoDelete"`
	QueueExclusive    bool                                 `mapstructure:"queueExclusive"`
	QueueNoWait       bool                                 `mapstructure:"queueNoWait"`
	QueueArgs         amqp.Table                           `mapstructure:"queueArgs"`
	Consumer          string                               `mapstructure:"consumer"`
	ConsumerAutoAck   bool                                 `mapstructure:"consumerAutoAck"`
	ConsumerExclusive bool                                 `mapstructure:"consumerExclusive"`
	ConsumerNoLocal   bool                                 `mapstructure:"consumerNoLocal"`
	ConsumerNoWait    bool                                 `mapstructure:"consumerNoWait"`
	ConsumerArgs      amqp.Table                           `mapstructure:"consumerArgs"`
	PublishMandatory  bool                                 `mapstructure:"publishMandatory"`
	PublishImmediate  bool                                 `mapstructure:"publishImmediate"`
	AMQPQueueName     string                               `mapstructure:"amqpQueueName"`
	DockerCACertPath  string                               `mapstructure:"dockerCACertPath"`
	DockerCertPath    string                               `mapstructure:"dockerCertPath"`
	DockerKeyPath     string                               `mapstructure:"dockerKeyPath"`
	VolumeDriver      string                               `mapstructure:"volumeDriver"`
	VoluemDriverOpts  map[string]string                    `mapstructure:"volumeDriverOpts"`
	Verbosity         string                               `mapstructure:"verbosity"`
}

func (c Config) GetQueueConfig() entity.QueueConfig {
	return entity.QueueConfig{
		Durable:    c.QueueDurable,
		AutoDelete: c.QueueAutoDelete,
		Exclusive:  c.QueueExclusive,
		NoWait:     c.QueueNoWait,
		Args:       c.QueueArgs,
	}
}

func (c Config) GetConsumeConfig() entity.ConsumeConfig {
	return entity.ConsumeConfig{
		Consumer:  c.Consumer,
		AutoAck:   c.ConsumerAutoAck,
		Exclusive: c.ConsumerExclusive,
		NoLocal:   c.ConsumerNoLocal,
		NoWait:    c.ConsumerNoWait,
		Args:      c.ConsumerArgs,
	}
}

func (c Config) GetPublishConfig() entity.PublishConfig {
	return entity.PublishConfig{
		Mandatory: c.PublishMandatory,
		Immediate: c.PublishImmediate,
	}
}

func (c Config) GetAMQPConfig() entity.AMQPConfig {
	return entity.AMQPConfig{
		QueueName: c.AMQPQueueName,
		Queue:     c.GetQueueConfig(),
		Consume:   c.GetConsumeConfig(),
		Publish:   c.GetPublishConfig(),
	}
}

func (c Config) GetDockerConfig() entity.DockerConfig {
	return entity.DockerConfig{
		CACertPath: c.DockerCACertPath,
		CertPath:   c.DockerCertPath,
		KeyPath:    c.DockerKeyPath,
	}
}

func (c Config) GetVolumeConfig() entity.VolumeConfig {
	return entity.VolumeConfig{
		Driver:     c.VolumeDriver,
		DriverOpts: c.VoluemDriverOpts,
	}
}

//NodesPerCluster represents the maximum number of nodes allowed in a cluster
var NodesPerCluster uint32

var conf = new(Config)

func setViperEnvBindings() {
	viper.BindEnv("queueDurable", "QUEUE_DURABLE")
	viper.BindEnv("queueAutoDelete", "QUEUE_AUTO_DELETE")
	viper.BindEnv("queueExclusive", "QUEUE_EXCLUSIVE")
	viper.BindEnv("queueNoWait", "QUEUE_NO_WAIT")
	viper.BindEnv("queueArgs", "QUEUE_ARGS")
	viper.BindEnv("consumer", "CONSUMER")
	viper.BindEnv("consumerAutoAck", "CONSUMER_AUTO_ACK")
	viper.BindEnv("consumerExclusive", "CONSUMER_EXCLUSIVE")
	viper.BindEnv("consumerNoLocal", "CONSUMER_NO_LOCAL")
	viper.BindEnv("consumerNoWait", "CONSUMER_NO_WAIT")
	viper.BindEnv("consumerArgs", "CONSUMER_ARGS")
	viper.BindEnv("publishMandatory", "PUBLISH_MANDATORY")
	viper.BindEnv("publishImmediate", "PUBLISH_IMMEDIATE")
	viper.BindEnv("amqpQueueName", "AMQP_QUEUE_NAME")
	viper.BindEnv("dockerCACertPath", "DOCKER_CACERT_PATH")
	viper.BindEnv("dockerCertPath", "DOCKER_CERT_PATH")
	viper.BindEnv("dockerKeyPath", "DOCKER_KEY_PATH")
	viper.BindEnv("volumeDriver", "VOLUME_DRIVER")
	viper.BindEnv("volumeDriverOpts", "VOLUME_DRIVER_OPTS")
	viper.BindEnv("verbosity", "VERBOSITY")
}

func setViperDefaults() { //todo am i missing anything?
	viper.SetDefault("verbosity", "INFO")
	viper.SetDefault("queueDurable", true)
	viper.SetDefault("queueAutoDelete", false)
	viper.SetDefault("queueExclusive", false)
	viper.SetDefault("consumerAutoAck", false)
	viper.SetDefault("consumerExclusive", false)
}

// GCPFormatter enables the ability to use genesis logging with Stackdriver
type GCPFormatter struct {
	//todo: does this stay the same?
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

	lvl, err := log.ParseLevel(conf.Verbosity)
	if err != nil {
		log.SetLevel(log.InfoLevel)
		log.Warn(err)
	}
	log.SetLevel(lvl)
}

// GetConfig gets a pointer to the global config object.
// Do not modify conf object
func GetConfig() *Config {
	return conf
}
