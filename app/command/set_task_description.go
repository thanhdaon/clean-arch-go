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

type SetTaskDescription struct {
	TaskId      string
	Description string
	Updater     user.User
}

type SetTaskDescriptionHandler decorator.CommandHandler[SetTaskDescription]

type setTaskDescriptionHandler struct {
	tasks TaskRepository
}

func NewSetTaskDescriptionHandler(taskRepository TaskRepository, logger *logrus.Entry) SetTaskDescriptionHandler {
	if taskRepository == nil {
		log.Fatalln("nil taskRepository")
	}
	handler := setTaskDescriptionHandler{tasks: taskRepository}
	return decorator.ApplyCommandDecorators(handler, logger)
}

func (h setTaskDescriptionHandler) Handle(ctx context.Context, cmd SetTaskDescription) error {
	op := errors.Op("cmd.SetTaskDescription")

	updateFn := func(ctx context.Context, t task.Task) (task.Task, error) {
		if err := t.SetDescription(cmd.Updater, cmd.Description); err != nil {
			return nil, errors.E(errors.Op("cmd.SetTaskDescription.updateFn"), err)
		}
		return t, nil
	}

	if err := h.tasks.UpdateByID(ctx, cmd.TaskId, updateFn); err != nil {
		return errors.E(op, err)
	}
	return nil
}
