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

type UpdateTaskTitle struct {
	TaskId  string
	Title   string
	Updater user.User
}

type UpdateTaskTitleHandler decorator.CommandHandler[UpdateTaskTitle]

type updateTaskTitleHandler struct {
	tasks      TaskRepository
	activities ActivityRepository
}

func NewUpdateTaskTitleHandler(taskRepository TaskRepository, activities ActivityRepository, logger *logrus.Entry) UpdateTaskTitleHandler {
	if taskRepository == nil {
		log.Fatalln("nil taskRepository")
	}
	if activities == nil {
		log.Fatalln("nil activities")
	}
	return decorator.ApplyCommandDecorators(updateTaskTitleHandler{tasks: taskRepository, activities: activities}, logger)
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

	a, err := activity.New(cmd.TaskId, cmd.Updater.UUID(), activity.TypeTitleUpdated,
		map[string]any{"title": cmd.Title})
	if err != nil {
		return errors.E(op, err)
	}

	if err := h.activities.Add(ctx, a); err != nil {
		return errors.E(op, err)
	}
	return nil
}
