package sensu

import (
	"github.com/streadway/amqp"
	"log"
	"os"
	"time"
)

type Processor interface {
	Init(MessageQueuer, *Config) error
	Start()
	Stop(force bool)
}

type Client struct {
	config    *Config
	processes []Processor
	q         *Rabbitmq
	debug     bool
}

func NewClient(c *Config, p []Processor) *Client {
	return &Client{
		config:    c,
		processes: p,
		debug:     "" != os.Getenv("DEBUG"),
	}
}

func (c *Client) Start(stop chan bool) {
	var disconnected chan *amqp.Error
	connected := make(chan bool)

	c.q = NewRabbitmq(c.config.Rabbitmq)
	go c.q.Connect(connected)

	for {
		select {
		case <-connected:
			for _, proc := range c.processes {
				err := proc.Init(c.q, c.config)
				if err != nil {
					log.Printf("FAIL: Failed to start process %+v", proc)
				} else {
					go proc.Start()
					if c.debug {
						log.Printf("%T was started", proc)
					}
				}
			}
			// Enable disconnect channel
			disconnected = c.q.Disconnected()

		case errd := <-disconnected:
			// Disable disconnect channel
			disconnected = nil

			log.Printf("RabbitMQ disconnected: %s", errd)
			c.Stop(false) // tell our processors to stop if they don't need an amqp connection

			time.Sleep(10 * time.Second)
			go c.q.Connect(connected)

		case <-stop:
			c.Shutdown()
			return
		}
	}
}

// this asks our processors to stop.
// force=true if you really want them to shutdown. this is so we can continue to collect stats in the background
func (c *Client) Stop(force bool) {
	var extra string
	if force {
		extra = "... forcefully!"
	}
	log.Print("STOP: Closing down processes", extra)
	for _, proc := range c.processes {
		if c.debug {
			log.Printf("%T is being stopped %s", proc, extra)
		}
		proc.Stop(force)
	}
}

func (c *Client) Shutdown() {
	// Disconnect rabbitmq
	c.q.Disconnect()
	c.Stop(true)
}
