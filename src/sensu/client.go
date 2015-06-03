package sensu

import (
	"github.com/streadway/amqp"
	"log"
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
}

func NewClient(c *Config, p []Processor) *Client {
	return &Client{
		config:    c,
		processes: p,
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
					panic(err) //TODO: Add recovery error handling
				}
				go proc.Start()
			}
			// Enable disconnect channel
			disconnected = c.q.Disconnected()

		case errd := <-disconnected:
			// Disable disconnect channel
			disconnected = nil

			log.Printf("RabbitMQ disconnected: %s", errd)
			c.Stop(false)

			go c.q.Connect(connected)

		case <-stop:
			c.Shutdown()
			return
		}
	}
}

func (c *Client) Stop(force bool) {
	log.Print("STOP: Closing down processes")
	for _, proc := range c.processes {
		proc.Stop(force)
	}
}

func (c *Client) Shutdown() {
	// Disconnect rabbitmq
	c.q.Disconnect()
	c.Stop(true)
}
