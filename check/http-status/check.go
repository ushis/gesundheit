package mtime

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/ushis/gesundheit/check"
)

type Check struct {
	Url    string
	Status int
}

func init() {
	check.Register("http-status", New)
}

func New(configure func(interface{}) error) (check.Check, error) {
	chk := Check{}

	if err := configure(&chk); err != nil {
		return nil, err
	}
	if chk.Url == "" {
		return nil, errors.New("missing Url")
	}
	if chk.Status == 0 {
		return nil, errors.New("missing Status")
	}
	return chk, nil
}

func (c Check) Exec() (string, error) {
	resp, err := http.Get(c.Url)

	if err != nil {
		return "", fmt.Errorf("failed to get %s: %s", c.Url, err.Error())
	}
	if resp.StatusCode != c.Status {
		return "", fmt.Errorf("%s responded with \"%s\"", c.Url, resp.Status)
	}
	return fmt.Sprintf("%s responded with \"%s\"", c.Url, resp.Status), nil
}
