package mtime

import (
	"errors"
	"fmt"

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

func (c Check) Exec() (string, error) {
	resp, err := http.Request(c.HttpConf)

	if err != nil {
		return "", fmt.Errorf("failed to get %s: %s", c.HttpConf, err.Error())
	}
	if resp.StatusCode != c.Status {
		return "", fmt.Errorf("%s responded with \"%s\"", c.HttpConf, resp.Status)
	}
	return fmt.Sprintf("%s responded with \"%s\"", c.HttpConf, resp.Status), nil
}
