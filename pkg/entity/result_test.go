/*
	Copyright 2019 whiteblock Inc.
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
	expected := Result{
		Error: nil,
		Type:  SuccessType,
	}

	expectedUnsuccessful := Result{
		Error: errors.New("test"),
		Type:  FatalType,
	}

	expectedUnsuccessful2 := Result{
		Error: errors.New("test"),
		Type:  SuccessType,
	}

	assert.Equal(t, expected, NewSuccessResult())
	assert.NotEqual(t, expectedUnsuccessful, NewSuccessResult())
	assert.NotEqual(t, expectedUnsuccessful2, NewSuccessResult())
}

func TestNewFatalResult(t *testing.T) {
	expected := Result{
		Error: errors.New("fatal test"),
		Type:  FatalType,
	}

	expectedUnsuccessful := Result{
		Error: errors.New("test"),
		Type:  TooSoonType,
	}

	expectedUnsuccessful2 := Result{
		Error: errors.New("test"),
		Type:  SuccessType,
	}

	expectedUnsuccessful3 := Result{
		Error: nil,
		Type:  SuccessType,
	}

	assert.Equal(t, expected, NewFatalResult(expected.Error))
	assert.NotEqual(t, expectedUnsuccessful, NewFatalResult(expectedUnsuccessful.Error))
	assert.NotEqual(t, expectedUnsuccessful2, NewFatalResult(expectedUnsuccessful2.Error))
	assert.NotEqual(t, expectedUnsuccessful3, NewFatalResult(expectedUnsuccessful3.Error))
}

func TestNewErrorResult(t *testing.T) {
	expected := Result{
		Error: errors.New("fatal test"),
		Type:  ErrorType,
	}

	expectedUnsuccessful := Result{
		Error: errors.New("test"),
		Type:  ErrorType,
	}

	expectedUnsuccessful2 := Result{
		Error: errors.New("test"),
		Type:  ErrorType,
	}

	assert.Equal(t, expected, NewErrorResult(expected.Error))
	assert.Equal(t, expectedUnsuccessful, NewErrorResult(expectedUnsuccessful.Error))
	assert.Equal(t, expectedUnsuccessful2, NewErrorResult(expectedUnsuccessful2.Error))
}

func TestNewAllDoneResult(t *testing.T) {
	expected := Result{
		Error: nil,
		Type:  AllDoneType,
	}
	assert.Equal(t, expected, NewAllDoneResult())
}
