package checks

import ()

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

type DisplayStats struct {
	continue_gathering bool
}

func (display *DisplayStats) Init(config CheckConfigType) (string, error) {
	display.continue_gathering = true
	return "display_metrics", nil
}

func (display *DisplayStats) Gather(r *Result) error {
	output, err := display.createPayload(r.ShortName(), r.StartTime())
	r.SetOutput(output)
	return err
}
