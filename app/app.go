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
	AddUser            command.AddUserHandler
	UpdateUserRole     command.UpdateUserRoleHandler
	DeleteUser         command.DeleteUserHandler
	UpdateUserProfile  command.UpdateUserProfileHandler
	CreateTask         command.CreateTaskHandler
	ChangeTaskStatus   command.ChangeTaskStatusHandler
	AssignTask         command.AssignTaskHandler
	UpdateTaskTitle    command.UpdateTaskTitleHandler
	UnassignTask       command.UnassignTaskHandler
	ReopenTask         command.ReopenTaskHandler
	DeleteTask         command.DeleteTaskHandler
	ArchiveTask        command.ArchiveTaskHandler
	SetTaskPriority    command.SetTaskPriorityHandler
	SetTaskDueDate     command.SetTaskDueDateHandler
	SetTaskDescription command.SetTaskDescriptionHandler
	AddTaskTag         command.AddTaskTagHandler
	RemoveTaskTag      command.RemoveTaskTagHandler
	AddComment         command.AddCommentHandler
	UpdateComment      command.UpdateCommentHandler
	DeleteComment      command.DeleteCommentHandler
}

type Queries struct {
	Tasks          query.TasksHandler
	User           query.UserHandler
	Users          query.UsersHandler
	Login          query.LoginHandler
	TaskActivities query.TaskActivitiesHandler
}
