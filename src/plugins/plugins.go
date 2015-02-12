package plugins

import (
	"time"
)

// Handles any of our checks reimplemented in Golang

const (
	OK       Status = iota
	WARNING  Status = iota
	CRITICAL Status = iota
	UNKNOWN  Status = iota
)

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

type Result struct {
	output       []ResultStat
	runStatus    Status
	noOutputWrap bool
}

type ResultStat struct {
	Output    string
	Time      time.Time
	TimeIsSet bool
}

var statusLookupTable = map[Status]string{
	OK:       `OK`,
	WARNING:  `WARNING`,
	CRITICAL: `CRITICAL`,
	UNKNOWN:  `UNKNOWN`,
}

var pluginList = map[string]SensuPluginInterface{}

func Register(handle string, plugin SensuPluginInterface) {
	pluginList[handle] = plugin
}

func GetPlugin(name string) SensuPluginInterface {
	return pluginList[name]
}

func (r *Result) Add(output string) {
	stat := ResultStat{Output: output, TimeIsSet: false}
	r.output = append(r.output, stat)
}

func (r *Result) AddWithTime(output string, t time.Time) {
	stat := ResultStat{Output: output, Time: t, TimeIsSet: true}
	r.output = append(r.output, stat)
}

func (r *Result) Output() []ResultStat {
	return r.output
}

func (r *Result) OutputAsStrings() []string {
	var o []string
	for _, stat := range r.output {
		o = append(o, stat.Output)
	}
	return o
}

func (r *Result) SetStatus(runStatus Status) {
	r.runStatus = runStatus
}

func (r *Result) Status() string {
	return r.runStatus.ToString()
}

func (r *Result) SetNoWrapOutput() {
	r.noOutputWrap = true
}
func (r Result) IsNoWrapOutput() bool {
	return r.noOutputWrap
}

func (s Status) ToString() string {
	return statusLookupTable[s]
}

func (s Status) ToInt() int {
	return int(s)
}
