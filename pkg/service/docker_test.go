/*
	Copyright 2019 Whiteblock Inc.
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

package service

import (
	"fmt"
	//"strings"
	"testing"

	entityMock "github.com/whiteblock/genesis/mocks/pkg/entity"
	externalsMock "github.com/whiteblock/genesis/mocks/pkg/externals"
	//fileMock "github.com/whiteblock/genesis/mocks/pkg/file"
	repoMock "github.com/whiteblock/genesis/mocks/pkg/repository"
	"github.com/whiteblock/genesis/pkg/config"
	"github.com/whiteblock/genesis/pkg/entity"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/api/types/network"
	dockerVolume "github.com/docker/docker/api/types/volume"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/whiteblock/definition/command"
	"github.com/whiteblock/utility/utils"
)

func TestNewDockerService(t *testing.T) {
	assert.NotNil(t, NewDockerService(nil, config.Docker{}, nil, nil))
}

func TestDockerService_CreateContainer(t *testing.T) {
	testNetwork := types.NetworkResource{Name: "Testnet", ID: "id1"}
	testContainer := command.Container{
		EntryPoint: "/bin/bash",
		Environment: map[string]string{
			"FOO": "BAR",
		},
		Name:    "TEST",
		Network: "Testnet",
		Ports:   map[int]int{8888: 8889},
		Volumes: []command.Mount{command.Mount{Name: "volume1", Directory: "/foo/bar", ReadOnly: false}}, //TODO
		Image:   "alpine",
		Args:    []string{"test"},
	}
	testContainer.Cpus = "2.5"
	testContainer.Memory = "5gb"

	cli := new(entityMock.Client)
	cli.On("ContainerCreate", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(
		container.ContainerCreateCreatedBody{}, nil).Run(func(args mock.Arguments) {
		require.Len(t, args, 5)
		assert.Nil(t, args.Get(0))
		{
			config, ok := args.Get(1).(*container.Config)
			require.True(t, ok)
			require.NotNil(t, config)
			require.Len(t, config.Entrypoint, 2)
			assert.Contains(t, config.Env, "FOO=BAR")
			assert.Equal(t, testContainer.EntryPoint, config.Entrypoint[0])
			assert.Equal(t, testContainer.Args[0], config.Entrypoint[1])
			assert.Equal(t, testContainer.Name, config.Hostname)
			assert.NotNil(t, config.Labels)
			assert.Equal(t, testContainer.Image, config.Image)
			{
				_, exists := config.ExposedPorts["8889/tcp"]
				assert.True(t, exists)
			}
		}
		{
			hostConfig, ok := args.Get(2).(*container.HostConfig)
			require.True(t, ok)
			require.NotNil(t, hostConfig)
			assert.Equal(t, int64(2500000000), hostConfig.NanoCPUs)
			assert.Equal(t, int64(5*utils.Gibi), hostConfig.Memory)
			{ //Port bindings
				bindings, exists := hostConfig.PortBindings["8889/tcp"]
				assert.True(t, exists)
				require.NotNil(t, bindings)
				require.Len(t, bindings, 1)
				assert.Equal(t, bindings[0].HostIP, "0.0.0.0")
				assert.Equal(t, bindings[0].HostPort, "8888")
			}

			require.NotNil(t, hostConfig.Mounts)
			require.Len(t, hostConfig.Mounts, len(testContainer.Volumes))
			for i, vol := range testContainer.Volumes {
				assert.Equal(t, vol.Name, hostConfig.Mounts[i].Source)
				assert.Equal(t, vol.Directory, hostConfig.Mounts[i].Target)
				assert.Equal(t, vol.ReadOnly, hostConfig.Mounts[i].ReadOnly)
				assert.Equal(t, mount.TypeVolume, hostConfig.Mounts[i].Type)
				assert.Nil(t, hostConfig.Mounts[i].TmpfsOptions)
				assert.Nil(t, hostConfig.Mounts[i].BindOptions)
			}
			assert.True(t, hostConfig.AutoRemove)
		}
		{
			networkingConfig, ok := args.Get(3).(*network.NetworkingConfig)
			require.True(t, ok)
			require.NotNil(t, networkingConfig)
			require.NotNil(t, networkingConfig.EndpointsConfig)
			netconf, ok := networkingConfig.EndpointsConfig[testContainer.Network]
			require.True(t, ok)
			require.NotNil(t, netconf)
			assert.Equal(t, testNetwork.Name, netconf.NetworkID)
			assert.Nil(t, netconf.Links)
			assert.Nil(t, netconf.Aliases)
			assert.Nil(t, netconf.IPAMConfig)
			assert.Empty(t, netconf.IPv6Gateway)
			assert.Empty(t, netconf.GlobalIPv6Address)
			assert.Empty(t, netconf.EndpointID)
			assert.Empty(t, netconf.Gateway)
			assert.Empty(t, netconf.IPAddress)
			assert.Nil(t, netconf.DriverOpts)
		}
		containerName, ok := args.Get(4).(string)
		require.True(t, ok)
		assert.Equal(t, testContainer.Name, containerName)
	})

	repo := new(repoMock.DockerRepository)
	repo.On("EnsureImagePulled", mock.Anything, mock.Anything,
		mock.Anything, mock.Anything).Return(nil).Run(func(args mock.Arguments) {

		require.Len(t, args, 4)
		assert.Nil(t, args.Get(0))
		assert.NotNil(t, args.Get(1))
		assert.Equal(t, testContainer.Image, args.String(2))
	})

	ds := NewDockerService(repo, config.Docker{}, nil, logrus.New())
	res := ds.CreateContainer(nil, entity.DockerCli{
		Client: cli,
		Labels: map[string]string{
			"FOO": "BAR",
		},
	}, testContainer)
	assert.NoError(t, res.Error)
}

func TestDockerService_StartContainer_Success(t *testing.T) {
	scCommand := command.StartContainer{Name: "TEST"}
	cli := new(entityMock.Client)
	cli.On("ContainerStart", mock.Anything, mock.Anything, mock.Anything).Return(nil).Run(
		func(args mock.Arguments) {

			require.Len(t, args, 3)
			assert.Nil(t, args.Get(0))
			assert.Equal(t, scCommand.Name, args.Get(1))
			assert.Equal(t, types.ContainerStartOptions{}, args.Get(2))
		}).Once()
	conn := new(externalsMock.NetConn)
	conn.On("SetDeadline", mock.Anything).Return(nil).Maybe()
	conn.On("Close", mock.Anything).Return(nil).Maybe()
	//conn.On("Read",mock.Anything).Return(0,io.EOF)

	cli.On("ContainerAttach", mock.Anything, mock.Anything, mock.Anything).Return(
		types.HijackedResponse{Conn: conn}, nil).Run(
		func(args mock.Arguments) {

			require.Len(t, args, 3)
			assert.Nil(t, args.Get(0))
			assert.Equal(t, scCommand.Name, args.Get(1))
		}).Maybe()

	repo := new(repoMock.DockerRepository)
	ds := NewDockerService(repo, config.Docker{}, nil, logrus.New())
	res := ds.StartContainer(nil, entity.DockerCli{Client: cli}, scCommand)
	assert.NoError(t, res.Error)
	cli.AssertExpectations(t)
	conn.AssertExpectations(t)
}

func TestDockerService_CreateNetwork_Success(t *testing.T) {
	testNetwork := command.Network{
		Name:    "testnet",
		Global:  true,
		Gateway: "10.14.0.1",
		Subnet:  "10.14.0.0/16",
	}
	cli := new(entityMock.Client)
	cli.On("NetworkCreate", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(
		types.NetworkCreateResponse{}, nil).Run(func(args mock.Arguments) {

		require.Len(t, args, 3)
		assert.Nil(t, args.Get(0))
		assert.Equal(t, testNetwork.Name, args.String(1))

		networkCreate, ok := args.Get(2).(types.NetworkCreate)
		require.True(t, ok)
		require.NotNil(t, networkCreate)
		assert.True(t, networkCreate.CheckDuplicate)
		assert.True(t, networkCreate.Attachable)
		assert.False(t, networkCreate.Ingress)
		assert.False(t, networkCreate.Internal)
		assert.False(t, networkCreate.EnableIPv6)
		assert.NotNil(t, networkCreate.Labels)
		assert.Nil(t, networkCreate.ConfigFrom)

		require.NotNil(t, networkCreate.IPAM)
		assert.Equal(t, "default", networkCreate.IPAM.Driver)
		assert.Nil(t, networkCreate.IPAM.Options)
		require.NotNil(t, networkCreate.IPAM.Config)
		require.Len(t, networkCreate.IPAM.Config, 1)
		assert.Equal(t, testNetwork.Subnet, networkCreate.IPAM.Config[0].Subnet)
		assert.Equal(t, testNetwork.Gateway, networkCreate.IPAM.Config[0].Gateway)

		if testNetwork.Global {
			assert.Equal(t, "overlay", networkCreate.Driver)
			assert.Equal(t, "swarm", networkCreate.Scope)
		} else {
			assert.Equal(t, "bridge", networkCreate.Driver)
			assert.Equal(t, "local", networkCreate.Scope)
			bridgeName, ok := networkCreate.Options["com.docker.network.bridge.name"]
			assert.True(t, ok)
			assert.Equal(t, testNetwork.Name, bridgeName)
		}
	}).Twice()

	repo := new(repoMock.DockerRepository)
	ds := NewDockerService(repo, config.Docker{}, nil, logrus.New())

	res := ds.CreateNetwork(nil, entity.DockerCli{
		Client: cli,
		Labels: map[string]string{
			"FOO": "BAR",
		},
	}, testNetwork)
	assert.NoError(t, res.Error)

	testNetwork.Global = false
	res = ds.CreateNetwork(nil, entity.DockerCli{
		Client: cli,
		Labels: map[string]string{
			"FOO": "BAR",
		},
	}, testNetwork)
	assert.NoError(t, res.Error)

	cli.AssertExpectations(t)
}

func TestDockerService_CreateNetwork_Failure(t *testing.T) {
	testNetwork := command.Network{
		Name:   "testnet",
		Global: true,
		Labels: map[string]string{
			"FOO": "BAR",
		},
		Gateway: "10.14.0.1",
		Subnet:  "10.14.0.0/16",
	}
	cli := new(entityMock.Client)
	cli.On("NetworkCreate", mock.Anything, mock.Anything, mock.Anything).Return(
		types.NetworkCreateResponse{}, fmt.Errorf("error")).Once()

	repo := new(repoMock.DockerRepository)
	ds := NewDockerService(repo, config.Docker{}, nil, logrus.New())

	res := ds.CreateNetwork(nil, entity.DockerCli{Client: cli}, testNetwork)
	assert.Error(t, res.Error)

	cli.AssertExpectations(t)
}

func TestDockerService_RemoveNetwork_Success(t *testing.T) {
	cli := new(entityMock.Client)
	networks := []types.NetworkResource{
		types.NetworkResource{Name: "test1"},
		types.NetworkResource{Name: "test2"},
	}

	for _, net := range networks {
		cli.On("NetworkRemove", mock.Anything, net.Name).Return(nil).Run(
			func(args mock.Arguments) {

				require.Len(t, args, 2)
				assert.Nil(t, args.Get(0))
			}).Once()
	}

	ds := NewDockerService(nil, config.Docker{}, nil, logrus.New())

	for _, net := range networks {
		res := ds.RemoveNetwork(nil, entity.DockerCli{Client: cli}, net.Name)
		assert.NoError(t, res.Error)
	}

	cli.AssertExpectations(t)
}

func TestDockerService_RemoveNetwork_NetworkList_Failure(t *testing.T) {
	cli := new(entityMock.Client)
	cli.On("NetworkRemove", mock.Anything, mock.Anything).Return(fmt.Errorf("test")).Once()

	ds := NewDockerService(nil, config.Docker{}, nil, logrus.New())

	res := ds.RemoveNetwork(nil, entity.DockerCli{Client: cli}, "")
	assert.Error(t, res.Error)

	cli.AssertExpectations(t)
}

func TestDockerService_RemoveNetwork_NetworkRemove_Failure(t *testing.T) {
	cli := new(entityMock.Client)
	networks := []types.NetworkResource{
		types.NetworkResource{Name: "test1"},
		types.NetworkResource{Name: "test2"},
	}

	for _, net := range networks {
		cli.On("NetworkRemove", mock.Anything, net.Name).Return(fmt.Errorf("err")).Once()
	}

	ds := NewDockerService(nil, config.Docker{}, nil, logrus.New())

	for _, net := range networks {
		res := ds.RemoveNetwork(nil, entity.DockerCli{Client: cli}, net.Name)
		assert.Error(t, res.Error)
	}

	cli.AssertExpectations(t)
}

func TestDockerService_RemoveContainer(t *testing.T) {
	cli := new(entityMock.Client)
	cntrs := []types.Container{
		types.Container{Names: []string{"test1", "test3"}, ID: "id1"},
		types.Container{Names: []string{"test2", "test4"}, ID: "id2"},
	}

	for _, cntr := range cntrs {
		cli.On("ContainerRemove", mock.Anything, cntr.Names[0], mock.Anything).Return(nil).Run(
			func(args mock.Arguments) {

				require.Len(t, args, 3)
				assert.Nil(t, args.Get(0))
				opts, ok := args.Get(2).(types.ContainerRemoveOptions)
				require.True(t, ok)
				assert.False(t, opts.RemoveVolumes)
				assert.False(t, opts.RemoveLinks)
				assert.True(t, opts.Force)

			}).Once()
	}

	ds := NewDockerService(nil, config.Docker{}, nil, logrus.New())

	for _, cntr := range cntrs {
		res := ds.RemoveContainer(nil, entity.DockerCli{Client: cli}, cntr.Names[0])
		assert.NoError(t, res.Error)
	}
	cli.AssertExpectations(t)
}

/*
func TestDockerService_PlaceFileInContainer(t *testing.T) {
	testDir := "/pkg"
	fileID := "barfoo"
	testContainer := types.Container{Names: []string{"test1"}, ID: "id1"}

	cli := new(entityMock.Client)
	cli.On("CopyToContainer", mock.Anything, mock.Anything,
		mock.Anything, mock.Anything, mock.Anything).Return(nil).Run(
		func(args mock.Arguments) {
			require.Len(t, args, 5)
			assert.Nil(t, args.Get(0))
			assert.Equal(t, testContainer.Names[0], args.String(1))
			assert.Equal(t, testDir, args.String(2))
			assert.NotNil(t, args.Get(3))
			{
				opts, ok := args.Get(4).(types.CopyToContainerOptions)
				require.True(t, ok)
				assert.True(t, opts.AllowOverwriteDirWithFile)
				assert.False(t, opts.CopyUIDGID)
			}
		}).Once()

	fs := new(fileMock.RemoteSources)
	fs.On("GetTarReader", mock.Anything, mock.Anything).Return(strings.NewReader("barfoo"), nil).Once()
	ds := NewDockerService(nil, config.Docker{}, fs, logrus.New())

	res := ds.PlaceFileInContainer(nil, entity.DockerCli{Client: cli},
		testContainer.Names[0], command.File{
			Destination: testDir,
			ID:          fileID,
		})
	assert.NoError(t, res.Error)

	cli.AssertExpectations(t)
	fs.AssertExpectations(t)
}*/

func TestDockerService_AttachNetwork(t *testing.T) {
	cn := command.ContainerNetwork{
		Network:       "test2",
		ContainerName: "test1",
	}
	cli := new(entityMock.Client)
	cli.On("NetworkConnect", mock.Anything, mock.Anything, mock.Anything,
		mock.Anything).Return(nil).Run(func(args mock.Arguments) {

		require.Len(t, args, 4)
		assert.Nil(t, args.Get(0))
		assert.Equal(t, cn.Network, args.Get(1))
		assert.Equal(t, cn.ContainerName, args.Get(2))
		epSettings, ok := args.Get(3).(*network.EndpointSettings)
		require.True(t, ok)
		require.NotNil(t, epSettings)
	}).Once()

	ds := NewDockerService(nil, config.Docker{}, nil, logrus.New())

	res := ds.AttachNetwork(nil, entity.DockerCli{Client: cli}, cn)
	assert.NoError(t, res.Error)

	cli.AssertExpectations(t)
}

func TestDockerService_DetachNetwork(t *testing.T) {
	netName := "test2"
	cntrName := "test1"

	cli := new(entityMock.Client)
	cli.On("NetworkDisconnect", mock.Anything, mock.Anything, mock.Anything, mock.Anything,
		mock.Anything).Return(nil).Run(func(args mock.Arguments) {

		require.Len(t, args, 4)
		assert.Nil(t, args.Get(0))
		assert.Equal(t, netName, args.String(1))
		assert.Equal(t, cntrName, args.String(2))
		assert.True(t, args.Bool(3))
	}).Once()

	ds := NewDockerService(nil, config.Docker{}, nil, logrus.New())

	res := ds.DetachNetwork(nil, entity.DockerCli{Client: cli}, netName, cntrName)
	assert.NoError(t, res.Error)

	cli.AssertExpectations(t)
}

func TestDockerService_CreateVolume(t *testing.T) {
	volume := types.Volume{
		Name:   "test_volume",
		Labels: map[string]string{"foo": "bar"},
	}

	cli := new(entityMock.Client)
	cli.On("VolumeCreate", mock.Anything, mock.Anything).Return(volume, nil).Run(
		func(args mock.Arguments) {
			require.Len(t, args, 2)
			assert.Nil(t, args.Get(0))

			vol, ok := args.Get(1).(dockerVolume.VolumeCreateBody)
			require.True(t, ok)
			require.NotNil(t, vol)
			assert.Contains(t, vol.Labels, "foo")

			assert.Equal(t, volume.Name, vol.Name)
			assert.Equal(t, volume.Labels["foo"], vol.Labels["foo"])

		}).Once()

	repo := new(repoMock.DockerRepository)

	ds := NewDockerService(repo, config.Docker{}, nil, logrus.New())

	res := ds.CreateVolume(nil, entity.DockerCli{Client: cli}, command.Volume{
		Name:   "test_volume",
		Labels: map[string]string{"foo": "bar"},
	})
	assert.NoError(t, res.Error)

	cli.AssertExpectations(t)
}

func TestDockerService_RemoveVolume(t *testing.T) {
	name := "test_volume"

	cli := new(entityMock.Client)
	cli.On("VolumeRemove", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil).Run(
		func(args mock.Arguments) {
			require.Len(t, args, 3)
			assert.Nil(t, args.Get(0))
			assert.Equal(t, name, args.Get(1))
			assert.Equal(t, true, args.Get(2))
		}).Once()

	repo := new(repoMock.DockerRepository)

	ds := NewDockerService(repo, config.Docker{}, nil, logrus.New())

	res := ds.RemoveVolume(nil, entity.DockerCli{Client: cli}, name)
	assert.NoError(t, res.Error)

	cli.AssertExpectations(t)
}

func TestRandomMacAddress(t *testing.T) {
	_, err := generateMacAddress()
	if err != nil {
		t.Fatal(err)
	}
}
