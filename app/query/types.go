package query

import "time"

type Task struct {
	UUID       string    `json:"uuid"`
	Title      string    `json:"title"`
	Status     string    `json:"status"`
	CreatedBy  string    `json:"createdBy"`
	AssignedTo string    `json:"assignedTo"`
	CreatedAt  time.Time `json:"createdAt"`
	UpdatedAt  time.Time `json:"updatedAt"`
}
