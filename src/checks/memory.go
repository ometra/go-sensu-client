package checks

import ()

// Memory Stats for Linux based machines
//
// DESCRIPTION
//  This plugin gets the Memory stats from linux machines and puts them on the wire without prompting for sensu
//
// OUTPUT
//   Graphite plain-text format (name value timestamp\n)
//
// PLATFORMS
//   Linux

type MemoryStats struct{}

func (mem *MemoryStats) Init(config CheckConfigType) (string, error) {
	return "memory_metrics", nil
}

func (mem *MemoryStats) Gather(r *Result) error {
	output, err := mem.createPayload(r.ShortName(), r.StartTime())
	r.SetOutput(output)
	return err
}
