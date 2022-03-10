package diskspace

import (
	"syscall"

	"github.com/ushis/gesundheit/check"
	"github.com/ushis/gesundheit/check/size"
	"github.com/ushis/gesundheit/result"
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

func New(_ check.Database, configure func(interface{}) error) (check.Check, error) {
	conf := Config{}

	if err := configure(&conf); err != nil {
		return nil, err
	}
	minAvailable, err := size.Parse(conf.MinAvailable)

	if err != nil {
		return nil, err
	}
	return &Check{MountPoint: conf.MountPoint, MinAvailable: minAvailable}, nil
}

func (c Check) Exec() result.Result {
	var stat syscall.Statfs_t

	if err := syscall.Statfs(c.MountPoint, &stat); err != nil {
		return result.Fail("failed to stat %s: %s", c.MountPoint, err)
	}
	if stat.Bsize < 1 {
		return result.Fail("unexpected block size: %d", stat.Bsize)
	}
	total := size.B(uint64(stat.Bsize)).Mul(size.N(stat.Blocks))
	avail := size.B(uint64(stat.Bsize)).Mul(size.N(stat.Bavail))
	availPercent := avail.Mul(size.N(100)).DivSize(total)

	if avail.CompareTo(c.MinAvailable) < 0 {
		return result.Fail("%s is running out of available disk space: %s (%s%%)", c.MountPoint, avail, availPercent)
	}
	return result.OK("%s has %s (%s%%) of disk space available", c.MountPoint, avail, availPercent)
}
