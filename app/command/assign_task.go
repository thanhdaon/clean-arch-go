package command

import (
	"clean-arch-go/common/errors"
	"clean-arch-go/domain/errkind"
	"clean-arch-go/domain/task"
	"clean-arch-go/domain/user"
	"context"
	"log"
)

type AssignTask struct {
	TaskId     string
	AssigneeId string
	Assigner   user.User
}

type AssignTaskHandler struct {
	taskRepository TaskRepository
	userRepository UserRepository
}

func NewAssignTaskHandler(taskRepository TaskRepository, userRepository UserRepository) AssignTaskHandler {
	if taskRepository == nil {
		log.Fatalln("nil taskRepository")
	}

	if userRepository == nil {
		log.Fatalln("nil userRepository")
	}

	return AssignTaskHandler{
		taskRepository: taskRepository,
		userRepository: userRepository,
	}
}

func (h AssignTaskHandler) Handle(ctx context.Context, cmd AssignTask) error {
	op := errors.Op("cmd.AssignTask")

	assignee, err := h.userRepository.FindById(ctx, cmd.AssigneeId)
	if err != nil {
		if errors.Is(errkind.NotExist, err) {
			return errors.E(op, errkind.NotExist, err)
		}

		return errors.E(op, err)
	}

	updateFn := func(ctx context.Context, t task.Task) (task.Task, error) {
		if err := t.AssignTo(cmd.Assigner, assignee); err != nil {
			return nil, errors.E(errors.Op("cmd.AssignTask.updateFn"), err)
		}

		return t, nil
	}

	if err := h.taskRepository.UpdateByID(ctx, cmd.TaskId, updateFn); err != nil {
		return errors.E(op, err)
	}

	return nil
}
