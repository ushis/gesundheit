package filesystem

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/ushis/gesundheit/db"
	"github.com/ushis/gesundheit/db/memory"
	"github.com/ushis/gesundheit/result"
)

func init() {
	db.Register("filesystem", New)
}

const walFilename = "gesundheit.wal"
const tmpWalFilename = "_gesundheit.wal"

type Database struct {
	*sync.Mutex
	db               db.Database
	log              *os.File
	directory        string
	vacuumInterval   time.Duration
	cancelAutoVacuum context.CancelFunc
}

type Config struct {
	Directory      string
	VacuumInterval string
}

func New(configure func(interface{}) error) (db.Database, error) {
	conf := Config{Directory: ".", VacuumInterval: "24h"}

	if err := configure(&conf); err != nil {
		return nil, err
	}
	vacuumInterval, err := time.ParseDuration(conf.VacuumInterval)

	if err != nil {
		return nil, err
	}
	memoryDB, err := memory.New(nil)

	if err != nil {
		return nil, err
	}
	log, err := os.OpenFile(
		filepath.Join(conf.Directory, walFilename),
		os.O_APPEND|os.O_CREATE|os.O_RDWR,
		0755,
	)
	if err != nil {
		return nil, err
	}
	if err := readEvents(log, memoryDB.InsertEvent); err != nil {
		return nil, err
	}
	ctx, cancel := context.WithCancel(context.Background())
	db := &Database{&sync.Mutex{}, memoryDB, log, conf.Directory, vacuumInterval, cancel}
	go db.autoVacuum(ctx)
	return db, nil
}

func (db *Database) Close() error {
	db.Lock()
	defer db.Unlock()

	db.cancelAutoVacuum()
	return db.log.Close()
}

func (db *Database) InsertEvent(e result.Event) (bool, error) {
	db.Lock()
	defer db.Unlock()

	if err := json.NewEncoder(db.log).Encode(e); err != nil {
		return false, err
	}
	return db.db.InsertEvent(e)
}

func (db *Database) GetEvents() ([]result.Event, error) {
	return db.db.GetEvents()
}

func (db *Database) GetEventsByNode(name string) ([]result.Event, error) {
	return db.db.GetEventsByNode(name)
}

func (db *Database) autoVacuum(ctx context.Context) {
	for {
		select {
		case <-time.After(db.vacuumInterval):
		case <-ctx.Done():
			return
		}
		log.Println("db: vacuum: start")

		if err := db.vacuum(); err != nil {
			log.Println("db: vacuum: failed:", err)
		} else {
			log.Println("db: vacuum: done")
		}
	}
}

func (db *Database) vacuum() error {
	db.Lock()
	defer db.Unlock()

	newLog, err := os.OpenFile(
		filepath.Join(db.directory, tmpWalFilename),
		os.O_CREATE|os.O_TRUNC|os.O_WRONLY,
		0755,
	)
	if err != nil {
		return err
	}
	events, err := db.db.GetEvents()

	if err != nil {
		return err
	}
	w := json.NewEncoder(newLog)

	for _, event := range events {
		if err := w.Encode(event); err != nil {
			return err
		}
	}
	db.log.Close()
	db.log = newLog

	return os.Rename(
		filepath.Join(db.directory, tmpWalFilename),
		filepath.Join(db.directory, walFilename),
	)
}

func readEvents(r io.Reader, insertEvent func(result.Event) (bool, error)) error {
	decoder := json.NewDecoder(r)

	for {
		event := result.Event{}

		if err := decoder.Decode(&event); err == io.EOF {
			return nil
		} else if err != nil {
			return err
		}
		if _, err := insertEvent(event); err != nil {
			return err
		}
	}
}
