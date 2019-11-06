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

//Result is the last known status of the command, contains a type and possibly an error

package entity

type Result struct {
	Error error
	Type  string
}

//IsSuccess returns whether or not the result indicates success
func (res Result) IsSuccess() bool {
	return res.Error == nil
}

//IsFatal returns true if there is an errr and it is marked as a fatal error,
//meaning it should not be reattempted
func (res Result) IsFatal() bool {
	return res.Error != nil && res.Type == FatalType
}

const (
	//SuccessType is the type of a successful result
	SuccessType = "Success"
	//TooSoonType is the type of a result from a cmdwhich tried to execute too soon
	TooSoonType = "TooSoon"
	//FatalType is the type of a result which indicates a fatal error
	FatalType = "Fatal"
)
