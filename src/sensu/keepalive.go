package sensu

import (
	"fmt"
	"github.com/bitly/go-simplejson"
	"github.com/streadway/amqp"
	"io"
	"log"
	"time"
)

type Keepalive struct {
	q      MessageQueuer
	config *Config
	close  chan bool
	logger *log.Logger
}

func NewKeepalive(w io.Writer) *Keepalive {
	k := new(Keepalive)
	k.logger = log.New(w, "Keepalive: ", log.LstdFlags)
	return k
}

const keepaliveInterval = 20 * time.Second

func (k *Keepalive) Init(q MessageQueuer, config *Config) error {
	if err := q.ExchangeDeclare(
		"keepalives",
		"direct",
	); err != nil {
		return fmt.Errorf("Exchange Declare: %s", err)
	}

	k.q = q
	k.config = config
	k.close = make(chan bool)

	return nil
}

func (k *Keepalive) Start() {
	clientConfig := k.config.Data().Get("client")
	reset := make(chan bool)
	timer := time.AfterFunc(0, func() {
		payload := createKeepalivePayload(clientConfig, time.Now())
		k.publish(payload)
		reset <- true
	})
	defer timer.Stop()

	for {
		select {
		case <-reset:
			timer.Reset(keepaliveInterval)
		case <-k.close:
			return
		}
	}
}

func (k *Keepalive) Stop(force bool) {
	k.logger.Print("STOP: Shutting Down")
	k.close <- true
}

func (k *Keepalive) publish(payload amqp.Publishing) {
	if err := k.q.Publish(
		"keepalives",
		"",
		payload,
	); err != nil {
		k.logger.Printf("keepalive.publish: %v", err)
		return
	}
	k.logger.Print("Keepalive published")
}

func createKeepalivePayload(clientConfig *simplejson.Json, timestamp time.Time) amqp.Publishing {
	payload := clientConfig
	payload.Set("timestamp", int64(timestamp.Unix()))
	body, _ := payload.MarshalJSON()
	return amqp.Publishing{
		ContentType:  "application/octet-stream",
		Body:         body,
		DeliveryMode: amqp.Transient,
	}
}
