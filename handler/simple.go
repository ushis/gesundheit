package handler

import (
	"log"
	"sync"

	"github.com/ushis/gesundheit/result"
)

type Simple interface {
	Handle(result.Event) error
}

func Wrap(s Simple) Handler {
	return wrapper{s}
}

type wrapper struct {
	Simple
}

func (w wrapper) Run(wg *sync.WaitGroup) (chan<- result.Event, error) {
	chn := make(chan result.Event)
	wg.Add(1)

	go func() {
		w.run(chn)
		wg.Done()
	}()

	return chn, nil
}

func (w wrapper) run(chn <-chan result.Event) {
	for e := range chn {
		if err := w.Handle(e); err != nil {
			log.Println(e)
		}
	}
}
