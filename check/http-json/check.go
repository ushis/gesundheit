package mtime

import (
	"encoding/json"
	"errors"

	"github.com/itchyny/gojq"
	"github.com/ushis/gesundheit/check"
	"github.com/ushis/gesundheit/check/http"
	"github.com/ushis/gesundheit/result"
)

type Check struct {
	HttpConf http.Config
	Query    *gojq.Query
	Value    interface{}
}

type Config struct {
	http.Config
	Query string
	Value interface{}
}

func init() {
	check.Register("http-json", New)
}

func New(_ check.Database, configure func(interface{}) error) (check.Check, error) {
	conf := Config{}

	if err := configure(&conf); err != nil {
		return nil, err
	}
	if conf.Url == "" {
		return nil, errors.New("missing Url")
	}
	if conf.Query == "" {
		return nil, errors.New("missing Query")
	}
	if conf.Value == nil {
		return nil, errors.New("missing Value")
	}
	query, err := gojq.Parse(conf.Query)

	if err != nil {
		return nil, err
	}
	if n, ok := conf.Value.(int); ok {
		conf.Value = int64(n)
	}
	return &Check{HttpConf: conf.Config, Query: query, Value: conf.Value}, nil
}

func (c Check) Exec() result.Result {
	resp, err := http.Request(c.HttpConf)

	if err != nil {
		return result.Fail("failed to %s: %s", c.HttpConf, err)
	}
	defer resp.Body.Close()

	body := make(map[string]interface{})

	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return result.Fail("failed to decode response: %s", err.Error())
	}
	iter := c.Query.Run(body)
	n := 0

	for {
		v, ok := iter.Next()

		if !ok {
			if n == 0 {
				return result.Fail("%s -> \"%s\" returned no values", c.HttpConf, c.Query)
			}
			return result.OK("%s -> \"%s\" returned %#v", c.HttpConf, c.Query, c.Value)
		}
		if n > 1 {
			return result.Fail("%s -> \"%s\" returned multiple values", c.HttpConf, c.Query)
		}
		if n, ok := v.(int); ok {
			v = int64(n)
		}
		if v != c.Value {
			return result.Fail("%s -> \"%s\" returned %#v", c.HttpConf, c.Query, v)
		}
		n += 1
	}
}
