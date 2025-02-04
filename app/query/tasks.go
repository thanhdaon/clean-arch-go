package query

import (
	"clean-arch-go/core/decorator"
	"clean-arch-go/domain/user"
	"context"
	"log"

	"github.com/sirupsen/logrus"
)

type Tasks struct {
	User user.User
}

type TasksHandler decorator.QueryHandler[Tasks, []Task]

type tasksHandler struct {
	readmodel TasksReadModel
}

type TasksReadModel interface {
	AllTasks(ctx context.Context) ([]Task, error)
	FindTasksForUser(ctx context.Context, userUUID string) ([]Task, error)
}

func NewTaskHandler(readmodel TasksReadModel, logger *logrus.Entry) TasksHandler {
	if readmodel == nil {
		log.Fatalln("nil readmodel")
	}

	handler := tasksHandler{readmodel: readmodel}

	return decorator.ApplyQueryDecorators(handler, logger)
}

func (h tasksHandler) Handle(ctx context.Context, query Tasks) ([]Task, error) {
	if query.User.Role() == user.RoleEmployee {
		tasks, err := h.readmodel.FindTasksForUser(ctx, query.User.UUID())
		if err != nil {
			return nil, err
		}

		return tasks, nil
	}

	tasks, err := h.readmodel.AllTasks(ctx)
	if err != nil {
		return nil, err
	}

	return tasks, nil
}
