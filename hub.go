package main

import (
	"log"
	"sync"

	"github.com/ushis/gesundheit/check"
	"github.com/ushis/gesundheit/handler"
)

type hub struct {
	checkRunners   []*check.Runner
	handlerRunners []*handler.Runner
	events         chan check.Event
	done           chan struct{}
	wg             sync.WaitGroup
}

func newHub() *hub {
	return &hub{
		checkRunners:   []*check.Runner{},
		handlerRunners: []*handler.Runner{},
		events:         make(chan check.Event),
		done:           make(chan struct{}),
		wg:             sync.WaitGroup{},
	}
}

func (h *hub) registerCheckRunner(fn func(chan<- check.Event) *check.Runner) {
	h.checkRunners = append(h.checkRunners, fn(h.events))
}

func (h *hub) registerHandlerRunner(fn func() *handler.Runner) {
	h.handlerRunners = append(h.handlerRunners, fn())
}

func (h *hub) run() {
	for _, r := range h.checkRunners {
		h.wg.Add(1)
		go r.Run()
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
				h.wg.Done()
			}
			for _, r := range h.handlerRunners {
				r.Close()
			}
			return
		}
	}
}

func (h *hub) stop() {
	close(h.done)
	h.wg.Wait()
}
