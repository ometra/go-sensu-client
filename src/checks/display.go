package checks

import (
	"sensu"
)

// Display Status for Linux based machines
//
// DESCRIPTION
//  This plugin gathers stats about the display
//
// OUTPUT
//   Graphite plain-text format (name value timestamp\n)
//
// PLATFORMS
//   Linux

type DisplayStats struct{}

func (load *DisplayStats) Init(config *sensu.Config) (string, error) {
	return "display_metrics", nil
}

func (load *DisplayStats) Gather(r *Result) error {
	r.SetCommand("display-metrics.rb")
	output, err := load.createPayload(r.ShortName(), r.StartTime())
	r.SetOutput(output)
	return err
}
