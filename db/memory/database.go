package memory

import (
	"sync"

	"github.com/ushis/gesundheit/result"
	"github.com/ushis/gesundheit/db"
)

type Database struct {
	events events
	mutex  sync.RWMutex
}

type events map[string]nodeEvents
type nodeEvents map[string]result.Event

func init() {
	db.Register("memory", New)
}

func New(configure func(interface{}) error) (db.Database, error) {
	return &Database{events: make(events)}, nil
}

func (db *Database) Handle(e result.Event) error {
	db.mutex.Lock()
	defer db.mutex.Unlock()

	nodeEvents, ok := db.events[e.NodeName]

	if !ok {
		nodeEvents = make(map[string]result.Event)
		db.events[e.NodeName] = nodeEvents
	}
	nodeEvents[e.CheckId] = e
	return nil
}

func (db *Database) GetEvents() []result.Event {
	db.mutex.RLock()
	defer db.mutex.RUnlock()

	events := []result.Event{}

	for _, nodeEvents := range db.events {
		for _, event := range nodeEvents {
			events = append(events, event)
		}
	}
	return events
}

func (db *Database) GetEventsByNode(name string) []result.Event {
	db.mutex.RLock()
	defer db.mutex.RUnlock()

	nodeEvents, ok := db.events[name]

	if !ok {
		return []result.Event{}
	}
	events := []result.Event{}

	for _, event := range nodeEvents {
		events = append(events, event)
	}
	return events
}

func (db *Database) GetLatestEventByNode(name string) (event result.Event, ok bool) {
	db.mutex.RLock()
	defer db.mutex.RUnlock()

	events := db.GetEventsByNode(name)

	if len(events) == 0 {
		return event, false
	}
	event = events[0]

	for _, e := range events[1:] {
		if e.Timestamp.After(event.Timestamp) {
			event = e
		}
	}
	return event, true
}
