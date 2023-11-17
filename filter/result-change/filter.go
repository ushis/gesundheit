package resultchange

import (
	"github.com/ushis/gesundheit/db"
	"github.com/ushis/gesundheit/filter"
	"github.com/ushis/gesundheit/result"
)

type Filter struct {
	db db.Database
}

func init() {
	filter.Register("result-change", New)
}

func New(db db.Database, _ func(interface{}) error) (filter.Filter, error) {
	return Filter{db}, nil
}

func (f Filter) Filter(e result.Event) (result.Event, bool) {
	events, err := f.db.GetEventsByCheck(e.NodeName, e.CheckId)

	if err != nil || len(events) == 0 {
		return e, e.Status == result.StatusFail
	}
	index := indexOfEvent(events, e)

	if index == 0 {
		return e, e.Status == result.StatusFail
	}
	if index < 0 {
		index = len(events)
	}
	return e, e.Status != events[index-1].Status
}

func indexOfEvent(events []result.Event, e result.Event) int {
	for i, event := range events {
		if event.Id == e.Id {
			return i
		}
	}
	return -1
}
