package amqp

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/streadway/amqp"
	"github.com/ushis/gesundheit/input"
	"github.com/ushis/gesundheit/result"
)

type Input struct {
	Url      string
	Exchange string
	Queue    string
}

func init() {
	input.Register("amqp", New)
}

func New(configure func(interface{}) error) (input.Input, error) {
	input := Input{}

	if err := configure(&input); err != nil {
		return nil, err
	}
	return input, nil
}

func (i Input) Run(ctx context.Context, wg *sync.WaitGroup, out chan<- result.Event) error {
	wg.Add(1)

	go func() {
		i.run(ctx, wg, out)
		wg.Done()
	}()

	return nil
}

const reconnectDelay = 4 * time.Second

func (i Input) run(ctx context.Context, wg *sync.WaitGroup, out chan<- result.Event) {
	for {
		if err := i.connectAndReceive(ctx, wg, out); err != nil {
			log.Println(err)
		}
		select {
		case <-time.After(reconnectDelay):
		case <-ctx.Done():
			return
		}
	}
}

func (i Input) connectAndReceive(ctx context.Context, wg *sync.WaitGroup, out chan<- result.Event) error {
	conn, in, err := connect(i.Url, i.Exchange, i.Queue)

	if err != nil {
		return err
	}
	wg.Add(1)

	go func() {
		i.receive(out, in)
		wg.Done()
	}()

	connClosed := make(chan *amqp.Error)
	conn.NotifyClose(connClosed)

	select {
	case err := <-connClosed:
		return err
	case <-ctx.Done():
		return conn.Close()
	}
}

func (i Input) receive(out chan<- result.Event, in <-chan amqp.Delivery) {
	for d := range in {
		e := result.Event{}

		if err := json.Unmarshal(d.Body, &e); err != nil {
			log.Println("amqp: failed to decode event:", err)
			d.Nack(false, false)
		} else {
			out <- e
			d.Ack(false)
		}
	}
}

func connect(url, exchange, queue string) (*amqp.Connection, <-chan amqp.Delivery, error) {
	conn, err := amqp.Dial(url)

	if err != nil {
		return nil, nil, fmt.Errorf("amqp: failed to connect: %s", err)
	}
	chn, err := conn.Channel()

	if err != nil {
		conn.Close()
		return nil, nil, fmt.Errorf("amqp: failed to create channel: %s", err)
	}
	if err := chn.ExchangeDeclare(exchange, "fanout", true, false, false, false, nil); err != nil {
		conn.Close()
		return nil, nil, fmt.Errorf("amqp: failed to declare exchange: %s", err)
	}
	if err := chn.Confirm(false); err != nil {
		conn.Close()
		return nil, nil, fmt.Errorf("amqp: failed to put channel in confirmation mode: %s", err)
	}
	if _, err := chn.QueueDeclare(queue, true, false, false, false, nil); err != nil {
		conn.Close()
		return nil, nil, fmt.Errorf("amqp: failed to declare queue: %s", err)
	}
	if err := chn.QueueBind(queue, "", exchange, false, nil); err != nil {
		conn.Close()
		return nil, nil, fmt.Errorf("amqp: failed to bind queue to exchange: %s", err)
	}
	c, err := chn.Consume(queue, "", false, false, false, false, nil)

	if err != nil {
		conn.Close()
		return nil, nil, fmt.Errorf("amqp: failed to consume: %s", err)
	}
	return conn, c, nil
}
