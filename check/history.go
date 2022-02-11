package check

const OK = 0
const FAIL = 1

type Result uint32

type History uint32

func (h History) Last() Result {
	return Result(h & 1)
}
