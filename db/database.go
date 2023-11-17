package db

import (
	"github.com/ushis/gesundheit/result"
)

type Database interface {
	Close() error
	InsertEvent(e result.Event) (bool, error)
	GetEvents() ([]result.Event, error)
	GetEventsByNode(nodeName string) ([]result.Event, error)
	GetEventsByCheck(nodeName, checkId string) ([]result.Event, error)
}
