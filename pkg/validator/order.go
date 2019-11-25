/*
	Copyright 2019 whiteblock Inc.
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
	"strconv"

	"github.com/whiteblock/genesis/pkg/command"
)

//OrderValidator is a series of functions which validate different orders
type OrderValidator interface {
	ValidateContainer(cntr command.Container) error
}

type orderValidator struct {
}

//NewOrderValidator creates a new OrderValidator
func NewOrderValidator() OrderValidator {
	return &orderValidator{}
}

func (ov *orderValidator) ValidateContainer(cntr command.Container) error {
	if len(cntr.Name) == 0 {
		return errors.New(`missing field "name"`)
	}

	for hostP, cntrP := range cntr.Ports {
		if hostP >= 49152 {
			return errors.New(`host port mapping cannot exceed 49152`)
		}
		if cntrP >= 49152 {
			return errors.New(`container port mapping cannot exceed 49152`)
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
		return errors.New(`missing field "image"`)
	}
	return nil
}
