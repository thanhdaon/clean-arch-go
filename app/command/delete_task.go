package command

import (
	"clean-arch-go/core/decorator"
	"clean-arch-go/core/errors"
	"clean-arch-go/domain/activity"
	"clean-arch-go/domain/errkind"
	"clean-arch-go/domain/task"
	"clean-arch-go/domain/user"
	"context"
	"log"
	"strings"

	"github.com/sirupsen/logrus"
)

type DeleteTask struct {
	TaskId string
	Caller user.User
}

type DeleteTaskHandler decorator.CommandHandler[DeleteTask]

type deleteTaskHandler struct {
	tasks      TaskRepository
	activities ActivityRepository
}

func NewDeleteTaskHandler(taskRepository TaskRepository, activities ActivityRepository, logger *logrus.Entry) DeleteTaskHandler {
	if taskRepository == nil {
		log.Fatalln("nil taskRepository")
	}
	if activities == nil {
		log.Fatalln("nil activities")
	}
	return decorator.ApplyCommandDecorators(deleteTaskHandler{tasks: taskRepository, activities: activities}, logger)
}

func (h deleteTaskHandler) Handle(ctx context.Context, cmd DeleteTask) error {
	op := errors.Op("cmd.DeleteTask")

	updateFn := func(ctx context.Context, t task.Task) (task.Task, error) {
		if err := t.Delete(cmd.Caller); err != nil {
			if strings.Contains(err.Error(), "only employer") || strings.Contains(err.Error(), "only task creator") {
				return nil, errors.E(errors.Op("cmd.DeleteTask.updateFn"), errkind.Permission, err)
			}
			return nil, errors.E(errors.Op("cmd.DeleteTask.updateFn"), err)
		}
		return t, nil
	}

	if err := h.tasks.UpdateByID(ctx, cmd.TaskId, updateFn); err != nil {
		return errors.E(op, err)
	}

	a, err := activity.New(cmd.TaskId, cmd.Caller.UUID(), activity.TypeArchived, nil)
	if err != nil {
		return errors.E(op, err)
	}

	return errors.E(op, h.activities.Add(ctx, a))
}
