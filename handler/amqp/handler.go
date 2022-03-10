package amqp

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/streadway/amqp"
	"github.com/ushis/gesundheit/handler"
	"github.com/ushis/gesundheit/result"
)

type Handler struct {
	Url      string
	Exchange string
}

func init() {
	handler.Register("amqp", New)
}

func New(configure func(interface{}) error) (handler.Handler, error) {
	handler := Handler{}

	if err := configure(&handler); err != nil {
		return nil, err
	}
	return handler, nil
}

func (h Handler) Run(wg *sync.WaitGroup) (chan<- result.Event, error) {
	client := newClient(h.Url, h.Exchange)
	ctx, cancel := context.WithCancel(context.Background())
	chn := make(chan result.Event)
	wg.Add(2)

	go func() {
		h.send(client, chn)
		cancel()
		wg.Done()
	}()

	go func() {
		client.run(ctx)
		wg.Done()
	}()

	return chn, nil
}

func (h Handler) send(c *client, in <-chan result.Event) {
	for e := range in {
		if err := c.send(e); err != nil {
			log.Println(err)
		}
	}
}

type client struct {
	ready     bool
	url       string
	exchange  string
	conn      *amqp.Connection
	chn       *amqp.Channel
	chnClosed chan *amqp.Error
	confirms  chan amqp.Confirmation
}

func newClient(url, exchange string) *client {
	return &client{ready: false, url: url, exchange: exchange}
}

const reconnectDelay = 4 * time.Second

func (c *client) run(ctx context.Context) {
	for {
		if err := c.connect(); err != nil {
			log.Println(err)

			select {
			case <-time.After(reconnectDelay):
			case <-ctx.Done():
				return
			}
		} else {
			c.ready = true

			select {
			case err := <-c.chnClosed:
				c.ready = false
				log.Println(err)
			case <-ctx.Done():
				c.ready = false
				c.conn.Close()
				return
			}
		}
	}
}

func (c *client) connect() (err error) {
	c.conn, err = amqp.Dial(c.url)

	if err != nil {
		return fmt.Errorf("amqp: failed to connect: %s", err)
	}
	c.chn, err = c.conn.Channel()

	if err != nil {
		c.conn.Close()
		return fmt.Errorf("amqp: failed to create channel: %s", err)
	}
	if err := c.chn.ExchangeDeclare(c.exchange, "fanout", true, false, false, false, nil); err != nil {
		c.conn.Close()
		return fmt.Errorf("amqp: failed to declare exchange: %s", err)
	}
	if err := c.chn.Confirm(false); err != nil {
		c.conn.Close()
		return fmt.Errorf("amqp: failed to put channel in confirmation mode: %s", err)
	}
	c.chnClosed = make(chan *amqp.Error)
	c.chn.NotifyClose(c.chnClosed)

	c.confirms = make(chan amqp.Confirmation)
	c.chn.NotifyPublish(c.confirms)

	return nil
}

func (c *client) send(e result.Event) error {
	if !c.ready {
		return fmt.Errorf("amqp: failed to send event: connection closed")
	}
	body, err := json.Marshal(e)

	if err != nil {
		return fmt.Errorf("amqp: failed to encode event: %s", err)
	}
	msg := amqp.Publishing{
		ContentType:  "application/json; charset=utf-8",
		DeliveryMode: amqp.Persistent,
		MessageId:    e.Id,
		Timestamp:    e.Timestamp,
		Body:         body,
	}
	if err := c.chn.Publish(c.exchange, "", false, false, msg); err != nil {
		return fmt.Errorf("amqp: failed to send event: %s", err)
	}
	if conf := <-c.confirms; !conf.Ack {
		return fmt.Errorf("amqp: broker rejected message: %s", e.Id)
	}
	return nil
}
