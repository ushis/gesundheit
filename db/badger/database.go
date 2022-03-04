package badger

import (
	"bytes"
	"encoding/json"
	"sync"
	"time"

	badger "github.com/dgraph-io/badger/v3"
	"github.com/ushis/gesundheit/db"
	"github.com/ushis/gesundheit/result"
)

type Database struct {
	badger *badger.DB
	wg     *sync.WaitGroup
}

type Opts struct {
	Persistent bool
	Path       string
}

func New(opts Opts) (db.Database, error) {
	badgerOpts := badger.DefaultOptions(opts.Path)
	badgerOpts = badgerOpts.WithInMemory(!opts.Persistent)
	badgerOpts = badgerOpts.WithLoggingLevel(badger.WARNING)

	badgerDB, err := badger.Open(badgerOpts)

	if err != nil {
		return nil, err
	}
	wg := &sync.WaitGroup{}
	wg.Add(1)

	go func() {
		badgerDB.RunValueLogGC(0.5)
		wg.Done()
	}()

	return Database{badgerDB, wg}, nil
}

func (db Database) Close() error {
	db.wg.Wait()
	return db.badger.Close()
}

func (db Database) Handle(e result.Event) error {
	key := buildKey("event", e.NodeName, e.CheckId)
	ttl := time.Duration(e.CheckInterval) * time.Second * 2
	return db.update(key, e, ttl)
}

func (db Database) update(key []byte, value interface{}, ttl time.Duration) error {
	val, err := json.Marshal(value)

	if err != nil {
		return err
	}
	return db.badger.Update(func(txn *badger.Txn) error {
		return txn.SetEntry(badger.NewEntry(key, val).WithTTL(ttl))
	})
}

func (db Database) getEvents(path ...string) ([]result.Event, error) {
	path = append([]string{"event"}, path...)
	prefix := buildKeyPrefix(path...)
	events := []result.Event{}

	txn := db.badger.NewTransaction(false)
	defer txn.Discard()

	it := txn.NewIterator(badger.DefaultIteratorOptions)
	defer it.Close()

	for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
		err := it.Item().Value(func(val []byte) error {
			event := result.Event{}
			err := json.Unmarshal(val, &event)
			events = append(events, event)
			return err
		})
		if err != nil {
			return nil, err
		}
	}
	return events, nil
}

func (db Database) GetEvents() ([]result.Event, error) {
	return db.getEvents()
}

func (db Database) GetEventsByNode(name string) ([]result.Event, error) {
	return db.getEvents(name)
}

func (db Database) GetLatestEventByNode(name string) (e result.Event, ok bool, err error) {
	events, err := db.GetEventsByNode(name)

	if err != nil || len(events) == 0 {
		return e, false, err
	}
	e = events[0]

	for _, ev := range events[1:] {
		if ev.Timestamp.After(e.Timestamp) {
			e = ev
		}
	}
	return e, true, nil
}

const pathSep = ":"

func buildKey(path ...string) []byte {
	if len(path) == 0 {
		return []byte{}
	}
	prefix := buildKeyPrefix(path...)
	return prefix[:len(prefix)-len(pathSep)]
}

func buildKeyPrefix(path ...string) []byte {
	n := len(path) * len(pathSep)

	for _, p := range path {
		n += len(p)
	}
	b := bytes.Buffer{}
	b.Grow(n)

	for _, p := range path {
		b.WriteString(p)
		b.WriteString(pathSep)
	}
	return b.Bytes()
}
