/**
 * Copyright 2019 Whiteblock Inc. All rights reserved.
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 */

package repository

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/whiteblock/genesis/pkg/entity"

	"github.com/docker/cli/cli/command"
	"github.com/docker/distribution/reference"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/tlsconfig"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	def "github.com/whiteblock/definition/command"
)

//DockerRepository provides extra functions for docker service, which could be placed inside of docker
//service, but would make the testing more difficult
type DockerRepository interface {
	//WithTLSClientConfig provides the opt for TLS auth
	WithTLSClientConfig(cacertPath, certPath, keyPath string) client.Opt

	//EnsureImagePulled checks if the docker host contains an image and pulls it if it does not
	EnsureImagePulled(ctx context.Context, cli entity.Client,
		imageName string, auth def.Credentials) error

	//GetContainerByName attempts to find a container with the given name and return information on it.
	GetContainerByName(ctx context.Context, cli entity.Client, containerName string) (types.Container, error)

	//GetNetworkByName attempts to find a network with the given name and return information on it.
	GetNetworkByName(ctx context.Context, cli entity.Client, networkName string) (types.NetworkResource, error)

	//HostHasImage returns true if the docker host has an image matching what was given
	HostHasImage(ctx context.Context, cli entity.Client, image string) (bool, error)

	//Exec is sort of like docker exec
	Exec(ctx context.Context, cli entity.Client, containerName string, details entity.Exec) error
}

type dockerRepository struct {
	log logrus.Ext1FieldLogger
}

//NewDockerRepository creates a new DockerRepository instance
func NewDockerRepository(log logrus.Ext1FieldLogger) DockerRepository {
	return &dockerRepository{log: log}
}

func (da dockerRepository) WithTLSClientConfig(cacertPath, certPath, keyPath string) client.Opt {
	return func(c *client.Client) error {
		opts := tlsconfig.Options{
			CAFile:             cacertPath,
			CertFile:           certPath,
			KeyFile:            keyPath,
			ExclusiveRootPools: true,
			InsecureSkipVerify: true,
		}
		config, err := tlsconfig.Client(opts)
		if err != nil {
			return errors.Wrap(err, "failed to create tls config")
		}
		if transport, ok := c.HTTPClient().Transport.(*http.Transport); ok {
			transport.TLSClientConfig = config
			return nil
		}
		return errors.Errorf("cannot apply tls config to transport: %T", c.HTTPClient().Transport)
	}
}

//HostHasImage returns true if the docker host has an image matching what was given
func (da dockerRepository) HostHasImage(ctx context.Context, cli entity.Client, image string) (bool, error) {
	imgs, err := cli.ImageList(ctx, types.ImageListOptions{All: false})
	if err != nil {
		return false, err
	}
	for _, img := range imgs {
		for _, tag := range img.RepoTags {
			if tag == image {
				return true, nil
			}
		}
		for _, digest := range img.RepoDigests {
			if digest == image {
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
	b64, err := command.EncodeAuthToBase64(types.AuthConfig{
		Username:      auth.Username,
		Password:      auth.Password,
		RegistryToken: auth.RegistryToken,
	})
	if err != nil {
		da.log.WithField("error", err).Error("unable to base64 encode the credentials")
		return ""
	}
	return b64
}

//EnsureImagePulled checks if the docker host contains an image and pulls it if it does not
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
	rd, err := cli.ImagePull(ctx, name, types.ImagePullOptions{
		Platform:     "Linux",
		RegistryAuth: da.handleCredentials(auth),
	})
	if err != nil {
		return err
	}
	defer rd.Close()
	_, err = ioutil.ReadAll(rd)
	return err
}

//GetNetworkByName attempts to find a network with the given name and return information on it.
func (da dockerRepository) GetNetworkByName(ctx context.Context, cli entity.Client,
	networkName string) (types.NetworkResource, error) {

	nets, err := cli.NetworkList(ctx, types.NetworkListOptions{})
	if err != nil {
		return types.NetworkResource{}, err
	}
	for _, net := range nets {
		if net.Name == networkName {
			return net, nil
		}
	}
	return types.NetworkResource{}, fmt.Errorf("could not find the network \"%s\"", networkName)
}

//GetContainerByName attempts to find a container with the given name and return information on it.
func (da dockerRepository) GetContainerByName(ctx context.Context, cli entity.Client,
	containerName string) (types.Container, error) {

	cntrs, err := cli.ContainerList(ctx, types.ContainerListOptions{})
	if err != nil {
		return types.Container{}, err
	}
	for _, cntr := range cntrs {
		for _, name := range cntr.Names {
			if strings.Trim(name, "/") == strings.Trim(containerName, "/") {
				return cntr, nil
			}
		}
	}
	return types.Container{}, fmt.Errorf("could not find the container \"%s\"", containerName)
}

func (da dockerRepository) exec(ctx context.Context, cli entity.Client,
	containerName string, details entity.Exec) error {

	da.log.WithFields(logrus.Fields{
		"command": strings.Join(details.Cmd, " "),
	}).Debug("executing a command")
	idRes, err := cli.ContainerExecCreate(ctx, containerName, types.ExecConfig{
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
	err = cli.ContainerExecStart(ctx, idRes.ID, types.ExecStartCheck{})
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
				return fmt.Errorf(`command "%s" exited with exit code %d`, strings.Join(details.Cmd, " "), res.ExitCode)
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
	if strings.Contains(err.Error(), "connect to the Docker daemon") { //bypass to help get out the dead things
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
