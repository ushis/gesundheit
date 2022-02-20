package handler

import (
	"github.com/ushis/gesundheit/check"
	"github.com/ushis/gesundheit/filter"
)

type Runner struct {
	handler Handler
	filters []filter.Filter
}

func NewRunner(handler Handler, filters []filter.Filter) *Runner {
	return &Runner{
		handler: handler,
		filters: filters,
	}
}

func (r *Runner) Handle(e check.Event) error {
	if e, ok := r.filter(e); ok {
		return r.handler.Handle(e)
	}
	return nil
}

func (r *Runner) filter(e check.Event) (check.Event, bool) {
	ok := true

	for _, filter := range r.filters {
		if e, ok = filter.Filter(e); !ok {
			return e, ok
		}
	}
	return e, ok
}
