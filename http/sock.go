package http

import (
	"encoding/json"
	"net"
	"sync"

	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
	"github.com/ushis/gesundheit/check"
)

type sockPool struct {
	mutex *sync.Mutex
	pool  []sock
}

type sock struct {
	c net.Conn
	w *wsutil.Writer
	e *json.Encoder
}

func newSockPool() *sockPool {
	return &sockPool{
		mutex: &sync.Mutex{},
		pool:  []sock{},
	}
}

func (p *sockPool) serve(conn net.Conn) {
	writer := wsutil.NewWriter(conn, ws.StateServerSide, ws.OpText)
	encoder := json.NewEncoder(writer)
	sock := sock{conn, writer, encoder}

	p.mutex.Lock()
	p.pool = append(p.pool, sock)
	p.mutex.Unlock()

	for {
		if _, _, err := wsutil.ReadClientData(conn); err != nil {
			p.close(sock)
			return
		}
	}
}

func (p *sockPool) broadcast(e check.Event) {
	p.mutex.Lock()

	for _, sock := range p.pool {
		if err := sock.e.Encode(e); err != nil {
			go p.close(sock)
		}
		sock.w.Flush()
	}
	p.mutex.Unlock()
}

func (p *sockPool) close(sock sock) {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	for i, s := range p.pool {
		if s == sock {
			p.pool[i] = p.pool[len(p.pool)-1]
			p.pool = p.pool[:len(p.pool)-1]
			s.c.Close()
			return
		}
	}
}

func (p *sockPool) closeAll() {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	for _, sock := range p.pool {
		sock.c.Close()
	}
	p.pool = p.pool[:0]
}
