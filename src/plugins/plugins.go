package plugins

import (
	"time"
)

// executes the configured checks at the configured intervals

type SensuPluginInterface interface {
	Init(PluginConfig) (name string, err error)
	Gather(*Result) error
	GetStatus() string
}

type PluginConfig struct {
	Type       string        `json:"type"`
	Name       string        `json:"name"`
	Command    string        `json:"command"`
	Args       []string      `json:"args"`
	Handlers   []string      `json:"handlers"`
	Standalone bool          `json:"standalone"`
	Interval   time.Duration `json:"interval"`
}

type Status int // check status - not used for metrics

const (
	OK       Status = iota
	WARNING  Status = iota
	CRITICAL Status = iota
	UNKNOWN  Status = iota
)

var statusLookupTable = map[Status]string{
	OK:       `OK`,
	WARNING:  `WARNING`,
	CRITICAL: `CRITICAL`,
	UNKNOWN:  `UNKNOWN`,
}

type Result struct {
	output    []string
	runStatus Status
}

func (r *Result) Add(output string) {
	r.output = append(r.output, output)
}

func (r *Result) Output() []string {
	return r.output
}

func (r *Result) SetStatus(runStatus Status) {
	r.runStatus = runStatus
}

func (r *Result) Status() string {
	return r.runStatus.ToString()
}

func (s Status) ToString() string {
	return statusLookupTable[s]
}

func (s Status) ToInt() int {
	return int(s)
}
