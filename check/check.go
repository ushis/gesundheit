package check

type Check interface {
	Exec() Result
}
