package filemtime

import (
	"os/exec"
	"syscall"

	"github.com/ushis/gesundheit/check"
	"github.com/ushis/gesundheit/result"
)

type Check struct {
	Command string
	Args    []string
}

func init() {
	check.Register("exec", New)
}

func New(_ check.Database, configure func(interface{}) error) (check.Check, error) {
	check := Check{}

	if err := configure(&check); err != nil {
		return nil, err
	}
	return check, nil
}

func (c Check) Exec() result.Result {
	cmd := exec.Command(c.Command, c.Args...)
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true, Pgid: 0}
	out, err := cmd.CombinedOutput()

	if err == nil {
		return result.OK(string(out))
	}
	if _, ok := err.(*exec.ExitError); ok {
		return result.Fail(string(out))
	}
	return result.Fail("failed to exec: %s: %s", c.Command, err)
}
