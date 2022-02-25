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
	"syscall"
	"time"

	_ "github.com/ushis/gesundheit/check/disk-space"
	_ "github.com/ushis/gesundheit/check/file-mtime"
	_ "github.com/ushis/gesundheit/check/file-presence"
	_ "github.com/ushis/gesundheit/check/http-json"
	_ "github.com/ushis/gesundheit/check/http-status"
	_ "github.com/ushis/gesundheit/check/memory"
	_ "github.com/ushis/gesundheit/check/mtime"
	"github.com/ushis/gesundheit/crypto"
	"github.com/ushis/gesundheit/db"
	_ "github.com/ushis/gesundheit/db/memory"
	_ "github.com/ushis/gesundheit/filter/office-hours"
	_ "github.com/ushis/gesundheit/filter/result-change"
	"github.com/ushis/gesundheit/handler"
	_ "github.com/ushis/gesundheit/handler/gotify"
	_ "github.com/ushis/gesundheit/handler/log"
	_ "github.com/ushis/gesundheit/handler/remote"
	"github.com/ushis/gesundheit/http"
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

	conf, err := loadConf(confPath)

	if err != nil {
		log.Fatalln("failed to load config:", err)
	}
	f, err := openLog(conf.Log.Path)

	if err != nil {
		log.Fatalln("failed to open log file:", err)
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
		log.Fatalln("failed to load module config:", err)
	}
	dbFunc, _ := db.Get("memory")
	db, _ := dbFunc(func(_ interface{}) error { return nil })
	h.registerHandlerRunner(handler.NewRunner(db, nil))

	ctx, stop := context.WithCancel(context.Background())

	s := http.NewServer(db, ":8080")
	h.registerHandlerRunner(handler.NewRunner(s, nil))

	httpDone, _ := s.Run(ctx)

	done, err := h.run(ctx)

	if err != nil {
		log.Fatalln("failed to start:", err)
	}

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	<-sig

	stop()
	<-httpDone
	<-done
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
