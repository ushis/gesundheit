package main

import (
	"context"
	"log"
	"sync"

	"github.com/ushis/gesundheit/check"
	"github.com/ushis/gesundheit/db"
	"github.com/ushis/gesundheit/handler"
	"github.com/ushis/gesundheit/input"
	"github.com/ushis/gesundheit/result"
)

type hub struct {
	db           db.Database
	checkRunners []check.Runner
	inputRunners []input.Runner
	handlers     []handler.Handler
}

func (h *hub) registerCheckRunner(r check.Runner) {
	h.checkRunners = append(h.checkRunners, r)
}

func (h *hub) registerInputRunner(r input.Runner) {
	h.inputRunners = append(h.inputRunners, r)
}

func (h *hub) registerHandler(r handler.Handler) {
	h.handlers = append(h.handlers, r)
}

func (h *hub) run(ctx context.Context, wg *sync.WaitGroup) error {
	ctx, cancelRunners := context.WithCancel(ctx)
	runnersWg := sync.WaitGroup{}
	events := make(chan result.Event)

	if err := h.runRunners(ctx, &runnersWg, events); err != nil {
		cancelRunners()
		runnersWg.Wait()
		return err
	}
	wg.Add(2)

	go func() {
		for e := range events {
			h.dispatch(e)
		}
		wg.Done()
	}()

	go func() {
		<-ctx.Done()
		cancelRunners()
		runnersWg.Wait()
		close(events)
		wg.Done()
	}()

	return nil
}

func (h *hub) runRunners(ctx context.Context, wg *sync.WaitGroup, events chan<- result.Event) error {
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

func (h *hub) dispatch(e result.Event) {
	if ok, err := h.db.InsertEvent(e); err != nil {
		log.Println(err)
	} else if !ok {
		return
	}

	for _, r := range h.handlers {
		if err := r.Handle(e); err != nil {
			log.Println(err)
		}
	}
}
