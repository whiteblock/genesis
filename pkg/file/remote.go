/**
 * Copyright 2019 Whiteblock Inc. All rights reserved.
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
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
	return &http.Client{Timeout: rf.conf.FileHandler.APITimeout}
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

func (rf remoteSources) getReader(testnetID string, file command.File) (io.ReadCloser, error) {
	if rf.conf.LocalMode {
		rf.log.Info("reading a file locally")
		f, err := os.Open(file.ID)
		return f, err
	}
	client := rf.getClient()
	ctx, cancel := rf.getContext()
	defer cancel()
	req, err := rf.getRequest(ctx, testnetID, file.ID)
	if err != nil {
		return nil, err
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != 200 {
		rf.log.WithFields(logrus.Fields{
			"file":       file.ID,
			"dest":       file.Destination,
			"code":       resp.StatusCode,
			"definition": testnetID}).Warn("got back a non-200 http code")
		res, _ := ioutil.ReadAll(resp.Body)
		resp.Body.Close()
		return nil, fmt.Errorf(string(res))

	}
	rf.log.WithFields(logrus.Fields{
		"file": file.ID, "Destination": file.Destination}).Debug("copying a file")
	return resp.Body, nil

}

// GetTarReader fetches the file from the file handler service and converts it to a tar reader
func (rf remoteSources) GetTarReader(testnetID string, file command.File) (io.Reader, error) {
	fileReader, err := rf.getReader(testnetID, file)
	if err != nil {
		return nil, err
	}
	defer fileReader.Close()
	res, _ := ioutil.ReadAll(fileReader)
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
