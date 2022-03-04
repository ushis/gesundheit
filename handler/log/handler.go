package log

import (
	"log"

	"github.com/ushis/gesundheit/handler"
	"github.com/ushis/gesundheit/result"
)

type Handler struct{}

func init() {
	handler.Register("log", New)
}

func New(_ func(interface{}) error) (handler.Handler, error) {
	return Handler{}, nil
}

func (h Handler) Handle(e result.Event) error {
	log.Printf(
		"%s: %s %s: %s",
		e.NodeName,
		e.CheckDescription,
		e.Status,
		e.Message,
	)
	return nil
}
