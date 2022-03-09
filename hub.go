package main

import (
	"context"
	"log"
	"sync"

	"github.com/ushis/gesundheit/db"
	"github.com/ushis/gesundheit/result"
)

type producer interface {
	Run(context.Context, *sync.WaitGroup, chan<- result.Event) error
}

type consumer interface {
	Run(*sync.WaitGroup) (chan<- result.Event, error)
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
	ctx, cancel := context.WithCancel(ctx)
	prodsWg := sync.WaitGroup{}
	in, err := h.runProducers(ctx, &prodsWg)

	if err != nil {
		cancel()
		prodsWg.Wait()
		close(in)
		return err
	}
	consWg := sync.WaitGroup{}
	outs, err := h.runConsumers(&consWg)

	if err != nil {
		cancel()
		prodsWg.Wait()
		close(in)
		closeAll(outs)
		consWg.Wait()
		return err
	}
	wg.Add(2)

	go func() {
		h.dispatch(outs, in)
		wg.Done()
	}()

	go func() {
		<-ctx.Done()
		cancel()
		prodsWg.Wait()
		close(in)
		closeAll(outs)
		consWg.Wait()
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
	chans := make([]chan<- result.Event, len(h.consumers))

	for i, c := range h.consumers {
		if out, err := c.Run(wg); err != nil {
			return chans[:i], err
		} else {
			chans[i] = out
		}
	}
	return chans, nil
}

func (h *hub) dispatch(outs []chan<- result.Event, in <-chan result.Event) {
	for e := range in {
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

func closeAll(chans []chan<- result.Event) {
	for _, chn := range chans {
		close(chn)
	}
}
