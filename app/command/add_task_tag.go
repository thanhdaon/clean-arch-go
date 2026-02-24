package command

import (
	"clean-arch-go/core/decorator"
	"clean-arch-go/core/errors"
	"clean-arch-go/domain/tag"
	"context"
	"log"

	"github.com/sirupsen/logrus"
)

type AddTaskTag struct {
	TaskId  string
	TagName string
}

type AddTaskTagHandler decorator.CommandHandler[AddTaskTag]

type addTaskTagHandler struct {
	tasks TaskRepository
}

func NewAddTaskTagHandler(taskRepository TaskRepository, logger *logrus.Entry) AddTaskTagHandler {
	if taskRepository == nil {
		log.Fatalln("nil taskRepository")
	}
	handler := addTaskTagHandler{tasks: taskRepository}
	return decorator.ApplyCommandDecorators(handler, logger)
}

func (h addTaskTagHandler) Handle(ctx context.Context, cmd AddTaskTag) error {
	op := errors.Op("cmd.AddTaskTag")

	newTag, err := tag.NewTag(cmd.TaskId, cmd.TagName)
	if err != nil {
		return errors.E(op, err)
	}

	if err := h.tasks.AddTag(ctx, newTag); err != nil {
		return errors.E(op, err)
	}
	return nil
}
