package command

import (
	"clean-arch-go/core/decorator"
	"clean-arch-go/core/errors"
	"clean-arch-go/domain/task"
	"clean-arch-go/domain/user"
	"context"
	"log"
	"time"

	"github.com/sirupsen/logrus"
)

type SetTaskDueDate struct {
	TaskId  string
	DueDate time.Time
	Updater user.User
}

type SetTaskDueDateHandler decorator.CommandHandler[SetTaskDueDate]

type setTaskDueDateHandler struct {
	tasks TaskRepository
}

func NewSetTaskDueDateHandler(taskRepository TaskRepository, logger *logrus.Entry) SetTaskDueDateHandler {
	if taskRepository == nil {
		log.Fatalln("nil taskRepository")
	}
	handler := setTaskDueDateHandler{tasks: taskRepository}
	return decorator.ApplyCommandDecorators(handler, logger)
}

func (h setTaskDueDateHandler) Handle(ctx context.Context, cmd SetTaskDueDate) error {
	op := errors.Op("cmd.SetTaskDueDate")

	updateFn := func(ctx context.Context, t task.Task) (task.Task, error) {
		if err := t.SetDueDate(cmd.Updater, cmd.DueDate); err != nil {
			return nil, errors.E(errors.Op("cmd.SetTaskDueDate.updateFn"), err)
		}
		return t, nil
	}

	if err := h.tasks.UpdateByID(ctx, cmd.TaskId, updateFn); err != nil {
		return errors.E(op, err)
	}
	return nil
}
