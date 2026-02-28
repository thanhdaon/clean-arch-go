package query

import (
	"clean-arch-go/core/decorator"
	"context"
	"time"

	"github.com/sirupsen/logrus"
)

type ActivityDTO struct {
	UUID      string         `json:"uuid"`
	TaskID    string         `json:"task_id"`
	ActorID   string         `json:"actor_id"`
	ActorName string         `json:"actor_name"`
	Type      string         `json:"type"`
	Payload   map[string]any `json:"payload"`
	CreatedAt time.Time      `json:"created_at"`
}

type TaskActivities struct {
	TaskID string
	Limit  int
	Offset int
}

type TaskActivitiesHandler decorator.QueryHandler[TaskActivities, []ActivityDTO]

type TaskActivitiesReadModel interface {
	ActivityForTask(ctx context.Context, taskID string, limit, offset int) ([]ActivityDTO, error)
}

type taskActivitiesHandler struct {
	readModel TaskActivitiesReadModel
}

func NewTaskActivitiesHandler(readModel TaskActivitiesReadModel, logger *logrus.Entry) TaskActivitiesHandler {
	return decorator.ApplyQueryDecorators(taskActivitiesHandler{readModel: readModel}, logger)
}

func (h taskActivitiesHandler) Handle(ctx context.Context, q TaskActivities) ([]ActivityDTO, error) {
	limit := q.Limit
	if limit == 0 {
		limit = 20
	}
	return h.readModel.ActivityForTask(ctx, q.TaskID, limit, q.Offset)
}
