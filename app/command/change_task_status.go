package command

import (
	"clean-arch-go/domain/task"
	"clean-arch-go/domain/user"
	"context"
	"errors"
)

type ChangeTaskStatus struct {
	TaskId  string
	Status  task.Status
	Changer user.User
}

type ChangeTaskStatusHandler struct {
	repo TaskRepository
}

func (h ChangeTaskStatusHandler) Handle(ctx context.Context, cmd ChangeTaskStatus) error {
	if cmd.Changer == nil {
		return errors.New("This command need to be authenticated")
	}

	updateFn := func(ctx context.Context, t task.Task) (task.Task, error) {
		if err := t.ChangeStatus(cmd.Changer, cmd.Status); err != nil {
			return nil, err
		}
		return t, nil
	}

	if err := h.repo.UpdateByID(ctx, cmd.TaskId, updateFn); err != nil {
		return err
	}

	return nil
}
