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

	interval time.Duration
	started  bool
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
	k.interval = keepaliveInterval

	// did the user set a custom interval for keep alive?
	if json := config.Data().GetPath("client", "keepalive", "interval"); json != nil {
		if d, err := json.Uint64(); err == nil {
			k.interval = time.Duration(d) * time.Second
		}
	}
	k.logger.Printf("Keepalive Interval: %d seconds", k.interval/time.Second)

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
    k.started = true

	for {
		select {
		case <-reset:
			timer.Reset(k.interval)
		case <-k.close:
			return
		}
	}
}

func (k *Keepalive) Stop(force bool) {
    if k.started {
        k.logger.Print("STOP: Shutting Down")
        k.close <- true
    }
    k.started = false
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
