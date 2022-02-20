package remote

import (
	"encoding/json"
	"errors"
	"net"

	"github.com/ushis/gesundheit/check"
	"github.com/ushis/gesundheit/crypto"
	"github.com/ushis/gesundheit/input"
)

type Input struct {
	Addr  string
	Peers []crypto.Cipher
}

type Config struct {
	Listen     string
	PrivateKey string
	Peers      []PeerConfig
}

type PeerConfig struct {
	PublicKey string
}

func init() {
	input.Register("remote", New)
}

func New(configure func(interface{}) error) (input.Input, error) {
	conf := Config{}

	if err := configure(&conf); err != nil {
		return nil, err
	}
	privKey, err := crypto.DecodePrivKey(conf.PrivateKey)

	if err != nil {
		return nil, err
	}
	peers := []crypto.Cipher{}

	for _, peerConf := range conf.Peers {
		pubKey, err := crypto.DecodeKey(peerConf.PublicKey)

		if err != nil {
			return nil, err
		}
		cipher, err := privKey.Cipher(pubKey)

		if err != nil {
			return nil, err
		}
		peers = append(peers, cipher)
	}
	return &Input{Addr: conf.Listen, Peers: peers}, nil
}

func (i *Input) Run(events chan<- check.Event) {
	conn, err := net.ListenPacket("udp", i.Addr)

	if err != nil {
		panic(err) // TODO
	}
	defer conn.Close()

	buf := make([]byte, 4096)

	for {
		n, _, err := conn.ReadFrom(buf)

		if n > 0 {
			e, err := i.decodePacket(buf[:n])

			if err != nil {
				print(err)
			} else {
				events <- e
			}
		}
		if err != nil {
			print(err)
		}
	}
}

func (i Input) decodePacket(ciphertext []byte) (e check.Event, err error) {
	buf := make([]byte, 4096)

	for _, peer := range i.Peers {
		plaintext, err := crypto.Decrypt(peer, buf, ciphertext)

		if err != nil {
			print(err)
			continue
		}
		err = json.Unmarshal(plaintext, &e)
		return e, err
	}
	return e, errors.New("failed to decode packet")
}

func (i *Input) Close() {
	// TODO
}
