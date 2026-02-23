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

type UserDTO struct {
	UUID  string `json:"uuid"`
	Role  string `json:"role"`
	Name  string `json:"name"`
	Email string `json:"email"`
}
