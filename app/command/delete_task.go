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

type DeleteTask struct {
	TaskId string
	Caller user.User
}

type DeleteTaskHandler decorator.CommandHandler[DeleteTask]

type deleteTaskHandler struct {
	tasks TaskRepository
}

func NewDeleteTaskHandler(taskRepository TaskRepository, logger *logrus.Entry) DeleteTaskHandler {
	if taskRepository == nil {
		log.Fatalln("nil taskRepository")
	}

	handler := deleteTaskHandler{
		tasks: taskRepository,
	}

	return decorator.ApplyCommandDecorators(handler, logger)
}

func (h deleteTaskHandler) Handle(ctx context.Context, cmd DeleteTask) error {
	op := errors.Op("cmd.DeleteTask")

	updateFn := func(ctx context.Context, t task.Task) (task.Task, error) {
		if err := t.Delete(cmd.Caller); err != nil {
			return nil, errors.E(errors.E("cmd.DeleteTask.updateFn"), err)
		}
		return t, nil
	}

	if err := h.tasks.UpdateByID(ctx, cmd.TaskId, updateFn); err != nil {
		return errors.E(op, err)
	}

	return nil
}
