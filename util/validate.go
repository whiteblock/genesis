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

package util

import (
	"fmt"
	"strings"
)

// ValidateASCII checks if the given string only contains standard ASCII characters, which can fit
// in a signed char
func ValidateASCII(str string) error {
	for _, c := range str {
		if c > 127 {
			return fmt.Errorf("character %c is not ASCII", c)
		}
	}
	return nil
}

// ValidateNormalASCII is similar to ValidateAscii, except that it excludes control characters from the set of acceptable characters.
// Any character 127 > c > 31 is considered valid
func ValidateNormalASCII(str string) error {
	for _, c := range str {
		if c > 126 || c < 32 {
			return fmt.Errorf("invalid character %c", c)
		}
	}
	return nil
}

// ValidateFilePath check to make sure there is nothing malicous in the file path
func ValidateFilePath(path string) error {
	if len(path) == 0 {
		return fmt.Errorf("cannot be empty")
	}
	trimmedPath := strings.Trim(path, " \n\t\v\r\"\\/")
	if len(trimmedPath) == 0 {
		return fmt.Errorf("effective cannot be empty")
	}
	if strings.Contains(path, "..") {
		return fmt.Errorf("cannot contain \"..\"")
	}
	if strings.ContainsAny(path, ";\\*$#") {
		return fmt.Errorf("given path contains unusual characters")
	}

	return ValidateNormalASCII(path)
}

// ValidNormalCharacter checks to make sure a character is within a safe range to naively
// prevent most bash injects (Not for security, only for debugging)
func ValidNormalCharacter(chr rune) bool {
	return (chr >= '+' && chr <= ':') ||
		(chr >= 'A' && chr <= 'Z') ||
		(chr >= 'a' && chr <= 'z') ||
		(chr == ' ' || chr == '_' || chr == '@')
}

// ValidateCommandLine naively checks for ppntential accidental bash injections. Like ValidNormalCharacter,
// is not too be considered useful for security, only for picking up on potential bugs.
func ValidateCommandLine(str string) error {
	for _, c := range str {
		if !ValidNormalCharacter(c) {
			return fmt.Errorf("\"%s\" contains invalid character '%c'", str, c)
		}
	}
	return nil
}
