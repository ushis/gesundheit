package db

import (
	"github.com/ushis/gesundheit/check"
	"github.com/ushis/gesundheit/handler"
)

type Database interface {
	handler.Handler
	GetEvents() []check.Event
}
