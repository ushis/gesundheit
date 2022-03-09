package handler

import (
	"sync"

	"github.com/ushis/gesundheit/result"
)

type Handler interface {
	Run(*sync.WaitGroup) (chan<- result.Event, error)
}
