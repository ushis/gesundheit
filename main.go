package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	_ "github.com/ushis/gesundheit/check/mtime"
	_ "github.com/ushis/gesundheit/filter/result-change"
	_ "github.com/ushis/gesundheit/handler/gotify"
	_ "github.com/ushis/gesundheit/handler/log"
)

func main() {
	h := newHub()

	if err := loadConfDir(h, "conf"); err != nil {
		log.Fatalf("failed to load module config: %s", err.Error())
	}
	go h.run()

	chn := make(chan os.Signal, 1)
	signal.Notify(chn, syscall.SIGINT, syscall.SIGTERM)
	<-chn
	h.stop()
}
