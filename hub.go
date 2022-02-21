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
	checkRunners   []*check.Runner
	handlerRunners []*handler.Runner
	inputRunners   []*input.Runner
}

func newHub() *hub {
	return &hub{
		checkRunners:   []*check.Runner{},
		handlerRunners: []*handler.Runner{},
		inputRunners:   []*input.Runner{},
	}
}

func (h *hub) registerCheckRunner(r *check.Runner) {
	h.checkRunners = append(h.checkRunners, r)
}

func (h *hub) registerHandlerRunner(r *handler.Runner) {
	h.handlerRunners = append(h.handlerRunners, r)
}

func (h *hub) registerInputRunner(r *input.Runner) {
	h.inputRunners = append(h.inputRunners, r)
}

func (h *hub) run(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()

	events := make(chan check.Event)
	defer close(events)

	rwg := sync.WaitGroup{}
	rwg.Add(len(h.checkRunners))

	for _, r := range h.checkRunners {
		go r.Run(ctx, &rwg, events)
	}
	for _, r := range h.inputRunners {
		go r.Run(events)
	}
	for {
		select {
		case e := <-events:
			for _, r := range h.handlerRunners {
				if err := r.Handle(e); err != nil {
					log.Println(err)
				}
			}
		case <-ctx.Done():
			rwg.Wait()
			return
		}
	}
}
