/**
 * Copyright 2019 Whiteblock Inc. All rights reserved.
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 */

package repository

import (
	"fmt"
	"io"
	"strings"
	"testing"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types/network"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/whiteblock/definition/command"

	entityMock "github.com/whiteblock/genesis/mocks/pkg/entity"
)

func TestDockerRepository_GetNetworkByName_Success(t *testing.T) {
	results := []network.Summary{
		{Name: "test1", ID: "id1"},
		{Name: "test2", ID: "id2"},
	}
	cli := new(entityMock.Client)
	cli.On("NetworkList", mock.Anything, mock.Anything).Return(results, nil).Run(
		func(args mock.Arguments) {

			require.Len(t, args, 2)
			assert.Nil(t, args.Get(0))
		}).Times(len(results) + 1)
	ds := NewDockerRepository(logrus.New())

	for _, result := range results {
		net, err := ds.GetNetworkByName(nil, cli, result.Name)
		assert.NoError(t, err)
		assert.Equal(t, result, net)
	}

	_, err := ds.GetNetworkByName(nil, cli, "DNE")
	assert.Error(t, err)

	cli.AssertExpectations(t)
}

func TestDockerRepository_GetNetworkByName_Failure(t *testing.T) {
	cli := new(entityMock.Client)
	cli.On("NetworkList", mock.Anything, mock.Anything).Return(nil, fmt.Errorf("eerrr")).Once()
	ds := NewDockerRepository(logrus.New())
	_, err := ds.GetNetworkByName(nil, cli, "foo")
	assert.Error(t, err)

	cli.AssertExpectations(t)
}

func TestDockerRepository_HostHasImage_Success(t *testing.T) {
	testImageList := []image.Summary{
		{RepoDigests: []string{"test0"}, RepoTags: []string{"test2"}},
		{RepoDigests: []string{"test3"}, RepoTags: []string{"test4"}},
		{RepoDigests: []string{"test5"}, RepoTags: []string{"test6"}},
		{RepoDigests: []string{"test7"}, RepoTags: []string{"test8"}},
	}

	existingImageTags := []string{"test2", "test6"}
	existingImageDigests := []string{"test0", "test3"}
	noneExistingImageTags := []string{"a", "b"}
	noneExistingImageDigests := []string{"c", "d"}

	cli := new(entityMock.Client)
	cli.On("ImageList", mock.Anything, mock.Anything).Return(testImageList, nil).Run(
		func(args mock.Arguments) {

			require.Len(t, args, 2)
			assert.Nil(t, args.Get(0))
		})

	ds := NewDockerRepository(logrus.New())

	for _, term := range append(existingImageTags, existingImageDigests...) {
		exists, err := ds.HostHasImage(nil, cli, term)
		assert.NoError(t, err)
		assert.True(t, exists)
	}

	for _, term := range append(noneExistingImageTags, noneExistingImageDigests...) {
		exists, err := ds.HostHasImage(nil, cli, term)
		assert.NoError(t, err)
		assert.False(t, exists)
	}
}

func TestDockerRepository_HostHasImage_Failure(t *testing.T) {

	cli := new(entityMock.Client)
	cli.On("ImageList", mock.Anything, mock.Anything).Return(nil, fmt.Errorf("err"))

	ds := NewDockerRepository(logrus.New())
	exists, err := ds.HostHasImage(nil, cli, "foo")
	assert.Error(t, err)
	assert.False(t, exists)
}

func TestDockerRepository_EnsureImagePulled(t *testing.T) {
	testImageList := []image.Summary{
		{RepoDigests: []string{"test0"}, RepoTags: []string{"test2"}},
		{RepoDigests: []string{"test3"}, RepoTags: []string{"test4"}},
		{RepoDigests: []string{"test5"}, RepoTags: []string{"test6"}},
		{RepoDigests: []string{"test7"}, RepoTags: []string{"test8"}},
	}

	existingImages := []string{"test7", "test6"}
	nonExistingImages := []string{"a", "b"}
	testReader := strings.NewReader("TTTTTTTTTTTTTTTTTTTTTTTTTTTTTTT")

	cli := new(entityMock.Client)
	cli.On("ImageList", mock.Anything, mock.Anything).Return(testImageList, nil).Run(
		func(args mock.Arguments) {
			require.Len(t, args, 2)
			assert.Nil(t, args.Get(0))
		}).Times(2 * (len(nonExistingImages) + len(existingImages)))

	cli.On("ImagePull", mock.Anything, mock.Anything, mock.Anything).Return(
		io.NopCloser(testReader), nil).Run(func(args mock.Arguments) {
		testReader.Reset("TTTTTTTTTTTTTTTTTTTTTTTTTTTTTTT")
		require.Len(t, args, 3)
		assert.Nil(t, args.Get(0))
		ipo, ok := args.Get(2).(image.PullOptions)
		require.True(t, ok)
		assert.Equal(t, "Linux", ipo.Platform)
	}).Times(len(nonExistingImages))

	ds := NewDockerRepository(logrus.New())

	for _, img := range existingImages {
		err := ds.EnsureImagePulled(nil, cli, img, command.Credentials{})
		assert.NoError(t, err)
	}

	for _, img := range nonExistingImages {
		err := ds.EnsureImagePulled(nil, cli, img, command.Credentials{})
		assert.NoError(t, err)
	}
	cli.AssertExpectations(t)
}

func TestDockerRepository_EnsureImagePulled_ImagePull_Failure(t *testing.T) {
	testImageList := []image.Summary{
		{RepoDigests: []string{"test0"}, RepoTags: []string{"test2"}},
		{RepoDigests: []string{"test3"}, RepoTags: []string{"test4"}},
	}

	cli := new(entityMock.Client)
	cli.On("ImageList", mock.Anything, mock.Anything).Return(testImageList, nil).Times(2)

	cli.On("ImagePull", mock.Anything, mock.Anything, mock.Anything).Return(
		nil, fmt.Errorf("err")).Once()

	ds := NewDockerRepository(logrus.New())

	err := ds.EnsureImagePulled(nil, cli, "foobar", command.Credentials{})
	assert.Error(t, err)
	cli.AssertExpectations(t)
}

func TestDockerRepository_GetContainerByName_Success(t *testing.T) {
	results := []container.Summary{
		{Names: []string{"test1", "test3"}, ID: "id1"},
		{Names: []string{"test2", "test4"}, ID: "id2"},
	}
	cli := new(entityMock.Client)
	cli.On("ContainerList", mock.Anything, mock.Anything).Return(results, nil).Run(
		func(args mock.Arguments) {

			require.Len(t, args, 2)
			assert.Nil(t, args.Get(0))
		}).Times((2 * len(results)) + 1)
	ds := NewDockerRepository(logrus.New())

	for _, result := range results {
		for _, name := range result.Names {
			cntr, err := ds.GetContainerByName(nil, cli, name)
			assert.NoError(t, err)
			assert.Equal(t, result, cntr)
		}

	}

	_, err := ds.GetContainerByName(nil, cli, "DNE")
	assert.Error(t, err)

	cli.AssertExpectations(t)
}

func TestDockerRepository_GetContainerByName_Failure(t *testing.T) {
	cli := new(entityMock.Client)
	cli.On("ContainerList", mock.Anything, mock.Anything).Return(nil, fmt.Errorf("err")).Once()
	ds := NewDockerRepository(logrus.New())
	_, err := ds.GetContainerByName(nil, cli, "DNE")
	assert.Error(t, err)

	cli.AssertExpectations(t)
}
