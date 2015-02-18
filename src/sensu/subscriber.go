package sensu

import (
	"encoding/json"
	"fmt"
	"github.com/streadway/amqp"
	"io"
	"log"
	"plugins"
	"time"
)

type Subscriber struct {
	deliveries <-chan amqp.Delivery
	done       chan error
	logger     *log.Logger
	config     *Config
	q          MessageQueuer
}

func NewSubscriber(w io.Writer) *Subscriber {
	s := new(Subscriber)
	s.logger = log.New(w, "Subscriptions: ", log.LstdFlags)
	return s
}

func (s *Subscriber) Init(q MessageQueuer, c *Config) error {

	s.config = c
	s.q = q
	config_name := c.Client.Name
	config_ver := c.Client.Version

	queue_name := fmt.Sprintf("%s-%s-%d", config_name, config_ver, time.Now().Unix())
	s.logger.Printf("Declaring Queue: %s", queue_name)
	queue, err := q.QueueDeclare(queue_name)
	if err != nil {
		return fmt.Errorf("Queue Declare: %s", err)
	}
	s.logger.Printf("declared Queue")

	var subscriptions []string
	subscriptions, err = c.Data().GetPath("client", "subscriptions").StringArray()
	if err != nil {
		return fmt.Errorf("Subscriptions are not in a string array format")
	}

	for _, sub := range subscriptions {
		s.logger.Printf("declaring Exchange (%q)", sub)
		err = q.ExchangeDeclare(sub, "fanout")
		if err != nil {
			return fmt.Errorf("Exchange Declare: %s", err)
		}

		s.logger.Printf("Binding %s to Exchange %q", queue.Name, sub)
		err = q.QueueBind(queue.Name, "", sub)
		if err != nil {
			return fmt.Errorf("Queue Bind: %s", err)
		}
	}

	s.logger.Printf("Starting Consume on queue: " + queue.Name)
	s.deliveries, err = q.Consume(queue.Name, "")
	if err != nil {
		return fmt.Errorf("Queue Consume: %s", err)
	}

	s.done = make(chan error)
	return nil
}

func (s *Subscriber) Start() {
	//go s.handle(s.deliveries, s.done)
	var d amqp.Delivery
	for {
		select {
		case d = <-s.deliveries:
			go s.handle(d)
		case <-s.done:
			return
		}
	}
}

func (s *Subscriber) Stop() {
	s.logger.Print("STOP: Shutting down subscribers")
	s.done <- nil
}

func (s *Subscriber) handle(d amqp.Delivery) {
	clientConfig := s.config.Client

	defer func() {
		if r := recover(); r != nil {
			s.logger.Printf("Caught Panic on Close. %+v", r)
		}
	}()

	if nil == d.Body {
		s.logger.Println("Delivery had nil body")
		d.Reject(false) // discard this message
		return
	}

	checkConfig := new(plugins.PluginConfig)
	err := json.Unmarshal(d.Body, checkConfig)
	if nil != err {
		s.logger.Printf("Unable to decode message, skipping...")
		d.Reject(false)
		return
	}

	//s.logger.Printf("Our check consists of: %+v", checkConfig)
	s.logger.Printf("Running '%s'", checkConfig.Name)

	theJob := getCheckHandler(checkConfig.Name, checkConfig.Type)

	result := NewResult(clientConfig, checkConfig.Name)
	result.SetCommand(checkConfig.Command)

	presult := new(plugins.Result)

	theJob.Init(*checkConfig)

	err = theJob.Gather(presult)
	result.SetOutput(presult.Output())
	result.SetCheckStatus(theJob.GetStatus())

	if nil != err {
		// returned an error - we should stop this job from running
		s.logger.Printf("Failed to gather stat: %s. %v", checkConfig.Name, err)
		return
	}

	// and now send it back
	if result.HasOutput() {
		if err = s.q.Publish(RESULTS_QUEUE, "", result.GetPayload()); err != nil {
			s.logger.Printf("Error Publishing Stats: %v. %v", err, result)
		}
	}

	d.Ack(false)

}
