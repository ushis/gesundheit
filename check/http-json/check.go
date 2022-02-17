package mtime

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/itchyny/gojq"
	"github.com/ushis/gesundheit/check"
)

type Check struct {
	Url   string
	Query *gojq.Query
	Value interface{}
}

type Config struct {
	Url   string
	Query string
	Value interface{}
}

func init() {
	check.Register("http-json", New)
}

func New(configure func(interface{}) error) (check.Check, error) {
	cfg := Config{}

	if err := configure(&cfg); err != nil {
		return nil, err
	}
	if cfg.Url == "" {
		return nil, errors.New("missing Url")
	}
	if cfg.Query == "" {
		return nil, errors.New("missing Query")
	}
	if cfg.Value == nil {
		return nil, errors.New("missing Value")
	}
	query, err := gojq.Parse(cfg.Query)

	if err != nil {
		return nil, err
	}
	return &Check{Url: cfg.Url, Query: query, Value: cfg.Value}, nil
}

func (c Check) Exec() (string, error) {
	resp, err := http.Get(c.Url)

	if err != nil {
		return "", fmt.Errorf("failed to get %s: %s", c.Url, err.Error())
	}
	body := make(map[string]interface{})

	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return "", fmt.Errorf("failed to decode response: %s", err.Error())
	}
	iter := c.Query.Run(body)
	n := 0

	for {
		v, ok := iter.Next()

		if !ok {
			if n == 0 {
				return "", fmt.Errorf("%s -> \"%s\" returned no values", c.Url, c.Query)
			}
			return fmt.Sprintf("%s -> \"%s\" returned %#v", c.Url, c.Query, c.Value), nil
		}
		if n > 1 {
			return "", fmt.Errorf("%s -> \"%s\" returned multiple values", c.Url, c.Query)
		}
		if v != c.Value {
			return "", fmt.Errorf("%s -> \"%s\" returned %#v", c.Url, c.Query, v)
		}
		n += 1
	}
}
