/**
 * Copyright 2019 Whiteblock Inc. All rights reserved.
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 */

package externals

import (
	"context"
	"net"
)

//NetConn is just net.Conn pulled into another interface for mocking
type NetConn interface {
	net.Conn
}

//Context is context.Context pulled into this interface for mocking
type Context interface {
	context.Context
}
