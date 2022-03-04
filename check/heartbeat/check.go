package heartbeat

import (
	"github.com/ushis/gesundheit/check"
	"github.com/ushis/gesundheit/result"
)

type Check struct{}

func init() {
	check.Register("heartbeat", New)
}

func New(_ check.Database, _ func(interface{}) error) (check.Check, error) {
	return Check{}, nil
}

func (c Check) Exec() result.Result {
	return result.OK("i am alive")
}
