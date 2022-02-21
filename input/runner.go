package input

import (
	"context"
	"sync"

	"github.com/ushis/gesundheit/check"
)

type Runner struct {
	Input
}

func NewRunner(i Input) *Runner {
	return &Runner{i}
}

func (r *Runner) Run(ctx context.Context, wg *sync.WaitGroup, events chan<- check.Event) error {
	return r.Input.Run(ctx, wg, events)
}
