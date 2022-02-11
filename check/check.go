package check

type Check interface {
	Exec() (string, error)
}
