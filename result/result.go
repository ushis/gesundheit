package result

import "fmt"

type Result struct {
	Status  Status
	Message string
}

func OK(format string, args ...interface{}) Result {
	return Result{StatusOK, fmt.Sprintf(format, args...)}
}

func Fail(format string, args ...interface{}) Result {
	return Result{StatusFail, fmt.Sprintf(format, args...)}
}
