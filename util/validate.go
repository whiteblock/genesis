package util

import (
	"fmt"
	"strings"
)

/*
   Check if the given string only contains standard ASCII characters, which can fit
   in a signed char
*/
func ValidateAscii(str string) error {
	for _, c := range str {
		if c > 127 {
			return fmt.Errorf("character %c is not ASCII", c)
		}
	}
	return nil
}

/*
   Similar to ValidateAscii, except that it excludes control characters from the set of acceptable characters.
   Any character 127 > c > 31 is considered valid
*/
func ValidateNormalAscii(str string) error {
	for _, c := range str {
		if c > 126 || c < 32 {
			return fmt.Errorf("invalid character %c", c)
		}
	}
	return nil
}

/*
   Check to make sure there is nothing malicous in the file path
*/
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

	return ValidateNormalAscii(path)
}

func ValidNormalCharacter(chr rune) bool {
	return (chr >= '+' && chr <= ':') ||
		(chr >= 'A' && chr <= 'Z') ||
		(chr >= 'a' && chr <= 'z') ||
		(chr == ' ' || chr == '_')
}

func ValidateCommandLine(image string) error {
	for _, c := range image {
		if !ValidNormalCharacter(c) {
			return fmt.Errorf("docker image contains invalid character \"%c\"", c)
		}
	}
	return nil
}
