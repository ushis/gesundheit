package memory

import (
	"sync"

	"github.com/ushis/gesundheit/check"
	"github.com/ushis/gesundheit/db"
)

type Database struct {
	events events
	mutex  sync.RWMutex
}

type events map[string]nodeEvents
type nodeEvents map[string]check.Event

func init() {
	db.Register("memory", New)
}

func New(configure func(interface{}) error) (db.Database, error) {
	return &Database{events: make(events)}, nil
}

func (db *Database) Handle(e check.Event) error {
	db.mutex.Lock()
	defer db.mutex.Unlock()

	nodeEvents, ok := db.events[e.NodeName]

	if !ok {
		nodeEvents = make(map[string]check.Event)
		db.events[e.NodeName] = nodeEvents
	}
	nodeEvents[e.CheckId] = e
	return nil
}

func (db *Database) GetEvents() []check.Event {
	db.mutex.RLock()
	defer db.mutex.RUnlock()

	events := []check.Event{}

	for _, nodeEvents := range db.events {
		for _, event := range nodeEvents {
			events = append(events, event)
		}
	}
	return events
}
