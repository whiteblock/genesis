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

package aion

import (
	"encoding/json"
	"fmt"
	"github.com/whiteblock/genesis/protocols/helpers"
	"github.com/whiteblock/genesis/protocols/services"
)

// AConf represents the settings for the aion build
type AConf struct {
	CorsEnabled    bool   `xml:"corsEnabled" json:"corsEnabled"`
	SecureConnect  bool   `xml:"secureConnect" json:"secureConnect"`
	NRGDefault     int64  `xml:"nrgDefault" json:"nrgDefault"`
	NRGMax         int64  `xml:"nrgMax" json:"nrgMax"`
	OracleEnabled  bool   `xml:"oracleEnabled" json:"oracleEnabled"`
	BlocksQueueMax int64  `xml:"blocksQueueMax"` //TODO continue adding JSON tags
	ShowStatus     bool   `xml:"showStatus"`
	ShowStatistics bool   `xml:"showStatistics"`
	CompactEnabled bool   `xml:"compactEnabled"`
	SlowImport     int64  `xml:"slowImport"`
	Frequency      int64  `xml:"frequency"`
	Mining         bool   `xml:"mining"`
	MineThreads    int64  `xml:"mineThreads"`
	ExtraData      string `xml:"extraData"`
	ClampedDecayUB int64  `xml:"clampedDecayUB"`
	ClampedDecayLB int64  `xml:"clampedDecayLB"`
	Database       string `xml:"database"`
	CheckIntegrity bool   `xml:"checkIntegrity"`
	StateStorage   string `xml:"stateStorage"`
	Vendor         string `xml:"vendor"`
	DBCompression  bool   `xml:"dbCompression"`
	LogFile        bool   `xml:"logFile"`
	LogPath        string `xml:"logPath"`
	GenLogs        string `xml:"genLogs`
	VMLogs         string `xml:"vmLogs"`
	APILogs        string `xml:"apiLogs"`
	SyncLogs       string `xml:"syncLogs"`
	DBLogs         string `xml:"dbLogs"`
	ConsLogs       string `xml:"consLogs"`
	P2PLogs        string `xml:"p2plogs"`
	CacheMax       int64  `xml:"cacheMax"`

	InitBalance   string `json:"initBalance"`
	EnergyLimit   int64  `json:"energyLimit"`
	Nonce         int64  `json:"nonce"`
	Difficulty    int64  `json:"difficulty"`
	TimeStamp     int64  `json:"timeStamp"`
	ChainID       int64  `json:"chainId"`
	ExtraAccounts int64  `json:"extraAccounts"`
}

/**
 * Fills in the defaults for missing parts,
 */
func newConf(data map[string]interface{}) (*AConf, error) {
	out := new(AConf)
	err := helpers.HandleBlockchainConfig(blockchain, data, out)
	if err != nil || data == nil {
		return out, err
	}

	initBalance, exists := data["initBalance"]
	if exists && initBalance != nil {
		switch initBalance.(type) {
		case json.Number:
			out.InitBalance = initBalance.(json.Number).String()
		case string:
			out.InitBalance = initBalance.(string)
		default:
			return nil, fmt.Errorf("incorrect type for initBalance given")
		}
	}
	return out, nil
}

//NewAionConf creates the configuration for aion
func NewAionConf(data map[string]interface{}) (*AConf, error) {
	out := new(AConf)
	return out, helpers.HandleBlockchainConfig(blockchain, data, out)
}

// GetServices returns the services which are used by artemis
func GetServices() []services.Service {
	return nil
}
