package remote

import (
	"bytes"
	"encoding/json"
	"net"

	"github.com/ushis/gesundheit/result"
	"github.com/ushis/gesundheit/crypto"
	"github.com/ushis/gesundheit/handler"
)

func init() {
	handler.Register("remote", New)
}

type Handler struct {
	Cipher crypto.Cipher
	Addr   *net.UDPAddr
}

type Config struct {
	PrivateKey string
	PublicKey  string
	Address    string
}

func New(configure func(interface{}) error) (handler.Handler, error) {
	conf := Config{}

	if err := configure(&conf); err != nil {
		return nil, err
	}
	privKey, err := crypto.DecodePrivKey(conf.PrivateKey)

	if err != nil {
		return nil, err
	}
	pubKey, err := crypto.DecodePubKey(conf.PublicKey)

	if err != nil {
		return nil, err
	}
	cipher, err := privKey.Cipher(pubKey)

	if err != nil {
		return nil, err
	}
	addr, err := net.ResolveUDPAddr("udp", conf.Address)

	if err != nil {
		return nil, err
	}
	return Handler{Cipher: cipher, Addr: addr}, nil
}

func (h Handler) Handle(e result.Event) error {
	buf := bytes.NewBuffer(nil)

	if err := json.NewEncoder(buf).Encode(e); err != nil {
		return err
	}
	ciphertext, err := crypto.Encrypt(h.Cipher, buf.Bytes())

	if err != nil {
		return err
	}
	conn, err := net.DialUDP("udp", nil, h.Addr)

	if err != nil {
		return err
	}
	defer conn.Close()

	_, err = conn.Write(ciphertext)
	return err
}
