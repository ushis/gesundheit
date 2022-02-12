package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"

	_ "github.com/ushis/gesundheit/check/http-json"
	_ "github.com/ushis/gesundheit/check/mtime"
	_ "github.com/ushis/gesundheit/filter/office-hours"
	_ "github.com/ushis/gesundheit/filter/result-change"
	_ "github.com/ushis/gesundheit/handler/gotify"
	_ "github.com/ushis/gesundheit/handler/log"
)

var (
	confDir string
)

func init() {
	flag.StringVar(&confDir, "confdir", "/etc/gesundheit", "configuration directory")
}

func main() {
	flag.Parse()
	h := newHub()

	if err := loadConfDir(h, confDir); err != nil {
		log.Fatalf("failed to load module config: %s", err.Error())
	}
	go h.run()

	chn := make(chan os.Signal, 1)
	signal.Notify(chn, syscall.SIGINT, syscall.SIGTERM)
	<-chn
	h.stop()
}
