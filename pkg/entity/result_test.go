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
		res Result
		expected bool
	}{
		{
			res: Result{
				Error: nil,
				Type: SuccessType,
			},
			expected: true,
		},
		{
			res: Result{
				Error: errors.New("Some error"),
				Type: TooSoonType,
			},
			expected: false,
		},
		{
			res: Result{
				Error: nil,
				Type: TooSoonType, //todo should this be a success?
			},
			expected: true,
		},
	}

	for i, tt := range tests {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			assert.Equal(t, tt.res.IsSuccess(), tt.expected)
		})
	}
}

func TestResult_IsFatal(t *testing.T) {
	var tests = []struct {
		res Result
		expected bool
	}{
		{
			res: Result{
				Error: nil,
				Type: SuccessType,
			},
			expected: false,
		},
		{
			res: Result{
				Error: errors.New("Some error"),
				Type: TooSoonType,
			},
			expected: false,
		},
		{
			res: Result{
				Error: errors.New("Some error"),
				Type: FatalType,
			},
			expected: true,
		},
	}

	for i, tt := range tests {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			assert.Equal(t, tt.res.IsFatal(), tt.expected)
		})
	}
}

func TestResult_IsRequeue(t *testing.T) {
	var tests = []struct {
		res Result
		expected bool
	}{
		{
			res: Result{
				Error: nil,
				Type: SuccessType,
			},
			expected: false,
		},
		{
			res: Result{
				Error: errors.New("Some error"),
				Type: TooSoonType,
			},
			expected: true,
		},
		{
			res: Result{
				Error: errors.New("Some error"),
				Type: FatalType,
			},
			expected: false,
		},
	}

	for i, tt := range tests {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			assert.Equal(t, tt.res.IsRequeue(), tt.expected)
		})
	}
}

func TestNewSuccessResult(t *testing.T) {
	expected := Result {
		Error: nil,
		Type: SuccessType,
	}

	expectedUnsuccessful := Result {
		Error: errors.New("test"),
		Type: FatalType,
	}

	expectedUnsuccessful2 := Result {
		Error: errors.New("test"),
		Type: SuccessType,
	}

	assert.Equal(t, NewSuccessResult(), expected)
	assert.NotEqual(t, NewSuccessResult(), expectedUnsuccessful)
	assert.NotEqual(t, NewSuccessResult(), expectedUnsuccessful2)
}

func TestNewFatalResult(t *testing.T) {
	expected := Result {
		Error: errors.New("fatal test"),
		Type: FatalType,
	}

	expectedUnsuccessful := Result {
		Error: errors.New("test"),
		Type: TooSoonType,
	}

	expectedUnsuccessful2 := Result {
		Error: errors.New("test"),
		Type: SuccessType,
	}

	expectedUnsuccessful3 := Result {
		Error: nil,
		Type: SuccessType,
	}

	assert.Equal(t, NewFatalResult(expected.Error), expected)
	assert.NotEqual(t, NewFatalResult(expectedUnsuccessful.Error), expectedUnsuccessful)
	assert.NotEqual(t, NewFatalResult(expectedUnsuccessful2.Error), expectedUnsuccessful2)
	assert.NotEqual(t, NewFatalResult(expectedUnsuccessful3.Error), expectedUnsuccessful3)
}

func TestNewErrorResult(t *testing.T) {
	expected := Result {
		Error: errors.New("fatal test"),
		Type: ErrorType,
	}

	expectedUnsuccessful := Result {
		Error: errors.New("test"),
		Type: ErrorType,
	}

	expectedUnsuccessful2 := Result {
		Error: errors.New("test"),
		Type: ErrorType,
	}

	expectedUnsuccessful3 := Result {
		Error: nil,
		Type: ErrorType,
	}

	assert.Equal(t, NewErrorResult(expected.Error), expected) //todo shouldn't this not pass?
	assert.Equal(t, NewErrorResult(expectedUnsuccessful.Error), expectedUnsuccessful)
	assert.Equal(t, NewErrorResult(expectedUnsuccessful2.Error), expectedUnsuccessful2)
	assert.Equal(t, NewErrorResult(expectedUnsuccessful3.Error), expectedUnsuccessful3)
}
