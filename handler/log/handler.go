package log

import (
	"log"

	"github.com/ushis/gesundheit/check"
	"github.com/ushis/gesundheit/handler"
)

type Handler struct{}

func init() {
	handler.Register("log", New)
}

func New(configure func(interface{}) error) (handler.Handler, error) {
	return Handler{}, nil
}

func (h Handler) Handle(e check.Event) error {
	log.Printf(
		"%s: %s %s: %s",
		e.NodeName,
		e.CheckDescription,
		e.Result.Status,
		e.Result.Message,
	)
	return nil
}
