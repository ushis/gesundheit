package check

import "fmt"

type Status uint8

const (
	STATUS_OK   Status = 0
	STATUS_FAIL Status = 1
)

func (s Status) String() string {
	if s == STATUS_OK {
		return "OK"
	}
	return "FAIL"
}

type StatusHistory uint32

func (h *StatusHistory) Append(s Status) {
	if s > 1 {
		panic("status out of bounds")
	}
	*h = (*h << 1) | StatusHistory(s)
}

func (h *StatusHistory) Last() Status {
	return Status(*h & 1)
}

type Result struct {
	Status  Status
	Message string
}

func OK(format string, args ...interface{}) Result {
	return Result{STATUS_OK, fmt.Sprintf(format, args...)}
}

func Fail(format string, args ...interface{}) Result {
	return Result{STATUS_FAIL, fmt.Sprintf(format, args...)}
}
