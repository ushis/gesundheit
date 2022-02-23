package filepresence

import (
	"fmt"
	"os"

	"github.com/ushis/gesundheit/check"
)

type Check struct {
	Path    string
	Present bool
}

func init() {
	check.Register("file-presence", New)
}

func New(configure func(interface{}) error) (check.Check, error) {
	check := Check{Present: true}

	if err := configure(&check); err != nil {
		return nil, err
	}
	return check, nil
}

func (c Check) Exec() (string, error) {
	_, err := os.Stat(c.Path)

	if os.IsNotExist(err) {
		if c.Present {
			return "", fmt.Errorf("%s is absent", c.Path)
		}
		return fmt.Sprintf("%s is absent", c.Path), nil
	}
	if err != nil {
		return "", fmt.Errorf("failed to stat %s: %s", c.Path, err)
	}
	if !c.Present {
		return "", fmt.Errorf("%s is present", c.Path)
	}
	return fmt.Sprintf("%s is present", c.Path), nil
}
