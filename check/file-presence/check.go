package filepresence

import (
	"os"

	"github.com/ushis/gesundheit/check"
	"github.com/ushis/gesundheit/result"
)

type Check struct {
	Path    string
	Present bool
}

func init() {
	check.Register("file-presence", New)
}

func New(_ check.Database, configure func(interface{}) error) (check.Check, error) {
	check := Check{Present: true}

	if err := configure(&check); err != nil {
		return nil, err
	}
	return check, nil
}

func (c Check) Exec() result.Result {
	_, err := os.Stat(c.Path)

	if os.IsNotExist(err) {
		if c.Present {
			return result.Fail("%s is absent", c.Path)
		}
		return result.OK("%s is absent", c.Path)
	}
	if err != nil {
		return result.Fail("failed to stat %s: %s", c.Path, err)
	}
	if !c.Present {
		return result.Fail("%s is present", c.Path)
	}
	return result.OK("%s is present", c.Path)
}
