package main

// this is a front end to check-procs. it allows us to build the check
// as a stand alone function

import (
	"checks"
	"fmt"
	"os"
	"strings"
)

var check = checks.CheckConfigType{
	Type:       "check",
	Command:    "",
	Handlers:   []string{},
	Standalone: true,
	Interval:   15,
}

func main() {
	procCheck := new(checks.ProcessCheck)

	check.Command = strings.Join(os.Args, " ")
	check.Args = os.Args
	if len(check.Args) <= 1 {
		procCheck.Usage()
	} else {
		procCheck.Init(check)
		fmt.Printf("%+v\n", procCheck)
	}
}
