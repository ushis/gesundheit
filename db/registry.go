package db

import "errors"

type DatabaseFunc func(func(interface{}) error) (Database, error)

type Registry map[string]DatabaseFunc

func (r Registry) Register(name string, fn DatabaseFunc) {
	if _, ok := r[name]; ok {
		panic("Database already registered: " + name)
	}
	r[name] = fn
}

func (r Registry) Get(name string) (DatabaseFunc, error) {
	if fn, ok := r[name]; ok {
		return fn, nil
	}
	return nil, errors.New("unknown Database: " + name)
}

var defaultRegistry = make(Registry)

func Register(name string, fn DatabaseFunc) {
	defaultRegistry.Register(name, fn)
}

func Get(name string) (DatabaseFunc, error) {
	return defaultRegistry.Get(name)
}
