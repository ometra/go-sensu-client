package sensu

import (
	"crypto/tls"
	"fmt"
	"log"
	"net/url"
	"strconv"
	"time"

	"github.com/streadway/amqp"
)

type MessageQueuer interface {
	Connect(connected chan bool)
	Disconnected() chan *amqp.Error
	ExchangeDeclare(name string, kind string) error
	QueueDeclare(name string) (amqp.Queue, error)
	QueueBind(name, key, source string) error
	Consume(name, consumer string) (<-chan amqp.Delivery, error)
	Publish(exchange string, key string, msg amqp.Publishing) error
}

type Rabbitmq struct {
	uri          string
	tlsConfig    *tls.Config
	conn         *amqp.Connection
	channel      *amqp.Channel
	disconnected chan *amqp.Error
	connected    bool
}

// back off logic
const rabbitmqRetryInterval = 2
const rabbitmqRetryIntervalMax = 120

func NewRabbitmq(cfg RabbitmqConfig) *Rabbitmq {
	isTLS := false
	tlsConfig := new(tls.Config)
	if cfg.Ssl.CertChainFile != "" && cfg.Ssl.PrivateKeyFile != "" {
		log.Printf("In CERTS: %s", cfg.Ssl)
		tlsConfig.InsecureSkipVerify = true
		if cert, err := tls.LoadX509KeyPair(cfg.Ssl.CertChainFile, cfg.Ssl.PrivateKeyFile); err == nil {
			tlsConfig.Certificates = append(tlsConfig.Certificates, cert)
			isTLS = true
		}
	}

	uri := createRabbitmqUri(cfg, isTLS)

	return &Rabbitmq{uri: uri, tlsConfig: tlsConfig}
}

func (r *Rabbitmq) Connect(connected chan bool) {
	reset := make(chan bool)
	done := make(chan bool)
	timer := time.AfterFunc(0, func() {
		r.connect(r.uri, done)
		reset <- true
	})
	defer timer.Stop()

	var backoffIntervalCounter, backoffInterval int64

	for {
		select {
		case <-done:
			log.Println("RabbitMQ connected and channel established")
			r.connected = true
			connected <- true
			backoffIntervalCounter = 0
			backoffInterval = 0
			return
		case <-reset:
			r.connected = false
			backoffIntervalCounter++
			if 0 == backoffInterval {
				backoffInterval = rabbitmqRetryInterval
			} else {
				backoffInterval = backoffInterval * rabbitmqRetryInterval
			}

			if backoffInterval > rabbitmqRetryIntervalMax {
				backoffInterval = rabbitmqRetryIntervalMax
			}

			log.Printf("Failed to connect, attempt %d, Retrying in %d seconds", backoffIntervalCounter, backoffInterval)

			timer.Reset(time.Duration(backoffInterval) * time.Second)
		}
	}
}

func (r *Rabbitmq) Disconnect() {
	if r.connected {
		r.conn.Close()
	}
	r.connected = false
}

func (r *Rabbitmq) Disconnected() chan *amqp.Error {
	return r.disconnected
}

func (r *Rabbitmq) ExchangeDeclare(name, kind string) error {
	return r.channel.ExchangeDeclare(
		name,
		kind,
		false, // All exchanges are not declared durable
		false,
		false,
		false,
		nil,
	)
}

func (r *Rabbitmq) QueueDeclare(name string) (amqp.Queue, error) {
	return r.channel.QueueDeclare(
		name,
		false,
		true,
		false,
		false,
		nil,
	)
}

func (r *Rabbitmq) QueueBind(name, key, source string) error {
	return r.channel.QueueBind(
		name,
		key,
		source,
		false,
		nil,
	)
}

func (r *Rabbitmq) Consume(name, consumer string) (<-chan amqp.Delivery, error) {
	return r.channel.Consume(
		name,
		consumer,
		false,
		false,
		false,
		false,
		nil,
	)
}

func (r *Rabbitmq) Publish(exchange, key string, msg amqp.Publishing) error {
	return r.channel.Publish(
		exchange,
		key,
		false,
		false,
		msg,
	)
}

func (r *Rabbitmq) connect(uri string, done chan bool) {
	var err error

	log.Printf("Dialing %q", uri)
	if len(r.tlsConfig.Certificates) > 0 {
		r.conn, err = amqp.DialTLS(uri, r.tlsConfig)
	} else {
		r.conn, err = amqp.Dial(uri)
	}
	if err != nil {
		log.Printf("Dial: %s", err)
		return
	}

	log.Printf("Connection established, getting Channel")
	r.channel, err = r.conn.Channel()
	if err != nil {
		log.Printf("Channel: %s", err)
		return
	}

	// Notify disconnect channel when disconnected
	r.disconnected = make(chan *amqp.Error)
	r.channel.NotifyClose(r.disconnected)

	done <- true
}

func createRabbitmqUri(cfg RabbitmqConfig, isTLS bool) string {
	scheme := "amqp"
	if isTLS {
		scheme = "amqps"
	}

	u := url.URL{
		Scheme: scheme,
		Host:   fmt.Sprintf("%s:%s", cfg.Host, strconv.FormatInt(int64(cfg.Port), 10)),
		Path:   fmt.Sprintf("/%s", cfg.Vhost),
		User:   url.UserPassword(cfg.User, cfg.Password),
	}
	return u.String()
}
