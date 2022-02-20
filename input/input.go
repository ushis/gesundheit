package input

import "github.com/ushis/gesundheit/check"

type Input interface {
	Run(chan<- check.Event)
	Close()
}
