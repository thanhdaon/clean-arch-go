package command

import (
	"clean-arch-go/domain/task"
	"clean-arch-go/domain/user"
	"context"
	"log"
)

type AssignTask struct {
	TaskId   string
	Assigner user.User
	Assignee user.User
}

type AssignTaskHandler struct {
	repo TaskRepository
}

func NewAssignTaskHandler(repo TaskRepository) AssignTaskHandler {
	if repo == nil {
		log.Fatalln("nil repo")
	}

	return AssignTaskHandler{repo: repo}
}

func (h AssignTaskHandler) Handle(ctx context.Context, cmd AssignTask) error {
	updateFn := func(ctx context.Context, t task.Task) (task.Task, error) {
		if err := t.AssignTo(cmd.Assigner, cmd.Assignee); err != nil {
			return nil, err
		}

		return t, nil
	}

	if err := h.repo.UpdateByID(ctx, cmd.TaskId, updateFn); err != nil {
		return err
	}

	return nil
}
