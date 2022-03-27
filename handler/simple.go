package handler

import (
	"log"
	"sync"

	"github.com/ushis/gesundheit/result"
)

type Simple interface {
	Handle(result.Event) error
}

type simpleWrapper struct {
	Simple
}

func (w simpleWrapper) Run(wg *sync.WaitGroup, chn <-chan result.Event) error {
	wg.Add(1)

	go func() {
		w.run(chn)
		wg.Done()
	}()

	return nil
}

func (w simpleWrapper) run(chn <-chan result.Event) {
	for e := range chn {
		if err := w.Handle(e); err != nil {
			log.Println(err)
		}
	}
}
