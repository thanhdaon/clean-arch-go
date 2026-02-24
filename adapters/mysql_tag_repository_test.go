package adapters_test

import (
	"clean-arch-go/adapters"
	"clean-arch-go/domain/tag"
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMysqlTagRepository_Add(t *testing.T) {
	t.Parallel()

	taskRepo := newMysqlTaskRepository(t)
	tagRepo := newMysqlTagRepository(t)

	creator := newExampleEmployer(t)
	tk := newExampleTask(t, creator)
	err := taskRepo.Add(context.Background(), tk)
	require.NoError(t, err)

	tg, err := tag.NewTag(tk.UUID(), "bug")
	require.NoError(t, err)

	err = tagRepo.Add(context.Background(), tg)
	require.NoError(t, err)

	found, err := tagRepo.FindById(context.Background(), tg.UUID())
	require.NoError(t, err)
	require.Equal(t, tg.UUID(), found.UUID())
	require.Equal(t, tg.Name(), found.Name())
	require.Equal(t, tg.TaskID(), found.TaskID())
}

func TestMysqlTagRepository_Remove(t *testing.T) {
	t.Parallel()

	taskRepo := newMysqlTaskRepository(t)
	tagRepo := newMysqlTagRepository(t)

	creator := newExampleEmployer(t)
	tk := newExampleTask(t, creator)
	err := taskRepo.Add(context.Background(), tk)
	require.NoError(t, err)

	tg, err := tag.NewTag(tk.UUID(), "feature")
	require.NoError(t, err)

	err = tagRepo.Add(context.Background(), tg)
	require.NoError(t, err)

	err = tagRepo.Remove(context.Background(), tg.UUID())
	require.NoError(t, err)

	_, err = tagRepo.FindById(context.Background(), tg.UUID())
	require.Error(t, err)
}

func TestMysqlTagRepository_FindByTaskId(t *testing.T) {
	t.Parallel()

	taskRepo := newMysqlTaskRepository(t)
	tagRepo := newMysqlTagRepository(t)

	creator := newExampleEmployer(t)
	tk := newExampleTask(t, creator)
	err := taskRepo.Add(context.Background(), tk)
	require.NoError(t, err)

	tag1, _ := tag.NewTag(tk.UUID(), "bug")
	tag2, _ := tag.NewTag(tk.UUID(), "feature")

	err = tagRepo.Add(context.Background(), tag1)
	require.NoError(t, err)
	err = tagRepo.Add(context.Background(), tag2)
	require.NoError(t, err)

	tags, err := tagRepo.FindByTaskId(context.Background(), tk.UUID())
	require.NoError(t, err)
	require.Len(t, tags, 2)
}

func newMysqlTagRepository(t *testing.T) adapters.MysqlTagRepository {
	t.Helper()
	db, err := adapters.NewMySQLConnection()
	require.NoError(t, err)
	return adapters.NewMysqlTagRepository(db)
}
