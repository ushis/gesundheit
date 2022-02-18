package http

import (
	"fmt"
	"net/http"
)

type Config struct {
	Method string
	Url    string
	Header http.Header
}

func (c Config) method() string {
	if c.Method == "" {
		return "GET"
	}
	return c.Method
}

func (c Config) String() string {
	return fmt.Sprintf("%s %s", c.method(), c.Url)
}

func Request(c Config) (*http.Response, error) {
	req, err := http.NewRequest(c.method(), c.Url, nil)

	if err != nil {
		return nil, err
	}
	req.Header = c.Header

	return (&http.Client{}).Do(req)
}
