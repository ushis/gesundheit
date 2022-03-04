package handler

import (
	"github.com/ushis/gesundheit/check"
)

type Handler interface {
	Handle(check.Event) error
}
