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
	Url   string
	Queue string
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
	ctx, cancel := context.WithCancel(context.Background())
	in := make(chan result.Event)
	out := make(chan result.Event)
	wg.Add(2)

	go func() {
		for e := range in {
			out <- e
		}
		cancel()
		wg.Done()
	}()

	go func() {
		h.run(ctx, out)
		close(out)
		wg.Done()
	}()

	return in, nil
}

const reconnectDelay = 4 * time.Second

func (h Handler) run(ctx context.Context, in <-chan result.Event) {
	for {
		if err := h.publish(ctx, in); err != nil {
			log.Println(err)
		}
		select {
		case <-time.After(reconnectDelay):
		case <-ctx.Done():
			return
		}
	}
}

func (h Handler) publish(ctx context.Context, in <-chan result.Event) error {
	conn, chn, err := connect(h.Url, h.Queue)

	if err != nil {
		return err
	}
	defer conn.Close()

	chnClosed := make(chan *amqp.Error)
	chn.NotifyClose(chnClosed)

	msgConfirmed := make(chan amqp.Confirmation)
	chn.NotifyPublish(msgConfirmed)

	for {
		select {
		case e := <-in:
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
			if err := chn.Publish("", h.Queue, false, false, msg); err != nil {
				return fmt.Errorf("amqp: failed to publish event: %s", err)
			}
			if c := <-msgConfirmed; !c.Ack {
				return fmt.Errorf("amqp: broker rejected message: %s", e.Id)
			}
		case err := <-chnClosed:
			return err
		case <-ctx.Done():
			return nil
		}
	}
}

func connect(url, queue string) (*amqp.Connection, *amqp.Channel, error) {
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
	return conn, chn, nil
}
