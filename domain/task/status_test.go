package task_test

import (
	"clean-arch-go/domain/task"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestStatusFromString_ValidStatuses(t *testing.T) {
	t.Parallel()

	status, err := task.StatusFromString("todo")
	require.NoError(t, err)
	require.Equal(t, task.StatusTodo, status)

	status, err = task.StatusFromString("pending")
	require.NoError(t, err)
	require.Equal(t, task.StatusPending, status)

	status, err = task.StatusFromString("inprogress")
	require.NoError(t, err)
	require.Equal(t, task.StatusInProgress, status)

	status, err = task.StatusFromString("completed")
	require.NoError(t, err)
	require.Equal(t, task.StatusCompleted, status)
}

func TestStatusFromString_InvalidStatus(t *testing.T) {
	t.Parallel()

	status, err := task.StatusFromString("invalid")
	require.Error(t, err)
	require.Equal(t, task.StatusUnknow, status)
	require.EqualError(t, err, "unknow status: invalid")
}

func TestStatus_StringMethod(t *testing.T) {
	t.Parallel()

	require.Equal(t, "todo", task.StatusTodo.String())
	require.Equal(t, "pending", task.StatusPending.String())
	require.Equal(t, "inprogress", task.StatusInProgress.String())
	require.Equal(t, "completed", task.StatusCompleted.String())
}

func TestStatus_IsZeroMethod(t *testing.T) {
	t.Parallel()

	require.True(t, task.StatusUnknow.IsZero())
	require.False(t, task.StatusTodo.IsZero())
	require.False(t, task.StatusPending.IsZero())
	require.False(t, task.StatusInProgress.IsZero())
	require.False(t, task.StatusCompleted.IsZero())
}
