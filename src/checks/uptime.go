package checks

import ()

// uptime stats for Linux based machines
//
// DESCRIPTION
//  This plugin gets the uptime for the system and network
//
// OUTPUT
//   Graphite plain-text format (name value timestamp\n)
//
// PLATFORMS
//   Linux

type UptimeStats struct {
}

func (u *UptimeStats) Init(config checkConfigType) (string, error) {
	return "uptime_metrics", nil
}

func (u *UptimeStats) Gather(r *Result) error {
	r.SetCommand("uptime-metrics.rb")

	output, err := u.createPayload(r.ShortName(), r.StartTime())
	r.SetOutput(output)
	return err
}
