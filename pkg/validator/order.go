/**
 * Copyright 2019 Whiteblock Inc. All rights reserved.
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
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

	// ErrMissingImage means missing image field
	ErrMissingImage = errors.New(`missing field "image"`)

	// ErrHostPortTooHigh means the host port number is too high
	ErrHostPortTooHigh = fmt.Errorf(`host port mapping cannot exceed %d`, UnixMinEphemeralPort)

	// ErrContainerPortTooHigh means the container port number is too high
	ErrContainerPortTooHigh = fmt.Errorf(`container port mapping cannot exceed %d`, UnixMinEphemeralPort)
)

// Container validates a container command payload
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
