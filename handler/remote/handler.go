package remote

import "github.com/ushis/gesundheit/handler"

func init() {
	handler.Register("remote", New)
}

type Handler struct{}

type Config struct {
	PrivateKey string
	PublicKey  string
	Address    string
}

func New(configure func(interface{}) error) (handler.Handler, error) {
	return nil, nil
}
