/**
 * Copyright 2019 Whiteblock Inc. All rights reserved.
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 */

package repository

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/distribution/reference"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/api/types/registry"
	"github.com/sirupsen/logrus"
	def "github.com/whiteblock/definition/command"

	"github.com/whiteblock/genesis/pkg/entity"
)

// DockerRepository provides extra functions for docker service, which could be placed inside of docker
// service, but would make the testing more difficult
type DockerRepository interface {
	// EnsureImagePulled checks if the docker host contains an image and pulls it if it does not
	EnsureImagePulled(ctx context.Context, cli entity.Client,
		imageName string, auth def.Credentials) error

	// GetContainerByName attempts to find a container with the given name and return information on it.
	GetContainerByName(ctx context.Context, cli entity.Client, containerName string) (container.Summary, error)

	// GetNetworkByName attempts to find a network with the given name and return information on it.
	GetNetworkByName(ctx context.Context, cli entity.Client, networkName string) (network.Summary, error)

	// HostHasImage returns true if the docker host has an image matching what was given
	HostHasImage(ctx context.Context, cli entity.Client, image string) (bool, error)

	// Exec is sort of like docker exec
	Exec(ctx context.Context, cli entity.Client, containerName string, details entity.Exec) error
}

type dockerRepository struct {
	log logrus.Ext1FieldLogger
}

// NewDockerRepository creates a new DockerRepository instance
func NewDockerRepository(log logrus.Ext1FieldLogger) DockerRepository {
	return &dockerRepository{log: log}
}

// HostHasImage returns true if the docker host has an image matching what was given
func (da dockerRepository) HostHasImage(ctx context.Context, cli entity.Client, img string) (bool, error) {
	imgs, err := cli.ImageList(ctx, image.ListOptions{All: false})
	if err != nil {
		return false, err
	}
	for _, summary := range imgs {
		for _, tag := range summary.RepoTags {
			if tag == img {
				return true, nil
			}
		}
		for _, digest := range summary.RepoDigests {
			if digest == img {
				return true, nil
			}
		}
	}
	return false, nil
}

func (da dockerRepository) handleCredentials(auth def.Credentials) string {
	if auth.Empty() {
		return ""
	}
	ac := registry.AuthConfig{
		Username:      auth.Username,
		Password:      auth.Password,
		RegistryToken: auth.RegistryToken,
	}
	b, err := json.Marshal(ac)
	if err != nil {
		da.log.WithField("error", err).Error("unable to marshal the credentials")
		return ""
	}
	return base64.URLEncoding.EncodeToString(b)
}

// EnsureImagePulled checks if the docker host contains an image and pulls it if it does not
func (da dockerRepository) EnsureImagePulled(ctx context.Context, cli entity.Client,
	imageName string, auth def.Credentials) error {
	distributionRef, err := reference.ParseNormalizedNamed(imageName)
	if err != nil {
		return err
	}
	name := distributionRef.String()
	exists, err := da.HostHasImage(ctx, cli, name)
	if exists || err != nil {
		return err
	}
	exists2, err := da.HostHasImage(ctx, cli, imageName)
	if exists2 || err != nil {
		return err
	}
	rd, err := cli.ImagePull(ctx, name, image.PullOptions{
		Platform:     "Linux",
		RegistryAuth: da.handleCredentials(auth),
	})
	if err != nil {
		return err
	}
	defer rd.Close()
	_, err = io.Copy(io.Discard, rd)
	return err
}

// GetNetworkByName attempts to find a network with the given name and return information on it.
func (da dockerRepository) GetNetworkByName(ctx context.Context, cli entity.Client,
	networkName string) (network.Summary, error) {

	nets, err := cli.NetworkList(ctx, network.ListOptions{})
	if err != nil {
		return network.Summary{}, err
	}
	for _, net := range nets {
		if net.Name == networkName {
			return net, nil
		}
	}
	return network.Summary{}, fmt.Errorf("could not find the network %q", networkName)
}

// GetContainerByName attempts to find a container with the given name and return information on it.
func (da dockerRepository) GetContainerByName(ctx context.Context, cli entity.Client,
	containerName string) (container.Summary, error) {

	cntrs, err := cli.ContainerList(ctx, container.ListOptions{})
	if err != nil {
		return container.Summary{}, err
	}
	for _, cntr := range cntrs {
		for _, name := range cntr.Names {
			if strings.Trim(name, "/") == strings.Trim(containerName, "/") {
				return cntr, nil
			}
		}
	}
	return container.Summary{}, fmt.Errorf("could not find the container %q", containerName)
}

func (da dockerRepository) exec(ctx context.Context, cli entity.Client,
	containerName string, details entity.Exec) error {

	da.log.WithFields(logrus.Fields{
		"command": strings.Join(details.Cmd, " "),
	}).Debug("executing a command")
	idRes, err := cli.ContainerExecCreate(ctx, containerName, container.ExecOptions{
		User:         "",
		Privileged:   details.Privileged,
		Tty:          false,
		AttachStdin:  false,
		AttachStderr: false,
		AttachStdout: false,
		Detach:       true,
		DetachKeys:   "",
		Env:          nil,
		WorkingDir:   "",
		Cmd:          details.Cmd,
	})
	if err != nil {
		return err
	}
	err = cli.ContainerExecStart(ctx, idRes.ID, container.ExecStartOptions{})
	if err != nil {
		return err
	}
	for {
		res, err := cli.ContainerExecInspect(ctx, idRes.ID)
		if err != nil {
			return err
		}
		if !res.Running {
			if res.ExitCode != 0 {
				return fmt.Errorf(`command %q exited with exit code %d`, strings.Join(details.Cmd, " "), res.ExitCode)
			}
			break
		}
	}

	return nil
}

func (da dockerRepository) Exec(ctx context.Context, cli entity.Client,
	containerName string, details entity.Exec) error {
	err := da.exec(ctx, cli, containerName, details)
	if err == nil {
		return nil
	}
	if strings.Contains(err.Error(), "connect to the Docker daemon") { // bypass to help get out the dead things
		return err
	}

	for i := 0; i < details.Retries; i++ {
		if details.Delay != 0 {
			time.Sleep(details.Delay)
		}
		da.log.WithFields(logrus.Fields{
			"command": details.Cmd,
			"attempt": i + 1,
		}).Debug("retrying a command")
		err = da.exec(ctx, cli, containerName, details)
		if err == nil {
			break
		}
	}
	return err
}
