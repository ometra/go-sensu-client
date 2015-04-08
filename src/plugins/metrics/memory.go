package metrics

import (
	"plugins"
)

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

func init() {
	plugins.Register("memory_metrics", new(MemoryStats))
}

func (mem *MemoryStats) Init(config plugins.PluginConfig) (string, error) {
	return "memory_metrics", nil
}

func (mem *MemoryStats) Gather(r *plugins.Result) error {
	return mem.createPayload(r)
}

func (mem *MemoryStats) GetStatus() string {
	return ""
}
