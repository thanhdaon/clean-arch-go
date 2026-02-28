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

type ChangeTaskStatus struct {
	TaskId  string
	Status  string
	Changer user.User
}

type ChangeTaskStatusHandler decorator.CommandHandler[ChangeTaskStatus]

type changeTaskStatusHandler struct {
	tasks      TaskRepository
	activities ActivityRepository
}

func NewChangeTaskStatusHandler(taskRepository TaskRepository, activities ActivityRepository, logger *logrus.Entry) ChangeTaskStatusHandler {
	if taskRepository == nil {
		log.Fatalln("nil taskRepository")
	}
	if activities == nil {
		log.Fatalln("nil activities")
	}
	return decorator.ApplyCommandDecorators(changeTaskStatusHandler{tasks: taskRepository, activities: activities}, logger)
}

func (h changeTaskStatusHandler) Handle(ctx context.Context, cmd ChangeTaskStatus) error {
	op := errors.Op("cmd.ChangeTaskStatus")

	status, err := task.StatusFromString(cmd.Status)
	if err != nil {
		return errors.E(op, err)
	}

	updateFn := func(ctx context.Context, t task.Task) (task.Task, error) {
		if err := t.ChangeStatus(cmd.Changer, status); err != nil {
			return nil, errors.E(errors.E("cmd.ChangeTaskStatus.updateFn"), err)
		}
		return t, nil
	}

	if err := h.tasks.UpdateByID(ctx, cmd.TaskId, updateFn); err != nil {
		return errors.E(op, err)
	}

	a, err := activity.New(cmd.TaskId, cmd.Changer.UUID(), activity.TypeStatusChanged,
		map[string]any{"status": cmd.Status})
	if err != nil {
		return errors.E(op, err)
	}

	return errors.E(op, h.activities.Add(ctx, a))
}
