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

package usecase

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	mockService "github.com/whiteblock/genesis/mocks/pkg/service"
	mockUseCase "github.com/whiteblock/genesis/mocks/pkg/usecase"
	"github.com/whiteblock/genesis/pkg/command"
	"github.com/whiteblock/genesis/pkg/entity"
)

func TestNewDockerUseCase(t *testing.T) {
	service := new(mockService.DockerService)
	cmdService := new(mockService.CommandService)

	expected := &dockerUseCase{
		service:    service,
		cmdService: cmdService,
	}

	duc, err := NewDockerUseCase(service, cmdService)
	assert.NoError(t, err)

	assert.Equal(t, expected, duc)
}

func TestDockerUseCase_dependencyCheck_tooSoon(t *testing.T) {

	cmd := command.Command{
		ID:        "test",
		Timestamp: 17737688240,
	}
	cmdService := new(mockService.CommandService)
	cmdService.On("CheckDependenciesExecuted", mock.Anything).Return(true, nil)

	duc, err := NewDockerUseCase(nil, cmdService)
	assert.NoError(t, err)
	res, ok := duc.(*dockerUseCase).dependencyCheck(cmd)
	assert.False(t, ok)
	assert.Error(t, res.Error)
	assert.Equal(t, entity.TooSoonType, res.Type)
}

func TestDockerUseCase_dependencyCheck_fail(t *testing.T) {

	cmd := command.Command{
		ID:        "test",
		Timestamp: 0,
	}
	cmdService := new(mockService.CommandService)
	cmdService.On("CheckDependenciesExecuted", mock.Anything).Return(false, nil)

	duc, err := NewDockerUseCase(nil, cmdService)
	assert.NoError(t, err)
	res, ok := duc.(*dockerUseCase).dependencyCheck(cmd)
	assert.False(t, ok)
	assert.Error(t, res.Error)
}

func TestDockerUseCase_dependencyCheck_error(t *testing.T) {

	cmd := command.Command{
		ID:        "test",
		Timestamp: 0,
	}
	cmdService := new(mockService.CommandService)
	cmdService.On("CheckDependenciesExecuted", mock.Anything).Return(false, fmt.Errorf("err"))

	duc, err := NewDockerUseCase(nil, cmdService)
	assert.NoError(t, err)
	res, ok := duc.(*dockerUseCase).dependencyCheck(cmd)
	assert.False(t, ok)
	assert.Error(t, res.Error)
}

func TestDockerUseCase_TimeSupplier(t *testing.T) {
	usecase := new(mockUseCase.DockerUseCase)
	usecase.On("TimeSupplier").Return(int64(5)).Once()

	assert.Equal(t, usecase.TimeSupplier(), int64(5))
	usecase.AssertExpectations(t)
}

func TestDockerUseCase_Run_CreateContainer(t *testing.T) {
	service := new(mockService.DockerService)
	service.On("CreateClient", mock.Anything, mock.Anything).Return(nil, nil)
	service.On("CreateContainer", mock.Anything, mock.Anything, mock.Anything).Return(entity.Result{Type: entity.SuccessType})

	cmdService := new(mockService.CommandService)
	cmdService.On("CheckDependenciesExecuted", mock.Anything).Return(true, nil)

	usecase, err := NewDockerUseCase(service, cmdService)
	assert.NoError(t, err)

	res := usecase.Run(command.Command{
		ID:        "TEST",
		Timestamp: time.Now().Unix() - 5,
		Timeout:   0,
		Target:    command.Target{IP: "0.0.0.0"},
		Order: command.Order{
			Type:    "createContainer",
			Payload: map[string]interface{}{},
		},
	})
	assert.Equal(t, res.Error, nil)
	assert.True(t, service.AssertNumberOfCalls(t, "CreateContainer", 1))
}

func TestDockerUseCase_Run_StartContainer(t *testing.T) {
	service := new(mockService.DockerService)
	service.On("CreateClient", mock.Anything, mock.Anything).Return(nil, nil)
	service.On("StartContainer", mock.Anything, mock.Anything, mock.Anything).Return(entity.Result{Type: entity.SuccessType})

	cmdService := new(mockService.CommandService)
	cmdService.On("CheckDependenciesExecuted", mock.Anything).Return(true, nil)

	usecase, err := NewDockerUseCase(service, cmdService)
	assert.NoError(t, err)

	res := usecase.Run(command.Command{
		ID:        "TEST",
		Timestamp: time.Now().Unix() - 5,
		Timeout:   0,
		Target:    command.Target{IP: "0.0.0.0"},
		Order: command.Order{
			Type:    "startContainer",
			Payload: map[string]interface{}{"name": "test"},
		},
	})
	assert.Equal(t, res.Error, nil)
	assert.True(t, service.AssertNumberOfCalls(t, "StartContainer", 1))
}

func TestDockerUseCase_Run_RemoveContainer_Success(t *testing.T) {
	service := new(mockService.DockerService)
	containerName := "testContainer"
	service.On("CreateClient", mock.Anything, mock.Anything).Return(nil, nil).Once()
	service.On("RemoveContainer", mock.Anything, mock.Anything, mock.Anything).Return(
		entity.Result{Type: entity.SuccessType}).Run(
		func(args mock.Arguments) {

			require.Len(t, args, 3)
			assert.NotNil(t, args.Get(0))
			assert.Nil(t, args.Get(1))
			assert.Equal(t, containerName, args.String(2))

		}).Once()

	cmdService := new(mockService.CommandService)
	cmdService.On("CheckDependenciesExecuted", mock.Anything).Return(true, nil).Once()

	usecase, err := NewDockerUseCase(service, cmdService)
	assert.NoError(t, err)

	res := usecase.Run(command.Command{
		ID:        "TEST",
		Timestamp: 0,
		Timeout:   0,
		Target:    command.Target{IP: "0.0.0.0"},
		Order: command.Order{
			Type: "removeContainer",
			Payload: map[string]interface{}{
				"name": containerName,
			},
		},
	})
	assert.Equal(t, res.Error, nil)
	service.AssertExpectations(t)
	cmdService.AssertExpectations(t)
}

func TestDockerUseCase_Run_RemoveContainer_Failure(t *testing.T) {
	service := new(mockService.DockerService)
	service.On("CreateClient", mock.Anything, mock.Anything).Return(nil, nil)
	service.On("RemoveContainer", mock.Anything, mock.Anything, mock.Anything).Return(
		entity.Result{Type: entity.SuccessType}).Maybe()

	cmdService := new(mockService.CommandService)
	cmdService.On("CheckDependenciesExecuted", mock.Anything).Return(true, nil).Once()

	usecase, err := NewDockerUseCase(service, cmdService)
	assert.NoError(t, err)

	res := usecase.Run(command.Command{
		ID:        "TEST",
		Timestamp: time.Now().Unix() - 5,
		Timeout:   0,
		Target:    command.Target{IP: "0.0.0.0"},
		Order: command.Order{
			Type:    "removeContainer",
			Payload: map[string]interface{}{},
		},
	})
	assert.Error(t, res.Error)
	service.AssertExpectations(t)
	cmdService.AssertExpectations(t)
}

func TestDockerUseCase_Run_CreateNetwork(t *testing.T) {
	service := new(mockService.DockerService)
	service.On("CreateClient", mock.Anything, mock.Anything).Return(nil, nil)
	service.On("CreateNetwork", mock.Anything, mock.Anything, mock.Anything).Return(
		entity.Result{Type: entity.SuccessType}).Once()

	cmdService := new(mockService.CommandService)
	cmdService.On("CheckDependenciesExecuted", mock.Anything).Return(true, nil)

	usecase, err := NewDockerUseCase(service, cmdService)
	assert.NoError(t, err)

	res := usecase.Run(command.Command{
		ID:        "TEST",
		Timestamp: time.Now().Unix() - 5,
		Timeout:   0,
		Target:    command.Target{IP: "0.0.0.0"},
		Order: command.Order{
			Type:    "createNetwork",
			Payload: map[string]interface{}{"name": "test"},
		},
	})
	assert.Equal(t, res.Error, nil)
	service.AssertExpectations(t)
}

func TestDockerUseCase_Run_AttachNetwork(t *testing.T) {
	service := new(mockService.DockerService)
	service.On("CreateClient", mock.Anything, mock.Anything).Return(nil, nil)
	service.On("AttachNetwork", mock.Anything, mock.Anything, mock.Anything,
		mock.Anything, mock.Anything).Return(entity.Result{Type: entity.SuccessType}).Once()

	cmdService := new(mockService.CommandService)
	cmdService.On("CheckDependenciesExecuted", mock.Anything).Return(true, nil)

	usecase, err := NewDockerUseCase(service, cmdService)
	assert.NoError(t, err)

	res := usecase.Run(command.Command{
		ID:        "TEST",
		Timestamp: time.Now().Unix() - 5,
		Timeout:   0,
		Target:    command.Target{IP: "0.0.0.0"},
		Order: command.Order{
			Type: "attachNetwork",
			Payload: map[string]interface{}{
				"container": "test",
				"network":   "testnet",
			},
		},
	})
	assert.NoError(t, res.Error)
	service.AssertExpectations(t)
}

func TestDockerUseCase_Run_DetachNetwork(t *testing.T) {
	service := new(mockService.DockerService)
	service.On("CreateClient", mock.Anything, mock.Anything).Return(nil, nil).Once()
	service.On("DetachNetwork", mock.Anything, mock.Anything, mock.Anything,
		mock.Anything, mock.Anything).Return(entity.NewSuccessResult()).Once()

	cmdService := new(mockService.CommandService)
	cmdService.On("CheckDependenciesExecuted", mock.Anything).Return(true, nil)

	usecase, err := NewDockerUseCase(service, cmdService)
	assert.NoError(t, err)

	res := usecase.Run(command.Command{
		ID:        "TEST",
		Timestamp: time.Now().Unix() - 5,
		Timeout:   0,
		Target:    command.Target{IP: "0.0.0.0"},
		Order: command.Order{
			Type: "detachNetwork",
			Payload: map[string]interface{}{
				"container": "test",
				"network":   "testnet",
			},
		},
	})
	assert.NoError(t, res.Error)
	service.AssertExpectations(t)
}

func TestDockerUseCase_RemoveNetwork(t *testing.T) {
	service := new(mockService.DockerService)
	service.On("CreateClient", mock.Anything, mock.Anything).Return(nil, nil).Once()
	service.On("RemoveNetwork", mock.Anything, mock.Anything, mock.Anything).Return(
		entity.NewSuccessResult()).Once()

	cmdService := new(mockService.CommandService)
	cmdService.On("CheckDependenciesExecuted", mock.Anything).Return(true, nil)

	usecase, err := NewDockerUseCase(service, cmdService)
	assert.NoError(t, err)

	res := usecase.Run(command.Command{
		ID:        "TEST",
		Timestamp: time.Now().Unix() - 5,
		Timeout:   0,
		Target:    command.Target{IP: "0.0.0.0"},
		Order: command.Order{
			Type: "removeNetwork",
			Payload: map[string]interface{}{
				"name": "test",
			},
		},
	})
	assert.NoError(t, res.Error)
	assert.True(t, service.AssertNumberOfCalls(t, "RemoveNetwork", 1))
	service.AssertExpectations(t)
}

func TestDockerUseCase_Run_CreateVolume(t *testing.T) {
	service := new(mockService.DockerService)
	service.On("CreateClient", mock.Anything, mock.Anything).Return(nil, nil)
	service.On("CreateVolume", mock.Anything, mock.Anything, mock.Anything).Return(entity.Result{Type: entity.SuccessType})

	cmdService := new(mockService.CommandService)
	cmdService.On("CheckDependenciesExecuted", mock.Anything).Return(true, nil)

	usecase, err := NewDockerUseCase(service, cmdService)
	assert.NoError(t, err)

	res := usecase.Run(command.Command{
		ID:        "TEST",
		Timestamp: time.Now().Unix() - 5,
		Timeout:   0,
		Target:    command.Target{IP: "0.0.0.0"},
		Order: command.Order{
			Type:    "createVolume",
			Payload: map[string]interface{}{"volume": map[string]string{"name": "test"}},
		},
	})
	assert.Equal(t, res.Error, nil)
	assert.True(t, service.AssertNumberOfCalls(t, "CreateVolume", 1))
}

func TestDockerUseCase_Run_RemoveVolume(t *testing.T) {
	service := new(mockService.DockerService)
	service.On("CreateClient", mock.Anything, mock.Anything).Return(nil, nil)
	service.On("RemoveVolume", mock.Anything, mock.Anything, mock.Anything).Return(entity.Result{Type: entity.SuccessType})

	cmdService := new(mockService.CommandService)
	cmdService.On("CheckDependenciesExecuted", mock.Anything).Return(true, nil)

	usecase, err := NewDockerUseCase(service, cmdService)
	assert.NoError(t, err)

	res := usecase.Run(command.Command{
		ID:        "TEST",
		Timestamp: time.Now().Unix() - 5,
		Timeout:   0,
		Target:    command.Target{IP: "0.0.0.0"},
		Order: command.Order{
			Type:    "removeVolume",
			Payload: map[string]interface{}{"name": "test"},
		},
	})
	assert.Equal(t, res.Error, nil)
	assert.True(t, service.AssertNumberOfCalls(t, "RemoveVolume", 1))
}

func TestDockerUseCase_Run_PutFileInContainer(t *testing.T) {
	service := new(mockService.DockerService)
	service.On("CreateClient", mock.Anything, mock.Anything).Return(nil, nil)
	service.On("PlaceFileInContainer", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(entity.Result{Type: entity.SuccessType})

	cmdService := new(mockService.CommandService)
	cmdService.On("CheckDependenciesExecuted", mock.Anything).Return(true, nil)

	usecase, err := NewDockerUseCase(service, cmdService)
	assert.NoError(t, err)

	res := usecase.Run(command.Command{
		ID:        "TEST",
		Timestamp: time.Now().Unix() - 5,
		Timeout:   0,
		Target:    command.Target{IP: "0.0.0.0"},
		Order: command.Order{
			Type:    "putFileInContainer",
			Payload: map[string]interface{}{"container": "test", "file": map[string]interface{}{"path": "test/path/", "data": []byte("content")}},
		},
	})
	assert.Equal(t, res.Error, nil)
	assert.True(t, service.AssertNumberOfCalls(t, "PlaceFileInContainer", 1))
}

func TestDockerUseCase_Run_Emulation(t *testing.T) {
	service := new(mockService.DockerService)
	service.On("CreateClient", mock.Anything, mock.Anything).Return(nil, nil)
	service.On("Emulation", mock.Anything, mock.Anything, mock.Anything).Return(entity.Result{Type: entity.SuccessType})

	cmdService := new(mockService.CommandService)
	cmdService.On("CheckDependenciesExecuted", mock.Anything).Return(true, nil)

	usecase, err := NewDockerUseCase(service, cmdService)
	assert.NoError(t, err)

	res := usecase.Run(command.Command{
		ID:        "TEST",
		Timestamp: time.Now().Unix() - 5,
		Timeout:   0,
		Target:    command.Target{IP: "0.0.0.0"},
		Order: command.Order{
			Type: "emulation",
			Payload: map[string]interface{}{
				"limit":     4,
				"loss":      float64(2),
				"delay":     4,
				"rate":      "0",
				"duplicate": float64(2),
				"corrupt":   float64(0),
				"reorder":   float64(0)},
		},
	})
	assert.Equal(t, res.Error, nil)
	assert.True(t, service.AssertNumberOfCalls(t, "Emulation", 1))
}

func TestDockerUseCase_Execute_CreateContainer(t *testing.T) {
	service := new(mockService.DockerService)
	service.On("CreateClient", mock.Anything, mock.Anything).Return(nil, nil)
	service.On("CreateContainer", mock.Anything, mock.Anything, mock.Anything).Return(entity.Result{Type: entity.SuccessType})

	usecase, err := NewDockerUseCase(service, nil)
	assert.NoError(t, err)

	res := usecase.Execute(context.TODO(), command.Command{
		ID:        "TEST",
		Timestamp: 1234567,
		Timeout:   5 * time.Second,
		Target:    command.Target{IP: "0.0.0.0"},
		Order: command.Order{
			Type:    "createContainer",
			Payload: map[string]interface{}{},
		},
	})
	assert.Equal(t, res.Error, nil)
	assert.True(t, service.AssertNumberOfCalls(t, "CreateContainer", 1))
}

func TestDockerUseCase_Execute_StartContainer(t *testing.T) {
	service := new(mockService.DockerService)
	service.On("CreateClient", mock.Anything, mock.Anything).Return(nil, nil)
	service.On("StartContainer", mock.Anything, mock.Anything, mock.Anything).Return(entity.Result{Type: entity.SuccessType})

	usecase, err := NewDockerUseCase(service, nil)
	assert.NoError(t, err)

	res := usecase.Execute(context.TODO(), command.Command{
		ID:        "TEST",
		Timestamp: 1234567,
		Timeout:   5 * time.Second,
		Target:    command.Target{IP: "0.0.0.0"},
		Order: command.Order{
			Type:    "startContainer",
			Payload: map[string]interface{}{"name": "test"},
		},
	})
	assert.Equal(t, res.Error, nil)
	assert.True(t, service.AssertNumberOfCalls(t, "StartContainer", 1))
}

func TestDockerUseCase_Execute_RemoveContainer_Success(t *testing.T) {
	service := new(mockService.DockerService)
	containerName := "testContainer"
	service.On("CreateClient", mock.Anything, mock.Anything).Return(nil, nil).Once()
	service.On("RemoveContainer", mock.Anything, mock.Anything, mock.Anything).Return(
		entity.Result{Type: entity.SuccessType}).Run(
		func(args mock.Arguments) {

			require.Len(t, args, 3)
			assert.NotNil(t, args.Get(0))
			assert.Nil(t, args.Get(1))
			assert.Equal(t, containerName, args.String(2))

		}).Once()

	usecase, err := NewDockerUseCase(service, nil)
	assert.NoError(t, err)

	res := usecase.Execute(context.TODO(), command.Command{
		ID:        "TEST",
		Timestamp: 1234567,
		Timeout:   5 * time.Second,
		Target:    command.Target{IP: "0.0.0.0"},
		Order: command.Order{
			Type: "removeContainer",
			Payload: map[string]interface{}{
				"name": containerName,
			},
		},
	})
	assert.Equal(t, res.Error, nil)
	service.AssertExpectations(t)
}

func TestDockerUseCase_Execute_RemoveContainer_Failure(t *testing.T) {
	service := new(mockService.DockerService)
	service.On("CreateClient", mock.Anything, mock.Anything).Return(nil, nil).Once()

	usecase, err := NewDockerUseCase(service, nil)
	assert.NoError(t, err)

	res := usecase.Execute(context.TODO(), command.Command{
		ID:        "TEST",
		Timestamp: 1234567,
		Timeout:   5 * time.Second,
		Target:    command.Target{IP: "0.0.0.0"},
		Order: command.Order{
			Type:    "removeContainer",
			Payload: map[string]interface{}{},
		},
	})
	assert.Error(t, res.Error)
	service.AssertExpectations(t)
}

func TestDockerUseCase_Execute_CreateNetwork(t *testing.T) {
	service := new(mockService.DockerService)
	service.On("CreateClient", mock.Anything, mock.Anything).Return(nil, nil)
	service.On("CreateNetwork", mock.Anything, mock.Anything, mock.Anything).Return(entity.Result{Type: entity.SuccessType})

	usecase, err := NewDockerUseCase(service, nil)
	assert.NoError(t, err)

	res := usecase.Execute(context.TODO(), command.Command{
		ID:        "TEST",
		Timestamp: 1234567,
		Timeout:   5 * time.Second,
		Target:    command.Target{IP: "0.0.0.0"},
		Order: command.Order{
			Type:    "createNetwork",
			Payload: map[string]interface{}{},
		},
	})
	assert.Equal(t, res.Error, nil)
	assert.True(t, service.AssertNumberOfCalls(t, "CreateNetwork", 1))
}

func TestDockerUseCase_Execute_AttachNetwork_Success(t *testing.T) {
	containerName := "tester"
	networkName := "testnet"
	service := new(mockService.DockerService)
	service.On("CreateClient", mock.Anything, mock.Anything).Return(nil, nil).Once()
	service.On("AttachNetwork", mock.Anything, mock.Anything, mock.Anything,
		mock.Anything).Return(entity.Result{Type: entity.SuccessType}).Run(
		func(args mock.Arguments) {

			require.Len(t, args, 4)
			assert.NotNil(t, args.Get(0))
			assert.Nil(t, args.Get(1))
			assert.Equal(t, networkName, args.String(2))
			assert.Equal(t, containerName, args.String(3))

		}).Once()

	usecase, err := NewDockerUseCase(service, nil)
	assert.NoError(t, err)

	res := usecase.Execute(context.TODO(), command.Command{
		ID:        "TEST",
		Timestamp: 1234567,
		Timeout:   5 * time.Second,
		Target:    command.Target{IP: "0.0.0.0"},
		Order: command.Order{
			Type: "attachNetwork",
			Payload: map[string]interface{}{
				"network":   networkName,
				"container": containerName,
			},
		},
	})
	assert.Equal(t, res.Error, nil)
	service.AssertExpectations(t)
}

func TestDockerUseCase_Execute_AttachNetwork_Failure(t *testing.T) {
	service := new(mockService.DockerService)
	service.On("CreateClient", mock.Anything, mock.Anything).Return(nil, nil).Once()

	usecase, err := NewDockerUseCase(service, nil)
	assert.NoError(t, err)

	res := usecase.Execute(context.TODO(), command.Command{
		ID:        "TEST",
		Timestamp: 1234567,
		Timeout:   5 * time.Second,
		Target:    command.Target{IP: "0.0.0.0"},
		Order: command.Order{
			Type:    "attachNetwork",
			Payload: map[string]interface{}{},
		},
	})
	assert.Error(t, res.Error)
	service.AssertExpectations(t)
}

func TestDockerUseCase_Execute_CreateVolume(t *testing.T) {
	service := new(mockService.DockerService)
	service.On("CreateClient", mock.Anything, mock.Anything).Return(nil, nil)
	service.On("CreateVolume", mock.Anything, mock.Anything, mock.Anything).Return(entity.Result{Type: entity.SuccessType})

	usecase, err := NewDockerUseCase(service, nil)
	assert.NoError(t, err)

	res := usecase.Execute(context.TODO(), command.Command{
		ID:        "TEST",
		Timestamp: 1234567,
		Timeout:   5 * time.Second,
		Target:    command.Target{IP: "0.0.0.0"},
		Order: command.Order{
			Type:    "createVolume",
			Payload: map[string]interface{}{},
		},
	})
	assert.Equal(t, res.Error, nil)
	assert.True(t, service.AssertNumberOfCalls(t, "CreateVolume", 1))
}

func TestDockerUseCase_Execute_RemoveVolume_Success(t *testing.T) {
	volumeName := "vol"
	service := new(mockService.DockerService)
	service.On("CreateClient", mock.Anything, mock.Anything).Return(nil, nil)
	service.On("RemoveVolume", mock.Anything, mock.Anything, mock.Anything).Return(
		entity.Result{Type: entity.SuccessType}).Run(
		func(args mock.Arguments) {

			require.Len(t, args, 3)
			assert.NotNil(t, args.Get(0))
			assert.Nil(t, args.Get(1))
			assert.Equal(t, volumeName, args.String(2))

		}).Once()

	usecase, err := NewDockerUseCase(service, nil)
	assert.NoError(t, err)

	res := usecase.Execute(context.TODO(), command.Command{
		ID:        "TEST",
		Timestamp: 1234567,
		Timeout:   5 * time.Second,
		Target:    command.Target{IP: "0.0.0.0"},
		Order: command.Order{
			Type:    "removeVolume",
			Payload: map[string]interface{}{"name": volumeName},
		},
	})
	assert.NoError(t, res.Error)
	service.AssertExpectations(t)
}

func TestDockerUseCase_Execute_RemoveVolume_Failure(t *testing.T) {
	service := new(mockService.DockerService)
	service.On("CreateClient", mock.Anything, mock.Anything).Return(nil, nil).Once()

	usecase, err := NewDockerUseCase(service, nil)
	assert.NoError(t, err)

	res := usecase.Execute(context.TODO(), command.Command{
		ID:        "TEST",
		Timestamp: 1234567,
		Timeout:   5 * time.Second,
		Target:    command.Target{IP: "0.0.0.0"},
		Order: command.Order{
			Type:    "removeVolume",
			Payload: map[string]interface{}{},
		},
	})
	assert.Error(t, res.Error, nil)
	service.AssertExpectations(t)
}

func TestDockerUseCase_Execute_PutFileInContainer_Success(t *testing.T) {
	mockFile := map[string]interface{}{"mode": 0777, "destination": "/test/path/", "data": []byte("contents")}
	containerName := "tester"
	service := new(mockService.DockerService)
	service.On("CreateClient", mock.Anything, mock.Anything).Return(nil, nil)
	service.On("PlaceFileInContainer", mock.Anything, mock.Anything, mock.Anything,
		mock.Anything).Return(entity.Result{Type: entity.SuccessType}).Run(
		func(args mock.Arguments) {

			require.Len(t, args, 4)
			assert.NotNil(t, args.Get(0))
			assert.Nil(t, args.Get(1))
			assert.Equal(t, containerName, args.String(2))
			file, ok := args.Get(3).(command.File)
			require.True(t, ok)
			assert.Equal(t, int64(0777), file.Mode)
			assert.Equal(t, mockFile["destination"], file.Destination)
			assert.ElementsMatch(t, mockFile["data"], file.Data)
		}).Once()

	usecase, err := NewDockerUseCase(service, nil)
	assert.NoError(t, err)

	res := usecase.Execute(context.TODO(), command.Command{
		ID:        "TEST",
		Timestamp: 1234567,
		Timeout:   5 * time.Second,
		Target:    command.Target{IP: "0.0.0.0"},
		Order: command.Order{
			Type:    "putFileInContainer",
			Payload: map[string]interface{}{"container": containerName, "file": mockFile},
		},
	})
	assert.Equal(t, res.Error, nil)
	service.AssertExpectations(t)
}

func TestDockerUseCase_putFileInContainerShim_MissingFields(t *testing.T) {
	cmdService := new(mockService.CommandService)
	service := new(mockService.DockerService)
	service.On("PlaceFileInContainer", mock.Anything, mock.Anything, mock.Anything,
		mock.Anything).Return(entity.Result{Type: entity.SuccessType})

	duc := &dockerUseCase{service: service, cmdService: cmdService}

	cmd := command.Command{
		Order: command.Order{
			Payload: map[string]interface{}{},
		},
	}
	res := duc.putFileInContainerShim(nil, nil, cmd)
	assert.Error(t, res.Error)

	cmd.Order.Payload = map[string]interface{}{"file": map[string]interface{}{"destination": 2, "Data": []byte("contents")}}
	res = duc.putFileInContainerShim(nil, nil, cmd)
	assert.Error(t, res.Error)

	cmd.Order.Payload = map[string]interface{}{"file": map[string]interface{}{
		"destination": "/test/path/",
		"data":        []byte("contents")}}
	res = duc.putFileInContainerShim(nil, nil, cmd)
	assert.Error(t, res.Error)
}

func TestDockerUseCase_Execute_Emulation(t *testing.T) {
	service := new(mockService.DockerService)
	service.On("CreateClient", mock.Anything, mock.Anything).Return(nil, nil)
	service.On("Emulation", mock.Anything, mock.Anything, mock.Anything).Return(entity.Result{Type: entity.SuccessType})

	usecase, err := NewDockerUseCase(service, nil)
	assert.NoError(t, err)

	res := usecase.Execute(context.TODO(), command.Command{
		ID:        "TEST",
		Timestamp: 1234567,
		Timeout:   5 * time.Second,
		Target:    command.Target{IP: "0.0.0.0"},
		Order: command.Order{
			Type: "emulation",
			Payload: map[string]interface{}{
				"limit":     4,
				"loss":      float64(2),
				"delay":     4,
				"rate":      "0",
				"duplicate": float64(2),
				"corrupt":   float64(0),
				"reorder":   float64(0)},
		},
	})
	assert.Equal(t, res.Error, nil)
	assert.True(t, service.AssertNumberOfCalls(t, "Emulation", 1))
}
