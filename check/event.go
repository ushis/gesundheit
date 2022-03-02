package check

import "time"

type Event struct {
	NodeName         string
	CheckId          string
	CheckDescription string
	StatusHistory    StatusHistory
	Result           Result
	Timestamp        time.Time
}
