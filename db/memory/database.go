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
	db map[string]map[string]*cappedList[result.Event]
}

func New(_ func(interface{}) error) (db.Database, error) {
	return Database{&sync.RWMutex{}, make(map[string]map[string]*cappedList[result.Event])}, nil
}

func (db Database) Close() error {
	return nil
}

func (db Database) InsertEvent(e result.Event) (bool, error) {
	db.Lock()
	defer db.Unlock()

	checks, ok := db.db[e.NodeName]

	if !ok {
		checks = map[string]*cappedList[result.Event]{}
		db.db[e.NodeName] = checks
	}
	events, ok := checks[e.CheckId]

	if !ok {
		events = &cappedList[result.Event]{}
		checks[e.CheckId] = events
	}
	for _, prevEvent := range events.slice() {
		if prevEvent.Id == e.Id {
			return false, nil
		}
	}
	events.push(e)
	return true, nil
}

func (db Database) GetEvents() ([]result.Event, error) {
	db.RLock()
	defer db.RUnlock()

	now := time.Now()
	result := []result.Event{}

	for _, checks := range db.db {
		for _, events := range checks {
			for _, event := range events.slice() {
				if event.ExpiresAt.After(now) {
					result = append(result, event)
				}
			}
		}
	}
	return result, nil
}

func (db Database) GetEventsByNode(nodeName string) ([]result.Event, error) {
	db.RLock()
	defer db.RUnlock()

	checks, ok := db.db[nodeName]

	if !ok {
		return []result.Event{}, nil
	}
	now := time.Now()
	result := []result.Event{}

	for _, events := range checks {
		for _, event := range events.slice() {
			if event.ExpiresAt.After(now) {
				result = append(result, event)
			}
		}
	}
	return result, nil
}

func (db Database) GetEventsByCheck(nodeName, checkId string) ([]result.Event, error) {
	db.RLock()
	defer db.RUnlock()

	checks, ok := db.db[nodeName]

	if !ok {
		return []result.Event{}, nil
	}
	events, ok := checks[checkId]

	if !ok {
		return []result.Event{}, nil
	}
	now := time.Now()
	result := []result.Event{}

	for _, event := range events.slice() {
		if event.ExpiresAt.After(now) {
			result = append(result, event)
		}
	}
	return result, nil
}

const cappedListBufLen = 64
const cappedListCap = 6

type cappedList[T any] struct {
	buffer [cappedListBufLen]T
	offset int
	length int
}

func (l *cappedList[T]) push(v T) {
	index := l.offset + l.length

	if index >= len(l.buffer) {
		copy(l.buffer[:], l.buffer[l.offset+1:])
		l.offset = 0
		l.length -= 1
		index = l.length
	}
	l.buffer[index] = v

	if l.length < cappedListCap {
		l.length += 1
	} else {
		l.offset += 1
	}
}

func (l *cappedList[T]) slice() []T {
	return l.buffer[l.offset : l.offset+l.length]
}
