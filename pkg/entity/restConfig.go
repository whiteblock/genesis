/**
 * Copyright 2019 Whiteblock Inc. All rights reserved.
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 */

package entity

//RestConfig represents the configuration needed for the REST API
type RestConfig struct {
	//Listen is the socket to listen on
	Listen string `json:"listen"`
}
