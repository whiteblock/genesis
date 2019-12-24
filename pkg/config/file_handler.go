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
