package checks

import (
	"sensu"
)

// wireless network station stats for Linux based machines
//
// DESCRIPTION
//  This plugin gets the interface metrics
//
// OUTPUT
//   Graphite plain-text format (name value timestamp\n)
//
// PLATFORMS
//   Linux

type WirelessStats struct {
	files   []string
	exclude []string
}

func (ws *WirelessStats) Init(config *sensu.Config) (string, error) {
	return "wireless_metrics", ws.setup()
}

func (ws *WirelessStats) Gather(r *Result) error {
	r.SetCommand("wireless-metrics.rb")

	output, err := ws.createPayload(r.ShortName(), r.StartTime())
	r.SetOutput(output)
	return err
}
