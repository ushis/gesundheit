package gotify

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"

	"github.com/ushis/gesundheit/check"
	"github.com/ushis/gesundheit/handler"
)

type Handler struct {
	Url      string
	Token    string
	Priority int
}

func init() {
	handler.Register("gotify", New)
}

func New(configure func(interface{}) error) (handler.Handler, error) {
	h := Handler{Priority: 4}

	if err := configure(&h); err != nil {
		return nil, err
	}
	url, err := url.Parse(h.Url)

	if err != nil {
		return nil, err
	}
	url.Path = "/message"
	h.Url = url.String()
	return h, nil
}

type Message struct {
	Title    string `json:"title"`
	Message  string `json:"message"`
	Priority int    `json:"priority"`
}

func (h Handler) Handle(e check.Event) error {
	msg := Message{
		Title:    fmt.Sprintf("(%s) %s %s", e.NodeName, e.CheckDescription, e.Result.Status),
		Message:  e.Result.Message,
		Priority: h.Priority,
	}
	buf := bytes.NewBuffer(nil)

	if err := json.NewEncoder(buf).Encode(msg); err != nil {
		return err
	}
	req, err := http.NewRequest("POST", h.Url, buf)

	if err != nil {
		return err
	}
	req.Header.Set("X-Gotify-Key", h.Token)
	req.Header.Set("Content-Type", "application/json")
	resp, err := (&http.Client{}).Do(req)

	if err != nil {
		return err
	}
	if resp.StatusCode != 200 {
		return errors.New("unexpected response status: " + resp.Status)
	}
	return nil
}
