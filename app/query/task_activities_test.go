package query_test

import (
	"clean-arch-go/app/query"
	"context"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
)

type fakeActivityReadModel struct{}

func (f fakeActivityReadModel) ActivityForTask(_ context.Context, taskID string, _, _ int) ([]query.ActivityDTO, error) {
	return []query.ActivityDTO{
		{UUID: "a1", TaskID: taskID, ActorID: "u1", ActorName: "Alice", Type: "status_changed",
			Payload: map[string]any{}, CreatedAt: time.Now()},
	}, nil
}

func TestTaskActivitiesHandler(t *testing.T) {
	t.Parallel()
	logger := logrus.NewEntry(logrus.New())
	handler := query.NewTaskActivitiesHandler(fakeActivityReadModel{}, logger)
	result, err := handler.Handle(context.Background(), query.TaskActivities{TaskID: "task-1", Limit: 10, Offset: 0})
	require.NoError(t, err)
	require.Len(t, result, 1)
	require.Equal(t, "task-1", result[0].TaskID)
}
