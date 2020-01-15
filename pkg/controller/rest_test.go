/**
 * Copyright 2019 Whiteblock Inc. All rights reserved.
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 */

package controller

import (
	"testing"

	"github.com/whiteblock/genesis/pkg/entity"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestRestController(t *testing.T) {
	assert.NotNil(t, NewRestController(entity.RestConfig{}, nil, nil, logrus.New()))
}
