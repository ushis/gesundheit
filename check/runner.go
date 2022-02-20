package check

import (
	"context"
	"math/rand"
	"sync"
	"time"
)

type Runner struct {
	description string
	interval    time.Duration
	history     History
	check       Check
}

func NewRunner(description string, interval time.Duration, check Check) *Runner {
	return &Runner{
		description: description,
		interval:    interval,
		history:     OK,
		check:       check,
	}
}

func (r *Runner) Run(ctx context.Context, wg *sync.WaitGroup, events chan<- Event) {
	defer wg.Done()

	maxJitter := r.interval / 60
	jitter := time.Duration(rand.Uint64() & uint64(2*maxJitter))
	interval := r.interval + jitter - maxJitter
	delay := time.Duration(rand.Uint64() & uint64(interval))

	select {
	case <-time.After(delay):
	case <-ctx.Done():
		return
	}
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case events <- r.exec():
		case <-ctx.Done():
			return
		}
		select {
		case <-ticker.C:
		case <-ctx.Done():
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
