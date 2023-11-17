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
