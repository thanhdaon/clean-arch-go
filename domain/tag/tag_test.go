package tag_test

import (
	"clean-arch-go/domain/tag"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewTag(t *testing.T) {
	t.Parallel()

	tg, err := tag.NewTag("task-id-1", "bug")
	require.NoError(t, err)
	require.NotNil(t, tg)
	require.NotEmpty(t, tg.UUID())
	require.Equal(t, "task-id-1", tg.TaskID())
	require.Equal(t, "bug", tg.Name())
}

func TestNewTag_EmptyTaskID(t *testing.T) {
	t.Parallel()

	_, err := tag.NewTag("", "bug")
	require.Error(t, err)
	require.EqualError(t, err, "empty task id")
}

func TestNewTag_EmptyName(t *testing.T) {
	t.Parallel()

	_, err := tag.NewTag("task-id-1", "")
	require.Error(t, err)
	require.EqualError(t, err, "empty tag name")
}

func TestFrom(t *testing.T) {
	t.Parallel()

	tg, err := tag.From("tag-id-1", "task-id-1", "feature")
	require.NoError(t, err)
	require.Equal(t, "tag-id-1", tg.UUID())
	require.Equal(t, "task-id-1", tg.TaskID())
	require.Equal(t, "feature", tg.Name())
}

func TestFrom_EmptyID(t *testing.T) {
	t.Parallel()

	_, err := tag.From("", "task-id-1", "feature")
	require.Error(t, err)
	require.EqualError(t, err, "empty tag id")
}
