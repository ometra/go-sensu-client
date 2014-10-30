package main

// this is a front end to check-procs. it allows us to build the check
// as a stand alone function

import (
	"fmt"
	"os"
	"plugins"
	"plugins/checks"
	"strings"
)

var the_check = plugins.PluginConfig{
	Type:       "check",
	Command:    "",
	Handlers:   []string{},
	Standalone: true,
	Interval:   15,
}

func main() {
	procCheck := new(checks.ProcessCheck)

	the_check.Command = strings.Join(os.Args, " ")
	the_check.Args = os.Args
	procCheck.Init(the_check)
	if procCheck.ShowHelp {
		procCheck.Usage()
	} else {
		r := new(plugins.Result)
		procCheck.Gather(r)
		fmt.Printf("%s: %s\n", procCheck.GetStatus(), strings.Join(r.Output(), ""))
	}
}
