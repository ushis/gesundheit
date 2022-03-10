package handler

import "errors"

type HandlerFunc func(func(interface{}) error) (Handler, error)

type SimpleFunc func(func(interface{}) error) (Simple, error)

type Registry map[string]HandlerFunc

func (r Registry) Register(name string, fn HandlerFunc) {
	if _, ok := r[name]; ok {
		panic("handler already registered: " + name)
	}
	r[name] = fn
}

func (r Registry) RegisterSimple(name string, fn SimpleFunc) {
	r.Register(name, wrapSimpleFunc(fn))
}

func (r Registry) Get(name string) (HandlerFunc, error) {
	if fn, ok := r[name]; ok {
		return fn, nil
	}
	return nil, errors.New("unknown handler: " + name)
}

func wrapSimpleFunc(fn SimpleFunc) HandlerFunc {
	return func(configure func(interface{}) error) (Handler, error) {
		simple, err := fn(configure)
		return simpleWrapper{simple}, err
	}
}

var defaultRegistry = make(Registry)

func Register(name string, fn HandlerFunc) {
	defaultRegistry.Register(name, fn)
}

func RegisterSimple(name string, fn SimpleFunc) {
	defaultRegistry.RegisterSimple(name, fn)
}

func Get(name string) (HandlerFunc, error) {
	return defaultRegistry.Get(name)
}
