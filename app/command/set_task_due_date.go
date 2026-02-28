package command

import (
	"clean-arch-go/core/decorator"
	"clean-arch-go/core/errors"
	"clean-arch-go/domain/activity"
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
	tasks      TaskRepository
	activities ActivityRepository
}

func NewSetTaskDueDateHandler(taskRepository TaskRepository, activities ActivityRepository, logger *logrus.Entry) SetTaskDueDateHandler {
	if taskRepository == nil {
		log.Fatalln("nil taskRepository")
	}
	if activities == nil {
		log.Fatalln("nil activities")
	}
	return decorator.ApplyCommandDecorators(setTaskDueDateHandler{tasks: taskRepository, activities: activities}, logger)
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

	a, err := activity.New(cmd.TaskId, cmd.Updater.UUID(), activity.TypeDueDateSet,
		map[string]any{"due_date": cmd.DueDate})
	if err != nil {
		return errors.E(op, err)
	}

	if err := h.activities.Add(ctx, a); err != nil {
		return errors.E(op, err)
	}
	return nil
}
