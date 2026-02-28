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

	"github.com/sirupsen/logrus"
)

type AssignTask struct {
	TaskId     string
	AssigneeId string
	Assigner   user.User
}

type AssignTaskHandler decorator.CommandHandler[AssignTask]

type assignTaskHandler struct {
	taskRepository TaskRepository
	userRepository UserRepository
	activities     ActivityRepository
}

func NewAssignTaskHandler(taskRepository TaskRepository, userRepository UserRepository, activities ActivityRepository, logger *logrus.Entry) AssignTaskHandler {
	if taskRepository == nil {
		log.Fatalln("nil taskRepository")
	}
	if userRepository == nil {
		log.Fatalln("nil userRepository")
	}
	if activities == nil {
		log.Fatalln("nil activities")
	}
	return decorator.ApplyCommandDecorators(assignTaskHandler{
		taskRepository: taskRepository,
		userRepository: userRepository,
		activities:     activities,
	}, logger)
}

func (h assignTaskHandler) Handle(ctx context.Context, cmd AssignTask) error {
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

	a, err := activity.New(cmd.TaskId, cmd.Assigner.UUID(), activity.TypeAssigned,
		map[string]any{"assignee_id": cmd.AssigneeId})
	if err != nil {
		return errors.E(op, err)
	}

	return errors.E(op, h.activities.Add(ctx, a))
}
