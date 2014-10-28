package checks

import ()

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

func (iface *NetworkInterfaceStats) Init(config CheckConfigType) (string, error) {
	return "interface_metrics", nil
}

func (iface *NetworkInterfaceStats) Gather(r *Result) error {
	output, err := iface.createPayload(r.ShortName(), r.StartTime())
	r.SetOutput(output)
	return err
}
