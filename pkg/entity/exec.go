/**
 * Copyright 2019 Whiteblock Inc. All rights reserved.
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 */

package entity

import "time"

// Exec contains the information for an exec call
type Exec struct {
	Cmd        []string
	Privileged bool
	Retries    int
	Delay      time.Duration
}
