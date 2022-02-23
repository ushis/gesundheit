package input

import (
	"context"
	"sync"

	"github.com/ushis/gesundheit/check"
)

type Input interface {
	Run(ctx context.Context, wg *sync.WaitGroup, events chan<- check.Event) error
}
