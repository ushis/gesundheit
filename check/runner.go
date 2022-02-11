package check

import (
	"math/rand"
	"sync"
	"time"
)

type Runner struct {
	description string
	interval    time.Duration
	history     History
	check       Check
	events      chan<- Event
	stop        chan struct{}
	wg          sync.WaitGroup
}

func NewRunner(description string, interval time.Duration, check Check, events chan<- Event) *Runner {
	return &Runner{
		description: description,
		interval:    interval,
		history:     OK,
		check:       check,
		events:      events,
		stop:        make(chan struct{}),
		wg:          sync.WaitGroup{},
	}
}

func (r *Runner) Run() {
	r.wg.Add(1)
	maxJitter := r.interval / 100
	jitter := time.Duration(rand.Intn(int(2 * maxJitter)))
	interval := r.interval + time.Duration(jitter) - maxJitter

	select {
	case <-time.After(jitter):
	case <-r.stop:
		return
	}
	ticker := time.NewTicker(interval)

	for {
		if msg, err := r.check.Exec(); err != nil {
			r.events <- Event{
				Result:           FAIL,
				Message:          err.Error(),
				CheckDescription: r.description,
				CheckHistory:     r.history,
			}
			r.history = (r.history << 1) | FAIL
		} else {
			r.events <- Event{
				Result:           OK,
				Message:          msg,
				CheckDescription: r.description,
				CheckHistory:     r.history,
			}
			r.history = (r.history << 1) | OK
		}
		select {
		case <-ticker.C:
		case <-r.stop:
			ticker.Stop()
			r.wg.Done()
			return
		}
	}
}

func (r *Runner) Stop() {
	close(r.stop)
	r.wg.Wait()
}
