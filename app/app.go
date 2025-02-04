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
	AddUser          command.AddUserHandler
	CreateTask       command.CreateTaskHandler
	ChangeTaskStatus command.ChangeTaskStatusHandler
	AssignTask       command.AssignTaskHandler
}

type Queries struct {
	Tasks query.TasksHandler
}
