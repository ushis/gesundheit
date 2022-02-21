package main

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/ushis/gesundheit/check"
	"github.com/ushis/gesundheit/filter"
	"github.com/ushis/gesundheit/handler"
	"github.com/ushis/gesundheit/node"
)

type config struct {
	Node    node.Info
	Log     logConfig
	Modules modulesConfig
}

type logConfig struct {
	Path       string
	Timestamps bool
}

type modulesConfig struct {
	Config string
}

type modConfig struct {
	Check   *checkConfig
	Handler *handlerConfig
}

type checkConfig struct {
	Module      string
	Description string
	Interval    string
	Config      toml.Primitive
}

type handlerConfig struct {
	Module string
	Filter []*filterConfig
	Config toml.Primitive
}

type filterConfig struct {
	Module string
	Config toml.Primitive
}

func loadConf(path string) (config, error) {
	conf := config{
		Log:     logConfig{Path: "-", Timestamps: false},
		Modules: modulesConfig{Config: "modules.d"},
	}
	meta, err := toml.DecodeFile(path, &conf)

	if err != nil {
		return conf, err
	}
	if len(meta.Undecoded()) > 0 {
		return conf, fmt.Errorf("failed to load config: %s: unknown field %s", path, meta.Undecoded()[0])
	}
	if conf.Node.Name == "" {
		hostname, err := os.Hostname()

		if err != nil {
			return conf, fmt.Errorf("failed to determine hostname: %s", err)
		}
		conf.Node.Name = hostname
	}
	return conf, nil
}

type modConfLoader struct {
	node node.Info
	hub  *hub
}

func newModConfLoader(node node.Info, hub *hub) modConfLoader {
	return modConfLoader{node, hub}
}

func (l modConfLoader) loadAll(glob string) error {
	paths, err := filepath.Glob(glob)

	if err != nil {
		return err
	}
	for _, path := range paths {
		if err := l.load(path); err != nil {
			return err
		}
	}
	return nil
}

func (l modConfLoader) load(path string) error {
	mod := modConfig{}
	meta, err := toml.DecodeFile(path, &mod)

	if err != nil {
		return err
	}
	if mod.Check != nil {
		return l.loadCheck(mod.Check, path, meta)
	}
	if mod.Handler != nil {
		return l.loadHandler(mod.Handler, path, meta)
	}
	return fmt.Errorf("failed to load module config: %s: missing module configuration", path)
}

func (l modConfLoader) loadCheck(conf *checkConfig, path string, meta toml.MetaData) error {
	fn, err := check.Get(conf.Module)

	if err != nil {
		return fmt.Errorf("failed to load check config: %s: %s", path, err.Error())
	}
	chk, err := fn(func(cfg interface{}) error {
		return meta.PrimitiveDecode(conf.Config, cfg)
	})
	if err != nil {
		return fmt.Errorf("failed to load check config: %s: %s", path, err.Error())
	}
	if len(meta.Undecoded()) > 0 {
		return fmt.Errorf("failed to load check config: %s: unknown field %s", path, meta.Undecoded()[0])
	}
	if conf.Description == "" {
		return fmt.Errorf("failed to load check config: %s: missing Description", path)
	}
	interval, err := time.ParseDuration(conf.Interval)

	if err != nil {
		return fmt.Errorf("failed to load check config: %s: %s", path, err.Error())
	}
	l.hub.registerCheckRunner(check.NewRunner(l.node, conf.Description, interval, chk))

	return nil
}

func (l modConfLoader) loadHandler(conf *handlerConfig, path string, meta toml.MetaData) error {
	fn, err := handler.Get(conf.Module)

	if err != nil {
		return fmt.Errorf("failed to load handler config: %s: %s", path, err.Error())
	}
	hdl, err := fn(func(cfg interface{}) error {
		return meta.PrimitiveDecode(conf.Config, cfg)
	})
	if err != nil {
		return fmt.Errorf("failed to load handler config: %s: %s", path, err.Error())
	}
	filters := []filter.Filter{}

	for _, cfg := range conf.Filter {
		f, err := l.loadFilter(cfg, path, meta)

		if err != nil {
			return fmt.Errorf("failed to load handler config: %s: %s", path, err.Error())
		}
		filters = append(filters, f)
	}
	l.hub.registerHandlerRunner(handler.NewRunner(hdl, filters))

	if len(meta.Undecoded()) > 0 {
		return fmt.Errorf("failed to load handler config: %s: unknown field %s", path, meta.Undecoded()[0])
	}
	return nil
}

func (l modConfLoader) loadFilter(conf *filterConfig, path string, meta toml.MetaData) (filter.Filter, error) {
	fn, err := filter.Get(conf.Module)

	if err != nil {
		return nil, err
	}
	return fn(func(cfg interface{}) error {
		return meta.PrimitiveDecode(conf.Config, cfg)
	})
}
