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

type ArchiveTask struct {
	TaskId string
	Caller user.User
}

type ArchiveTaskHandler decorator.CommandHandler[ArchiveTask]

type archiveTaskHandler struct {
	tasks      TaskRepository
	activities ActivityRepository
}

func NewArchiveTaskHandler(taskRepository TaskRepository, activities ActivityRepository, logger *logrus.Entry) ArchiveTaskHandler {
	if taskRepository == nil {
		log.Fatalln("nil taskRepository")
	}
	if activities == nil {
		log.Fatalln("nil activities")
	}
	return decorator.ApplyCommandDecorators(archiveTaskHandler{tasks: taskRepository, activities: activities}, logger)
}

func (h archiveTaskHandler) Handle(ctx context.Context, cmd ArchiveTask) error {
	op := errors.Op("cmd.ArchiveTask")

	updateFn := func(ctx context.Context, t task.Task) (task.Task, error) {
		if err := t.Archive(cmd.Caller); err != nil {
			return nil, errors.E(errors.E("cmd.ArchiveTask.updateFn"), err)
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

	if err := h.activities.Add(ctx, a); err != nil {
		return errors.E(op, err)
	}
	return nil
}
