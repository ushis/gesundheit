package db

import (
	"github.com/ushis/gesundheit/result"
)

type Database interface {
	Close() error
	InsertEvent(e result.Event) (bool, error)
	GetEvents() ([]result.Event, error)
	GetEventsByNode(name string) ([]result.Event, error)
	GetLatestEventByNode(name string) (event result.Event, ok bool, err error)
}
