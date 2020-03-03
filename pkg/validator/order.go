/**
 * Copyright 2019 Whiteblock Inc. All rights reserved.
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 */

package validator

import (
	"errors"
	"strconv"

	"github.com/whiteblock/definition/command"
)

var (
	// ErrMissingName means missing name field
	ErrMissingName = errors.New(`missing field "name"`)

	// ErrMissingImage means missing image field
	ErrMissingImage = errors.New(`missing field "image"`)
)

// Container validates a container command payload
func Container(cntr command.Container) error {
	if len(cntr.Name) == 0 {
		return ErrMissingName
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
