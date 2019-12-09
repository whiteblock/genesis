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
	"strings"

	"github.com/whiteblock/genesis/pkg/entity"

	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"github.com/whiteblock/definition/command"
)

// Config groups all of the global configuration parameters into
// a single struct
type Config struct {
	MaxMessageRetries   int64  `mapstructure:"maxMessageRetries"`
	QueueMaxConcurrency int64  `mapstructure:"queueMaxConcurrency"`
	CompletionQueues    string `mapstructure:"completionQueues"`
	CommandQueueName    string `mapstructure:"commandQueueName"`
	DockerCACertPath    string `mapstructure:"dockerCACertPath"`
	DockerCertPath      string `mapstructure:"dockerCertPath"`
	DockerKeyPath       string `mapstructure:"dockerKeyPath"`
	//LocalMode indicates that Genesis is operating in standalone mode
	LocalMode        bool              `mapstructure:"localMode"`
	VolumeDriver     string            `mapstructure:"volumeDriver"`
	VolumeDriverOpts map[string]string `mapstructure:"volumeDriverOpts"`
	Verbosity        string            `mapstructure:"verbosity"`
	Listen           string            `mapstructure:"listen"`
}

//GetLogger gets a logger according to the config
func (c Config) GetLogger() (*logrus.Logger, error) {
	logger := logrus.New()
	lvl, err := logrus.ParseLevel(c.Verbosity)
	if err != nil {
		return nil, err
	}
	logger.SetLevel(lvl)
	logger.SetReportCaller(true)
	return logger, nil
}

func getAmqpBase() (AMQP, error) {
	queue, err := GetQueue()
	if err != nil {
		return AMQP{}, err
	}

	consume, err := GetConsume()
	if err != nil {
		return AMQP{}, err
	}

	publish, err := GetPublish()
	if err != nil {
		return AMQP{}, err
	}

	ep, err := GetAMQPEndpoint()
	if err != nil {
		return AMQP{}, err
	}

	return AMQP{
		Queue:    queue,
		Consume:  consume,
		Publish:  publish,
		Endpoint: ep,
	}, nil
}

//CompletionAMQP gets the AMQP for the completion queue
func (c Config) CompletionAMQP() ([]AMQP, error) {
	out := []AMQP{}
	queues := strings.Split(c.CompletionQueues, ",")
	for _, complQueue := range queues {
		conf, err := getAmqpBase()
		if err != nil {
			return nil, err
		}
		conf.QueueName = strings.TrimSpace(complQueue)
		out = append(out, conf)
	}

	return out, nil
}

//CommandAMQP gets the AMQP for the command queue
func (c Config) CommandAMQP() (AMQP, error) {
	conf, err := getAmqpBase()
	conf.QueueName = c.CommandQueueName
	return conf, err
}

// GetDockerConfig extracts the fields of this object representing DockerConfig
func (c Config) GetDockerConfig() entity.DockerConfig {
	return entity.DockerConfig{
		CACertPath: c.DockerCACertPath,
		CertPath:   c.DockerCertPath,
		KeyPath:    c.DockerKeyPath,
		LocalMode:  c.LocalMode,
	}
}

// GetVolumeConfig extracts the fields of this object representing VolumeConfig
func (c Config) GetVolumeConfig() command.VolumeConfig {
	return command.VolumeConfig{
		Driver:     c.VolumeDriver,
		DriverOpts: c.VolumeDriverOpts,
	}
}

//GetRestConfig extracts the fields of this object representing RestConfig
func (c Config) GetRestConfig() entity.RestConfig {
	return entity.RestConfig{Listen: c.Listen}
}

func setViperEnvBindings() {
	viper.BindEnv("maxMessageRetries", "MAX_MESSAGE_RETRIES")
	viper.BindEnv("queueMaxConcurrency", "QUEUE_MAX_CONCURRENCY")
	viper.BindEnv("dockerCACertPath", "DOCKER_CACERT_PATH")
	viper.BindEnv("dockerCertPath", "DOCKER_CERT_PATH")
	viper.BindEnv("dockerKeyPath", "DOCKER_KEY_PATH")
	viper.BindEnv("localMode", "LOCAL_MODE")
	viper.BindEnv("volumeDriver", "VOLUME_DRIVER")
	viper.BindEnv("volumeDriverOpts", "VOLUME_DRIVER_OPTS")
	viper.BindEnv("verbosity", "VERBOSITY")
	viper.BindEnv("listen", "LISTEN")
	viper.BindEnv("completionQueues", "COMPLETION_QUEUES")
	viper.BindEnv("commandQueueName", "COMMAND_QUEUE_NAME")
}

func setViperDefaults() {
	viper.SetDefault("completionQueues", "teardownRequests")
	viper.SetDefault("commandQueueName", "commands")
	viper.SetDefault("maxMessageRetries", 10)
	viper.SetDefault("queueMaxConcurrency", 20)
	viper.SetDefault("verbosity", "INFO")
	viper.SetDefault("listen", "0.0.0.0:8000")
	viper.SetDefault("localMode", true)
}

func init() {
	amqpInit()
	setViperDefaults()
	setViperEnvBindings()

	viper.AddConfigPath("/etc/whiteblock/")          // path to look for the config file in
	viper.AddConfigPath("$HOME/.config/whiteblock/") // call multiple times to add many search paths
	viper.SetConfigName("genesis")
	viper.SetConfigType("yaml")

}

// NewConfig creates a new config object from the global config
func NewConfig() (*Config, error) {
	conf := new(Config)
	_ = viper.ReadInConfig()

	return conf, viper.Unmarshal(&conf)
}
