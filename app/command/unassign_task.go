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

type UnassignTask struct {
	TaskId string
	Caller user.User
}

type UnassignTaskHandler decorator.CommandHandler[UnassignTask]

type unassignTaskHandler struct {
	tasks      TaskRepository
	activities ActivityRepository
}

func NewUnassignTaskHandler(taskRepository TaskRepository, activities ActivityRepository, logger *logrus.Entry) UnassignTaskHandler {
	if taskRepository == nil {
		log.Fatalln("nil taskRepository")
	}
	if activities == nil {
		log.Fatalln("nil activities")
	}
	return decorator.ApplyCommandDecorators(unassignTaskHandler{tasks: taskRepository, activities: activities}, logger)
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

	a, err := activity.New(cmd.TaskId, cmd.Caller.UUID(), activity.TypeUnassigned, nil)
	if err != nil {
		return errors.E(op, err)
	}

	return errors.E(op, h.activities.Add(ctx, a))
}
