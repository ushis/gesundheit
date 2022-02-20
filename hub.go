package main

import (
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
	events         chan check.Event
	done           chan struct{}
	wg             sync.WaitGroup
}

func newHub() *hub {
	return &hub{
		checkRunners:   []*check.Runner{},
		handlerRunners: []*handler.Runner{},
		inputRunners:   []*input.Runner{},
		events:         make(chan check.Event),
		done:           make(chan struct{}),
		wg:             sync.WaitGroup{},
	}
}

func (h *hub) registerCheckRunner(fn func() *check.Runner) {
	h.checkRunners = append(h.checkRunners, fn())
}

func (h *hub) registerHandlerRunner(fn func() *handler.Runner) {
	h.handlerRunners = append(h.handlerRunners, fn())
}

func (h *hub) registerInputRunner(fn func() *input.Runner) {
	h.inputRunners = append(h.inputRunners, fn())
}

func (h *hub) run() {
	h.wg.Add(1)

	for _, r := range h.checkRunners {
		go r.Run(h.events)
	}
	for _, r := range h.inputRunners {
		go r.Run(h.events)
	}
	for {
		select {
		case e := <-h.events:
			for _, r := range h.handlerRunners {
				if err := r.Handle(e); err != nil {
					log.Println(err)
				}
			}
		case <-h.done:
			for _, r := range h.checkRunners {
				r.Stop()
			}
			for _, r := range h.handlerRunners {
				r.Close()
			}
			h.wg.Done()
			return
		}
	}
}

func (h *hub) stop() {
	close(h.done)
	h.wg.Wait()
	close(h.events)
}
