package log

import (
	"io"
	"log"
	"os"

	"github.com/ushis/gesundheit/check"
	"github.com/ushis/gesundheit/handler"
)

type Handler struct {
	w io.WriteCloser
	l *log.Logger
}

type Config struct {
	Path      string
	Prefix    string
	Timestamp bool
}

func init() {
	handler.Register("log", New)
}

func New(configure func(interface{}) error) (handler.Handler, error) {
	conf := Config{Path: "-", Prefix: ""}

	if err := configure(&conf); err != nil {
		return nil, err
	}
	w := os.Stdout

	if conf.Path != "-" {
		f, err := os.OpenFile(conf.Path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0640)

		if err != nil {
			return nil, err
		}
		w = f
	}
	flags := 0

	if conf.Timestamp {
		flags = log.Ldate | log.Ltime
	}
	return Handler{w: w, l: log.New(w, conf.Prefix, flags)}, nil
}

func (h Handler) Handle(e check.Event) error {
	var result string

	if e.Result == check.OK {
		result = "succeeded"
	} else {
		result = "failed"
	}
	h.l.Printf("%s %s: %s", e.CheckDescription, result, e.Message)
	return nil
}

func (h Handler) Close() {
	if h.w != os.Stdout {
		h.w.Close()
	}
}
