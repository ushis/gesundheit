package filter

import (
	"github.com/ushis/gesundheit/check"
	"github.com/ushis/gesundheit/handler"
)

type Filter interface {
	Filter(check.Event) (check.Event, bool)
}

type filterHandler struct {
	handler handler.Handler
	filters []Filter
}

func (h filterHandler) Handle(e check.Event) error {
	if e, ok := h.filter(e); ok {
		return h.handler.Handle(e)
	}
	return nil
}

func (h filterHandler) filter(e check.Event) (check.Event, bool) {
	ok := true

	for _, filter := range h.filters {
		if e, ok = filter.Filter(e); !ok {
			return e, ok
		}
	}
	return e, ok
}

func Handler(handler handler.Handler, filters []Filter) handler.Handler {
	if len(filters) == 0 {
		return handler
	}
	return filterHandler{handler, filters}
}
