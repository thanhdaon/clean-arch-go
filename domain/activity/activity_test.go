package activity_test

import (
	"clean-arch-go/domain/activity"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewActivity(t *testing.T) {
	t.Parallel()
	a, err := activity.New("task-id-1", "actor-id-1", activity.TypeStatusChanged, map[string]any{"from": "todo", "to": "in_progress"})
	require.NoError(t, err)
	require.NotEmpty(t, a.UUID())
	require.Equal(t, "task-id-1", a.TaskID())
	require.Equal(t, "actor-id-1", a.ActorID())
	require.Equal(t, activity.TypeStatusChanged, a.Type())
	require.NotNil(t, a.Payload())
	require.False(t, a.CreatedAt().IsZero())
}

func TestNewActivity_MissingTaskID(t *testing.T) {
	t.Parallel()
	_, err := activity.New("", "actor-id-1", activity.TypeStatusChanged, nil)
	require.EqualError(t, err, "empty task id")
}

func TestNewActivity_MissingActorID(t *testing.T) {
	t.Parallel()
	_, err := activity.New("task-id-1", "", activity.TypeStatusChanged, nil)
	require.EqualError(t, err, "empty actor id")
}
