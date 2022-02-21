package main

import (
	"fmt"
	"path/filepath"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/ushis/gesundheit/check"
	"github.com/ushis/gesundheit/filter"
	"github.com/ushis/gesundheit/handler"
)

type config struct {
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

type moduleConfig struct {
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
	return conf, nil
}

func loadModuleConfs(hub *hub, glob string) error {
	paths, err := filepath.Glob(glob)

	if err != nil {
		return err
	}
	for _, path := range paths {
		if err := loadModuleConf(hub, path); err != nil {
			return err
		}
	}
	return nil
}

func loadModuleConf(hub *hub, path string) error {
	mod := moduleConfig{}
	meta, err := toml.DecodeFile(path, &mod)

	if err != nil {
		return err
	}
	if mod.Check != nil {
		return loadCheckModule(hub, mod.Check, path, meta)
	}
	if mod.Handler != nil {
		return loadHandlerModule(hub, mod.Handler, path, meta)
	}
	return fmt.Errorf("failed to load module config: %s: missing module configuration", path)
}

func loadCheckModule(hub *hub, conf *checkConfig, path string, meta toml.MetaData) error {
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
	hub.registerCheckRunner(check.NewRunner(conf.Description, interval, chk))

	return nil
}

func loadHandlerModule(hub *hub, conf *handlerConfig, path string, meta toml.MetaData) error {
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
		f, err := loadFilterModule(cfg, path, meta)

		if err != nil {
			return fmt.Errorf("failed to load handler config: %s: %s", path, err.Error())
		}
		filters = append(filters, f)
	}
	hub.registerHandlerRunner(handler.NewRunner(hdl, filters))

	if len(meta.Undecoded()) > 0 {
		return fmt.Errorf("failed to load handler config: %s: unknown field %s", path, meta.Undecoded()[0])
	}
	return nil
}

func loadFilterModule(conf *filterConfig, path string, meta toml.MetaData) (filter.Filter, error) {
	fn, err := filter.Get(conf.Module)

	if err != nil {
		return nil, err
	}
	return fn(func(cfg interface{}) error {
		return meta.PrimitiveDecode(conf.Config, cfg)
	})
}
