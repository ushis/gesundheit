package input

import (
	"context"
	"sync"

	"github.com/ushis/gesundheit/result"
)

type Input interface {
	Run(ctx context.Context, wg *sync.WaitGroup, events chan<- result.Event) error
}
