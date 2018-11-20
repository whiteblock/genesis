package main

import (
	"encoding/json"
	"log"
)

type Response struct {
	Error	error	`json:"error"`
	Result	string	`json:"result"`
}

func (r Response) Marshal() ([]byte) {
	out,err := json.Marshal(r)
	if err != nil {
		log.Println(err)
	}
	return out
}