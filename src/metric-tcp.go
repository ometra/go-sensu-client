package main

// this is a front end to check-procs. it allows us to build the check
// as a stand alone function

import (
	"fmt"
	"os"
	"plugins"
	"plugins/metrics"
	"strings"
)

var the_metric = plugins.PluginConfig{
	Type:       "metric",
	Command:    "",
	Handlers:   []string{},
	Standalone: true,
	Interval:   15,
}

func main() {
	m := new(metrics.TcpStats)

	the_metric.Command = strings.Join(os.Args, " ")
	the_metric.Args = os.Args
	_, err := m.Init(the_metric)

	if nil != err {
		fmt.Println(err)
		m.ShowUsage()
		os.Exit(1)
	}

	r := new(plugins.Result)
	err = m.Gather(r)
	if nil != err {
		fmt.Println("Error:", err)
		os.Exit(2)
	}
	fmt.Println(strings.Join(r.OutputAsStrings(), "\n"))
}
