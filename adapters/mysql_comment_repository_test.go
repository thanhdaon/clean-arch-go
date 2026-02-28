package adapters_test

import (
	"clean-arch-go/adapters"
	"clean-arch-go/domain/comment"
	"clean-arch-go/domain/task"
	"clean-arch-go/domain/user"
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMysqlCommentRepository_AddAndUpdate(t *testing.T) {
	t.Parallel()
	db, err := adapters.NewMySQLConnection()
	require.NoError(t, err)
	repo := adapters.NewMysqlCommentRepository(db)
	taskRepo := adapters.NewMysqlTaskRepository(db)
	userRepo := adapters.NewMysqlUserRepository(db)

	ctx := context.Background()
	u, tk := seedUserAndTask(t, ctx, userRepo, taskRepo)

	c, err := comment.New(tk.UUID(), u.UUID(), "Hello @[user:uuid-123]")
	require.NoError(t, err)

	err = repo.Add(ctx, c)
	require.NoError(t, err)

	err = repo.UpdateByID(ctx, c.UUID(), func(ctx context.Context, existing comment.Comment) (comment.Comment, error) {
		return existing, existing.Update(u.UUID(), "Updated content")
	})
	require.NoError(t, err)
}

func seedUserAndTask(t *testing.T, ctx context.Context, userRepo adapters.MysqlUserRepository, taskRepo adapters.MysqlTaskRepository) (user.User, task.Task) {
	t.Helper()
	u, err := user.NewUser(adapters.NewID().New(), user.RoleEmployer, "Seed User", "seed@example.com")
	require.NoError(t, err)
	require.NoError(t, userRepo.Add(ctx, u))

	tk, err := task.NewTask(u, adapters.NewID().New(), "Seed Task")
	require.NoError(t, err)
	require.NoError(t, taskRepo.Add(ctx, tk))

	return u, tk
}
