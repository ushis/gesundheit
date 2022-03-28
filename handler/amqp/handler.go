package amqp

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sync"

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

func (h Handler) Run(wg *sync.WaitGroup, chn <-chan result.Event) error {
	client := newClient(h.Url, h.Exchange)
	wg.Add(1)

	go func() {
		h.send(client, chn)
		wg.Done()
	}()

	return nil
}

func (h Handler) send(c *client, in <-chan result.Event) {
	ctx, cancel := context.WithCancel(context.Background())

	for e := range in {
		if err := c.send(ctx, e); err != nil {
			log.Println(err)
		}
	}
	cancel()
}

type client struct {
	url      string
	exchange string
	conn     *clientConn
}

type clientConn struct {
	channel       *amqp.Channel
	confirmations chan amqp.Confirmation
}

func newClient(url, exchange string) *client {
	return &client{url: url, exchange: exchange, conn: nil}
}

func (c *client) connection(ctx context.Context) (*clientConn, error) {
	conn := c.conn

	if conn != nil {
		return conn, nil
	}
	return c.connect(ctx)
}

func (c *client) connect(ctx context.Context) (*clientConn, error) {
	conn, err := amqp.Dial(c.url)

	if err != nil {
		return nil, fmt.Errorf("amqp: failed to connect: %s", err)
	}
	channel, err := conn.Channel()

	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("amqp: failed to create channel: %s", err)
	}
	if err := channel.ExchangeDeclare(c.exchange, "fanout", true, false, false, false, nil); err != nil {
		conn.Close()
		return nil, fmt.Errorf("amqp: failed to declare exchange: %s", err)
	}
	if err := channel.Confirm(false); err != nil {
		conn.Close()
		return nil, fmt.Errorf("amqp: failed to put channel in confirmation mode: %s", err)
	}
	confirmations := make(chan amqp.Confirmation)
	channel.NotifyPublish(confirmations)

	channelClosed := make(chan *amqp.Error)
	channel.NotifyClose(channelClosed)

	clientConn := &clientConn{channel, confirmations}
	c.conn = clientConn

	go func() {
		select {
		case err := <-channelClosed:
			c.conn = nil
			conn.Close()
			log.Println(err)
		case <-ctx.Done():
			c.conn = nil
			conn.Close()
		}
	}()

	return clientConn, nil
}

func (c *client) send(ctx context.Context, e result.Event) error {
	conn, err := c.connection(ctx)

	if err != nil {
		return err
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
	if err := conn.channel.Publish(c.exchange, "", false, false, msg); err != nil {
		return fmt.Errorf("amqp: failed to send event: %s", err)
	}
	if conf := <-conn.confirmations; !conf.Ack {
		return fmt.Errorf("amqp: broker rejected message: %s", e.Id)
	}
	return nil
}
