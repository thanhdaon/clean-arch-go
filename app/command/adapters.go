package command

import (
	"clean-arch-go/domain/tag"
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
	AddTag(ctx context.Context, t tag.Tag) error
	RemoveTag(ctx context.Context, tagID string) error
}

type TagRepository interface {
	FindById(ctx context.Context, uuid string) (tag.Tag, error)
}

type UserRepository interface {
	Add(context.Context, user.User) error
	FindById(ctx context.Context, uuid string) (user.User, error)
	FindByEmail(ctx context.Context, email string) (user.User, error)
	FindAll(ctx context.Context) ([]user.User, error)
	UpdateByID(ctx context.Context, uuid string, updateFn UserUpdater) error
	DeleteByID(ctx context.Context, uuid string) error
}

type UserUpdater func(context.Context, user.User) (user.User, error)

type AuthService interface {
	CreateIDToken(data map[string]any) (string, error)
}

type VideoService interface {
	GetAll(context.Context) error
}
