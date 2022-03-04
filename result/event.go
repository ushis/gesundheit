package result

import "time"

type Event struct {
	NodeName         string
	CheckId          string
	CheckDescription string
	StatusHistory    StatusHistory
	Status           Status
	Message          string
	Timestamp        time.Time
}
