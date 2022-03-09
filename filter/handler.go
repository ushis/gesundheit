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

func (h filterHandler) Run(wg *sync.WaitGroup) (chan<- result.Event, error) {
	out, err := h.handler.Run(wg)

	if err != nil {
		return nil, err
	}
	in := make(chan result.Event)
	wg.Add(1)

	go func() {
		h.run(out, in)
		wg.Done()
	}()

	return in, nil
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
