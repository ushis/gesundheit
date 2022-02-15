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
	maxJitter := r.interval / 60
	jitter := time.Duration(rand.Uint64() & uint64(2*maxJitter))
	interval := r.interval + jitter - maxJitter

	select {
	case <-time.After(jitter):
	case <-r.stop:
		r.wg.Done()
		return
	}
	ticker := time.NewTicker(interval)

	for {
		select {
		case r.events <- r.exec():
		case <-r.stop:
			ticker.Stop()
			r.wg.Done()
			return
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

func (r *Runner) exec() (event Event) {
	if msg, err := r.check.Exec(); err != nil {
		event = Event{
			Result:           FAIL,
			Message:          err.Error(),
			CheckDescription: r.description,
			CheckHistory:     r.history,
		}
		r.history = (r.history << 1) | FAIL
	} else {
		event = Event{
			Result:           OK,
			Message:          msg,
			CheckDescription: r.description,
			CheckHistory:     r.history,
		}
		r.history = (r.history << 1) | OK
	}
	return event
}

func (r *Runner) Stop() {
	close(r.stop)
	r.wg.Wait()
}
