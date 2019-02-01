/*
    Handles global state that can be changed
 */
package state

/*
    CustomError is a custom wrapper for a go error, which 
    has What containing error.Error()
 */
type CustomError struct{
    What    string      `json:"what"`
    err     error
}


