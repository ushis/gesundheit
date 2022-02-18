package memory

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/ushis/gesundheit/check"
	"github.com/ushis/gesundheit/check/size"
)

type Check struct {
	MinAvailable size.Size
}

type Config struct {
	MinAvailable string
}

func init() {
	check.Register("memory", New)
}

func New(configure func(interface{}) error) (check.Check, error) {
	cfg := Config{}

	if err := configure(&cfg); err != nil {
		return nil, err
	}
	minAvailable, err := size.Parse(cfg.MinAvailable)

	if err != nil {
		return nil, err
	}
	return &Check{MinAvailable: minAvailable}, nil
}

func (c Check) Exec() (string, error) {
	f, err := os.Open("/proc/meminfo")

	if err != nil {
		return "", err
	}
	defer f.Close()

	avail, total, err := readMeminfo(f)

	if err != nil {
		return "", err
	}
	availPercent := avail.Mul(size.N(100)).DivSize(total)

	if avail.CompareTo(c.MinAvailable) < 0 {
		return "", fmt.Errorf("system running out of available memory: %s (%s%%)", avail, availPercent)
	}
	return fmt.Sprintf("system has %s (%s%%) of memory available", avail, availPercent), nil
}

func readMeminfo(r io.Reader) (avail size.Size, total size.Size, err error) {
	br := bufio.NewReader(r)
	info := make(map[string]size.Size)

	for {
		line, err := br.ReadString('\n')

		if err == io.EOF {
			break
		}
		if err != nil {
			return avail, total, err
		}
		fields := strings.SplitN(line, ":", 2)

		if len(fields) != 2 {
			return avail, total, errors.New("failed to parse meminfo")
		}
		size, err := size.Parse(strings.TrimSpace(fields[1]))

		if err != nil {
			return avail, total, err
		}
		info[fields[0]] = size
	}
	var ok bool

	if avail, ok = info["MemAvailable"]; !ok {
		return avail, total, errors.New("MemAvailable is missing in meminfo")
	}
	if total, ok = info["MemTotal"]; !ok {
		return avail, total, errors.New("MemTotal is missing in meminfo")
	}
	return avail, total, nil
}
