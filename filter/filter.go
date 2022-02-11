package filter

import "github.com/ushis/gesundheit/check"

type Filter interface {
	Filter(check.Event) (check.Event, bool)
}
