package memory

import (
	"github.com/ushis/gesundheit/db"
	"github.com/ushis/gesundheit/db/badger"
)

func init() {
	db.Register("memory", New)
}

func New(_ func(interface{}) error) (db.Database, error) {
	return badger.New(badger.Opts{Persistent: false, Path: ""})
}
