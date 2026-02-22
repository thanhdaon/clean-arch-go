package command

import (
	"clean-arch-go/core/decorator"
	"clean-arch-go/core/errors"
	"clean-arch-go/domain/task"
	"clean-arch-go/domain/user"
	"context"
	"log"

	"github.com/sirupsen/logrus"
)

type CreateTask struct {
	Title   string
	Creator user.User
}

type CreateTaskHandler decorator.CommandHandler[CreateTask]

type createTaskHandler struct {
	id   ID
	repo TaskRepository
}

func NewCreateTaskHandler(id ID, repo TaskRepository, logger *logrus.Entry) CreateTaskHandler {
	if id == nil {
		log.Fatalln("nil ID")
	}

	if repo == nil {
		log.Fatalln("nil repo")
	}

	handler := createTaskHandler{
		id:   id,
		repo: repo,
	}

	return decorator.ApplyCommandDecorators(handler, logger)
}

func (h createTaskHandler) Handle(ctx context.Context, cmd CreateTask) error {
	op := errors.Op("cmd.CreateTask")

	uuid := h.id.New()

	t, err := task.NewTask(cmd.Creator, uuid, cmd.Title)
	if err != nil {
		return errors.E(op, err)
	}

	if err := h.repo.Add(ctx, t); err != nil {
		return errors.E(op, err)
	}

	return nil
}
