package filter

import (
	"sync"

	"github.com/ushis/gesundheit/handler"
	"github.com/ushis/gesundheit/result"
)

func Handler(handler handler.Handler, filters []Filter) handler.Handler {
	if len(filters) == 0 {
		return handler
	}
	return filterHandler{handler, filters}
}

type filterHandler struct {
	handler handler.Handler
	filters []Filter
}

func (h filterHandler) Run(wg *sync.WaitGroup, in <-chan result.Event) error {
	out := make(chan result.Event)

	if err := h.handler.Run(wg, out); err != nil {
		close(out)
		return err
	}
	wg.Add(1)

	go func() {
		h.run(out, in)
		close(out)
		wg.Done()
	}()

	return nil
}

func (h filterHandler) run(out chan<- result.Event, in <-chan result.Event) {
	for e := range in {
		if e, ok := h.filter(e); ok {
			out <- e
		}
	}
}

func (h filterHandler) filter(e result.Event) (result.Event, bool) {
	ok := true

	for _, filter := range h.filters {
		if e, ok = filter.Filter(e); !ok {
			return e, ok
		}
	}
	return e, ok
}
