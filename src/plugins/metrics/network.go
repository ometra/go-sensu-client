package metrics

import (
	"plugins"
)

// network interfaces stats for Linux based machines
//
// DESCRIPTION
//  This plugin gets the interface metrics
//
// OUTPUT
//   Graphite plain-text format (name value timestamp\n)
//
// PLATFORMS
//   Linux

type NetworkInterfaceStats struct{}

func (iface *NetworkInterfaceStats) Init(config plugins.PluginConfig) (string, error) {
	return "interface_metrics", nil
}

func (iface *NetworkInterfaceStats) Gather(r *plugins.Result) error {
	return iface.createPayload(r)
}

func (iface *NetworkInterfaceStats) GetStatus() string {
	return ""
}
