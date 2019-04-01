package util

import(
    "fmt"
    "strings"
)

/*
    Check if the given string only contains standard ASCII characters, which can fit
    in a signed char
 */
func ValidateAscii(str string) error {
    for _,c := range str {
        if c > 127 {
            return fmt.Errorf("Character %c is not ASCII",c)
        }
    }
    return nil
}

/*
    Similar to ValidateAscii, except that it excludes control characters from the set of acceptable characters.
    Any character 127 > c > 31 is considered valid
 */
func ValidateNormalAscii(str string) error {
    for _,c := range str {
        if c > 126  || c < 32 {
            return fmt.Errorf("Character %c is not allowed",c)
        }
    }
    return nil
}

/*
    Check to make sure there is nothing malicous in the file path
 */
func ValidateFilePath(path string) error {
    if len(path) == 0 {
        return fmt.Errorf("Cannot be empty")
    }
    trimmedPath := strings.Trim(path," \n\t\v\r\"\\/")
    if len(trimmedPath) == 0 {
        return fmt.Errorf("Effective cannot be empty")
    }
    if strings.Contains(path,"..") {
        return fmt.Errorf("Cannot contain \"..\"")
    }
    if strings.ContainsAny(path,";\\*$#") {
        return fmt.Errorf("Given path contains unusual path characters.")
    }

    return ValidateNormalAscii(path) 
}

func ValidNormalCharacter(chr rune) bool{
    return  (chr >= '+' && chr <= ':') ||
            (chr >= 'A' && chr <= 'Z') || 
            (chr >= 'a' && chr <= 'z') ||
            (chr == ' ' || chr == '_')
}

func ValidateDockerImage(image string) error {
    for _,c := range image {
        if !ValidNormalCharacter(c) {
            return fmt.Errorf("Docker image contains invalid character \"%c\"",c)
        }
    }
    return nil
}