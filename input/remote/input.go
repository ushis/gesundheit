package remote

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net"
	"sync"

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
		pubKey, err := crypto.DecodePubKey(peerConf.PublicKey)

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

func (i *Input) Run(ctx context.Context, wg *sync.WaitGroup, events chan<- check.Event) error {
	conn, err := net.ListenPacket("udp", i.Addr)

	if err != nil {
		return err
	}
	wg.Add(1)

	go func() {
		i.serve(conn, events)
		wg.Done()
	}()

	go func() {
		<-ctx.Done()
		conn.Close()
	}()

	return nil
}

func (i Input) serve(conn net.PacketConn, events chan<- check.Event) {
	plain := make([]byte, 1024)
	packet := make([]byte, 1024)

	for {
		n, _, err := conn.ReadFrom(packet)

		if n > 0 {
			if e, err := i.decodePacket(plain, packet[:n]); err != nil {
				log.Println("failed to decode packet:", err)
			} else {
				events <- e
			}
		}
		if errors.Is(err, net.ErrClosed) {
			return
		}
		if err != nil {
			log.Println("failed to read packet", err)
		}
	}
}

func (i Input) decodePacket(buf, packet []byte) (e check.Event, err error) {
	plaintext, err := i.decryptPacket(buf[:0], packet)

	if err != nil {
		return e, err
	}
	return e, json.Unmarshal(plaintext, &e)
}

func (i Input) decryptPacket(dest, ciphertext []byte) ([]byte, error) {
	for _, peer := range i.Peers {
		plaintext, err := crypto.Decrypt(peer, dest, ciphertext)

		if err == nil {
			return plaintext, nil
		}
	}
	return nil, errors.New("failed to decrypt packet")
}
