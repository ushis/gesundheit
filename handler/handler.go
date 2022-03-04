package handler

import "github.com/ushis/gesundheit/result"

type Handler interface {
	Handle(result.Event) error
}
