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

//FileHandler is the configuration for execution
type FileHandler struct {
	APIEndpoint string        `mapstructure:"apiEndpoint"`
	APITimeout  time.Duration `mapstructure:"apiTimeout"`
}

//NewFileHandler creates a new FileHandler config from the given viper
func NewFileHandler(v *viper.Viper) (out FileHandler, err error) {
	return out, v.Unmarshal(&out)
}

func setFileHandlerBindings(v *viper.Viper) error {
	err := v.BindEnv("apiTimeout", "API_TIMEOUT")
	if err != nil {
		return err
	}
	return v.BindEnv("apiEndpoint", "API_ENDPOINT")
}

func setFileHandlerDefaults(v *viper.Viper) {
	v.SetDefault("apiEndpoint", "https://www.infra.whiteblock.io")
	v.SetDefault("apiTimeout", 10*time.Second)
}
