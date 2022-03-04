package result

const (
	StatusOK   Status = 0
	StatusFail Status = 1
)

type Status uint8

func (s Status) String() string {
	if s == StatusOK {
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
