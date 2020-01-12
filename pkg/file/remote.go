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

package file

import (
	"archive/tar"
	"bytes"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/whiteblock/genesis/pkg/config"

	"github.com/sirupsen/logrus"
	"github.com/whiteblock/definition/command"
)

//RemoteSources represents a remote file source
type RemoteSources interface {
	GetTarReader(testnetID string, file command.File) (io.Reader, error)
}

type remoteSources struct {
	log  logrus.Ext1FieldLogger
	conf config.Config
}

//NewRemoteSources creates a new instance of RemoteSources
func NewRemoteSources(conf config.Config, log logrus.Ext1FieldLogger) RemoteSources {
	return &remoteSources{conf: conf, log: log}
}

func (rf remoteSources) getTarHeader(file command.File, size int64) *tar.Header {
	name := filepath.Base(file.Destination)
	if file.Destination[len(file.Destination)-1] == '/' {
		name = filepath.Base(file.Meta.Filename)
	}
	rf.log.WithFields(logrus.Fields{
		"name": name,
		"mode": file.Mode,
		"size": size,
	}).Trace("got the tar header for a file")
	return &tar.Header{
		Name: name,
		Mode: file.Mode,
		Size: size,
	}
}

func (rf remoteSources) getClient() *http.Client {
	return http.DefaultClient
}

func (rf remoteSources) getRequest(ctx context.Context, testnetID, id string) (*http.Request, error) {

	return http.NewRequestWithContext(ctx, "GET",
		fmt.Sprintf("%s/api/v1/files/definitions/%s/%s", rf.conf.FileHandler.APIEndpoint, testnetID, id),
		strings.NewReader(""))
}

func (rf remoteSources) getContext() (context.Context, context.CancelFunc) {
	if rf.conf.FileHandler.APITimeout.Nanoseconds() == 0 {
		return context.WithCancel(context.Background())
	}
	return context.WithTimeout(context.Background(), rf.conf.FileHandler.APITimeout)
}

func (rf remoteSources) getReader(testnetID string, file command.File) (io.ReadCloser, int64, error) {
	if rf.conf.LocalMode {
		rf.log.Info("reading a file locally")
		f, err := os.Open(file.ID)
		return f, 0, err
	}
	client := rf.getClient()
	ctx, cancel := rf.getContext()
	defer cancel()
	req, err := rf.getRequest(ctx, testnetID, file.ID)
	if err != nil {
		return nil, 0, err
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, 0, err
	}
	if resp.StatusCode != 200 {
		rf.log.WithFields(logrus.Fields{
			"file":       file.ID,
			"dest":       file.Destination,
			"code":       resp.StatusCode,
			"definition": testnetID}).Warn("got back a non-200 http code")
		res, _ := ioutil.ReadAll(resp.Body)
		resp.Body.Close()
		return nil, 0, fmt.Errorf(string(res))

	}
	/*if resp.ContentLength == -1 {
		rf.log.WithFields(logrus.Fields{
			"file":       file.ID,
			"dest":       file.Destination,
			"code":       resp.StatusCode,
			"definition": testnetID}).Warn("got a -1 content length")
		res, _ := ioutil.ReadAll(resp.Body)
		return nil, 0, fmt.Errorf(string(res))
	}*/
	rf.log.WithFields(logrus.Fields{"size": resp.ContentLength,
		"file": file.ID, "Destination": file.Destination}).Debug("copying a file")
	return resp.Body, resp.ContentLength, nil

}

// GetTarReader fetches the file from the file handler service and converts it to a tar reader
func (rf remoteSources) GetTarReader(testnetID string, file command.File) (io.Reader, error) {
	fileReader, _, err := rf.getReader(testnetID, file)
	if err != nil {
		return nil, err
	}
	defer fileReader.Close()
	res, _ := ioutil.ReadAll(fileReader)
	rf.log.Error("I need to be reverted")
	rdr := bytes.NewReader(res)

	var buf bytes.Buffer
	buf.Grow(len(res))
	//might want to make a custom reader here for memory sake

	tr := tar.NewWriter(&buf)
	tr.WriteHeader(rf.getTarHeader(file, int64(len(res))))
	n, err := io.Copy(tr, rdr)
	rf.log.WithFields(logrus.Fields{
		"file":  file.ID,
		"dest":  file.Destination,
		"bytes": n,
		"error": err,
	}).Info("copy has been completed")

	return &buf, nil

}
