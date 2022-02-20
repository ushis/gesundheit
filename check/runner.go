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
	stop        chan struct{}
	wg          sync.WaitGroup
}

func NewRunner(description string, interval time.Duration, check Check) *Runner {
	return &Runner{
		description: description,
		interval:    interval,
		history:     OK,
		check:       check,
		stop:        make(chan struct{}),
		wg:          sync.WaitGroup{},
	}
}

func (r *Runner) Run(events chan<- Event) {
	r.wg.Add(1)
	maxJitter := r.interval / 60
	jitter := time.Duration(rand.Uint64() & uint64(2*maxJitter))
	interval := r.interval + jitter - maxJitter
	delay := time.Duration(rand.Uint64() & uint64(interval))

	select {
	case <-time.After(delay):
	case <-r.stop:
		r.wg.Done()
		return
	}
	ticker := time.NewTicker(interval)

	for {
		select {
		case events <- r.exec():
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
			Result:           CRITICAL,
			Message:          err.Error(),
			CheckDescription: r.description,
			CheckHistory:     r.history,
		}
		r.history = (r.history << 1) | CRITICAL
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
