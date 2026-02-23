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

type ReopenTask struct {
	TaskId string
	Caller user.User
}

type ReopenTaskHandler decorator.CommandHandler[ReopenTask]

type reopenTaskHandler struct {
	tasks TaskRepository
}

func NewReopenTaskHandler(taskRepository TaskRepository, logger *logrus.Entry) ReopenTaskHandler {
	if taskRepository == nil {
		log.Fatalln("nil taskRepository")
	}

	handler := reopenTaskHandler{
		tasks: taskRepository,
	}

	return decorator.ApplyCommandDecorators(handler, logger)
}

func (h reopenTaskHandler) Handle(ctx context.Context, cmd ReopenTask) error {
	op := errors.Op("cmd.ReopenTask")

	updateFn := func(ctx context.Context, t task.Task) (task.Task, error) {
		if err := t.Reopen(cmd.Caller); err != nil {
			return nil, errors.E(errors.E("cmd.ReopenTask.updateFn"), err)
		}
		return t, nil
	}

	if err := h.tasks.UpdateByID(ctx, cmd.TaskId, updateFn); err != nil {
		return errors.E(op, err)
	}

	return nil
}
