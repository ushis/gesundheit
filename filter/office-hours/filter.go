package officehours

import (
	"bytes"
	"errors"
	"strconv"
	"time"

	"github.com/ushis/gesundheit/db"
	"github.com/ushis/gesundheit/filter"
	"github.com/ushis/gesundheit/result"
)

type Filter struct {
	Hours []Hours
}

func init() {
	filter.Register("office-hours", New)
}

func New(_ db.Database, configure func(interface{}) error) (filter.Filter, error) {
	f := Filter{}

	if err := configure(&f); err != nil {
		return nil, err
	}
	if len(f.Hours) == 0 {
		return nil, errors.New("no office hours configured")
	}
	return f, nil
}

func (f Filter) Filter(e result.Event) (result.Event, bool) {
	now := Now()

	for _, hours := range f.Hours {
		if hours.Contains(now) {
			return e, true
		}
	}
	return e, false
}

type Hours struct {
	From Time
	To   Time
}

func (h Hours) Contains(t Time) bool {
	if h.From <= h.To {
		return h.From <= t && t <= h.To
	}
	return t <= h.To || h.From <= t
}

type Time int

const (
	Minute = 1
	Hour   = 60 * Minute
)

func Now() Time {
	now := time.Now()
	return Time(now.Hour()*Hour + now.Minute()*Minute)
}

func (t *Time) UnmarshalText(text []byte) error {
	i := bytes.IndexByte(text, ':')

	if i < 0 {
		return errors.New("invalid time: " + string(text))
	}
	minute, err := strconv.ParseInt(string(text[i+1:]), 10, 8)

	if err != nil {
		return err
	}
	if minute < 0 || 59 < minute {
		return errors.New("invalid time: " + string(text))
	}
	hour, err := strconv.ParseInt(string(text[:i]), 10, 8)

	if err != nil {
		return err
	}
	if hour < 0 || 24 < hour || (hour == 24 && minute != 0) {
		return errors.New("invalid time: " + string(text))
	}
	if hour == 24 {
		hour = 0
	}
	*t = Time(hour*Hour + minute*Minute)

	return nil
}
