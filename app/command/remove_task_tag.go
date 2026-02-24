package command

import (
	"clean-arch-go/core/decorator"
	"clean-arch-go/core/errors"
	"context"
	"log"

	"github.com/sirupsen/logrus"
)

type RemoveTaskTag struct {
	TagId string
}

type RemoveTaskTagHandler decorator.CommandHandler[RemoveTaskTag]

type removeTaskTagHandler struct {
	tasks TaskRepository
}

func NewRemoveTaskTagHandler(taskRepository TaskRepository, logger *logrus.Entry) RemoveTaskTagHandler {
	if taskRepository == nil {
		log.Fatalln("nil taskRepository")
	}
	handler := removeTaskTagHandler{tasks: taskRepository}
	return decorator.ApplyCommandDecorators(handler, logger)
}

func (h removeTaskTagHandler) Handle(ctx context.Context, cmd RemoveTaskTag) error {
	op := errors.Op("cmd.RemoveTaskTag")

	if err := h.tasks.RemoveTag(ctx, cmd.TagId); err != nil {
		return errors.E(op, err)
	}
	return nil
}
