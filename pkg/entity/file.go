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

package entity

import (
	"archive/tar"
	"bytes"
	"io"
	"path/filepath"
)

//TarWriter represents a writer that outputs a tar achive
type TarWriter interface {
	io.Closer
	io.Writer
	Flush() error
	WriteHeader(hdr *tar.Header) error
}

// File represents a file which will be placed inside either a docker container or volume
type File struct {
	//Mode is permission and mode bits
	Mode int64 `json:"mode"`
	//Destination is the mount point of the file
	Destination string `json:"destination"`
	//Data is the contents of the file
	Data []byte `json:"data"`
}

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

//GetTarReader returns a reader which reads this file as if it was in a tar archive
func (file File) GetTarReader() (io.Reader, error) {
	var buf bytes.Buffer
	return &buf, file.writeToTar(tar.NewWriter(&buf))
}
