package main

import (
	"context"
	"log"
	"sync"

	"github.com/ushis/gesundheit/check"
	"github.com/ushis/gesundheit/handler"
	"github.com/ushis/gesundheit/input"
)

type hub struct {
	checkRunners   []check.Runner
	inputRunners   []input.Runner
	handlerRunners []handler.Handler
}

func (h *hub) registerCheckRunner(r check.Runner) {
	h.checkRunners = append(h.checkRunners, r)
}

func (h *hub) registerInputRunner(r input.Runner) {
	h.inputRunners = append(h.inputRunners, r)
}

func (h *hub) registerHandlerRunner(r handler.Handler) {
	h.handlerRunners = append(h.handlerRunners, r)
}

func (h *hub) run(ctx context.Context) (<-chan struct{}, error) {
	ctx, cancel := context.WithCancel(ctx)
	wg := sync.WaitGroup{}
	events := make(chan check.Event)

	if err := h.runRunners(ctx, &wg, events); err != nil {
		cancel()
		wg.Wait()
		return nil, err
	}
	done := make(chan struct{})

	go func() {
		for e := range events {
			h.dispatch(e)
		}
		close(done)
	}()

	go func() {
		<-ctx.Done()
		cancel()
		wg.Wait()
		close(events)
	}()

	return done, nil
}

func (h *hub) runRunners(ctx context.Context, wg *sync.WaitGroup, events chan<- check.Event) error {
	for _, r := range h.inputRunners {
		if err := r.Run(ctx, wg, events); err != nil {
			return err
		}
	}
	for _, r := range h.checkRunners {
		if err := r.Run(ctx, wg, events); err != nil {
			return err
		}
	}
	return nil
}

func (h *hub) dispatch(e check.Event) {
	for _, r := range h.handlerRunners {
		if err := r.Handle(e); err != nil {
			log.Println(err)
		}
	}
}
