package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"math/rand"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"time"

	_ "github.com/ushis/gesundheit/check/disk-space"
	_ "github.com/ushis/gesundheit/check/exec"
	_ "github.com/ushis/gesundheit/check/file-mtime"
	_ "github.com/ushis/gesundheit/check/file-presence"
	_ "github.com/ushis/gesundheit/check/heartbeat"
	_ "github.com/ushis/gesundheit/check/http-json"
	_ "github.com/ushis/gesundheit/check/http-status"
	_ "github.com/ushis/gesundheit/check/memory"
	_ "github.com/ushis/gesundheit/check/node-alive"
	_ "github.com/ushis/gesundheit/check/tls-cert"
	"github.com/ushis/gesundheit/crypto"
	_ "github.com/ushis/gesundheit/db/memory"
	_ "github.com/ushis/gesundheit/db/redis"
	_ "github.com/ushis/gesundheit/filter/office-hours"
	_ "github.com/ushis/gesundheit/filter/result-change"
	_ "github.com/ushis/gesundheit/handler/gotify"
	_ "github.com/ushis/gesundheit/handler/log"
	_ "github.com/ushis/gesundheit/handler/remote"
	_ "github.com/ushis/gesundheit/input/remote"
)

func usage() {
	fmt.Fprintf(os.Stderr, "Usage: %s [<subcommand>] [<args>]\n\n", os.Args[0])
	fmt.Fprintf(os.Stderr, "Available subcommands:\n")
	fmt.Fprintf(os.Stderr, "  serve: Runs the gesundheit service (default)\n")
	fmt.Fprintf(os.Stderr, "    -conf <path> Path to config file\n")
	fmt.Fprintf(os.Stderr, "  genkey: Generates a new private key and writes it to stdout\n")
	fmt.Fprintf(os.Stderr, "  pubkey: Reads a private key from stdin and writes a public key to stdout\n")
}

func main() {
	var cmd string
	var cmdArgs []string

	if len(os.Args) < 2 || strings.HasPrefix(os.Args[1], "-") {
		cmd = "serve"
		cmdArgs = os.Args[1:]
	} else {
		cmd = os.Args[1]
		cmdArgs = os.Args[2:]
	}

	switch cmd {
	case "serve":
		cmdServe(cmdArgs)
	case "genkey":
		cmdGenkey(cmdArgs)
	case "pubkey":
		cmdPubkey(cmdArgs)
	default:
		usage()
		os.Exit(2)
	}
}

func cmdServe(args []string) {
	var confPath string

	flags := flag.NewFlagSet("", flag.ExitOnError)
	flags.Usage = usage
	flags.StringVar(&confPath, "conf", "/etc/gesundheit/gesundheit.toml", "config file")
	flags.Parse(args)

	if flags.NArg() > 0 {
		flags.Usage()
		os.Exit(2)
	}
	rand.Seed(time.Now().UnixNano())
	log.SetOutput(os.Stdout)
	log.SetFlags(0)

	conf, db, http, err := loadConf(confPath)

	if err != nil {
		log.Fatalln("failed to load config:", err)
	}
	defer db.Close()

	f, err := openLog(conf.Log.Path)

	if err != nil {
		log.Fatalln("failed to open log file:", err)
	}
	defer f.Close()

	log.SetOutput(f)

	if conf.Log.Timestamps {
		log.SetFlags(log.Ldate | log.Ltime)
	}
	h := &hub{db: db}

	confDir := filepath.Dir(confPath)
	modConfs := filepath.Join(confDir, conf.Modules.Config)
	modConfLoader := newModConfLoader(conf.Node, h, db)

	if err := modConfLoader.loadAll(modConfs); err != nil {
		log.Fatalln("failed to load module config:", err)
	}
	wg := sync.WaitGroup{}
	ctx, stop := context.WithCancel(context.Background())

	if http != nil {
		if err := http.Run(ctx, &wg); err != nil {
			log.Fatalln("failed to run http:", err)
		}
		h.registerHandler(http)
	}

	if err := h.run(ctx, &wg); err != nil {
		log.Fatalln("failed to start:", err)
	}

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	<-sig

	stop()
	wg.Wait()
}

func openLog(path string) (io.WriteCloser, error) {
	if path == "-" {
		return os.Stdout, nil
	}
	return os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0640)
}

func cmdGenkey(args []string) {
	flags := flag.NewFlagSet("", flag.ExitOnError)
	flags.Usage = usage
	flags.Parse(args)

	if flags.NArg() > 0 {
		flags.Usage()
		os.Exit(2)
	}
	priv, err := crypto.GeneratePrivKey()

	if err != nil {
		log.Fatalln("failed to generate key:", err)
	}
	fmt.Println(priv.Encode())
}

func cmdPubkey(args []string) {
	flags := flag.NewFlagSet("", flag.ExitOnError)
	flags.Usage = usage
	flags.Parse(args)

	if flags.NArg() > 0 {
		flags.Usage()
		os.Exit(2)
	}
	buf, err := ioutil.ReadAll(os.Stdin)

	if err != nil {
		log.Fatalln("failed to read:", err)
	}
	priv, err := crypto.DecodePrivKey(string(buf))

	if err != nil {
		log.Fatalln("key has not the correct length or format")
	}
	pub, err := priv.PubKey()

	if err != nil {
		log.Fatalln("failed to calculate pubkey:", err)
	}
	fmt.Println(pub.Encode())
}
