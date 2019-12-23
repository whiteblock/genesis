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

	mockService "github.com/whiteblock/genesis/mocks/pkg/service"
	mockValid "github.com/whiteblock/genesis/mocks/pkg/validator"
	"github.com/whiteblock/genesis/pkg/entity"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/whiteblock/definition/command"
)

const (
	testFileID = "foo"
)

func TestNewDockerUseCase(t *testing.T) {
	duc := NewDockerUseCase(nil, nil, logrus.New())
	assert.NotNil(t, duc)
}

func TestDockerUseCase_validationCheck_success(t *testing.T) {
	cmd := command.Command{
		Target: command.Target{IP: "127.0.0.1"},
	}
	duc := NewDockerUseCase(nil, nil, logrus.New())
	_, ok := duc.(*dockerUseCase).validationCheck(cmd)
	assert.True(t, ok)
}

func TestDockerUseCase_validationCheck_failure_special_ip(t *testing.T) {
	cmd := command.Command{
		Target: command.Target{IP: "0.0.0.0"},
	}

	duc := NewDockerUseCase(nil, nil, logrus.New())
	res, ok := duc.(*dockerUseCase).validationCheck(cmd)
	assert.False(t, ok)
	assert.Error(t, res.Error)
}

func TestDockerUseCase_validationCheck_failure_no_ip(t *testing.T) {
	cmd := command.Command{}
	duc := NewDockerUseCase(nil, nil, logrus.New())
	res, ok := duc.(*dockerUseCase).validationCheck(cmd)
	assert.False(t, ok)
	assert.Error(t, res.Error)
}

func TestDockerUseCase_Run_Failure_CreateClient(t *testing.T) {
	service := new(mockService.DockerService)
	service.On("CreateClient", mock.Anything, mock.Anything).Return(nil, fmt.Errorf("err")).Once()

	usecase := NewDockerUseCase(service, nil, logrus.New())

	res := usecase.Run(command.Command{Target: command.Target{IP: "127.0.0.1"}})
	assert.Error(t, res.Error)
	service.AssertExpectations(t)
}

func TestDockerUseCase_Run_Failure_Invalid_IP(t *testing.T) {
	usecase := NewDockerUseCase(nil, nil, logrus.New())

	res := usecase.Run(command.Command{Target: command.Target{IP: "0.0.0.0"}})
	assert.Error(t, res.Error)
}

func TestDockerUseCase_Run_CreateContainer_Success(t *testing.T) {
	service := new(mockService.DockerService)
	service.On("CreateClient", mock.Anything, mock.Anything).Return(nil, nil).Once()
	service.On("CreateContainer", mock.Anything, mock.Anything, mock.Anything).Return(
		entity.Result{Type: entity.SuccessType}).Once()

	valid := new(mockValid.OrderValidator)
	valid.On("ValidateContainer", mock.Anything).Return(nil).Once()

	usecase := NewDockerUseCase(service, valid, logrus.New())

	res := usecase.Run(command.Command{
		ID:     "TEST",
		Target: command.Target{IP: "127.0.0.1"},
		Order: command.Order{
			Type:    "createContainer",
			Payload: map[string]interface{}{},
		},
	})
	assert.NoError(t, res.Error)
	service.AssertExpectations(t)
	valid.AssertExpectations(t)

}

func TestDockerUseCase_Run_CreateContainer_Failure(t *testing.T) {
	service := new(mockService.DockerService)
	service.On("CreateClient", mock.Anything, mock.Anything).Return(nil, nil).Once()

	valid := new(mockValid.OrderValidator)
	valid.On("ValidateContainer", mock.Anything).Return(fmt.Errorf("err")).Once()

	usecase := NewDockerUseCase(service, valid, logrus.New())

	res := usecase.Run(command.Command{
		ID:     "TEST",
		Target: command.Target{IP: "127.0.0.1"},
		Order: command.Order{
			Type:    "createContainer",
			Payload: map[string]interface{}{},
		},
	})
	assert.Error(t, res.Error)
	service.AssertExpectations(t)

	valid.AssertExpectations(t)
}

func TestDockerUseCase_Run_StartContainer_Success(t *testing.T) {
	service := new(mockService.DockerService)
	service.On("CreateClient", mock.Anything, mock.Anything).Return(nil, nil).Once()
	service.On("StartContainer", mock.Anything, mock.Anything, mock.Anything).Return(
		entity.Result{Type: entity.SuccessType}).Once()

	valid := new(mockValid.OrderValidator)

	usecase := NewDockerUseCase(service, valid, logrus.New())

	res := usecase.Run(command.Command{
		ID:     "TEST",
		Target: command.Target{IP: "127.0.0.1"},
		Order: command.Order{
			Type:    command.Startcontainer,
			Payload: command.StartContainer{Name: "test"},
		},
	})
	assert.NoError(t, res.Error)
	service.AssertExpectations(t)

}

func TestDockerUseCase_Run_StartContainer_Failure_ExtraField(t *testing.T) {
	service := new(mockService.DockerService)
	service.On("CreateClient", mock.Anything, mock.Anything).Return(nil, nil).Once()

	valid := new(mockValid.OrderValidator)

	usecase := NewDockerUseCase(service, valid, logrus.New())

	res := usecase.Run(command.Command{
		ID:     "TEST",
		Target: command.Target{IP: "127.0.0.1"},
		Order: command.Order{
			Type:    command.Startcontainer,
			Payload: map[string]interface{}{"name": "test", "invalid": "field"},
		},
	})
	assert.Error(t, res.Error)
	service.AssertExpectations(t)

}

func TestDockerUseCase_Run_StartContainer_Failure_EmptyName(t *testing.T) {
	service := new(mockService.DockerService)
	service.On("CreateClient", mock.Anything, mock.Anything).Return(nil, nil).Once()

	valid := new(mockValid.OrderValidator)

	usecase := NewDockerUseCase(service, valid, logrus.New())

	res := usecase.Run(command.Command{
		ID:     "TEST",
		Target: command.Target{IP: "127.0.0.1"},
		Order: command.Order{
			Type:    command.Startcontainer,
			Payload: command.StartContainer{Name: ""},
		},
	})
	assert.Error(t, res.Error)
	service.AssertExpectations(t)
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
			assert.NotNil(t, args.Get(1))
			assert.Equal(t, containerName, args.String(2))

		}).Once()

	valid := new(mockValid.OrderValidator)

	usecase := NewDockerUseCase(service, valid, logrus.New())

	res := usecase.Run(command.Command{
		ID:     "TEST",
		Target: command.Target{IP: "127.0.0.1"},
		Order: command.Order{
			Type: command.Removecontainer,
			Payload: command.SimpleName{
				Name: containerName,
			},
		},
	})
	assert.NoError(t, res.Error)
	service.AssertExpectations(t)
}

func TestDockerUseCase_Run_RemoveContainer_Failure_EmptyName(t *testing.T) {
	service := new(mockService.DockerService)
	service.On("CreateClient", mock.Anything, mock.Anything).Return(nil, nil).Once()

	valid := new(mockValid.OrderValidator)

	usecase := NewDockerUseCase(service, valid, logrus.New())

	res := usecase.Run(command.Command{
		ID:     "TEST",
		Target: command.Target{IP: "127.0.0.1"},
		Order: command.Order{
			Type:    command.Removecontainer,
			Payload: command.SimpleName{},
		},
	})
	assert.Error(t, res.Error)
	service.AssertExpectations(t)

}

func TestDockerUseCase_Run_RemoveContainer_Failure_ExtraField(t *testing.T) {
	service := new(mockService.DockerService)
	service.On("CreateClient", mock.Anything, mock.Anything).Return(nil, nil).Once()

	valid := new(mockValid.OrderValidator)

	usecase := NewDockerUseCase(service, valid, logrus.New())

	res := usecase.Run(command.Command{
		ID:     "TEST",
		Target: command.Target{IP: "127.0.0.1"},
		Order: command.Order{
			Type:    command.Removecontainer,
			Payload: map[string]interface{}{"name": "tester", "extra": "field"},
		},
	})
	assert.Error(t, res.Error)
	service.AssertExpectations(t)
}

func TestDockerUseCase_Run_CreateNetwork(t *testing.T) {
	service := new(mockService.DockerService)
	service.On("CreateClient", mock.Anything, mock.Anything).Return(nil, nil).Once()
	service.On("CreateNetwork", mock.Anything, mock.Anything, mock.Anything).Return(
		entity.Result{Type: entity.SuccessType}).Once()

	valid := new(mockValid.OrderValidator)

	usecase := NewDockerUseCase(service, valid, logrus.New())

	res := usecase.Run(command.Command{
		ID:     "TEST",
		Target: command.Target{IP: "127.0.0.1"},
		Order: command.Order{
			Type:    "createNetwork",
			Payload: map[string]interface{}{"name": "test"},
		},
	})
	assert.NoError(t, res.Error)
	service.AssertExpectations(t)

}

func TestDockerUseCase_Run_AttachNetwork_Success(t *testing.T) {
	service := new(mockService.DockerService)
	service.On("CreateClient", mock.Anything, mock.Anything).Return(nil, nil).Once()
	service.On("AttachNetwork", mock.Anything, mock.Anything, mock.Anything,
		mock.Anything, mock.Anything).Return(entity.Result{Type: entity.SuccessType}).Once()

	valid := new(mockValid.OrderValidator)

	usecase := NewDockerUseCase(service, valid, logrus.New())

	res := usecase.Run(command.Command{
		ID:     "TEST",
		Target: command.Target{IP: "127.0.0.1"},
		Order: command.Order{
			Type: command.Attachnetwork,
			Payload: command.ContainerNetwork{
				ContainerName: "test",
				Network:       "testnet",
			},
		},
	})
	assert.NoError(t, res.Error)
	service.AssertExpectations(t)

}

func TestDockerUseCase_Run_AttachNetwork_Failure_EmptyContainerName(t *testing.T) {
	service := new(mockService.DockerService)
	service.On("CreateClient", mock.Anything, mock.Anything).Return(nil, nil).Once()

	valid := new(mockValid.OrderValidator)

	usecase := NewDockerUseCase(service, valid, logrus.New())

	res := usecase.Run(command.Command{
		ID:     "TEST",
		Target: command.Target{IP: "127.0.0.1"},
		Order: command.Order{
			Type: command.Attachnetwork,
			Payload: command.ContainerNetwork{
				ContainerName: "",
				Network:       "testnet",
			},
		},
	})
	assert.Error(t, res.Error)
	service.AssertExpectations(t)

}

func TestDockerUseCase_Run_AttachNetwork_Failure_EmptyNetworkName(t *testing.T) {
	service := new(mockService.DockerService)
	service.On("CreateClient", mock.Anything, mock.Anything).Return(nil, nil).Once()

	valid := new(mockValid.OrderValidator)

	usecase := NewDockerUseCase(service, valid, logrus.New())

	res := usecase.Run(command.Command{
		ID:     "TEST",
		Target: command.Target{IP: "127.0.0.1"},
		Order: command.Order{
			Type: command.Attachnetwork,
			Payload: command.ContainerNetwork{
				ContainerName: "tester",
				Network:       "",
			},
		},
	})
	assert.Error(t, res.Error)
	service.AssertExpectations(t)

}

func TestDockerUseCase_Run_AttachNetwork_Failure_ExtraField(t *testing.T) {
	service := new(mockService.DockerService)
	service.On("CreateClient", mock.Anything, mock.Anything).Return(nil, nil).Once()

	valid := new(mockValid.OrderValidator)

	usecase := NewDockerUseCase(service, valid, logrus.New())

	res := usecase.Run(command.Command{
		ID:     "TEST",
		Target: command.Target{IP: "127.0.0.1"},
		Order: command.Order{
			Type: command.Attachnetwork,
			Payload: map[string]interface{}{
				"container": "test",
				"network":   "testnet",
				"extra":     "field",
			},
		},
	})
	assert.Error(t, res.Error)
	service.AssertExpectations(t)

}

func TestDockerUseCase_Run_DetachNetwork(t *testing.T) {
	service := new(mockService.DockerService)
	service.On("CreateClient", mock.Anything, mock.Anything).Return(nil, nil).Once()
	service.On("DetachNetwork", mock.Anything, mock.Anything, mock.Anything,
		mock.Anything, mock.Anything).Return(entity.NewSuccessResult()).Once()

	valid := new(mockValid.OrderValidator)

	usecase := NewDockerUseCase(service, valid, logrus.New())

	res := usecase.Run(command.Command{
		ID:     "TEST",
		Target: command.Target{IP: "127.0.0.1"},
		Order: command.Order{
			Type:    "detachNetwork",
			Payload: command.ContainerNetwork{ContainerName: "test", Network: "testnet"},
		},
	})
	assert.NoError(t, res.Error)
	service.AssertExpectations(t)

}

func TestDockerUseCase_Run_DetachNetwork_Failure(t *testing.T) {
	service := new(mockService.DockerService)
	service.On("CreateClient", mock.Anything, mock.Anything).Return(nil, nil).Once()

	valid := new(mockValid.OrderValidator)

	usecase := NewDockerUseCase(service, valid, logrus.New())

	res := usecase.Run(command.Command{
		ID:     "TEST",
		Target: command.Target{IP: "127.0.0.1"},
		Order: command.Order{
			Type:    "detachNetwork",
			Payload: map[string]interface{}{"invalid": "value"},
		},
	})
	assert.Error(t, res.Error)
	service.AssertExpectations(t)

}

func TestDockerUseCase_RemoveNetwork_Success(t *testing.T) {
	service := new(mockService.DockerService)
	service.On("CreateClient", mock.Anything, mock.Anything).Return(nil, nil).Once()
	service.On("RemoveNetwork", mock.Anything, mock.Anything, mock.Anything).Return(
		entity.NewSuccessResult()).Once()

	valid := new(mockValid.OrderValidator)

	usecase := NewDockerUseCase(service, valid, logrus.New())

	res := usecase.Run(command.Command{
		ID:     "TEST",
		Target: command.Target{IP: "127.0.0.1"},
		Order: command.Order{
			Type: command.Removenetwork,
			Payload: command.SimpleName{
				Name: "test",
			},
		},
	})
	assert.NoError(t, res.Error)
	service.AssertExpectations(t)

}

func TestDockerUseCase_RemoveNetwork_Failure_EmptyName(t *testing.T) {
	service := new(mockService.DockerService)
	service.On("CreateClient", mock.Anything, mock.Anything).Return(nil, nil).Once()

	valid := new(mockValid.OrderValidator)

	usecase := NewDockerUseCase(service, valid, logrus.New())

	res := usecase.Run(command.Command{
		ID:     "TEST",
		Target: command.Target{IP: "127.0.0.1"},
		Order: command.Order{
			Type: command.Removenetwork,
			Payload: command.SimpleName{
				Name: "",
			},
		},
	})
	assert.Error(t, res.Error)
	service.AssertExpectations(t)

}

func TestDockerUseCase_RemoveNetwork_Failure_ExtraField(t *testing.T) {
	service := new(mockService.DockerService)
	service.On("CreateClient", mock.Anything, mock.Anything).Return(nil, nil).Once()

	valid := new(mockValid.OrderValidator)

	usecase := NewDockerUseCase(service, valid, logrus.New())

	res := usecase.Run(command.Command{
		ID:     "TEST",
		Target: command.Target{IP: "127.0.0.1"},
		Order: command.Order{
			Type: command.Removenetwork,
			Payload: map[string]interface{}{
				"name":    "test",
				"invalid": "field",
			},
		},
	})
	assert.Error(t, res.Error)
	service.AssertExpectations(t)

}

func TestDockerUseCase_Run_CreateVolume(t *testing.T) {
	service := new(mockService.DockerService)
	service.On("CreateClient", mock.Anything, mock.Anything).Return(nil, nil).Once()
	service.On("CreateVolume", mock.Anything, mock.Anything, mock.Anything).Return(
		entity.Result{Type: entity.SuccessType}).Once()

	valid := new(mockValid.OrderValidator)

	usecase := NewDockerUseCase(service, valid, logrus.New())

	res := usecase.Run(command.Command{
		ID:     "TEST",
		Target: command.Target{IP: "127.0.0.1"},
		Order: command.Order{
			Type: command.Createvolume,
			Payload: command.Volume{
				Name: "test",
			},
		},
	})
	assert.NoError(t, res.Error)
	service.AssertExpectations(t)

}

func TestDockerUseCase_Run_RemoveVolume(t *testing.T) {
	service := new(mockService.DockerService)
	service.On("CreateClient", mock.Anything, mock.Anything).Return(nil, nil)
	service.On("RemoveVolume", mock.Anything, mock.Anything, mock.Anything).Return(entity.Result{Type: entity.SuccessType})

	valid := new(mockValid.OrderValidator)

	usecase := NewDockerUseCase(service, valid, logrus.New())

	res := usecase.Run(command.Command{
		ID:     "TEST",
		Target: command.Target{IP: "127.0.0.1"},
		Order: command.Order{
			Type:    "removeVolume",
			Payload: command.SimpleName{Name: "test"},
		},
	})
	assert.NoError(t, res.Error)
	assert.True(t, service.AssertNumberOfCalls(t, "RemoveVolume", 1))
}

func TestDockerUseCase_Run_PutFileInContainer(t *testing.T) {
	service := new(mockService.DockerService)
	service.On("CreateClient", mock.Anything, mock.Anything).Return(nil, nil).Once()
	service.On("PlaceFileInContainer", mock.Anything, mock.Anything, mock.Anything, mock.Anything,
		mock.Anything).Return(entity.Result{Type: entity.SuccessType}).Once()

	valid := new(mockValid.OrderValidator)

	usecase := NewDockerUseCase(service, valid, logrus.New())

	res := usecase.Run(command.Command{
		ID:     "TEST",
		Target: command.Target{IP: "127.0.0.1"},
		Order: command.Order{
			Type: command.Putfileincontainer,
			Payload: command.FileAndContainer{
				ContainerName: "tester",
				File:          command.File{Destination: "test/path/", ID: testFileID},
			},
		},
	})
	assert.NoError(t, res.Error)
	service.AssertExpectations(t)
}

func TestDockerUseCase_Run_Emulation(t *testing.T) {
	service := new(mockService.DockerService)
	service.On("CreateClient", mock.Anything, mock.Anything).Return(nil, nil).Once()
	service.On("Emulation", mock.Anything, mock.Anything, mock.Anything).Return(
		entity.Result{Type: entity.SuccessType}).Once()

	valid := new(mockValid.OrderValidator)

	usecase := NewDockerUseCase(service, valid, logrus.New())

	res := usecase.Run(command.Command{
		ID:     "TEST",
		Target: command.Target{IP: "127.0.0.1"},
		Order: command.Order{
			Type: command.Emulation,
			Payload: command.Netconf{
				Container:   "test",
				Network:     "testnet",
				Limit:       4,
				Loss:        float64(2),
				Delay:       4,
				Rate:        "0",
				Duplication: float64(2),
				Corrupt:     float64(0),
				Reorder:     float64(0),
			},
		},
	})
	assert.NoError(t, res.Error)
	service.AssertExpectations(t)

}

func TestDockerUseCase_Execute_CreateContainer(t *testing.T) {
	service := new(mockService.DockerService)
	service.On("CreateClient", mock.Anything, mock.Anything).Return(nil, nil)
	service.On("CreateContainer", mock.Anything, mock.Anything, mock.Anything).Return(entity.Result{Type: entity.SuccessType})

	valid := new(mockValid.OrderValidator)
	valid.On("ValidateContainer", mock.Anything).Return(nil).Once()

	usecase := NewDockerUseCase(service, valid, logrus.New())

	res := usecase.Execute(context.TODO(), command.Command{
		ID:     "TEST",
		Target: command.Target{IP: "127.0.0.1"},
		Order: command.Order{
			Type:    command.Createcontainer,
			Payload: map[string]interface{}{},
		},
	})
	assert.NoError(t, res.Error)

	service.AssertExpectations(t)
	valid.AssertExpectations(t)
}

func TestDockerUseCase_Execute_StartContainer(t *testing.T) {
	service := new(mockService.DockerService)
	service.On("CreateClient", mock.Anything, mock.Anything).Return(nil, nil).Once()
	service.On("StartContainer", mock.Anything, mock.Anything, mock.Anything).Return(entity.Result{Type: entity.SuccessType}).Once()

	usecase := NewDockerUseCase(service, nil, logrus.New())

	res := usecase.Execute(context.TODO(), command.Command{
		ID:     "TEST",
		Target: command.Target{IP: "127.0.0.1"},
		Order: command.Order{
			Type:    command.Startcontainer,
			Payload: command.SimpleName{Name: "test"},
		},
	})
	assert.NoError(t, res.Error)
	service.AssertExpectations(t)
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
			assert.NotNil(t, args.Get(1))
			assert.Equal(t, containerName, args.String(2))

		}).Once()

	usecase := NewDockerUseCase(service, nil, logrus.New())

	res := usecase.Execute(context.TODO(), command.Command{
		ID:     "TEST",
		Target: command.Target{IP: "127.0.0.1"},
		Order: command.Order{
			Type: command.Removecontainer,
			Payload: command.SimpleName{
				Name: containerName,
			},
		},
	})
	assert.NoError(t, res.Error)
	service.AssertExpectations(t)
}

func TestDockerUseCase_Execute_RemoveContainer_Failure(t *testing.T) {
	service := new(mockService.DockerService)
	service.On("CreateClient", mock.Anything, mock.Anything).Return(nil, nil).Once()

	usecase := NewDockerUseCase(service, nil, logrus.New())

	res := usecase.Execute(context.TODO(), command.Command{
		ID:     "TEST",
		Target: command.Target{IP: "127.0.0.1"},
		Order: command.Order{
			Type:    command.Removecontainer,
			Payload: command.SimpleName{},
		},
	})
	assert.Error(t, res.Error)
	service.AssertExpectations(t)
}

func TestDockerUseCase_Execute_CreateNetwork(t *testing.T) {
	service := new(mockService.DockerService)
	service.On("CreateClient", mock.Anything, mock.Anything).Return(nil, nil)
	service.On("CreateNetwork", mock.Anything, mock.Anything, mock.Anything).Return(entity.Result{Type: entity.SuccessType})

	usecase := NewDockerUseCase(service, nil, logrus.New())

	res := usecase.Execute(context.TODO(), command.Command{
		ID:     "TEST",
		Target: command.Target{IP: "127.0.0.1"},
		Order: command.Order{
			Type:    "createNetwork",
			Payload: map[string]interface{}{},
		},
	})
	assert.NoError(t, res.Error)
	assert.True(t, service.AssertNumberOfCalls(t, "CreateNetwork", 1))
}

func TestDockerUseCase_Execute_AttachNetwork_Success(t *testing.T) {
	testCmd := command.Command{
		ID:     "TEST",
		Target: command.Target{IP: "127.0.0.1"},
		Order: command.Order{
			Type:    "attachNetwork",
			Payload: command.ContainerNetwork{ContainerName: "tester", Network: "testnet"},
		},
	}

	service := new(mockService.DockerService)
	service.On("CreateClient", mock.Anything, mock.Anything).Return(nil, nil).Once()
	service.On("AttachNetwork", mock.Anything, mock.Anything,
		mock.Anything).Return(entity.Result{Type: entity.SuccessType}).Run(
		func(args mock.Arguments) {

			require.Len(t, args, 3)
			assert.NotNil(t, args.Get(0))
			assert.NotNil(t, args.Get(1))
			assert.Equal(t, testCmd.Order.Payload, args.Get(2))
		}).Once()

	usecase := NewDockerUseCase(service, nil, logrus.New())

	res := usecase.Execute(context.TODO(), testCmd)
	assert.NoError(t, res.Error)
	service.AssertExpectations(t)
}

func TestDockerUseCase_Execute_AttachNetwork_Failure(t *testing.T) {
	service := new(mockService.DockerService)
	service.On("CreateClient", mock.Anything, mock.Anything).Return(nil, nil).Once()

	usecase := NewDockerUseCase(service, nil, logrus.New())

	res := usecase.Execute(context.TODO(), command.Command{
		ID:     "TEST",
		Target: command.Target{IP: "127.0.0.1"},
		Order: command.Order{
			Type:    command.Attachnetwork,
			Payload: command.ContainerNetwork{Network: "testnet"},
		},
	})
	assert.Error(t, res.Error)
	service.AssertExpectations(t)
}

func TestDockerUseCase_Execute_CreateVolume(t *testing.T) {
	service := new(mockService.DockerService)
	service.On("CreateClient", mock.Anything, mock.Anything).Return(nil, nil)
	service.On("CreateVolume", mock.Anything, mock.Anything, mock.Anything).Return(entity.Result{Type: entity.SuccessType})

	usecase := NewDockerUseCase(service, nil, logrus.New())

	res := usecase.Execute(context.TODO(), command.Command{
		ID:     "TEST",
		Target: command.Target{IP: "127.0.0.1"},
		Order: command.Order{
			Type:    command.Createvolume,
			Payload: command.Volume{},
		},
	})
	assert.NoError(t, res.Error)
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
			assert.NotNil(t, args.Get(1))
			assert.Equal(t, volumeName, args.String(2))

		}).Once()

	usecase := NewDockerUseCase(service, nil, logrus.New())

	res := usecase.Execute(context.TODO(), command.Command{
		ID:     "TEST",
		Target: command.Target{IP: "127.0.0.1"},
		Order: command.Order{
			Type:    command.Removevolume,
			Payload: command.SimpleName{Name: "vol"},
		},
	})
	assert.NoError(t, res.Error)
	service.AssertExpectations(t)
}

func TestDockerUseCase_Execute_RemoveVolume_Failure(t *testing.T) {
	service := new(mockService.DockerService)
	service.On("CreateClient", mock.Anything, mock.Anything).Return(nil, nil).Once()

	usecase := NewDockerUseCase(service, nil, logrus.New())

	res := usecase.Execute(context.TODO(), command.Command{
		ID:     "TEST",
		Target: command.Target{IP: "127.0.0.1"},
		Order: command.Order{
			Type: command.Removevolume,
			Payload: command.FileAndContainer{
				ContainerName: "tester",
				File: command.File{
					Destination: "/test/path/",
					ID:          testFileID,
					Mode:        0777}},
		},
	})
	assert.Error(t, res.Error, nil)
	service.AssertExpectations(t)
}

func TestDockerUseCase_Execute_PutFileInContainer_Success(t *testing.T) {
	mockFile := map[string]interface{}{"mode": 0777,
		"destination": "/test/path/", "id": testFileID}
	containerName := "tester"
	service := new(mockService.DockerService)
	service.On("CreateClient", mock.Anything, mock.Anything).Return(nil, nil)
	service.On("PlaceFileInContainer", mock.Anything, mock.Anything, mock.Anything,
		mock.Anything).Return(entity.Result{Type: entity.SuccessType}).Run(
		func(args mock.Arguments) {

			require.Len(t, args, 4)
			assert.NotNil(t, args.Get(0))
			assert.NotNil(t, args.Get(1))
			assert.Equal(t, containerName, args.String(2))
			file, ok := args.Get(3).(command.File)
			require.True(t, ok)
			assert.Equal(t, int64(0777), file.Mode)
			assert.Equal(t, mockFile["destination"], file.Destination)
			assert.Equal(t, mockFile["id"], file.ID)
		}).Once()

	usecase := NewDockerUseCase(service, nil, logrus.New())

	res := usecase.Execute(context.TODO(), command.Command{
		ID:     "TEST",
		Target: command.Target{IP: "127.0.0.1"},
		Order: command.Order{
			Type: command.Putfileincontainer,
			Payload: command.FileAndContainer{
				ContainerName: "tester",
				File: command.File{
					Destination: "/test/path/",
					ID:          testFileID,
					Mode:        0777}},
		},
	})
	assert.NoError(t, res.Error)
	service.AssertExpectations(t)
}

func TestDockerUseCase_putFileInContainerShim_MissingFields(t *testing.T) {
	valid := new(mockValid.OrderValidator)
	service := new(mockService.DockerService)
	service.On("PlaceFileInContainer", mock.Anything, mock.Anything, mock.Anything,
		mock.Anything).Return(entity.Result{Type: entity.SuccessType})

	duc := &dockerUseCase{service: service, valid: valid}

	cmd := command.Command{
		Order: command.Order{
			Payload: map[string]interface{}{},
		},
	}
	res := duc.putFileInContainerShim(nil, nil, cmd)
	assert.Error(t, res.Error)

	cmd.Order.Payload = map[string]interface{}{"file": map[string]interface{}{
		"destination": 2, "id": testFileID}}
	res = duc.putFileInContainerShim(nil, nil, cmd)
	assert.Error(t, res.Error)

	cmd.Order.Payload = map[string]interface{}{"file": map[string]interface{}{
		"destination": "/test/path/",
		"data":        testFileID}}
	res = duc.putFileInContainerShim(nil, nil, cmd)
	assert.Error(t, res.Error)
}

func TestDockerUseCase_Execute_Emulation(t *testing.T) {
	service := new(mockService.DockerService)
	service.On("CreateClient", mock.Anything, mock.Anything).Return(nil, nil).Once()
	service.On("Emulation", mock.Anything, mock.Anything, mock.Anything).Return(
		entity.Result{Type: entity.SuccessType}).Once()

	usecase := NewDockerUseCase(service, nil, logrus.New())

	res := usecase.Execute(context.TODO(), command.Command{
		ID:     "TEST",
		Target: command.Target{IP: "127.0.0.1"},
		Order: command.Order{
			Type: command.Emulation,
			Payload: command.Netconf{
				Limit:       4,
				Loss:        float64(2),
				Delay:       4,
				Rate:        "0",
				Duplication: float64(2),
				Corrupt:     float64(0),
				Reorder:     float64(0),
			},
		},
	})
	assert.NoError(t, res.Error)
	service.AssertExpectations(t)
}

func TestDockerUseCase_Execute_Emulation_Failure(t *testing.T) {
	service := new(mockService.DockerService)
	service.On("CreateClient", mock.Anything, mock.Anything).Return(nil, nil).Once()

	usecase := NewDockerUseCase(service, nil, logrus.New())

	res := usecase.Execute(context.TODO(), command.Command{
		ID:     "TEST",
		Target: command.Target{IP: "127.0.0.1"},
		Order: command.Order{
			Type: command.Emulation,
			Payload: map[string]interface{}{
				"invalid": "field",
			},
		},
	})
	assert.Error(t, res.Error)
	service.AssertExpectations(t)
}

func TestDockerUseCase_Execute_UnknownType_Failure(t *testing.T) {
	service := new(mockService.DockerService)
	service.On("CreateClient", mock.Anything, mock.Anything).Return(nil, nil).Once()

	usecase := NewDockerUseCase(service, nil, logrus.New())

	res := usecase.Execute(context.TODO(), command.Command{
		ID:     "TEST",
		Target: command.Target{IP: "127.0.0.1"},
		Order: command.Order{
			Type:    command.OrderType("I do not exist fdsfwerwerwR"),
			Payload: nil,
		},
	})
	assert.Error(t, res.Error)
	service.AssertExpectations(t)
}

func TestDockerUseCase_Execute_CreateContainer_Failure_ExtraField(t *testing.T) {
	service := new(mockService.DockerService)
	service.On("CreateClient", mock.Anything, mock.Anything).Return(nil, nil).Once()

	valid := new(mockValid.OrderValidator)

	usecase := NewDockerUseCase(service, valid, logrus.New())

	res := usecase.Run(command.Command{
		Target: command.Target{IP: "127.0.0.1"},
		Order: command.Order{
			Type: command.Createcontainer,
			Payload: map[string]interface{}{
				"extra": "field",
			},
		},
	})
	assert.Error(t, res.Error)
	service.AssertExpectations(t)
}

func TestDockerUseCase_Execute_CreateVolume_Failure_ExtraField(t *testing.T) {
	service := new(mockService.DockerService)
	service.On("CreateClient", mock.Anything, mock.Anything).Return(nil, nil).Once()

	valid := new(mockValid.OrderValidator)

	usecase := NewDockerUseCase(service, valid, logrus.New())

	res := usecase.Run(command.Command{
		Target: command.Target{IP: "127.0.0.1"},
		Order: command.Order{
			Type: command.Createvolume,
			Payload: map[string]interface{}{
				"extra": "field",
			},
		},
	})
	assert.Error(t, res.Error)
	service.AssertExpectations(t)

}

func TestDockerUseCase_Execute_PutFileInContainer_NoName_Fail(t *testing.T) {
	service := new(mockService.DockerService)
	service.On("CreateClient", mock.Anything, mock.Anything).Return(nil, nil)

	usecase := NewDockerUseCase(service, nil, logrus.New())

	res := usecase.Execute(context.TODO(), command.Command{
		Target: command.Target{IP: "127.0.0.1"},
		Order: command.Order{
			Type: command.Putfileincontainer,
			Payload: command.FileAndContainer{
				ContainerName: "",
				File: command.File{
					Destination: "/test/path/",
					ID:          testFileID,
					Mode:        0777,
				},
			},
		},
	})
	assert.Error(t, res.Error)
	service.AssertExpectations(t)
}
