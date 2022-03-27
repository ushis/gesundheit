package memory

import (
	"sync"
	"time"

	"github.com/ushis/gesundheit/db"
	"github.com/ushis/gesundheit/result"
)

func init() {
	db.Register("memory", New)
}

type Database struct {
	*sync.RWMutex
	db map[string]map[string]result.Event
}

func New(_ func(interface{}) error) (db.Database, error) {
	return Database{&sync.RWMutex{}, make(map[string]map[string]result.Event)}, nil
}

func (db Database) Close() error {
	return nil
}

func (db Database) InsertEvent(e result.Event) (bool, error) {
	db.Lock()
	defer db.Unlock()

	checks, ok := db.db[e.NodeName]

	if !ok {
		db.db[e.NodeName] = map[string]result.Event{e.CheckId: e}
		return true, nil
	}
	prevE, ok := checks[e.CheckId]

	if !ok || prevE.Id != e.Id {
		checks[e.CheckId] = e
		return true, nil
	}
	return false, nil
}

func (db Database) GetEvents() ([]result.Event, error) {
	db.RLock()
	defer db.RUnlock()

	now := time.Now()
	events := []result.Event{}

	for _, checks := range db.db {
		for _, event := range checks {
			if event.ExpiresAt.After(now) {
				events = append(events, event)
			}
		}
	}
	return events, nil
}

func (db Database) GetEventsByNode(name string) ([]result.Event, error) {
	db.RLock()
	defer db.RUnlock()

	checks, ok := db.db[name]

	if !ok {
		return []result.Event{}, nil
	}
	now := time.Now()
	events := []result.Event{}

	for _, event := range checks {
		if event.ExpiresAt.After(now) {
			events = append(events, event)
		}
	}
	return events, nil
}
