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
	"testing"
)

func TestValidateASCII(t *testing.T) {
	tests := map[string]bool{
		"\u0432\u8977":            true,
		"helloworld":              false,
		"f\n\r\t\v":               false,
		"how are you doing\u8333": true,
	}
	for test, expected := range tests {
		err := ValidateASCII(test)
		if (err != nil) != expected {
			if expected {
				t.Errorf("ValidateAscii(\"%s\") passed when should have failed", test)
			} else {
				t.Errorf("ValidateAscii(\"%s\") failed when should have passed", test)
			}
		}
	}

}

func TestValidateNormalASCII(t *testing.T) {
	tests := map[string]bool{
		"\u0432\u8977":            true,
		"helloworld":              false,
		"f\n\r\t\v":               true,
		"how are you doing\u8333": true,
	}
	for test, expected := range tests {
		err := ValidateNormalASCII(test)
		if (err != nil) != expected {
			if expected {
				t.Errorf("ValidateNormalAscii(\"%s\") passed when should have failed", test)
			} else {
				t.Errorf("ValidateNormalAscii(\"%s\") failed when should have passed", test)
			}

		}
	}
}

func TestValidateFilePath(t *testing.T) {
	//test --> invalid?
	tests := map[string]bool{
		"../../../":              true,
		"genesis.json; rm -rf /": true,
		"config.ini":             false,
		"parity/genesis.json":    false,
		"\rhello":                true,
	}
	for test, expected := range tests {
		err := ValidateFilePath(test)
		if (err != nil) != expected {
			if expected {
				t.Errorf("ValidateFilePath(\"%s\") passed when should have failed", test)
			} else {
				t.Errorf("ValidateFilePath(\"%s\") failed when should have passed", test)
			}

		}
	}
}

func TestValidateCommandLine(t *testing.T) {
	//test --> invalid?
	tests := map[string]bool{
		"../../../":              false,
		"genesis.json; rm -rf /": true,
		"config.ini":             false,
		"parity/genesis.json":    false,
		"\rhello":                true,
		"test\";rm -rf /":        true,
	}
	for test, expected := range tests {
		err := ValidateCommandLine(test)
		if (err != nil) != expected {
			if expected {
				t.Errorf("ValidateDockerImage(\"%s\") passed when should have failed", test)
			} else {
				t.Errorf("ValidateDockerImage(\"%s\") failed when should have passed", test)
			}

		}
	}
}
