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

type SetTaskPriority struct {
	TaskId   string
	Priority task.Priority
	Updater  user.User
}

type SetTaskPriorityHandler decorator.CommandHandler[SetTaskPriority]

type setTaskPriorityHandler struct {
	tasks      TaskRepository
	activities ActivityRepository
}

func NewSetTaskPriorityHandler(taskRepository TaskRepository, activities ActivityRepository, logger *logrus.Entry) SetTaskPriorityHandler {
	if taskRepository == nil {
		log.Fatalln("nil taskRepository")
	}
	if activities == nil {
		log.Fatalln("nil activities")
	}
	return decorator.ApplyCommandDecorators(setTaskPriorityHandler{tasks: taskRepository, activities: activities}, logger)
}

func (h setTaskPriorityHandler) Handle(ctx context.Context, cmd SetTaskPriority) error {
	op := errors.Op("cmd.SetTaskPriority")

	updateFn := func(ctx context.Context, t task.Task) (task.Task, error) {
		if err := t.SetPriority(cmd.Updater, cmd.Priority); err != nil {
			return nil, errors.E(errors.Op("cmd.SetTaskPriority.updateFn"), err)
		}
		return t, nil
	}

	if err := h.tasks.UpdateByID(ctx, cmd.TaskId, updateFn); err != nil {
		return errors.E(op, err)
	}

	a, err := activity.New(cmd.TaskId, cmd.Updater.UUID(), activity.TypePriorityChanged,
		map[string]any{"priority": int(cmd.Priority)})
	if err != nil {
		return errors.E(op, err)
	}

	return errors.E(op, h.activities.Add(ctx, a))
}
