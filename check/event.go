package check

import "time"

type Event struct {
	NodeName         string
	CheckId          string
	CheckDescription string
	Result           Result
	History          History
	Message          string
	Timestamp        time.Time
}
