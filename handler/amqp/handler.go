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
	session := newSession(h.Url, h.Exchange)
	ctx, cancel := context.WithCancel(context.Background())
	chn := make(chan result.Event)
	wg.Add(2)

	go func() {
		h.send(session, chn)
		cancel()
		wg.Done()
	}()

	go func() {
		session.run(ctx)
		wg.Done()
	}()

	return chn, nil
}

func (h Handler) send(s *session, in <-chan result.Event) {
	for e := range in {
		if err := s.send(e); err != nil {
			log.Println(err)
		}
	}
}

type session struct {
	ready     bool
	url       string
	exchange  string
	conn      *amqp.Connection
	chn       *amqp.Channel
	chnClosed chan *amqp.Error
	confirms  chan amqp.Confirmation
}

func newSession(url, exchange string) *session {
	return &session{ready: false, url: url, exchange: exchange}
}

const reconnectDelay = 4 * time.Second

func (s *session) run(ctx context.Context) {
	for {
		if err := s.connect(); err != nil {
			log.Println(err)

			select {
			case <-time.After(reconnectDelay):
			case <-ctx.Done():
				return
			}
		} else {
			s.ready = true

			select {
			case err := <-s.chnClosed:
				s.ready = false
				log.Println(err)
			case <-ctx.Done():
				s.ready = false
				s.conn.Close()
				return
			}
		}
	}
}

func (s *session) connect() (err error) {
	s.conn, err = amqp.Dial(s.url)

	if err != nil {
		return fmt.Errorf("amqp: failed to connect: %s", err)
	}
	s.chn, err = s.conn.Channel()

	if err != nil {
		s.conn.Close()
		return fmt.Errorf("amqp: failed to create channel: %s", err)
	}
	if err := s.chn.ExchangeDeclare(s.exchange, "fanout", true, false, false, false, nil); err != nil {
		s.conn.Close()
		return fmt.Errorf("amqp: failed to declare exchange: %s", err)
	}
	if err := s.chn.Confirm(false); err != nil {
		s.conn.Close()
		return fmt.Errorf("amqp: failed to put channel in confirmation mode: %s", err)
	}
	s.chnClosed = make(chan *amqp.Error)
	s.chn.NotifyClose(s.chnClosed)

	s.confirms = make(chan amqp.Confirmation)
	s.chn.NotifyPublish(s.confirms)

	return nil
}

func (s *session) send(e result.Event) error {
	if !s.ready {
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
	if err := s.chn.Publish(s.exchange, "", false, false, msg); err != nil {
		return fmt.Errorf("amqp: failed to send event: %s", err)
	}
	if c := <-s.confirms; !c.Ack {
		return fmt.Errorf("amqp: broker rejected message: %s", e.Id)
	}
	return nil
}
