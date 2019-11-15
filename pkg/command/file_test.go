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

package command

import (
	"archive/tar"
	"errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"path/filepath"
	"testing"
)

func TestFile_writeToTar_Success(t *testing.T) {
	testFile := File{
		Mode:        0644,
		Destination: "pkg/test",
		Data:        []byte("test"),
	}
	tw := new(mocks.TarWriter)
	tw.On("WriteHeader", mock.Anything).Return(nil).Run(func(args mock.Arguments) {
		require.Len(t, args, 1)
		hdr, ok := args.Get(0).(*tar.Header)
		require.True(t, ok)
		require.NotNil(t, hdr)
		assert.Equal(t, filepath.Base(testFile.Destination), hdr.Name)
		assert.Equal(t, testFile.Mode, hdr.Mode)
		assert.Equal(t, int64(len(testFile.Data)), hdr.Size)
	}).Once()

	tw.On("Write", mock.Anything).Return(len(testFile.Data), nil).Run(func(args mock.Arguments) {
		require.Len(t, args, 1)
		data, ok := args.Get(0).([]byte)
		require.True(t, ok)
		require.NotNil(t, data)
		assert.ElementsMatch(t, testFile.Data, data)
	}).Once()

	err := testFile.writeToTar(tw)
	assert.NoError(t, err)
	tw.AssertExpectations(t)
}

func TestFile_writeToTar_Failure(t *testing.T) {
	testFile := File{
		Mode:        0644,
		Destination: "pkg/test",
		Data:        []byte("test"),
	}
	tw := new(mocks.TarWriter)
	tw.On("WriteHeader", mock.Anything).Return(errors.New("err")).Run(func(args mock.Arguments) {
		require.Len(t, args, 1)
		hdr, ok := args.Get(0).(*tar.Header)
		require.True(t, ok)
		require.NotNil(t, hdr)
		assert.Equal(t, filepath.Base(testFile.Destination), hdr.Name)
		assert.Equal(t, testFile.Mode, hdr.Mode)
		assert.Equal(t, int64(len(testFile.Data)), hdr.Size)
	}).Once()

	err := testFile.writeToTar(tw)
	assert.Error(t, err)
	tw.AssertExpectations(t)
}

func TestFile_GetTarReader(t *testing.T) {
	testFile := File{
		Mode:        0644,
		Destination: "pkg/test",
		Data:        []byte("test"),
	}
	rdr, err := testFile.GetTarReader()
	assert.NoError(t, err)
	assert.NotNil(t, rdr)
}

func TestFile_GetDir(t *testing.T) {
	testFile := File{
		Mode:        0644,
		Destination: "pkg/test",
		Data:        []byte("test"),
	}
	dir := testFile.GetDir()
	assert.Equal(t, "pkg", dir)
}
