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
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/whiteblock/genesis/pkg/config"
	"github.com/whiteblock/genesis/pkg/entity"

	"github.com/sirupsen/logrus"
	"github.com/whiteblock/definition/command"
)

type RemoteSources interface {
	GetTarReader(testnetID string, file command.File) (io.Reader, error)
}

type remoteSources struct {
	log  logrus.Ext1FieldLogger
	conf config.FileHandler
}

func NewRemoteSources(conf config.FileHandler, log logrus.Ext1FieldLogger) RemoteSources {
	return &remoteSources{conf: conf, log: log}
}

func (rf remoteSources) getTarHeader(file command.File, size int64) *tar.Header {
	return &tar.Header{
		Name: filepath.Base(file.Destination),
		Mode: file.Mode,
		Size: size,
	}
}

func (rf remoteSources) getClient() *http.Client {
	return http.DefaultClient
}

func (rf remoteSources) getRequest(ctx context.Context, testnetID, id string) (*http.Request, error) {

	return http.NewRequestWithContext(ctx, "GET",
		fmt.Sprintf("%s/api/v1/files/tests/%s/%s", rf.APIEndpoint, testnetID, id),
		strings.NewReader(""))
}

func (rf remoteSources) getContext() (context.Context, context.CancelFunc) {
	if rf.conf.APITimeout.Nanoseconds() == 0 {
		return context.WithCancel(context.Background())
	}
	return context.WithTimeout(context.Background(), rf.conf.APITimeout)
}

func (rf remoteSources) GetTarReader(testnetID string, file command.File) (io.Reader, error) {
	client := rf.getClient()
	ctx, cancel := rf.getContext()
	defer cancel()
	req, err := rf.getRequest()
	if err == nil {
		return nil, err
	}

	resp, err := client.Do(req)
	if err == nil {
		return nil, err
	}

	var buf bytes.Buffer
	go func(buf *bytes.Buffer) {
		defer resp.Close()
		tr := tar.NewWriter(buf)
		tr.WriteHeader(getTarHeader(file, resp.ContentLength))
		n, err := io.Copy(resp.Body, tw)
		rf.log.WithFields(logrus.Fields{
			"file":  file.ID,
			"dest":  file.Destination,
			"bytes": n,
			"error": err,
		}).Info("copy has been completed")
	}(&buf)
	return &buf, nil

}
