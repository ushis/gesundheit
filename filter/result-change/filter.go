package resultchange

import (
	"github.com/ushis/gesundheit/check"
	"github.com/ushis/gesundheit/filter"
)

type Filter struct{}

func init() {
	filter.Register("result-change", New)
}

func New(configure func(interface{}) error) (filter.Filter, error) {
	return Filter{}, nil
}

func (f Filter) Filter(e check.Event) (check.Event, bool) {
	return e, e.Result.Status != e.StatusHistory.Last()
}
