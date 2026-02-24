package task_test

import (
	"clean-arch-go/domain/task"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPriorityString(t *testing.T) {
	t.Parallel()
	require.Equal(t, "low", task.PriorityLow.String())
	require.Equal(t, "medium", task.PriorityMedium.String())
	require.Equal(t, "high", task.PriorityHigh.String())
	require.Equal(t, "urgent", task.PriorityUrgent.String())
}

func TestPriorityFromString(t *testing.T) {
	t.Parallel()

	tests := []struct {
		input    string
		expected task.Priority
	}{
		{"low", task.PriorityLow},
		{"medium", task.PriorityMedium},
		{"high", task.PriorityHigh},
		{"urgent", task.PriorityUrgent},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			t.Parallel()
			p, err := task.PriorityFromString(tt.input)
			require.NoError(t, err)
			require.Equal(t, tt.expected, p)
		})
	}
}

func TestPriorityFromString_Invalid(t *testing.T) {
	t.Parallel()
	_, err := task.PriorityFromString("unknown")
	require.Error(t, err)
}

func TestPriorityIsZero(t *testing.T) {
	t.Parallel()
	require.True(t, task.PriorityZero.IsZero())
	require.False(t, task.PriorityLow.IsZero())
	require.False(t, task.PriorityMedium.IsZero())
}
