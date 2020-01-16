/**
 * Copyright 2019 Whiteblock Inc. All rights reserved.
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 */

package externals

import (
	"net"
)

//NetConn is just net.Conn pulled into another interface for mocking
type NetConn interface {
	net.Conn
}
