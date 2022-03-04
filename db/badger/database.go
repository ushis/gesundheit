package badger

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	badger "github.com/dgraph-io/badger/v3"
	"github.com/ushis/gesundheit/db"
	"github.com/ushis/gesundheit/result"
)

type Database struct {
	badger *badger.DB
	wg     *sync.WaitGroup
	close  func()
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
	ctx, close := context.WithCancel(context.Background())
	wg := &sync.WaitGroup{}
	wg.Add(1)

	go func() {
		defer wg.Done()

		ticker := time.NewTicker(time.Hour)
		defer ticker.Stop()

		for {
			badgerDB.RunValueLogGC(0.5)

			select {
			case <-ticker.C:
			case <-ctx.Done():
				return
			}
		}
	}()

	return Database{badgerDB, wg, close}, nil
}

func (db Database) Close() error {
	db.close()
	db.wg.Wait()
	return db.badger.Close()
}

func (db Database) Handle(e result.Event) error {
	val, err := encodeEvent(e)

	if err != nil {
		return err
	}
	key := []byte(fmt.Sprintf("event:%s:%s", e.NodeName, e.CheckId))
	ttl := time.Duration(e.CheckInterval) * time.Second * 2

	return db.badger.Update(func(txn *badger.Txn) error {
		return txn.SetEntry(badger.NewEntry(key, val).WithTTL(ttl))
	})
}

func (db Database) getEvents(p string) ([]result.Event, error) {
	prefix := []byte("event:" + p)
	events := []result.Event{}

	txn := db.badger.NewTransaction(false)
	defer txn.Discard()

	it := txn.NewIterator(badger.DefaultIteratorOptions)
	defer it.Close()

	for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
		err := it.Item().Value(func(val []byte) error {
			event, err := decodeEvent(val)
			events = append(events, event)
			return err
		})
		if err != nil {
			return nil, err
		}
	}
	return events, nil
}

func decodeEvent(buf []byte) (result.Event, error) {
	e := result.Event{}
	err := json.Unmarshal(buf, &e)
	return e, err
}

func encodeEvent(e result.Event) ([]byte, error) {
	return json.Marshal(e)
}

func (db Database) GetEvents() ([]result.Event, error) {
	return db.getEvents("")
}

func (db Database) GetEventsByNode(name string) ([]result.Event, error) {
	return db.getEvents(name + ":")
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
