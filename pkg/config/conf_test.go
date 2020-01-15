/**
 * Copyright 2019 Whiteblock Inc. All rights reserved.
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 */


package config

import (
	"strconv"
	"testing"

	"github.com/whiteblock/genesis/pkg/entity"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

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

	logger := conf.GetLogger()
	assert.NotNil(t, logger)
	assert.True(t, logger.IsLevelEnabled(logrus.InfoLevel))
	assert.False(t, logger.IsLevelEnabled(logrus.DebugLevel))
}

func TestConfig_GetLogger_Failure(t *testing.T) {
	conf := Config{
		Verbosity: "3434ds",
	}

	logger := conf.GetLogger()
	assert.NotNil(t, logger)
	assert.True(t, logger.IsLevelEnabled(logrus.InfoLevel))
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
