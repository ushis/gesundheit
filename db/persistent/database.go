package persistent

import (
	"github.com/ushis/gesundheit/db"
	"github.com/ushis/gesundheit/db/badger"
)

type Config struct {
	Path string
}

func init() {
	db.Register("persistent", New)
}

func New(configure func(interface{}) error) (db.Database, error) {
	conf := Config{}

	if err := configure(&conf); err != nil {
		return nil, err
	}
	return badger.New(badger.Opts{Persistent: true, Path: conf.Path})
}
