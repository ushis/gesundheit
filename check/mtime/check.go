package mtime

import (
	"log"
	"time"

	"github.com/ushis/gesundheit/check"
	filemtime "github.com/ushis/gesundheit/check/file-mtime"
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
	log.Println(
		"module \"mtime\" is deprecated and will be removed in the near future:",
		"use \"file-mtime\" instead.",
	)
	return filemtime.New(configure)
}
