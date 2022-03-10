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
	Url   string
	Queue string
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

func (i Input) Run(ctx context.Context, wg *sync.WaitGroup, events chan<- result.Event) error {
	wg.Add(1)

	go func() {
		i.run(ctx, events)
		wg.Done()
	}()

	return nil
}

const reconnectDelay = 4 * time.Second

func (i Input) run(ctx context.Context, out chan<- result.Event) {
	for {
		if err := i.receive(ctx, out); err != nil {
			log.Println(err)
		}
		select {
		case <-time.After(reconnectDelay):
		case <-ctx.Done():
			return
		}
	}
}

func (i Input) receive(ctx context.Context, out chan<- result.Event) error {
	conn, in, err := connect(i.Url, i.Queue)

	if err != nil {
		return err
	}
	defer conn.Close()

	connClosed := make(chan *amqp.Error)
	conn.NotifyClose(connClosed)

	go func() {
		select {
		case <-connClosed:
		case <-ctx.Done():
			conn.Close()
		}
	}()

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
	return nil
}

func connect(url, queue string) (*amqp.Connection, <-chan amqp.Delivery, error) {
	conn, err := amqp.Dial(url)

	if err != nil {
		return nil, nil, fmt.Errorf("amqp: failed to connect: %s", err)
	}
	chn, err := conn.Channel()

	if err != nil {
		conn.Close()
		return nil, nil, fmt.Errorf("amqp: failed to create channel: %s", err)
	}
	if err := chn.Confirm(false); err != nil {
		conn.Close()
		return nil, nil, fmt.Errorf("amqp: failed to put channel in confirmation mode: %s", err)
	}
	if _, err := chn.QueueDeclare(queue, true, false, false, false, nil); err != nil {
		conn.Close()
		return nil, nil, fmt.Errorf("amqp: failed to declare queue: %s", err)
	}
	c, err := chn.Consume(queue, "", false, false, false, false, nil)

	if err != nil {
		conn.Close()
		return nil, nil, fmt.Errorf("amqp: failed to consume: %s", err)
	}
	return conn, c, nil
}
