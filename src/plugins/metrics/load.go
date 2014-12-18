package metrics

import (
	"plugins"
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

func init() {
	plugins.Register("load_metrics", new(LoadStats))
}


func (load *LoadStats) Init(config plugins.PluginConfig) (string, error) {
	return "load_metrics", nil
}

func (load *LoadStats) Gather(r *plugins.Result) error {
	return load.createPayload(r)
}

func (load *LoadStats) GetStatus() string {
	return ""
}
