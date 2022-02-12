package mtime

import (
	"fmt"
	"os"
	"time"

	"github.com/ushis/gesundheit/check"
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
	check.Register("mtime", New)
}

func New(configure func(interface{}) error) (check.Check, error) {
	cfg := Config{}

	if err := configure(&cfg); err != nil {
		return nil, err
	}
	maxAge, err := time.ParseDuration(cfg.MaxAge)

	if err != nil {
		return nil, err
	}
	return &Check{Path: cfg.Path, MaxAge: maxAge}, nil
}

func (c Check) Exec() (string, error) {
	info, err := os.Stat(c.Path)

	if err != nil {
		return "", fmt.Errorf("failed to stat %s: %s", c.Path, err)
	}
	age := time.Since(info.ModTime()).Truncate(time.Second)

	if age > c.MaxAge {
		return "", fmt.Errorf("mtime of %s is %s overdue", c.Path, age-c.MaxAge)
	}
	return fmt.Sprintf("%s has been touched %s ago", c.Path, age), nil
}
