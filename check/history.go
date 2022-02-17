package check

const OK = 0
const CRITICAL = 1

type Result uint32

func (r Result) String() string {
	if r == OK {
		return "OK"
	}
	return "CRITICAL"
}

type History uint32

func (h History) Last() Result {
	return Result(h & 1)
}
