package check

type Event struct {
	Result           Result
	History          History
	Message          string
	CheckDescription string
	NodeName         string
}
