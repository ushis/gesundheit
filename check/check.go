package check

import "github.com/ushis/gesundheit/result"

type Check interface {
	Exec() result.Result
}
