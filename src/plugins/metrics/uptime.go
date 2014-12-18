package metrics

import (
	"plugins"
)

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

func init() {
	plugins.Register("uptime_metrics", new(UptimeStats))
}


func (u *UptimeStats) Init(config plugins.PluginConfig) (string, error) {
	return "uptime_metrics", nil
}

func (u *UptimeStats) Gather(r *plugins.Result) error {
	return u.createPayload(r)
}

func (u *UptimeStats) GetStatus() string {
	return ""
}
