package check

import (
	"context"
	"math/rand"
	"sync"
	"time"

	"github.com/ushis/gesundheit/node"
)

type Runner struct {
	node        node.Info
	description string
	interval    time.Duration
	check       Check
}

func NewRunner(node node.Info, description string, interval time.Duration, check Check) Runner {
	return Runner{
		node:        node,
		description: description,
		interval:    interval,
		check:       check,
	}
}

func (r Runner) Run(ctx context.Context, wg *sync.WaitGroup, events chan<- Event) error {
	wg.Add(1)

	go func() {
		r.run(ctx, events)
		wg.Done()
	}()

	return nil
}

func (r Runner) run(ctx context.Context, events chan<- Event) {
	maxJitter := r.interval / 60
	jitter := time.Duration(rand.Uint64() & uint64(2*maxJitter))
	interval := r.interval + jitter - maxJitter
	delay := time.Duration(rand.Uint64() & uint64(interval))

	select {
	case <-time.After(delay):
	case <-ctx.Done():
		return
	}
	history := History(OK)

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		event := r.exec(history)
		history.Append(event.Result)

		select {
		case events <- event:
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

func (r Runner) exec(history History) Event {
	e := Event{
		History:          history,
		CheckDescription: r.description,
		NodeName:         r.node.Name,
	}
	if msg, err := r.check.Exec(); err != nil {
		e.Result = CRITICAL
		e.Message = err.Error()
	} else {
		e.Result = OK
		e.Message = msg
	}
	return e
}
