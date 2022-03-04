package filemtime

import (
	"os"
	"time"

	"github.com/ushis/gesundheit/check"
	"github.com/ushis/gesundheit/result"
)

type Check struct {
	Path   string
	MaxAge time.Duration
}

type Config struct {
	Path   string
	MaxAge string
}

func init() {
	check.Register("file-mtime", New)
}

func New(_ check.Database, configure func(interface{}) error) (check.Check, error) {
	conf := Config{}

	if err := configure(&conf); err != nil {
		return nil, err
	}
	maxAge, err := time.ParseDuration(conf.MaxAge)

	if err != nil {
		return nil, err
	}
	return &Check{Path: conf.Path, MaxAge: maxAge}, nil
}

func (c Check) Exec() result.Result {
	info, err := os.Stat(c.Path)

	if err != nil {
		return result.Fail("failed to stat %s: %s", c.Path, err)
	}
	age := time.Since(info.ModTime()).Truncate(time.Second)

	if age > c.MaxAge {
		return result.Fail("mtime of %s is %s overdue", c.Path, age-c.MaxAge)
	}
	return result.OK("%s has been touched %s ago", c.Path, age)
}
