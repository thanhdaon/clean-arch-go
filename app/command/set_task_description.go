package command

import (
	"clean-arch-go/core/decorator"
	"clean-arch-go/core/errors"
	"clean-arch-go/domain/activity"
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
	tasks      TaskRepository
	activities ActivityRepository
}

func NewSetTaskDescriptionHandler(taskRepository TaskRepository, activities ActivityRepository, logger *logrus.Entry) SetTaskDescriptionHandler {
	if taskRepository == nil {
		log.Fatalln("nil taskRepository")
	}
	if activities == nil {
		log.Fatalln("nil activities")
	}
	return decorator.ApplyCommandDecorators(setTaskDescriptionHandler{tasks: taskRepository, activities: activities}, logger)
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

	a, err := activity.New(cmd.TaskId, cmd.Updater.UUID(), activity.TypeDescriptionSet, nil)
	if err != nil {
		return errors.E(op, err)
	}

	if err := h.activities.Add(ctx, a); err != nil {
		return errors.E(op, err)
	}
	return nil
}
