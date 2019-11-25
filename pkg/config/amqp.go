package config

import (
	"github.com/spf13/viper"
	"github.com/streadway/amqp"
)

//AMQP is the configuration for AMQP
type AMQP struct {
	//QueueName the name of the queue to connect to
	QueueName string
	//Queue is the configuration for the queue
	Queue Queue
	//Consume is the configuration of the consumer
	Consume Consume
	//Publish is the configuration for the publishing of messages
	Publish Publish

	Endpoint AMQPEndpoint
}

//Queue is the configuration of the queue to be created if the queue does not yet exist
type Queue struct {
	//Durable sets the queue to be persistent
	Durable bool `mapstructure:"queueDurable"`
	//AutoDelete tells the queue to drop messages if there are not any consumers
	AutoDelete bool `mapstructure:"queueAutoDelete"`
	//Exclusive queues are only accessible by the connection that declares them
	Exclusive bool
	//NoWait is true, the queue will assume to be declared on the server
	NoWait bool
	//Args contains addition arguments to be provided
	Args amqp.Table
}

//GetQueue gets Queue from the values in viper
func GetQueue() (Queue, error) {
	out := Queue{
		Exclusive: false,
		NoWait:    false,
	}
	return out, viper.Unmarshal(&out)
}

//Consume is the configuration of the consumer of messages from the queue
type Consume struct {
	//Consumer is the name of the consumer
	Consumer string `mapstructure:"consumer"`
	//AutoAck causes the server to acknowledge deliveries to this consumer prior to writing the delivery to the network
	AutoAck bool
	//Exclusive: when true, the server will ensure that this is the sole consumer from this queue
	//This should always be false.
	Exclusive bool
	//NoLocal is not supported by RabbitMQ
	NoLocal bool
	//NoWait: do not wait for the server to confirm the request and immediately begin deliveries
	NoWait bool `mapstructure:"consumerNoWait"`
	//Args contains addition arguments to be provided
	Args amqp.Table
}

//GetConsume gets Consume from the values in viper
func GetConsume() (Consume, error) {
	out := Consume{
		AutoAck:   false,
		Exclusive: false,
		NoLocal:   false,
	}
	return out, viper.Unmarshal(&out)
}

//Publish  is the configuration for the publishing of messages
type Publish struct {
	Mandatory bool   `mapstructure:"publishMandatory"`
	Immediate bool   `mapstructure:"publishImmediate"`
	Exchange  string `mapstructure:"exchange"`
}

//GetPublish gets Publish from the values in viper
func GetPublish() (Publish, error) {
	out := Publish{}
	return out, viper.Unmarshal(&out)
}

//AMQPEndpoint is the configuration needed to connect to a rabbitmq vhost
type AMQPEndpoint struct {
	QueueProtocol string `mapstructure:"queueProtocol"`
	QueueUser     string `mapstructure:"queueUser"`
	QueuePassword string `mapstructure:"queuePassword"`
	QueueHost     string `mapstructure:"queueHost"`
	QueuePort     int    `mapstructure:"queuePort"`
	QueueVHost    string `mapstructure:"queueVHost"`
}

//GetAMQPEndpoint gets AMQPEndpoint from the values in viper
func GetAMQPEndpoint() (AMQPEndpoint, error) {
	out := AMQPEndpoint{}
	return out, viper.Unmarshal(&out)
}

func amqpInit() {
	/** START Queue **/
	viper.BindEnv("queueDurable", "QUEUE_DURABLE")
	viper.BindEnv("queueAutoDelete", "QUEUE_AUTO_DELETE")
	viper.SetDefault("queueDurable", true)
	viper.SetDefault("queueAutoDelete", false)
	/** END Queue **/

	/** START Consume **/
	viper.BindEnv("consumer", "CONSUMER")
	viper.BindEnv("consumerNoWait", "CONSUMER_NO_WAIT")
	viper.SetDefault("consumer", "provisioner")
	viper.SetDefault("consumerNoWait", false)
	/** END Consume **/

	/** START Publish **/
	viper.BindEnv("publishMandatory", "PUBLISH_MANDATORU")
	viper.BindEnv("publishImmediate", "PUBLISH_IMMEDIATE")
	viper.BindEnv("exchange", "EXCHANGE")
	viper.SetDefault("exchange", "")
	viper.SetDefault("publishMandatory", false)
	viper.SetDefault("publishImmediate", false)
	/** END Publish **/

	/** START AMQPEndpoint **/
	viper.BindEnv("queueProtocol", "QUEUE_PROTOCOL")
	viper.BindEnv("queueUser", "QUEUE_USER")
	viper.BindEnv("queuePassword", "QUEUE_PASSWORD")
	viper.BindEnv("queueHost", "QUEUE_HOST")
	viper.BindEnv("queuePort", "QUEUE_PORT")
	viper.BindEnv("queueVHost", "QUEUE_VHOST")

	viper.SetDefault("queueProtocol", "amqp")
	viper.SetDefault("queueUser", "user")
	viper.SetDefault("queuePassword", "password")
	viper.SetDefault("queueHost", "localhost")
	viper.SetDefault("queuePort", 5672)
	viper.SetDefault("queueVHost", "/test")
	/** END AMQPEndpoint **/
}
