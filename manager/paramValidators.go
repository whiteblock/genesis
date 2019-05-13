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

package manager

import (
	"fmt"
	"github.com/whiteblock/genesis/db"
	"github.com/whiteblock/genesis/util"
	"log"
)

func validateResources(details *db.DeploymentDetails) error {
	for i, res := range details.Resources {
		err := res.ValidateAndSetDefaults()
		if err != nil {
			log.Println(err)
			return fmt.Errorf("%s. For node %d", err.Error(), i)
		}
	}
	return nil
}

func validateNumOfNodes(details *db.DeploymentDetails) error {
	if details.Nodes > conf.MaxNodes {
		err := fmt.Errorf("too many nodes: max of %d nodes", conf.MaxNodes)
		return err
	}

	if details.Nodes < 1 {
		err := fmt.Errorf("must have at least 1 node")
		return err
	}
	return nil
}

func validateImages(details *db.DeploymentDetails) error {
	for _, image := range details.Images {
		err := util.ValidateCommandLine(image)
		if err != nil {
			log.Println(err)
			return err
		}
	}
	return nil
}

func validateBlockchain(details *db.DeploymentDetails) error {
	err := util.ValidateCommandLine(details.Blockchain)
	if err != nil {
		log.Println(err)
		return err
	}
	return nil
}

func checkForNilOrMissing(details *db.DeploymentDetails) error {
	if details.Servers == nil {
		err := fmt.Errorf("servers cannot be null")
		if err != nil {
			log.Println(err)
			return err
		}
	}

	if len(details.Servers) == 0 {
		err := fmt.Errorf("servers cannot be empty")
		if err != nil {
			log.Println(err)
			return err
		}
	}

	if len(details.Blockchain) == 0 {
		err := fmt.Errorf("blockchain cannot be empty")
		if err != nil {
			log.Println(err)
			return err
		}
	}

	if details.Images == nil {
		err := fmt.Errorf("images cannot be null")
		if err != nil {
			log.Println(err)
			return err
		}
	}

	if len(details.Images) == 0 {
		err := fmt.Errorf("images cannot be empty")
		if err != nil {
			log.Println(err)
			return err
		}
	}

	return nil
}

func validate(details *db.DeploymentDetails) error {
	err := checkForNilOrMissing(details)
	if err != nil {
		log.Println(err)
		return err
	}

	err = validateResources(details)
	if err != nil {
		log.Println(err)
		return err
	}

	err = validateNumOfNodes(details)
	if err != nil {
		log.Println(err)
		return err
	}

	err = validateImages(details)
	if err != nil {
		log.Println(err)
		return err
	}

	return validateBlockchain(details)
}
