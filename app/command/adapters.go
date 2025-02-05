package command

import (
	"clean-arch-go/domain/task"
	"clean-arch-go/domain/user"
	"context"
)

type ID interface {
	New() string
	IsValid(id string) bool
}

type TaskUpdater func(context.Context, task.Task) (task.Task, error)

type TaskRepository interface {
	Add(context.Context, task.Task) error
	UpdateByID(ctx context.Context, uuid string, updateFn TaskUpdater) error
}

type UserRepository interface {
	Add(context.Context, user.User) error
	FindById(ctx context.Context, uuid string) (user.User, error)
}

type VideoService interface {
	GetAll(context.Context) error
}
