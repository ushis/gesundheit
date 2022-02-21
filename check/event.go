package check

type Event struct {
	Result           Result
	Message          string
	CheckDescription string
	CheckHistory     History
	NodeName         string
}
