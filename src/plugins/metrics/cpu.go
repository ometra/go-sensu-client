package metrics

import (
	"plugins"
)

// CPU Status for Linux based machines
//
// DESCRIPTION
//  This plugin gets the CPU stats from linux machines and puts them on the wire without prompting for sensu
//
// OUTPUT
//   Graphite plain-text format (name value timestamp\n)
//
// PLATFORMS
//   Linux

const CPU_STATS_NAME = "cpu_metrics"

type CpuStats struct {
	gather_frequency_stats   bool
	failed_freq_gather_count int
	cpu_count                int
}

func init() {
	plugins.Register("cpu_metrics", new(CpuStats))
}

func (cpu *CpuStats) Init(config plugins.PluginConfig) (string, error) {
	return CPU_STATS_NAME, cpu.setup() // os dependent part
}

func (cpu *CpuStats) Gather(r *plugins.Result) error {
	return cpu.createPayload(r)
}

func (cpu *CpuStats) GetStatus() string {
	return ""
}
