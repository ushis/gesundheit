package main

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/ushis/gesundheit/db"
	"github.com/ushis/gesundheit/result"
)

type producer interface {
	Run(context.Context, *sync.WaitGroup, chan<- result.Event) error
}

type consumer interface {
	Run(*sync.WaitGroup, <-chan result.Event) error
}

type hub struct {
	db        db.Database
	producers []producer
	consumers []consumer
}

func (h *hub) registerProducer(p producer) {
	h.producers = append(h.producers, p)
}

func (h *hub) registerConsumer(c consumer) {
	h.consumers = append(h.consumers, c)
}

func (h *hub) run(ctx context.Context, wg *sync.WaitGroup) error {
	ctx, cancelProds := context.WithCancel(ctx)
	prodsWg := sync.WaitGroup{}
	in, err := h.runProducers(ctx, &prodsWg)

	if err != nil {
		cancelProds()
		prodsWg.Wait()
		close(in)
		return err
	}
	consWg := sync.WaitGroup{}
	outs, err := h.runConsumers(&consWg)

	if err != nil {
		cancelProds()
		prodsWg.Wait()
		close(in)
		closeAll(outs)
		consWg.Wait()
		return err
	}
	wg.Add(2)

	go func() {
		h.dispatch(outs, in)
		closeAll(outs)
		consWg.Wait()
		wg.Done()
	}()

	go func() {
		<-ctx.Done()
		cancelProds()
		prodsWg.Wait()
		close(in)
		wg.Done()
	}()

	return nil
}

func (h *hub) runProducers(ctx context.Context, wg *sync.WaitGroup) (chan result.Event, error) {
	chn := make(chan result.Event)

	for _, p := range h.producers {
		if err := p.Run(ctx, wg, chn); err != nil {
			return chn, err
		}
	}
	return chn, nil
}

func (h *hub) runConsumers(wg *sync.WaitGroup) ([]chan<- result.Event, error) {
	chns := make([]chan<- result.Event, len(h.consumers))

	for i, c := range h.consumers {
		chn := make(chan result.Event)
		chns[i] = chn

		if err := c.Run(wg, chn); err != nil {
			return chns[:i+1], err
		}
	}
	return chns, nil
}

func (h *hub) dispatch(outs []chan<- result.Event, in <-chan result.Event) {
	for e := range in {
		if time.Now().After(e.ExpiresAt) {
			continue
		}
		ok, err := h.db.InsertEvent(e)

		if err != nil {
			log.Println(err)
		} else if !ok {
			continue
		}
		for _, out := range outs {
			out <- e
		}
	}
}

func closeAll[T any](chans []chan<- T) {
	for _, chn := range chans {
		close(chn)
	}
}
