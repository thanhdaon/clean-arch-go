package app

import (
	"clean-arch-go/app/command"
	"clean-arch-go/app/query"
)

type Application struct {
	Commands Commands
	Queries  Queries
}

type Commands struct {
	CreateTask command.CreateTaskHandler
	AssignTask command.AssignTaskHandler
	AddUser    command.AddUserHandler
}

type Queries struct {
	Tasks query.TasksHandler
}
