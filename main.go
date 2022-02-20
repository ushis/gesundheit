package main

import (
	"context"
	"flag"
	"fmt"
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
)

var (
	confPath string
)

func init() {
	flag.StringVar(&confPath, "conf", "/etc/gesundheit/gesundheit.toml", "config file")
}

func main() {
	flag.Parse()
	log.SetOutput(os.Stdout)
	log.SetFlags(0)
	rand.Seed(time.Now().UnixNano())

	conf, err := loadConf(confPath)

	if err != nil {
		log.Fatalf("failed to load config: %s", err)
	}
	if conf.Log.Path != "-" {
		f, err := os.OpenFile(conf.Log.Path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0640)

		if err != nil {
			log.Fatalf("failed to open log file: %s", err)
		}
		defer f.Close()

		log.SetOutput(f)
	}
	if conf.Log.Timestamps {
		log.SetFlags(log.Ldate | log.Ltime)
	}
	h := newHub()
	confDir := filepath.Dir(confPath)
	moduleConfigs := filepath.Join(confDir, conf.Modules.Config)

	if err := loadModuleConfigs(h, moduleConfigs); err != nil {
		log.Fatalf("failed to load module config: %s", err)
	}
	ctx, stop := context.WithCancel(context.Background())
	wg := sync.WaitGroup{}
	wg.Add(1)
	go h.run(ctx, &wg)

	chn := make(chan os.Signal, 1)
	signal.Notify(chn, syscall.SIGINT, syscall.SIGTERM)
	<-chn
	fmt.Println("exit...")
	stop()
	wg.Wait()
	fmt.Println("done")
}
