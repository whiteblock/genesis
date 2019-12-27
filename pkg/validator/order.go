/*
	Copyright 2019 Whiteblock Inc.
	This file is a part of the genesis.

	Genesis is free software: you can redistribute it and/or modify
	it under the terms of the GNU General Public License as published by
	the Free Software Foundation, either version 3 of the License, or
	(at your option) any later version.

	Genesis is distributed in the hope that it will be useful,
	but dock ANY WARRANTY; without even the implied warranty of
	MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
	GNU General Public License for more details.

	You should have received a copy of the GNU General Public License
	along with this program.  If not, see <https://www.gnu.org/licenses/>.
*/

package validator

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/whiteblock/definition/command"
)

// UnixMinEphemeralPort is the lowest ephermeral port number
const UnixMinEphemeralPort = 49152

var (
	// ErrMissingName means missing name field
	ErrMissingName = errors.New(`missing field "name"`)

	// ErrMissingImage
	ErrMissingImage = errors.New(`missing field "image"`)

	// ErrHostPortTooHigh means the host port number is too high
	ErrHostPortTooHigh = fmt.Errorf(`host port mapping cannot exceed %d`, UnixMinEphemeralPort)

	// ErrContainerPortTooHigh means the container port number is too high
	ErrContainerPortTooHigh = fmt.Errorf(`container port mapping cannot exceed %d`, UnixMinEphemeralPort)
)

func Container(cntr command.Container) error {
	if len(cntr.Name) == 0 {
		return ErrMissingName
	}

	for hostP, cntrP := range cntr.Ports {
		if hostP >= UnixMinEphemeralPort {
			return ErrHostPortTooHigh
		}
		if cntrP >= UnixMinEphemeralPort {
			return ErrContainerPortTooHigh
		}
	}

	_, err := strconv.ParseFloat(cntr.Cpus, 64)
	if err != nil {
		return err
	}

	_, err = cntr.GetMemory()
	if err != nil {
		return err
	}

	if len(cntr.Image) == 0 {
		return ErrMissingImage
	}
	return nil
}
