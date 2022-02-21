package main

import (
	"context"
	"flag"
	"io"
	"log"
	"math/rand"
	"os"
	"os/signal"
	"path/filepath"
	"sync"
	"syscall"
	"time"

	_ "github.com/ushis/gesundheit/check/disk-space"
	_ "github.com/ushis/gesundheit/check/http-json"
	_ "github.com/ushis/gesundheit/check/http-status"
	_ "github.com/ushis/gesundheit/check/memory"
	_ "github.com/ushis/gesundheit/check/mtime"
	_ "github.com/ushis/gesundheit/filter/office-hours"
	_ "github.com/ushis/gesundheit/filter/result-change"
	_ "github.com/ushis/gesundheit/handler/gotify"
	_ "github.com/ushis/gesundheit/handler/log"
	_ "github.com/ushis/gesundheit/handler/remote"
	_ "github.com/ushis/gesundheit/input/remote"
)

var (
	confPath string
)

func init() {
	flag.StringVar(&confPath, "conf", "/etc/gesundheit/gesundheit.toml", "config file")
}

func main() {
	log.SetOutput(os.Stdout)
	log.SetFlags(0)
	flag.Parse()
	rand.Seed(time.Now().UnixNano())

	conf, err := loadConf(confPath)

	if err != nil {
		log.Fatalf("failed to load config: %s", err)
	}
	f, err := openLog(conf.Log.Path)

	if err != nil {
		log.Fatalf("failed to open log file: %s", err)
	}
	defer f.Close()

	log.SetOutput(f)

	if conf.Log.Timestamps {
		log.SetFlags(log.Ldate | log.Ltime)
	}
	h := newHub()

	confDir := filepath.Dir(confPath)
	modConfs := filepath.Join(confDir, conf.Modules.Config)
	modConfLoader := newModConfLoader(conf.Node, h)

	if err := modConfLoader.loadAll(modConfs); err != nil {
		log.Fatalf("failed to load module config: %s", err)
	}
	wg := sync.WaitGroup{}
	ctx, stop := context.WithCancel(context.Background())

	wg.Add(1)
	go h.run(ctx, &wg)

	chn := make(chan os.Signal, 1)
	signal.Notify(chn, syscall.SIGINT, syscall.SIGTERM)
	<-chn

	stop()
	wg.Wait()
}

func openLog(path string) (io.WriteCloser, error) {
	if path == "-" {
		return os.Stdout, nil
	}
	return os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0640)
}
