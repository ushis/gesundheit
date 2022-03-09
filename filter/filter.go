package filter

import (
	"github.com/ushis/gesundheit/result"
)

type Filter interface {
	Filter(result.Event) (result.Event, bool)
}
