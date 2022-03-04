package db

import (
	"github.com/ushis/gesundheit/result"
)

type Database interface {
	Handle(e result.Event) error
	GetEvents() []result.Event
	GetEventsByNode(name string) []result.Event
	GetLatestEventByNode(name string) (event result.Event, ok bool)
}
