package sensu

import (
	amqp "github.com/streadway/amqp"
	//	"fmt"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"plugins"
	"strings"
	"time"
)

const RESULTS_QUEUE = "results"

type check struct {
	Name       string   `json:"name"`       // the check name in sensu
	Command    string   `json:"command"`    // the "command" that was run
	Executed   uint     `json:"executed"`   // timestamp for when this check was started
	Issued     uint     `json:"issued"`     // timestamp for when this check was sent
	Status     int      `json:"status"`     // the status for the check. 0 = success. > 0 = fail
	Output     string   `json:"output"`     // the output of the check script
	Duration   float64  `json:"duration"`   // how long it took to run the check. this needs to be transformed to "seconds.milliseconds
	CheckType  string   `json:"type"`       // "metric|???"
	Handlers   []string `json:"handlers"`   // how this data is processed
	Interval   int      `json:"interval"`   // how long between checks, in seconds
	Standalone bool     `json:"standalone"` // was this check unsolicited?

	Address string `json:"-"` // usage unknown

	// not used
	timeout    int
	started    time.Time
	time_taken time.Duration
}

type ResultInterface interface {
	HasOutput() bool
	GetPayload() amqp.Publishing
	toJson() []byte
}

type Result struct {
	Client            string `json:"client"` // DNS Name for this host
	client_short_name string
	Check             check  `json:"check"`
	checkStatus       string // for checks we want to know if they are critical/warning/unknown/ok
	wrapOutput        bool
}

type SavedResult struct {
	result []byte
}

// sets up the common result data
func NewResult(clientConfig ClientConfig, check_name string) *Result {
	result := new(Result)
	result.wrapOutput = true

	result.Client = clientConfig.Name
	// host name schema is stb.<location-name>.loc.swiftnetworks.com.au
	bits := strings.Split(result.Client, ".")
	if "stb" == bits[0] {
		result.client_short_name = fmt.Sprintf("%s.%s", bits[0], bits[1])
	} else {
		result.client_short_name = bits[0]
	}

	result.Check.Name = check_name
	result.Check.Address = clientConfig.Address
	result.Check.Executed = uint(time.Now().Unix())
	result.Check.Handlers = make([]string, 1)
	result.Check.Handlers[0] = "metrics"
	result.Check.CheckType = "metric"
	result.Check.Standalone = true
	result.Check.Status = 0

	result.Check.started = time.Now()

	return result
}

// takes each of our lines of output and prefixes the system we are checking
// and suffixes the timestamp when we checked
func (r *Result) SetOutput(rows []plugins.ResultStat) {

	switch r.Check.CheckType {
	case "metric":
		if !r.wrapOutput { // mainly for external metrics that provide their own fully qualified lines of output
			for _, row := range rows {
				r.Check.Output += row.Output + "\n"
			}
		} else {
			var t uint
			for _, row := range rows {
				if row.TimeIsSet {
					t = uint(row.Time.Unix())
				} else {
					t = r.StartTime()
				}

				r.Check.Output += fmt.Sprintf("%s.%s %d\n", r.ShortName(), row.Output, t)
			}
		}
	case "check":
		for _, row := range rows {
			r.Check.Output += row.Output
		}
	}
}

func (r *Result) SetCheckStatus(s string) {
	r.checkStatus = s
}

func (r *Result) Output() string {
	var output string
	switch r.Check.CheckType {
	case "metric":
		output = r.Check.Output
	case "check":
		output = r.checkStatus + ": " + r.Check.Output
	}
	return output
}

func (r *Result) HasOutput() bool {
	return "" != r.Check.Output
}

func (r *Result) SetInterval(interval time.Duration) {
	r.Check.Interval = int(interval / time.Second)
}

func (r *Result) ShortName() string {
	return r.client_short_name
}

func (r *Result) StartTime() uint {
	return r.Check.Executed
}

func (r *Result) SetStatus(status int) {
	r.Check.Status = status
}

func (r *Result) SetCommand(command string) {
	r.Check.Command = command
}

func (r *Result) SetType(checktype string) {
	r.Check.CheckType = checktype
}

func (r *Result) SetWrapOutput(wrapOutput bool) {
	r.wrapOutput = wrapOutput
}

func (r *Result) calculate_duration() {
	var duration = time.Now().Sub(r.Check.started)
	r.Check.Duration = duration.Seconds()
}

func (result *Result) toJson() []byte {
	result.calculate_duration()
	result.Check.Output += "\n"
	result.Check.Issued = uint(time.Now().Unix())

	json, err := json.Marshal(result)
	if nil != err {
		log.Panic(err)
	}
	if "" != os.Getenv("DEBUG") {
		log.Printf(string(json)) // handy json debug printing
	}
	return json
}

func (result *Result) GetPayload() amqp.Publishing {
	return getRabbitPayload(result.toJson())
}

func getRabbitPayload(json []byte) amqp.Publishing {
	return amqp.Publishing{
		ContentType:  "application/octet-stream",
		Body:         json,
		DeliveryMode: amqp.Transient,
	}

}

// Saved Results are just wrappers around JSON blobs

func (sr *SavedResult) SetResult(json string) {
	sr.result = []byte(json)
}

func (sr *SavedResult) HasOutput() bool {
	return len(sr.result) > 0
}

func (sr *SavedResult) GetPayload() amqp.Publishing {
	return getRabbitPayload(sr.result)
}

func (sr *SavedResult) toJson() []byte {
	return []byte(sr.result)
}
