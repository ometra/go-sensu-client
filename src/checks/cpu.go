package checks

import (
	"sensu"
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

type CpuStats struct {
	gather_frequency_stats bool
	cpu_count              int
}

func (cpu *CpuStats) Init(config *sensu.Config) (string, error) {
	return "cpu_metrics", cpu.setup() // os dependent part
}

func (cpu *CpuStats) Gather(r *Result) error {
	r.SetCommand("cpu-metrics.rb")
	output, err := cpu.createPayload(r.ShortName(), r.StartTime())
	r.SetOutput(output)
	return err
}
