package metrics

import (
	"plugins"
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

func (ws *WirelessStats) Init(config plugins.PluginConfig) (string, error) {
	return "wireless_metrics", ws.setup()
}

func (ws *WirelessStats) Gather(r *plugins.Result) error {
	return ws.createPayload(r)
}

func (ws *WirelessStats) GetStatus() string {
	return ""
}
