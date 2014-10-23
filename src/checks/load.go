package checks

import (
	"sensu"
)

// CPU Status for Linux based machines
//
// DESCRIPTION
//  This plugin gets the load average and reports it in graphite line format
//
// OUTPUT
//   Graphite plain-text format (name value timestamp\n)
//
// PLATFORMS
//   Linux

type LoadStats struct{}

func (load *LoadStats) Init(config *sensu.Config) (string, error) {
	return "load_metrics", nil
}

func (load *LoadStats) Gather(r *Result) error {
	r.SetCommand("load-metrics.rb")
	output, err := load.createPayload(r.ShortName(), r.StartTime())
	r.SetOutput(output)
	return err
}
