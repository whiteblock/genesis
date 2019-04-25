package manager

import (
	db "../db"
	util "../util"
	"fmt"
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
		err := fmt.Errorf("Too many nodes, max of %d nodes.", conf.MaxNodes)
		return err
	}

	if details.Nodes < 1 {
		err := fmt.Errorf("You must have atleast 1 node")
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
		err := fmt.Errorf("servers cannot be null.")
		if err != nil {
			log.Println(err)
			return err
		}
	}

	if len(details.Servers) == 0 {
		err := fmt.Errorf("servers cannot be empty.")
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
		err := fmt.Errorf("images cannot be null.")
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
