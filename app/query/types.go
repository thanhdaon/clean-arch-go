package query

import "time"

type Task struct {
	UUID       string    `json:"uuid"`
	Title      string    `json:"title"`
	Status     string    `json:"status"`
	CreatedBy  string    `json:"created_by"`
	AssignedTo string    `json:"assigned_to"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

type UserDTO struct {
	UUID  string `json:"uuid"`
	Role  string `json:"role"`
	Name  string `json:"name"`
	Email string `json:"email"`
}
