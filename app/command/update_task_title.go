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

type UpdateTaskTitle struct {
	TaskId  string
	Title   string
	Updater user.User
}

type UpdateTaskTitleHandler decorator.CommandHandler[UpdateTaskTitle]

type updateTaskTitleHandler struct {
	tasks TaskRepository
}

func NewUpdateTaskTitleHandler(taskRepository TaskRepository, logger *logrus.Entry) UpdateTaskTitleHandler {
	if taskRepository == nil {
		log.Fatalln("nil taskRepository")
	}

	handler := updateTaskTitleHandler{
		tasks: taskRepository,
	}

	return decorator.ApplyCommandDecorators(handler, logger)
}

func (h updateTaskTitleHandler) Handle(ctx context.Context, cmd UpdateTaskTitle) error {
	op := errors.Op("cmd.UpdateTaskTitle")

	updateFn := func(ctx context.Context, t task.Task) (task.Task, error) {
		if err := t.UpdateTitle(cmd.Updater, cmd.Title); err != nil {
			return nil, errors.E(errors.E("cmd.UpdateTaskTitle.updateFn"), err)
		}
		return t, nil
	}

	if err := h.tasks.UpdateByID(ctx, cmd.TaskId, updateFn); err != nil {
		return errors.E(op, err)
	}

	return nil
}
