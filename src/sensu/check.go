package sensu

import (
	"github.com/bitly/go-simplejson"
)

type Check struct {
	Name            string `json:"name"`
	Command         string `json:"command"`
	Executed        int
	Status          int
	Issued          int `json:"issued"`
	Output          string
	Duration        float64
	Timeout         int
	commandExecuted string
	data            *simplejson.Json
}
