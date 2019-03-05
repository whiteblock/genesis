package util

import(
    "fmt"
    "strings"
    "errors"
)

func ValidateAscii(str string) error {
    for _,c := range str {
        if c > 127 {
            return errors.New(fmt.Sprintf("Character %s is not ASCII",c))
        }
    }
    return nil
}

func ValidateNormalAscii(str string) error {
    for _,c := range str {
        if c > 127  || c < 32 {
            return errors.New(fmt.Sprintf("Character %s is not allowed",c))
        }
    }
    return nil
}

/*
    Check to make sure there is nothing malicous in the file path
 */
func ValidateFilePath(path string) error {
    if len(path) == 0 {
        return errors.New("Cannot be empty")
    }
    trimmedPath := strings.Trim(path," \n\t\v\r\"\\/")
    if len(trimmedPath) == 0 {
        return errors.New("Effective cannot be empty")
    }
    if strings.Contains(path,"..") {
        return errors.New("Cannot contain \"..\"")
    }

    err := ValidateNormalAscii(path) 


    return err

}