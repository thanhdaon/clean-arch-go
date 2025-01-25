package query

import "time"

type Task struct {
	UUID       string
	Title      string
	Status     string
	CreatedBy  string
	AssignedTo string
	CreatedAt  time.Time
	UpdatedAt  time.Time
}
