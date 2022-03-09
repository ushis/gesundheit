package http

import (
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
	db     db.Database
	listen string
}

func New(db db.Database, listen string) *Server {
	return &Server{db: db, listen: listen}
}

func (s *Server) Run(wg *sync.WaitGroup) (chan<- result.Event, error) {
	l, err := net.Listen("tcp", s.listen)

	if err != nil {
		return nil, err
	}
	socks := newSockPool()
	chn := make(chan result.Event)
	wg.Add(2)

	go func() {
		s.serve(l, socks)
		socks.closeAll()
		wg.Done()
	}()

	go func() {
		s.run(socks, chn)
		l.Close()
		wg.Done()
	}()

	return chn, nil
}

func (s *Server) run(socks *sockPool, chn <-chan result.Event) {
	for e := range chn {
		socks.broadcast(e)
	}
}

func (s *Server) serve(l net.Listener, socks *sockPool) {
	mux := http.NewServeMux()
	mux.Handle("/", http.FileServer(http.FS(uiFS)))

	mux.HandleFunc("/api/events", func(w http.ResponseWriter, r *http.Request) {
		if events, err := s.db.GetEvents(); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
		} else {
			w.Header().Set("Content-Type", "application/json; charset=utf-8")
			json.NewEncoder(w).Encode(events)
		}
	})

	mux.HandleFunc("/api/events/socket", func(w http.ResponseWriter, r *http.Request) {
		if conn, _, _, err := ws.UpgradeHTTP(r, w); err == nil {
			socks.serve(conn)
		}
	})

	http.Serve(l, mux)
}
