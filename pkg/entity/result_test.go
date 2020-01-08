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

package entity

import (
	"errors"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestResult_IsSuccess(t *testing.T) {
	var tests = []struct {
		res      Result
		expected bool
	}{
		{
			res: Result{
				Error: nil,
				Type:  SuccessType,
			},
			expected: true,
		},
		{
			res: Result{
				Error: errors.New("Some error"),
				Type:  TooSoonType,
			},
			expected: false,
		},
		{
			res: Result{
				Error: nil,
				Type:  TooSoonType,
			},
			expected: true,
		},
	}

	for i, tt := range tests {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.res.IsSuccess())
		})
	}
}

func TestResult_IsFatal(t *testing.T) {
	var tests = []struct {
		res      Result
		expected bool
	}{
		{
			res: Result{
				Error: nil,
				Type:  SuccessType,
			},
			expected: false,
		},
		{
			res: Result{
				Error: errors.New("Some error"),
				Type:  TooSoonType,
			},
			expected: false,
		},
		{
			res: Result{
				Error: errors.New("Some error"),
				Type:  FatalType,
			},
			expected: true,
		},
	}

	for i, tt := range tests {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.res.IsFatal())
		})
	}
}

func TestResult_IsRequeue(t *testing.T) {
	var tests = []struct {
		res      Result
		expected bool
	}{
		{
			res: Result{
				Error: nil,
				Type:  SuccessType,
			},
			expected: false,
		},
		{
			res: Result{
				Error: errors.New("Some error"),
				Type:  TooSoonType,
			},
			expected: true,
		},
		{
			res: Result{
				Error: errors.New("Some error"),
				Type:  FatalType,
			},
			expected: false,
		},
	}

	for i, tt := range tests {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.res.IsRequeue())
		})
	}
}

func TestResult_IsAllDone(t *testing.T) {
	var tests = []struct {
		res      Result
		expected bool
	}{
		{
			res:      NewAllDoneResult(),
			expected: true,
		},
		{
			res:      NewErrorResult("err"),
			expected: false,
		},
		{
			res:      NewSuccessResult(),
			expected: false,
		},
	}

	for i, tt := range tests {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.res.IsAllDone())
		})
	}
}

func TestNewSuccessResult(t *testing.T) {
	assert.True(t, NewSuccessResult().IsSuccess())
}

func TestNewFatalResult(t *testing.T) {
	assert.True(t, NewFatalResult("fatal test").IsFatal())
}

func TestNewErrorResult(t *testing.T) {
	assert.False(t, NewErrorResult(errors.New("fatal test")).IsSuccess())
	assert.False(t, NewErrorResult(errors.New("test")).IsSuccess())
}

func TestNewAllDoneResult(t *testing.T) {
	assert.True(t, NewAllDoneResult().IsAllDone())
}
