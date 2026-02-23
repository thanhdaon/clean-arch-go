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

type UnassignTask struct {
	TaskId string
	Caller user.User
}

type UnassignTaskHandler decorator.CommandHandler[UnassignTask]

type unassignTaskHandler struct {
	tasks TaskRepository
}

func NewUnassignTaskHandler(taskRepository TaskRepository, logger *logrus.Entry) UnassignTaskHandler {
	if taskRepository == nil {
		log.Fatalln("nil taskRepository")
	}

	handler := unassignTaskHandler{
		tasks: taskRepository,
	}

	return decorator.ApplyCommandDecorators(handler, logger)
}

func (h unassignTaskHandler) Handle(ctx context.Context, cmd UnassignTask) error {
	op := errors.Op("cmd.UnassignTask")

	updateFn := func(ctx context.Context, t task.Task) (task.Task, error) {
		if err := t.Unassign(cmd.Caller); err != nil {
			return nil, errors.E(errors.E("cmd.UnassignTask.updateFn"), err)
		}
		return t, nil
	}

	if err := h.tasks.UpdateByID(ctx, cmd.TaskId, updateFn); err != nil {
		return errors.E(op, err)
	}

	return nil
}
