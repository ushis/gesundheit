package memory

import (
	"time"

	"github.com/ushis/gesundheit/db"
	"github.com/ushis/gesundheit/result"
)

func init() {
	db.Register("memory", New)
}

type Database struct {
	db map[string]map[string]result.Event
}

func New(_ func(interface{}) error) (db.Database, error) {
	return Database{make(map[string]map[string]result.Event)}, nil
}

func (db Database) Close() error {
	return nil
}

func (db Database) Handle(e result.Event) error {
	if checks, ok := db.db[e.NodeName]; ok {
		checks[e.CheckId] = e
	} else {
		db.db[e.NodeName] = map[string]result.Event{e.CheckId: e}
	}
	return nil
}

func (db Database) GetEvents() ([]result.Event, error) {
	events := []result.Event{}

	for _, checks := range db.db {
		for _, event := range checks {
			if !isExpired(event) {
				events = append(events, event)
			}
		}
	}
	return events, nil
}

func (db Database) GetEventsByNode(name string) ([]result.Event, error) {
	checks, ok := db.db[name]

	if !ok {
		return []result.Event{}, nil
	}
	events := []result.Event{}

	for _, event := range checks {
		if !isExpired(event) {
			events = append(events, event)
		}
	}
	return events, nil
}

func (db Database) GetLatestEventByNode(name string) (event result.Event, ok bool, err error) {
	events, _ := db.GetEventsByNode(name)

	if len(events) == 0 {
		return event, false, nil
	}
	event = events[0]

	for _, e := range events[1:] {
		if e.Timestamp.After(event.Timestamp) {
			event = e
		}
	}
	return event, true, nil
}

func isExpired(e result.Event) bool {
	ttl := time.Duration(e.CheckInterval) * time.Second * 2
	return time.Since(e.Timestamp) > ttl
}
