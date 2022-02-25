package http

import (
	"context"
	"embed"
	"encoding/json"
	"io/fs"
	"net"
	"net/http"

	"github.com/ushis/gesundheit/check"
	"github.com/ushis/gesundheit/db"

	"github.com/gobwas/ws"
)

//go:embed ui/dist
var serverFS embed.FS
var uiFS fs.FS

func init() {
	subFS, err := fs.Sub(serverFS, "ui/dist")

	if err != nil {
		panic(err)
	}
	uiFS = subFS
}

type Server struct {
	addr    string
	db      db.Database
	sockets *sockPool
}

func NewServer(db db.Database, addr string) *Server {
	return &Server{addr: addr, db: db, sockets: newSockPool()}
}

func (s *Server) Run(ctx context.Context) (<-chan struct{}, error) {
	l, err := net.Listen("tcp", s.addr)

	if err != nil {
		return nil, err
	}
	done := make(chan struct{})

	go func() {
		s.run(l)
		s.sockets.closeAll()
		close(done)
	}()

	go func() {
		<-ctx.Done()
		l.Close()
	}()

	return done, nil
}

func (s *Server) Handle(e check.Event) error {
	s.sockets.broadcast(e)
	return nil
}

func (s *Server) run(l net.Listener) {
	mux := http.NewServeMux()
	mux.Handle("/", http.FileServer(http.FS(uiFS)))
	mux.HandleFunc("/api/events", s.serveEvents)
	mux.HandleFunc("/api/events/socket", s.serveEventsSocket)
	http.Serve(l, mux)
}

func (s *Server) serveEvents(w http.ResponseWriter, r *http.Request) {
	events := s.db.GetEvents()
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	json.NewEncoder(w).Encode(events)
}

func (s *Server) serveEventsSocket(w http.ResponseWriter, r *http.Request) {
	if conn, _, _, err := ws.UpgradeHTTP(r, w); err == nil {
		s.sockets.serve(conn)
	}
}
