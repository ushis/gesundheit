package mtime

import (
	"errors"

	"github.com/ushis/gesundheit/check"
	"github.com/ushis/gesundheit/check/http"
	"github.com/ushis/gesundheit/result"
)

type Check struct {
	HttpConf http.Config
	Status   int
}

type Config struct {
	http.Config
	Status int
}

func init() {
	check.Register("http-status", New)
}

func New(_ check.Database, configure func(interface{}) error) (check.Check, error) {
	conf := Config{}

	if err := configure(&conf); err != nil {
		return nil, err
	}
	if conf.Url == "" {
		return nil, errors.New("missing Url")
	}
	if conf.Status == 0 {
		return nil, errors.New("missing Status")
	}
	return Check{HttpConf: conf.Config, Status: conf.Status}, nil
}

func (c Check) Exec() result.Result {
	resp, err := http.Request(c.HttpConf)

	if err != nil {
		return result.Fail("failed to get %s: %s", c.HttpConf, err.Error())
	}
	if resp.StatusCode != c.Status {
		return result.Fail("%s responded with \"%s\"", c.HttpConf, resp.Status)
	}
	return result.OK("%s responded with \"%s\"", c.HttpConf, resp.Status)
}
