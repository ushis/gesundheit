package nodealive

import (
	"time"

	"github.com/ushis/gesundheit/check"
	"github.com/ushis/gesundheit/result"
)

type Check struct {
	db             check.Database
	NodeName       string
	MaxAbsenceTime time.Duration
}

type Config struct {
	Node           string
	MaxAbsenceTime string
}

func init() {
	check.Register("node-alive", New)
}

func New(db check.Database, configure func(interface{}) error) (check.Check, error) {
	conf := Config{}

	if err := configure(&conf); err != nil {
		return nil, err
	}
	maxAbsenceTime, err := time.ParseDuration(conf.MaxAbsenceTime)

	if err != nil {
		return nil, err
	}
	return Check{db, conf.Node, maxAbsenceTime}, nil
}

func (c Check) Exec() result.Result {
	event, ok := c.db.GetLatestEventByNode(c.NodeName)

	if !ok {
		return result.Fail("haven't seen %s at all", c.NodeName)
	}
	unseenPeriod := time.Since(event.Timestamp).Truncate(time.Second)

	if unseenPeriod > c.MaxAbsenceTime {
		return result.Fail("haven't seen %s for %s", c.NodeName, unseenPeriod)
	}
	return result.OK("saw %s %s ago", c.NodeName, unseenPeriod)
}
