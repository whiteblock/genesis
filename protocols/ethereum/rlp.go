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

package ethereum

import (
	"encoding/hex"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/whiteblock/genesis/util"
)

// RLPHexEncode converts an object into hex encoded RLP
func RLPHexEncode(v interface{}) (string, error) {
	rlp, err := rlp.EncodeToBytes(v)
	if err != nil {
		return "", util.LogError(err)
	}
	return hex.EncodeToString(rlp), nil
}
