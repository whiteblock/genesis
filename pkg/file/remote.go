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

package file

import (
	"archive/tar"
	"bytes"
	"context"
	"fmt"
	"io"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/whiteblock/genesis/pkg/config"
	"github.com/whiteblock/genesis/pkg/entity"
	"github.com/whiteblock/genesis/pkg/repository"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/api/types/strslice"
	"github.com/docker/docker/api/types/swarm"
	"github.com/docker/docker/api/types/volume"
	"github.com/docker/docker/client"
	"github.com/sirupsen/logrus"
	"github.com/whiteblock/definition/command"
)

func (file File) writeToTar(tw TarWriter) error {
	hdr := &tar.Header{
		Name: filepath.Base(file.Destination),
		Mode: file.Mode,
		Size: int64(len(file.Data)),
	}
	if err := tw.WriteHeader(hdr); err != nil {
		return err
	}
	_, err := tw.Write(file.Data)
	return err
}

// GetTarReader returns a reader which reads this file as if it was in a tar archive
func (file File) GetTarReader() (io.Reader, error) {
	var buf bytes.Buffer
	return &buf, file.writeToTar(tar.NewWriter(&buf))
}

// GetDir gets the destination directory
func (file File) GetDir() string {
	return filepath.Dir(file.Destination)
}

// TarWriter represents a writer that outputs a tar achive
type TarWriter interface {
	io.Closer
	io.Writer
	Flush() error
	WriteHeader(hdr *tar.Header) error
}

// IFile represents the operations needed for the file object
type IFile interface {
	GetTarReader() (io.Reader, error)
	GetDir() string
}
