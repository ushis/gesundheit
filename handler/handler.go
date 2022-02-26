package handler

import (
	"github.com/ushis/gesundheit/check"
	"github.com/ushis/gesundheit/filter"
)

type Handler interface {
	Handle(check.Event) error
}

type FilteredHandler struct {
	Handler Handler
	Filters []filter.Filter
}

func (h FilteredHandler) Handle(e check.Event) error {
	if e, ok := h.filter(e); ok {
		return h.Handler.Handle(e)
	}
	return nil
}

func (h FilteredHandler) filter(e check.Event) (check.Event, bool) {
	ok := true

	for _, filter := range h.Filters {
		if e, ok = filter.Filter(e); !ok {
			return e, ok
		}
	}
	return e, ok
}
