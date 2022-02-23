package input

import "errors"

type InputFunc func(func(interface{}) error) (Input, error)

type Registry map[string]InputFunc

func (r Registry) Register(name string, fn InputFunc) {
	if _, ok := r[name]; ok {
		panic("Input already registered: " + name)
	}
	r[name] = fn
}

func (r Registry) Get(name string) (InputFunc, error) {
	if fn, ok := r[name]; ok {
		return fn, nil
	}
	return nil, errors.New("unknown Input: " + name)
}

var defaultRegistry = make(Registry)

func Register(name string, fn InputFunc) {
	defaultRegistry.Register(name, fn)
}

func Get(name string) (InputFunc, error) {
	return defaultRegistry.Get(name)
}
