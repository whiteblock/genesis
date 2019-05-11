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

package util

/****Standard Data Structures****/

// KeyPair represents a cryptographic key pair
type KeyPair struct {
	PrivateKey string `json:"privateKey"`
	PublicKey  string `json:"publicKey"`
}

// Command represents a previously executed command
type Command struct {
	Cmdline  string
	Node     int
	ServerID int
}

// Service represents a service for a blockchain.
// All env variables will be passed to the container.
type Service struct {
	Name    string            `json:"name"`
	Image   string            `json:"image"`
	Env     map[string]string `json:"env"`
	Network string            `json:"network"`
}

// EndPoint represents an endpoint with basic auth
type EndPoint struct {
	URL  string `json:"url"`
	User string `json:"user"`
	Pass string `json:"pass"`
}
