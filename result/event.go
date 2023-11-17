package result

import "time"

type Event struct {
	NodeName         string
	CheckId          string
	CheckDescription string
	CheckInterval    uint64
	Id               string
	Status           Status
	Message          string
	Timestamp        time.Time
	ExpiresAt        time.Time
}
