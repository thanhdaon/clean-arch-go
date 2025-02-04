package command

import (
	"clean-arch-go/common/errors"
	"clean-arch-go/domain/task"
	"clean-arch-go/domain/user"
	"context"
)

type ChangeTaskStatus struct {
	TaskId  string
	Status  string
	Changer user.User
}

type ChangeTaskStatusHandler struct {
	tasks TaskRepository
}

func NewChangeTaskStatusHandler(taskRepository TaskRepository) ChangeTaskStatusHandler {
	return ChangeTaskStatusHandler{tasks: taskRepository}
}

func (h ChangeTaskStatusHandler) Handle(ctx context.Context, cmd ChangeTaskStatus) error {
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

	return nil
}
