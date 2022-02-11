package filter

import (
	"errors"
)

type FilterFunc func(func(interface{}) error) (Filter, error)

type Registry map[string]FilterFunc

func (r Registry) Register(name string, fn FilterFunc) {
	if _, ok := r[name]; ok {
		panic("filter already registered: " + name)
	}
	r[name] = fn
}

func (r Registry) Get(name string) (FilterFunc, error) {
	if fn, ok := r[name]; ok {
		return fn, nil
	}
	return nil, errors.New("unknown handler: " + name)
}

var defaultRegistry = make(Registry)

func Register(name string, fn FilterFunc) {
	defaultRegistry.Register(name, fn)
}

func Get(name string) (FilterFunc, error) {
	return defaultRegistry.Get(name)
}
