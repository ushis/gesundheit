package mtime

import (
	"errors"
	"io"

	"github.com/tidwall/gjson"
	"github.com/ushis/gesundheit/check"
	"github.com/ushis/gesundheit/check/http"
	"github.com/ushis/gesundheit/result"
)

type Check struct {
	HttpConf http.Config
	Query    string
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
	return &Check{HttpConf: conf.Config, Query: conf.Query, Value: conf.Value}, nil
}

func (c Check) Exec() result.Result {
	resp, err := http.Request(c.HttpConf)

	if err != nil {
		return result.Fail("failed to %s: %s", c.HttpConf, err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)

	if err != nil {
		return result.Fail("failed to read response body: %s", err.Error())
	}
	results := gjson.GetManyBytes(body, c.Query)

	if len(results) == 0 {
		return result.Fail("%s -> \"%s\" returned no values", c.HttpConf, c.Query)
	}
	if len(results) > 1 {
		return result.Fail("%s -> \"%s\" returned multiple values", c.HttpConf, c.Query)
	}
	res := results[0]

	if isEqual(c.Value, res) {
		return result.OK("%s -> \"%s\" returned %#v", c.HttpConf, c.Query, c.Value)
	}
	return result.Fail("%s -> \"%s\" returned %#v", c.HttpConf, c.Query, res.Value())
}

func isEqual(val interface{}, res gjson.Result) bool {
	switch v := val.(type) {
	case int:
		return int64(v) == res.Int()
	case int32:
		return int64(v) == res.Int()
	case int64:
		return v == res.Int()
	case float32:
		return float64(v) == res.Float()
	default:
		return v == res.Value()
	}
}
