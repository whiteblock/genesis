/**
 * Copyright 2019 Whiteblock Inc. All rights reserved.
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 */

package config

import (
	"github.com/spf13/viper"
	"time"
)

//Execution is the configuration for execution
type Execution struct {
	LimitPerTest      int64         `mapstructure:"executionLimitPerTest"`
	ConnectionRetries int           `mapstructure:"executionConnectionRetries"`
	RetryDelay        time.Duration `mapstructure:"executionRetryDelay"`
	// DebugMode causes Fatal errors to be replaced with trapping errors, which do
	// not signal completion
	DebugMode bool `mapstructure:"debugMode"`
}

//NewExecution creates a new Execution config from the given viper
func NewExecution(v *viper.Viper) (out Execution, err error) {
	return out, v.Unmarshal(&out)
}

func setExecutionBindings(v *viper.Viper) error {
	err := v.BindEnv("executionLimitPerTest", "EXECUTION_LIMIT_PER_TEST")
	if err != nil {
		return err
	}
	err = v.BindEnv("executionRetryDelay", "EXECUTION_RETRY_DELAY")
	if err != nil {
		return err
	}
	err = v.BindEnv("debugMode", "DEBUG_MODE")
	if err != nil {
		return err
	}
	return v.BindEnv("executionConnectionRetries", "EXECUTION_CONNECTION_RETRIES")
}

func setExecutionDefaults(v *viper.Viper) {
	v.SetDefault("executionLimitPerTest", 40)
	v.SetDefault("executionConnectionRetries", 5)
	v.SetDefault("executionRetryDelay", "10s")
	v.SetDefault("debugMode", true)
}
