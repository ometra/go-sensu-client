package checks

import (
	"fmt"
	"log"
	"sensu"
	"strconv"
	"strings"
	"time"
)

// executes the configured checks at the configured intervals

type SensuCheckOrMetric interface {
	Init(CheckConfigType) (name string, err error)
	Gather(*Result) error
}

type CheckConfigType struct {
	Type       string
	Command    string
	Args       []string
	Handlers   []string
	Standalone bool
	Interval   time.Duration
}

type Processor struct {
	q          sensu.MessageQueuer
	config     *sensu.Config
	jobs       map[string]SensuCheckOrMetric
	jobsConfig map[string]CheckConfigType
	close      chan bool
	results    chan *Result
}

var builtInChecksAndMetrics = map[string]SensuCheckOrMetric{
	"cpu_metrics":         new(CpuStats),
	"display_metrics":     new(DisplayStats),
	"interface_metrics":   new(NetworkInterfaceStats),
	"load_metrics":        new(LoadStats),
	"memory_metrics":      new(MemoryStats),
	"uptime_metrics":      new(UptimeStats),
	"wireless-ap_metrics": new(WirelessStats),
}

// used to create a new processor instance.
func NewProcessor() *Processor {
	proc := new(Processor)
	proc.jobs = make(map[string]SensuCheckOrMetric)
	proc.jobsConfig = make(map[string]CheckConfigType)
	proc.results = make(chan *Result, 500) // queue of 500 buffered results

	return proc
}

func newCheckConfig(json interface{}) CheckConfigType {
	var conf CheckConfigType

	converted := json.(map[string]interface{})

	if command, ok := converted["command"]; ok {
		conf.Command, _ = command.(string)
	}

	if args, ok := converted["args"]; ok {
		conf.Args, _ = args.([]string)
	} else {
		conf.Args = strings.Split(conf.Command, " ")
	}

	if handlers, ok := converted["handlers"]; ok {
		conf.Handlers, _ = handlers.([]string)
	}

	conf.Interval = 15 // default 15 second interval
	if interval, ok := converted["interval"]; ok {
		switch t := interval.(type) {
		default:
			i, err := strconv.ParseInt(fmt.Sprintf("%s", t), 10, 64)
			if nil == err {
				conf.Interval = time.Duration(i)
			}
		case int8, int16, int, int32, int64, uint8, uint16, uint, uint32, uint64:
			conf.Interval = time.Duration(interval.(int64))
			log.Println("is a typed int")
		}
	}

	conf.Standalone = true
	if standalone, ok := converted["standalone"]; ok {
		conf.Standalone, _ = standalone.(bool)
	}

	if conf_type, ok := converted["type"]; ok {
		conf.Type, _ = conf_type.(string)
	}

	return conf
}

// helper function to add a check to the queue of checks
func (p *Processor) AddJob(job SensuCheckOrMetric, checkConfig CheckConfigType) {
	name, err := job.Init(checkConfig)
	if nil != err {
		log.Printf("Failed to initialise check: (%s) %s\n", name, err)
		return
	}
	log.Printf("Scheduling job: %s (%s) every %d seconds", name, checkConfig.Command, checkConfig.Interval)

	p.jobs[name] = job
	p.jobsConfig[name] = checkConfig
}

// called to set things up
func (p *Processor) Init(q sensu.MessageQueuer, config *sensu.Config) error {
	if err := q.ExchangeDeclare(
		RESULTS_QUEUE,
		"direct",
	); err != nil {
		return fmt.Errorf("Exchange Declare: %s", err)
	}

	p.q = q
	p.config = config
	p.close = make(chan bool, len(p.jobs)+1)

	// load the checks we want to do
	checks := p.config.Data().Get("checks").MustMap()

	for check_type, checkConfigInterface := range checks {
		checkConfig, ok := checkConfigInterface.(map[string]interface{})
		if !ok {
			log.Printf("Failed to parse config: %", check_type)
			continue
		}

		if check, ok := builtInChecksAndMetrics[check_type]; ok {
			p.AddJob(check, newCheckConfig(checkConfig))
		} else {
			// check not built in
			log.Printf("External Check: %s", check_type)
		}

	}

	return nil
}

// gets the Gather of checks/metrics going
func (p *Processor) Start() {
	go p.publish()

	// start our result publisher thread
	for job_name, job := range p.jobs {
		go func(theJobName string, theJob SensuCheckOrMetric) {
			config := p.jobsConfig[job_name]

			log.Printf("Starting job: %s", theJobName)
			reset := make(chan bool)

			timer := time.AfterFunc(0, func() {
				log.Printf("Gathering: %s", theJobName)
				result := NewResult(p.config.Data().Get("client"), theJobName)
				result.SetCommand(config.Command)
				err := theJob.Gather(result)
				if nil != err {
					// returned an error - we should stop this job from running
					log.Printf("Failed to gather stat: %s. %v", theJobName, err)
					reset <- false
					return
				}

				// add it to the processing queue
				p.results <- result
				reset <- true
			})

			defer timer.Stop()
			for {
				select {
				case cont := <-reset:
					if cont {
						timer.Reset(config.Interval * time.Second) // need to grab this from the config - defaulting for 15 seconds
					} else {
						timer.Stop()
					}
				case <-p.close:
					return
				}
			}
		}(job_name, job)
	}
}

// Puts a halt to all of our checks/metrics gathering
func (p *Processor) Stop() {
	for i := 0; i < len(p.close); i++ {
		p.close <- true
	}
}

// our result publishing. will publish results until we call Processor.Stop()
func (p *Processor) publish() {

	for {
		select {
		case result := <-p.results:
			if result.HasOutput() {
				if err := p.q.Publish(RESULTS_QUEUE, "", result.GetPayload()); err != nil {
					log.Printf("Error Publishing Stats: %v. %v", err, result)
				}
			}
		case <-p.close:
			return
		}
	}
}
