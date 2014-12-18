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

func init() {
	plugins.Register("wireless-ap_metrics", new(WirelessStats))
}


func (ws *WirelessStats) Init(config plugins.PluginConfig) (string, error) {
	return "wireless-ap_metrics", ws.setup()
}

func (ws *WirelessStats) Gather(r *plugins.Result) error {
	return ws.createPayload(r)
}

func (ws *WirelessStats) GetStatus() string {
	return ""
}
