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
	"github.com/whiteblock/genesis/pkg/entity"

	joonix "github.com/joonix/log"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	queue "github.com/whiteblock/amqp"
	"github.com/whiteblock/definition/command"
)

// Config groups all of the global configuration parameters into
// a single struct
type Config struct {
	MaxMessageRetries   int64  `mapstructure:"maxMessageRetries"`
	QueueMaxConcurrency int64  `mapstructure:"queueMaxConcurrency"`
	CompletionQueueName string `mapstructure:"completionQueueName"`
	CommandQueueName    string `mapstructure:"commandQueueName"`
	ErrorQueueName      string `mapstructure:"errorQueueName"`
	StatusQueueName     string `mapstructure:"statusQueueName"`

	// LocalMode indicates that Genesis is operating in standalone mode
	LocalMode        bool              `mapstructure:"localMode"`
	VolumeDriver     string            `mapstructure:"volumeDriver"`
	VolumeDriverOpts map[string]string `mapstructure:"volumeDriverOpts"`
	Verbosity        string            `mapstructure:"verbosity"`
	FluentDLogging   bool              `mapstructure:"fluentDLogging"`
	Listen           string            `mapstructure:"listen"`

	Execution   Execution   `mapstructure:"-"`
	Docker      Docker      `mapstructure:"-"`
	FileHandler FileHandler `mapstructure:"-"`
}

// GetLogger gets a logger according to the config
func (c Config) GetLogger() *logrus.Logger {
	logger := logrus.New()
	lvl, err := logrus.ParseLevel(c.Verbosity)
	if err != nil {
		logger.SetLevel(logrus.InfoLevel)
	} else {
		logger.SetLevel(lvl)
	}

	logger.SetReportCaller(true)
	if c.FluentDLogging {
		logger.SetFormatter(joonix.NewFormatter())
	}

	return logger
}

// CompletionAMQP gets the AMQP for the completion queue
func (c Config) CompletionAMQP() (queue.AMQPConfig, error) {
	conf, err := queue.NewAMQPConfig(viper.GetViper())
	conf.QueueName = c.CompletionQueueName
	return conf, err
}

// CommandAMQP gets the AMQP for the command queue
func (c Config) CommandAMQP() (queue.AMQPConfig, error) {
	conf, err := queue.NewAMQPConfig(viper.GetViper())
	conf.QueueName = c.CommandQueueName
	return conf, err
}

// ErrorsAMQP gets the AMQP for the command queue
func (c Config) ErrorsAMQP() (queue.AMQPConfig, error) {
	conf, err := queue.NewAMQPConfig(viper.GetViper())
	conf.QueueName = c.ErrorQueueName
	return conf, err
}

// StatusAMQP gets the AMQP for the command queue
func (c Config) StatusAMQP() (queue.AMQPConfig, error) {
	conf, err := queue.NewAMQPConfig(viper.GetViper())
	conf.QueueName = c.StatusQueueName
	return conf, err
}

// GetVolumeConfig extracts the fields of this object representing VolumeConfig
func (c Config) GetVolumeConfig() command.VolumeConfig {
	return command.VolumeConfig{
		Driver:     c.VolumeDriver,
		DriverOpts: c.VolumeDriverOpts,
	}
}

// GetRestConfig extracts the fields of this object representing RestConfig
func (c Config) GetRestConfig() entity.RestConfig {
	return entity.RestConfig{Listen: c.Listen}
}

func setViperEnvBindings() {
	viper.BindEnv("statusQueueName", "STATUS_QUEUE_NAME")
	viper.BindEnv("fluentDLogging", "FLUENT_D_LOGGING")
	viper.BindEnv("maxMessageRetries", "MAX_MESSAGE_RETRIES")
	viper.BindEnv("queueMaxConcurrency", "QUEUE_MAX_CONCURRENCY")

	viper.BindEnv("localMode", "LOCAL_MODE")
	viper.BindEnv("volumeDriver", "VOLUME_DRIVER")
	viper.BindEnv("volumeDriverOpts", "VOLUME_DRIVER_OPTS")
	viper.BindEnv("verbosity", "VERBOSITY")
	viper.BindEnv("listen", "LISTEN")
	viper.BindEnv("completionQueueName", "COMPLETION_QUEUE_NAME")
	viper.BindEnv("commandQueueName", "COMMAND_QUEUE_NAME")
	viper.BindEnv("errorQueueName", "ERROR_QUEUE_NAME")
	setExecutionBindings(viper.GetViper())
	setDockerBindings(viper.GetViper())
	setFileHandlerBindings(viper.GetViper())
}

func setViperDefaults() {
	viper.SetDefault("statusQueueName", "status")
	viper.SetDefault("fluentDLogging", true)
	viper.SetDefault("completionQueueName", "teardownRequests")
	viper.SetDefault("commandQueueName", "commands")
	viper.SetDefault("maxMessageRetries", 5)
	viper.SetDefault("queueMaxConcurrency", 20)
	viper.SetDefault("verbosity", "INFO")
	viper.SetDefault("listen", "0.0.0.0:8000")
	viper.SetDefault("localMode", true)
	viper.SetDefault("errorQueueName", "errors")

	setExecutionDefaults(viper.GetViper())
	setDockerDefaults(viper.GetViper())
	setFileHandlerDefaults(viper.GetViper())
}

func init() {
	queue.SetConfig(viper.GetViper())
	setViperDefaults()
	setViperEnvBindings()

	viper.AddConfigPath("/etc/whiteblock/")          // path to look for the config file in
	viper.AddConfigPath("$HOME/.config/whiteblock/") // call multiple times to add many search paths
	viper.SetConfigName("genesis")
	viper.SetConfigType("yaml")

}

// NewConfig creates a new config object from the global config
func NewConfig() (conf Config, err error) {
	_ = viper.ReadInConfig()
	err = viper.Unmarshal(&conf)
	if err != nil {
		return
	}
	conf.Execution, err = NewExecution(viper.GetViper())
	if err != nil {
		return
	}

	conf.FileHandler, err = NewFileHandler(viper.GetViper())
	if err != nil {
		return
	}

	conf.Docker, err = NewDocker(viper.GetViper())
	return
}
