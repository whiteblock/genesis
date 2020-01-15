/**
 * Copyright 2019 Whiteblock Inc. All rights reserved.
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 */

package entity

// DockerCli is a wrapper around Client to provide extras such as labels
type DockerCli struct {
	Client
	Labels map[string]string
}
