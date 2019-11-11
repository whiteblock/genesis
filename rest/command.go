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

package rest

import (
	"encoding/json"
	"github.com/whiteblock/genesis/pkg/command"
	//"github.com/whiteblock/genesis/state"
	"github.com/whiteblock/genesis/util"
	"net/http"
)

func addCommand(w http.ResponseWriter, r *http.Request) {
	var commands []command.Command
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&commands)
	if err != nil {
		http.Error(w, util.LogError(err).Error(), 400)
		return
	}
	//go state.GetCommandState().AddCommands(commands...)

	w.Write([]byte("Success"))
}
