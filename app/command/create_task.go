package command

import (
	"clean-arch-go/domain/task"
	"clean-arch-go/domain/user"
	"context"
	"log"
)

type CreateTask struct {
	Title   string
	Creator user.User
}

type CreateTaskHandler struct {
	id   ID
	repo TaskRepository
}

func NewCreateTaskHandler(id ID, repo TaskRepository) CreateTaskHandler {
	if id == nil {
		log.Fatalln("nil ID")
	}

	if repo == nil {
		log.Fatalln("nil repo")
	}

	return CreateTaskHandler{id: id, repo: repo}
}

func (h CreateTaskHandler) Handle(ctx context.Context, cmd CreateTask) (err error) {
	uuid := h.id.New()

	task, err := task.NewTask(cmd.Creator, uuid, cmd.Title)
	if err != nil {
		return err
	}

	if err := h.repo.Add(ctx, task); err != nil {
		return err
	}

	return err
}
