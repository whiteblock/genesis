package state

type CustomError struct{
    What    string      `json:"what"`
    err     error
}


