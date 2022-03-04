package check

import (
	"errors"

	"github.com/ushis/gesundheit/result"
)

type Database interface {
	GetEvents() ([]result.Event, error)
	GetEventsByNode(name string) ([]result.Event, error)
	GetLatestEventByNode(name string) (event result.Event, ok bool, err error)
}

type CheckFunc func(Database, func(interface{}) error) (Check, error)

type Registry map[string]CheckFunc

func (r Registry) Register(name string, fn CheckFunc) {
	if _, ok := r[name]; ok {
		panic("check already registered: " + name)
	}
	r[name] = fn
}

func (r Registry) Get(name string) (CheckFunc, error) {
	if fn, ok := r[name]; ok {
		return fn, nil
	}
	return nil, errors.New("unknown check: " + name)
}

var defaultRegistry = make(Registry)

func Register(name string, fn CheckFunc) {
	defaultRegistry.Register(name, fn)
}

func Get(name string) (CheckFunc, error) {
	return defaultRegistry.Get(name)
}
