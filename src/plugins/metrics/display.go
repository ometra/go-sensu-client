package metrics

import (
	"plugins"
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

type DisplayStats struct {
	continue_gathering bool
}

func init() {
	plugins.Register("display_metrics", new(DisplayStats))
}

func (display *DisplayStats) Init(config plugins.PluginConfig) (string, error) {
	display.continue_gathering = true
	return "display_metrics", nil
}

func (display *DisplayStats) Gather(r *plugins.Result) error {
	return display.createPayload(r)
}

func (display *DisplayStats) GetStatus() string {
	return ""
}
