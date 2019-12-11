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
