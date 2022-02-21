package check

type Result uint8

const (
	OK       Result = 0
	CRITICAL Result = 1
)

func (r Result) String() string {
	if r == OK {
		return "OK"
	}
	return "CRITICAL"
}

type History uint32

func (h *History) Append(r Result) {
	if r > 1 {
		panic("result out of bounds")
	}
	*h = (*h << 1) | History(r)
}

func (h *History) Last() Result {
	return Result(*h & 1)
}
