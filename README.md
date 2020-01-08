Genesis
======
[![Build](https://travis-ci.org/whiteblock/genesis.svg?branch=dev)](https://www.travis-ci.org/Whiteblock/genesis/)
[![Maintainability](https://api.codeclimate.com/v1/badges/a30e833d3367ef530eaf/maintainability)](https://codeclimate.com/github/Whiteblock/genesis/maintainability)
[![Go report card](https://goreportcard.com/badge/github.com/whiteblock/genesis)](https://goreportcard.com/report/github.com/whiteblock/genesis)
[![codecov](https://codecov.io/gh/Whiteblock/genesis/branch/dev/graph/badge.svg)](https://codecov.io/gh/Whiteblock/genesis)
![Jenkins](https://jenkins-dev.whiteblock.io/buildStatus/icon?job=genesis)

![Version](https://img.shields.io/github/tag/whiteblock/genesis.svg)
[![API Reference](
https://camo.githubusercontent.com/915b7be44ada53c290eb157634330494ebe3e30a/68747470733a2f2f676f646f632e6f72672f6769746875622e636f6d2f676f6c616e672f6764646f3f7374617475732e737667
)](https://godoc.org/github.com/whiteblock/genesis)
[![Gitter](https://badges.gitter.im/whiteblock-io/community.svg)](https://gitter.im/whiteblock-io/community?utm_source=badge&utm_medium=badge&utm_campaign=pr-badge)
[![Docker](https://img.shields.io/badge/Docker%20Image-gcr.io%2Fwhiteblock%2Fgenesis%3Adev-brightgreen)](https://gcr.io/whiteblock/genesis:dev)

# Overview
The Whiteblock platform allows users to provision multiple fully-functioning nodes over which they have complete control within a private test network 


# Configuration
## General

| NAME                          | DEFAULT                                | DESCRIPTION                    |
| ------------------------------------- | ---------------------------- | ----------
| LOCAL_MODE | true | Puts Genesis into standalone mode for testing |
| VERBOSITY | INFO | The verbosity level of the logging |
| LISTEN | 0.0.0.0:8000 | The socket to listen on for the REST API

## RabbitMQ
| NAME                   | DEFAULT                    | DESCRIPTION         |
| ------------------------------------- | ---------------------------- | ----------
| COMPLETION_QUEUE_NAME | completion | The name of the completion queue |
| COMMAND_QUEUE_NAME | commands | The name of the commands queue |
| QUEUE_DURABLE | true | If Genesis creates the queue, should it be durable |
| QUEUE_AUTO_DELETE | false | If Genesis creates the queue, should it delete messages when there is no consumer |
| CONSUMER | genesis | The name of this consumer from the queue |
| CONSUMER_NO_WAIT | false | Enable no wait mode |
| QUEUE_PROTOCOL | amqp | The protocol to use to connect to the queue |
| QUEUE_USER | user | The user portion of the auth credentials |
| QUEUE_PASSWORD | password | The password portion of the auth credentials |
| QUEUE_HOST | localhost | The host address which hosts rabbitmq |
| QUEUE_PORT | 5672 | The port to connect to on the host address |
| QUEUE_VHOST | /test | The rabbitmq vhost to connect to |