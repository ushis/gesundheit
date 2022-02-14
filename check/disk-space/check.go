package mtime

import (
	"fmt"
	"syscall"

	"github.com/ushis/gesundheit/check"
	"github.com/ushis/gesundheit/util/size"
)

type Check struct {
	MountPoint   string
	MinAvailable size.Size
}

type Config struct {
	MountPoint   string
	MinAvailable string
}

func init() {
	check.Register("disk-space", New)
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
	return &Check{MountPoint: cfg.MountPoint, MinAvailable: minAvailable}, nil
}

func (c Check) Exec() (string, error) {
	var stat syscall.Statfs_t

	if err := syscall.Statfs(c.MountPoint, &stat); err != nil {
		return "", fmt.Errorf("failed to stat %s: %s", c.MountPoint, err)
	}
	if stat.Bsize < 1 {
		return "", fmt.Errorf("unexpected block size: %d", stat.Bsize)
	}
	total := size.B(uint64(stat.Bsize)).Mul(size.N(stat.Blocks))
	avail := size.B(uint64(stat.Bsize)).Mul(size.N(stat.Bavail))
	availPercent := avail.Mul(size.N(100)).DivSize(total)

	if avail.CompareTo(c.MinAvailable) < 0 {
		return "", fmt.Errorf("%s is running out of available disk space: %s (%s%%)", c.MountPoint, avail, availPercent)
	}
	return fmt.Sprintf("%s has %s (%s%%) of disk space available", c.MountPoint, avail, availPercent), nil
}
