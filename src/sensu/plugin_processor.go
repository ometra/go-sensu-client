package sensu

import (
	"fmt"
	"log"
	"plugins"
	"plugins/checks"
	"plugins/metrics"
	"strconv"
	"strings"
	"time"
)

type PluginProcessor struct {
	q          MessageQueuer
	config     *Config
	jobs       map[string]plugins.SensuPluginInterface
	jobsConfig map[string]plugins.PluginConfig
	close      chan bool
	results    chan *Result
}

// used to create a new processor instance.
func NewPluginProcessor() *PluginProcessor {
	proc := new(PluginProcessor)
	proc.jobs = make(map[string]plugins.SensuPluginInterface)
	proc.jobsConfig = make(map[string]plugins.PluginConfig)
	proc.results = make(chan *Result, 500) // queue of 500 buffered results

	return proc
}

func newCheckConfig(json interface{}) plugins.PluginConfig {
	var conf plugins.PluginConfig

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
func (p *PluginProcessor) AddJob(job plugins.SensuPluginInterface, checkConfig plugins.PluginConfig) {
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
func (p *PluginProcessor) Init(q MessageQueuer, config *Config) error {
	if err := q.ExchangeDeclare(
		RESULTS_QUEUE,
		"direct",
	); err != nil {
		return fmt.Errorf("Exchange Declare: %s", err)
	}

	p.q = q
	p.config = config
	p.close = make(chan bool, len(p.jobs)+1)
	var check plugins.SensuPluginInterface

	// load the checks we want to do
	checks_config := p.config.Data().Get("checks").MustMap()

	for check_type, checkConfigInterface := range checks_config {
		checkConfig, ok := checkConfigInterface.(map[string]interface{})
		if !ok {
			log.Printf("Failed to parse config: %", check_type)
			continue
		}

		config := newCheckConfig(checkConfig)

		// see if we can handle this check using one of our build in ones
		switch check_type {
		case "cpu_metrics":
			check = new(metrics.CpuStats)
		case "display_metrics":
			check = new(metrics.DisplayStats)
		case "interface_metrics":
			check = new(metrics.NetworkInterfaceStats)
		case "load_metrics":
			check = new(metrics.LoadStats)
		case "memory_metrics":
			check = new(metrics.MemoryStats)
		case "uptime_metrics":
			check = new(metrics.UptimeStats)
		case "wireless-ap_metrics":
			check = new(metrics.WirelessStats)
		case "check_procs":
			check = new(checks.ProcessCheck)
		default:
			if "metric" == config.Type {
				// we have a metric!
				check = new(metrics.ExternalMetric)
			} else {
				// we have a check!
				check = new(checks.ExternalCheck)
			}
		}

		config.Name = check_type
		p.AddJob(check, config)

	}

	return nil
}

// gets the Gather of checks/metrics going
func (p *PluginProcessor) Start() {
	go p.publish()

	// start our result publisher thread
	for job_name, job := range p.jobs {
		go func(theJobName string, theJob plugins.SensuPluginInterface) {
			config := p.jobsConfig[job_name]

			log.Printf("Starting job: %s", theJobName)
			reset := make(chan bool)

			timer := time.AfterFunc(0, func() {
				log.Printf("Gathering: %s", theJobName)
				result := NewResult(p.config.Data().Get("client"), theJobName)
				result.SetCommand(config.Command)

				presult := new(plugins.Result)

				err := theJob.Gather(presult)
				result.SetOutput(presult.Output())
				result.SetCheckStatus(theJob.GetStatus())

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
func (p *PluginProcessor) Stop() {
	for i := 0; i < len(p.close); i++ {
		p.close <- true
	}
}

// our result publishing. will publish results until we call PluginProcessor.Stop()
func (p *PluginProcessor) publish() {

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
