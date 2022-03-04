package check

import (
	"context"
	"math/rand"
	"sync"
	"time"

	"github.com/ushis/gesundheit/result"
	"github.com/ushis/gesundheit/node"
)

type Runner struct {
	node        node.Info
	id          string
	description string
	interval    time.Duration
	check       Check
}

func NewRunner(node node.Info, id, description string, interval time.Duration, check Check) Runner {
	return Runner{
		node:        node,
		id:          id,
		description: description,
		interval:    interval,
		check:       check,
	}
}

func (r Runner) Run(ctx context.Context, wg *sync.WaitGroup, events chan<- result.Event) error {
	wg.Add(1)

	go func() {
		r.run(ctx, events)
		wg.Done()
	}()

	return nil
}

func (r Runner) run(ctx context.Context, events chan<- result.Event) {
	maxJitter := r.interval / 60
	jitter := time.Duration(rand.Uint64() & uint64(2*maxJitter))
	interval := r.interval + jitter - maxJitter
	delay := time.Duration(rand.Uint64() & uint64(interval))

	select {
	case <-time.After(delay):
	case <-ctx.Done():
		return
	}
	var statusHistory result.StatusHistory

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		res := r.check.Exec()

		event := result.Event{
			NodeName:         r.node.Name,
			CheckId:          r.id,
			CheckDescription: r.description,
			StatusHistory:    statusHistory,
			Status:           res.Status,
			Message:          res.Message,
			Timestamp:        time.Now(),
		}
		statusHistory.Append(res.Status)

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
