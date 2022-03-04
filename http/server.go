package http

import (
	"context"
	"embed"
	"encoding/json"
	"io/fs"
	"net"
	"net/http"
	"sync"

	"github.com/ushis/gesundheit/db"
	"github.com/ushis/gesundheit/result"

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
	Listen  string
	db      db.Database
	sockets *sockPool
}

type Config struct {
	Listen string
}

func New(listen string, db db.Database) *Server {
	return &Server{Listen: listen, db: db, sockets: newSockPool()}
}

func (s *Server) Run(ctx context.Context, wg *sync.WaitGroup) error {
	l, err := net.Listen("tcp", s.Listen)

	if err != nil {
		return err
	}
	wg.Add(2)

	go func() {
		s.run(l)
		s.sockets.closeAll()
		wg.Done()
	}()

	go func() {
		<-ctx.Done()
		l.Close()
		wg.Done()
	}()

	return nil
}

func (s *Server) Handle(e result.Event) error {
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
	if events, err := s.db.GetEvents(); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	} else {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		json.NewEncoder(w).Encode(events)
	}
}

func (s *Server) serveEventsSocket(w http.ResponseWriter, r *http.Request) {
	if conn, _, _, err := ws.UpgradeHTTP(r, w); err == nil {
		s.sockets.serve(conn)
	}
}
