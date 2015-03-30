package plugins

import (
	"log"
	"os"
	"time"
)

// Handles any of our checks reimplemented in Golang

// run status types
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

// Used to initialise our built in checks and metrics
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

// Holds our Results from the plugins - each check/metric gets a new one each run
type Result struct {
	output       []ResultStat // the actual content with time stamps
	runStatus    Status       // whether or not the check failed
	noOutputWrap bool         // whether or not our check needs its own time stamp
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

// called by each built check/plugin in their own init() function
func Register(handle string, plugin SensuPluginInterface) {
	pluginList[handle] = plugin
}

// retrieves a plugin by name
func GetPlugin(name string) SensuPluginInterface {
	return pluginList[name]
}

// adds a result from a check
func (r *Result) Add(output string) {
	stat := ResultStat{Output: output, Time: time.Now()}
	r.output = append(r.output, stat)

	if "" != os.Getenv("DEBUG") {
		log.Println("Check/Metric: ", output) // handy json debug printing
	}
}

// grabs all of the results
func (r *Result) Output() []ResultStat {
	return r.output
}

// grabs all of the results without timestamp information
func (r *Result) OutputAsStrings() []string {
	var o []string
	for _, stat := range r.output {
		o = append(o, stat.Output)
	}
	return o
}

// for checks we can set the run status to indicate success or fail
func (r *Result) SetStatus(runStatus Status) {
	r.runStatus = runStatus
}

// whether or not we wrap the output of our check - used for external
func (r *Result) SetNoWrapOutput() {
	r.noOutputWrap = true
}
func (r Result) IsNoWrapOutput() bool {
	return r.noOutputWrap
}

// a nice text description of the status
func (r *Result) Status() string {
	return r.runStatus.ToString()
}

func (s Status) ToString() string {
	return statusLookupTable[s]
}

func (s Status) ToInt() int {
	return int(s)
}
