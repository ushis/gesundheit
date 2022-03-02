package mtime

import (
	"errors"

	"github.com/ushis/gesundheit/check"
	"github.com/ushis/gesundheit/check/http"
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

func New(configure func(interface{}) error) (check.Check, error) {
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

func (c Check) Exec() check.Result {
	resp, err := http.Request(c.HttpConf)

	if err != nil {
		return check.Fail("failed to get %s: %s", c.HttpConf, err.Error())
	}
	if resp.StatusCode != c.Status {
		return check.Fail("%s responded with \"%s\"", c.HttpConf, resp.Status)
	}
	return check.OK("%s responded with \"%s\"", c.HttpConf, resp.Status)
}
