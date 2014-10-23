package checks

import (
	"fmt"
	"log"
	"sensu"
	"time"
)

// executes the configured checks at the configured intervals

type SensuCheckOrMetric interface {
	Init(*sensu.Config) (name string, err error)
	Gather(*Result) error
}

type Processor struct {
	q       sensu.MessageQueuer
	config  *sensu.Config
	jobs    map[string]SensuCheckOrMetric
	close   chan bool
	results chan *Result
}

func (p *Processor) AddJob(job SensuCheckOrMetric) {
	name, err := job.Init(p.config)
	if nil != err {
		log.Printf("Failed to initialise check: %s\n", name)
		return
	}
	log.Printf("Adding job: %s", name)
	p.jobs[name] = job
}

func NewProcessor() *Processor {
	proc := new(Processor)
	proc.jobs = make(map[string]SensuCheckOrMetric)
	proc.results = make(chan *Result, 10) // queue of 20 buffered results

	proc.AddJob(new(CpuStats))
	proc.AddJob(new(LoadStats))
	proc.AddJob(new(NetworkInterfaceStats))
	proc.AddJob(new(MemoryStats))
	return proc
}

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
	return nil
}

// gets the Gather of checks/metrics going
func (p *Processor) Start() {
	go p.publish()

	// start our result publisher thread
	for job_name, job := range p.jobs {
		go func(theJobName string, theJob SensuCheckOrMetric) {
			log.Printf("Starting job: %s", theJobName)
			reset := make(chan bool)

			timer := time.AfterFunc(0, func() {
				log.Printf("Gathering: %s", theJobName)
				result := NewResult(p.config.Data().Get("client"), theJobName)
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
						timer.Reset(15 * time.Second) // need to grab this from the config - defaulting for 15 seconds
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
			if err := p.q.Publish(RESULTS_QUEUE, "", result.GetPayload()); err != nil {
				log.Printf("Error Publishing Stats: %v. %v", err, result)
			}
		case <-p.close:
			return
		}
	}
}
