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

package service

import (
	//"bufio"
	"fmt"
	//"strings"
	"testing"
	//"io"

	cmdMock "github.com/whiteblock/genesis/mocks/definition/command"
	entityMock "github.com/whiteblock/genesis/mocks/pkg/entity"
	externalsMock "github.com/whiteblock/genesis/mocks/pkg/externals"
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
	repo := new(repoMock.DockerRepository)
	assert.NotNil(t, NewDockerService(repo, config.Docker{}, nil))
}

func TestDockerService_CreateContainer(t *testing.T) {
	testNetwork := types.NetworkResource{Name: "Testnet", ID: "id1"}
	testContainer := command.Container{
		EntryPoint: "/bin/bash",
		Environment: map[string]string{
			"FOO": "BAR",
		},
		Name:    "TEST",
		Network: []string{"Testnet"},
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
			netconf, ok := networkingConfig.EndpointsConfig[testContainer.Network[0]]
			require.True(t, ok)
			require.NotNil(t, netconf)
			assert.Equal(t, netconf.NetworkID, testNetwork.ID)
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
	repo.On("GetNetworkByName", mock.Anything, mock.Anything, mock.Anything).Return(
		testNetwork, nil).Run(func(args mock.Arguments) {

		require.Len(t, args, 3)
		assert.Nil(t, args.Get(0))
		assert.NotNil(t, args.Get(1))
		assert.Contains(t, testContainer.Network, args.String(2))
	})

	ds := NewDockerService(repo, config.Docker{}, logrus.New())
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
	ds := NewDockerService(repo, config.Docker{}, logrus.New())
	res := ds.StartContainer(nil, entity.DockerCli{Client: cli}, scCommand)
	assert.NoError(t, res.Error)
	cli.AssertExpectations(t)
	conn.AssertExpectations(t)
}

func TestDockerService_StartContainer_Failure(t *testing.T) {
	/*scCommand := command.StartContainer{Name: "TEST"}
	cli := new(entityMock.Client)
	cli.On("ContainerStart", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(fmt.Errorf("error")).Once()

	repo := new(repoMock.DockerRepository)
	ds := NewDockerService( repo, config.Docker{}, logrus.New())
	res := ds.StartContainer(nil, cli, scCommand)
	assert.Error(t, res.Error)
	cli.AssertExpectations(t)*/
}

func TestDockerService_StartContainer_Attach_Success(t *testing.T) {
	/*scCommand := command.StartContainer{Name: "TEST", Attach: true}

	mockConn := new(externalsMock.NetConn)
	mockConn.On("Close").Return(nil).Once()

	mockResponse := types.HijackedResponse{Conn: mockConn, Reader: bufio.NewReader(strings.NewReader("tessssttt"))}
	cli := new(entityMock.Client)
	cli.On("ContainerStart", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil).Run(
		func(args mock.Arguments) {

			require.Len(t, args, 4)
			assert.Nil(t, args.Get(0))
			assert.Nil(t, args.Get(1))
			assert.Equal(t, scCommand.Name, args.Get(2))
			assert.Equal(t, types.ContainerStartOptions{}, args.Get(3))
		}).Once()

	cli.On("ContainerAttach", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(mockResponse, nil).Run(
		func(args mock.Arguments) {

			require.Len(t, args, 4)
			assert.Nil(t, args.Get(0))
			assert.Nil(t, args.Get(1))
			assert.Equal(t, scCommand.Name, args.Get(2))
			assert.Equal(t, types.ContainerAttachOptions{
				Stream: true,
				Stdin:  false,
				Stdout: true,
				Stderr: true,
				Logs:   true,
			}, args.Get(3))
		}).Once()

	repo := new(repoMock.DockerRepository)
	ds := NewDockerService( repo, config.Docker{}, logrus.New())
	res := ds.StartContainer(nil, cli, scCommand)
	assert.NoError(t, res.Error)
	cli.AssertExpectations(t)
	mockConn.AssertExpectations(t)*/
}

func TestDockerService_StartContainer_Attach_Failure(t *testing.T) {
	/*scCommand := command.StartContainer{Name: "TEST", Attach: true}

	cli := new(entityMock.Client)
	cli.On("ContainerStart", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil).Once()

	cli.On("ContainerAttach", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(
		types.HijackedResponse{}, fmt.Errorf("err")).Once()

	repo := new(repoMock.DockerRepository)
	ds := NewDockerService( repo, config.Docker{}, logrus.New())
	res := ds.StartContainer(nil, cli, scCommand)
	assert.Error(t, res.Error)
	cli.AssertExpectations(t)*/
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
	ds := NewDockerService(repo, config.Docker{}, logrus.New())

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
	ds := NewDockerService(repo, config.Docker{}, logrus.New())

	res := ds.CreateNetwork(nil, entity.DockerCli{Client: cli}, testNetwork)
	assert.Error(t, res.Error)

	cli.AssertExpectations(t)
}

func TestDockerService_RemoveNetwork_Success(t *testing.T) {
	cli := new(entityMock.Client)
	networks := []types.NetworkResource{
		types.NetworkResource{Name: "test1", ID: "id1"},
		types.NetworkResource{Name: "test2", ID: "id2"},
	}

	repo := new(repoMock.DockerRepository)
	for _, net := range networks {
		cli.On("NetworkRemove", mock.Anything, net.ID).Return(nil).Run(
			func(args mock.Arguments) {

				require.Len(t, args, 2)
				assert.Nil(t, args.Get(0))
			}).Once()
		repo.On("GetNetworkByName", mock.Anything, mock.Anything, net.Name).Return(
			types.NetworkResource{ID: net.ID}, nil).Once()
	}

	ds := NewDockerService(repo, config.Docker{}, logrus.New())

	for _, net := range networks {
		res := ds.RemoveNetwork(nil, entity.DockerCli{Client: cli}, net.Name)
		assert.NoError(t, res.Error)
	}

	cli.AssertExpectations(t)
}

func TestDockerService_RemoveNetwork_NetworkList_Failure(t *testing.T) {
	cli := new(entityMock.Client)
	cli.On("NetworkRemove", mock.Anything, mock.Anything).Return(fmt.Errorf("test")).Once()

	repo := new(repoMock.DockerRepository)
	repo.On("GetNetworkByName", mock.Anything, mock.Anything, mock.Anything).Return(
		types.NetworkResource{ID: ""}, nil).Once()
	ds := NewDockerService(repo, config.Docker{}, logrus.New())

	res := ds.RemoveNetwork(nil, entity.DockerCli{Client: cli}, "")
	assert.Error(t, res.Error)

	cli.AssertExpectations(t)
	repo.AssertExpectations(t)
}

func TestDockerService_RemoveNetwork_NetworkRemove_Failure(t *testing.T) {
	cli := new(entityMock.Client)
	networks := []types.NetworkResource{
		types.NetworkResource{Name: "test1", ID: "id1"},
		types.NetworkResource{Name: "test2", ID: "id2"},
	}

	repo := new(repoMock.DockerRepository)

	for _, net := range networks {
		cli.On("NetworkRemove", mock.Anything, net.ID).Return(fmt.Errorf("err")).Once()
		repo.On("GetNetworkByName", mock.Anything, mock.Anything, net.Name).Return(
			types.NetworkResource{ID: net.ID}, nil).Once()
	}

	ds := NewDockerService(repo, config.Docker{}, logrus.New())

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

	repo := new(repoMock.DockerRepository)
	for _, cntr := range cntrs {
		cli.On("ContainerRemove", mock.Anything, cntr.ID, mock.Anything).Return(nil).Run(
			func(args mock.Arguments) {

				require.Len(t, args, 3)
				assert.Nil(t, args.Get(0))
				opts, ok := args.Get(2).(types.ContainerRemoveOptions)
				require.True(t, ok)
				assert.False(t, opts.RemoveVolumes)
				assert.False(t, opts.RemoveLinks)
				assert.True(t, opts.Force)

			}).Once()
		repo.On("GetContainerByName", mock.Anything, mock.Anything, cntr.Names[0]).Return(
			types.Container{ID: cntr.ID}, nil).Once()
	}

	ds := NewDockerService(repo, config.Docker{}, logrus.New())

	for _, cntr := range cntrs {
		res := ds.RemoveContainer(nil, entity.DockerCli{Client: cli}, cntr.Names[0])
		assert.NoError(t, res.Error)
	}
	cli.AssertExpectations(t)
}

func TestDockerService_PlaceFileInContainer(t *testing.T) {
	testDir := "/pkg"
	testContainer := types.Container{Names: []string{"test1"}, ID: "id1"}

	file := new(cmdMock.IFile)
	file.On("GetDir").Return(testDir).Once()
	file.On("GetTarReader").Return(nil, nil).Once()

	cli := new(entityMock.Client)
	cli.On("CopyToContainer", mock.Anything, mock.Anything,
		mock.Anything, mock.Anything, mock.Anything).Return(nil).Run(
		func(args mock.Arguments) {
			require.Len(t, args, 5)
			assert.Nil(t, args.Get(0))
			assert.Equal(t, testContainer.ID, args.String(1))
			assert.Equal(t, testDir, args.String(2))
			assert.Nil(t, args.Get(3))
			{
				opts, ok := args.Get(4).(types.CopyToContainerOptions)
				require.True(t, ok)
				assert.False(t, opts.AllowOverwriteDirWithFile)
				assert.False(t, opts.CopyUIDGID)
			}
		}).Once()

	repo := new(repoMock.DockerRepository)
	repo.On("GetContainerByName", mock.Anything, mock.Anything, mock.Anything).Return(testContainer, nil).Run(
		func(args mock.Arguments) {

			require.Len(t, args, 3)
			assert.Nil(t, args.Get(0))
			assert.Equal(t, testContainer.Names[0], args.String(2))
		}).Once()

	ds := NewDockerService(repo, config.Docker{}, logrus.New())

	res := ds.PlaceFileInContainer(nil, entity.DockerCli{Client: cli}, testContainer.Names[0], file)
	assert.NoError(t, res.Error)

	cli.AssertExpectations(t)
	repo.AssertExpectations(t)
	file.AssertExpectations(t)
}

func TestDockerService_AttachNetwork(t *testing.T) {
	cntr := types.Container{Names: []string{"test1"}, ID: "id1"}
	net := types.NetworkResource{Name: "test2", ID: "id2"}

	repo := new(repoMock.DockerRepository)
	repo.On("GetContainerByName", mock.Anything, mock.Anything, mock.Anything).Return(cntr, nil).Run(
		func(args mock.Arguments) {

			require.Len(t, args, 3)
			assert.Nil(t, args.Get(0))
			assert.Equal(t, cntr.Names[0], args.String(2))

		}).Once()

	repo.On("GetNetworkByName", mock.Anything, mock.Anything, mock.Anything).Return(net, nil).Run(
		func(args mock.Arguments) {

			require.Len(t, args, 3)
			assert.Nil(t, args.Get(0))
			assert.Equal(t, net.Name, args.String(2))
		}).Once()

	cli := new(entityMock.Client)
	cli.On("NetworkConnect", mock.Anything, mock.Anything, mock.Anything,
		mock.Anything).Return(nil).Run(func(args mock.Arguments) {

		require.Len(t, args, 4)
		assert.Nil(t, args.Get(0))
		assert.Equal(t, net.ID, args.String(1))
		assert.Equal(t, cntr.ID, args.String(2))
		epSettings, ok := args.Get(3).(*network.EndpointSettings)
		require.True(t, ok)
		require.NotNil(t, epSettings)
		assert.Equal(t, net.ID, epSettings.NetworkID)
	}).Once()

	ds := NewDockerService(repo, config.Docker{}, logrus.New())

	res := ds.AttachNetwork(nil, entity.DockerCli{Client: cli}, net.Name, cntr.Names[0])
	assert.NoError(t, res.Error)

	cli.AssertExpectations(t)
	repo.AssertExpectations(t)
}

func TestDockerService_DetachNetwork(t *testing.T) {
	cntr := types.Container{Names: []string{"test1"}, ID: "id1"}
	net := types.NetworkResource{Name: "test2", ID: "id2"}

	repo := new(repoMock.DockerRepository)
	repo.On("GetContainerByName", mock.Anything, mock.Anything, mock.Anything).Return(cntr, nil).Run(
		func(args mock.Arguments) {

			require.Len(t, args, 3)
			assert.Nil(t, args.Get(0))
			assert.Equal(t, cntr.Names[0], args.String(2))

		}).Once()

	repo.On("GetNetworkByName", mock.Anything, mock.Anything, mock.Anything).Return(net, nil).Run(
		func(args mock.Arguments) {

			require.Len(t, args, 3)
			assert.Nil(t, args.Get(0))
			assert.Equal(t, net.Name, args.String(2))
		}).Once()

	cli := new(entityMock.Client)
	cli.On("NetworkDisconnect", mock.Anything, mock.Anything, mock.Anything, mock.Anything,
		mock.Anything).Return(nil).Run(func(args mock.Arguments) {

		require.Len(t, args, 4)
		assert.Nil(t, args.Get(0))
		assert.Equal(t, net.ID, args.String(1))
		assert.Equal(t, cntr.ID, args.String(2))
		assert.True(t, args.Bool(3))
	}).Once()

	ds := NewDockerService(repo, config.Docker{}, logrus.New())

	res := ds.DetachNetwork(nil, entity.DockerCli{Client: cli}, net.Name, cntr.Names[0])
	assert.NoError(t, res.Error)

	cli.AssertExpectations(t)
	repo.AssertExpectations(t)
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

	ds := NewDockerService(repo, config.Docker{}, logrus.New())

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

	ds := NewDockerService(repo, config.Docker{}, logrus.New())

	res := ds.RemoveVolume(nil, entity.DockerCli{Client: cli}, name)
	assert.NoError(t, res.Error)

	assert.True(t, cli.AssertNumberOfCalls(t, "VolumeRemove", 1))
	cli.AssertExpectations(t)
}

//func TestDockerService_Emulation(t *testing.T) {
/*cntr := types.Container{Names: []string{"test1"}, ID: "id1"}
net := types.NetworkResource{
	Name: "test2",
	ID:   "id2",
	IPAM: network.IPAM{
		Config: []network.IPAMConfig{
			network.IPAMConfig{
				Subnet: "10.0.0.0/8",
			},
		},
	},
}
testNetConf := command.Netconf{
	Container:   cntr.Names[0],
	Network:     net.Name,
	Limit:       1000,
	Loss:        5.5,
	Delay:       100000,
	Rate:        "10mbps",
	Duplication: 0.1,
	Corrupt:     0.1,
	Reorder:     0.1,
}
resultingContainerName := cntr.ID + "-" + net.ID

repo := new(repoMock.DockerRepository)
repo.On("GetContainerByName", mock.Anything, mock.Anything, mock.Anything).Return(cntr, nil).Run(
	func(args mock.Arguments) {

		require.Len(t, args, 3)
		assert.Nil(t, args.Get(0))
		assert.Nil(t, args.Get(1))
		assert.Equal(t, cntr.Names[0], args.String(2))

	}).Once()

repo.On("GetNetworkByName", mock.Anything, mock.Anything, mock.Anything).Return(net, nil).Run(
	func(args mock.Arguments) {

		require.Len(t, args, 3)
		assert.Nil(t, args.Get(0))
		assert.Nil(t, args.Get(1))
		assert.Equal(t, net.Name, args.String(2))
	}).Once()
repo.On("EnsureImagePulled", mock.Anything, mock.Anything, mock.Anything).Return(nil).Run(func(args mock.Arguments) {

	require.Len(t, args, 3)
	assert.Nil(t, args.Get(0))
	assert.Nil(t, args.Get(1))
	assert.Equal(t, "gaiadocker/iproute2:latest", args.String(2))
}).Once()*/

/*cli := new(entityMock.Client)
cli.On("ContainerCreate", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(
	container.ContainerCreateCreatedBody{}, nil).Run(func(args mock.Arguments) {
	require.Len(t, args, 6)
	assert.Nil(t, args.Get(0))
	assert.Nil(t, args.Get(1))
	{
		config, ok := args.Get(2).(*container.Config)
		require.True(t, ok)
		require.NotNil(t, config)
		require.Len(t, config.Entrypoint, 3)
		assert.Equal(t, "/bin/sh", config.Entrypoint[0])
		assert.Equal(t, "-c", config.Entrypoint[1])
		//
		assert.Contains(t, config.Entrypoint[2],
			fmt.Sprintf("tc qdisc add dev $(ip -o addr show to %s |"+*/
//		" sed -n 's/.*\\(eth[0-9]*\\).*/\\1/p') root netem", net.IPAM.Config[0].Subnet))
/*assert.Contains(t, config.Entrypoint[2], fmt.Sprintf(" limit %d", testNetConf.Limit))
	assert.Contains(t, config.Entrypoint[2], fmt.Sprintf(" loss %.4f", testNetConf.Loss))
	assert.Contains(t, config.Entrypoint[2], fmt.Sprintf(" delay %dus", testNetConf.Delay))
	assert.Contains(t, config.Entrypoint[2], fmt.Sprintf(" rate %s", testNetConf.Rate))
	assert.Contains(t, config.Entrypoint[2], fmt.Sprintf(" duplicate %.4f", testNetConf.Duplication))
	assert.Contains(t, config.Entrypoint[2], fmt.Sprintf(" corrupt %.4f", testNetConf.Duplication))
	assert.Contains(t, config.Entrypoint[2], fmt.Sprintf(" reorder %.4f", testNetConf.Reorder))

	assert.Nil(t, config.Labels)
	assert.Equal(t, "gaiadocker/iproute2:latest", config.Image)
}
{
	hostConfig, ok := args.Get(3).(*container.HostConfig)
	require.True(t, ok)
	require.NotNil(t, hostConfig)
	assert.Nil(t, hostConfig.PortBindings)
	require.NotNil(t, hostConfig.CapAdd)
	assert.Contains(t, hostConfig.CapAdd, "NET_ADMIN")
	require.Nil(t, hostConfig.Mounts)
	assert.True(t, hostConfig.AutoRemove)
}
{
	networkingConfig, ok := args.Get(4).(*network.NetworkingConfig)
	require.True(t, ok)
	require.NotNil(t, networkingConfig)
	require.Nil(t, networkingConfig.EndpointsConfig)
}
containerName, ok := args.Get(5).(string)
require.True(t, ok)
assert.Equal(t, resultingContainerName, containerName)*/
/*})

	cli.On("ContainerStart", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil).Run(
		func(args mock.Arguments) {

			require.Len(t, args, 4)
			assert.Nil(t, args.Get(0))
			assert.Nil(t, args.Get(1))
			assert.Equal(t, resultingContainerName, args.String(2))
			assert.Equal(t, types.ContainerStartOptions{}, args.Get(3))
		})

	ds := NewDockerService( repo, config.Docker{}, logrus.New())

	res := ds.Emulation(nil, cli, testNetConf)
	assert.NoError(t, res.Error)

	repo.AssertExpectations(t)
	cli.AssertExpectations(t)

}*/
