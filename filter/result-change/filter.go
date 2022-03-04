package resultchange

import (
	"github.com/ushis/gesundheit/filter"
	"github.com/ushis/gesundheit/result"
)

type Filter struct{}

func init() {
	filter.Register("result-change", New)
}

func New(_ func(interface{}) error) (filter.Filter, error) {
	return Filter{}, nil
}

func (f Filter) Filter(e result.Event) (result.Event, bool) {
	return e, e.Status != e.StatusHistory.Last()
}
