package adapters_test

import (
	"clean-arch-go/adapters"
	"clean-arch-go/domain/activity"
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMysqlActivityRepository_Add(t *testing.T) {
	t.Parallel()
	db, err := adapters.NewMySQLConnection()
	require.NoError(t, err)
	repo := adapters.NewMysqlActivityRepository(db)
	taskRepo := adapters.NewMysqlTaskRepository(db)
	userRepo := adapters.NewMysqlUserRepository(db)

	ctx := context.Background()
	u, tk := seedUserAndTask(t, ctx, userRepo, taskRepo)

	a, err := activity.New(tk.UUID(), u.UUID(), activity.TypeStatusChanged,
		map[string]any{"from": "todo", "to": "in_progress"})
	require.NoError(t, err)

	err = repo.Add(ctx, a)
	require.NoError(t, err)
}

func TestMysqlActivityRepository_ActivityForTask(t *testing.T) {
	t.Parallel()
	db, err := adapters.NewMySQLConnection()
	require.NoError(t, err)
	repo := adapters.NewMysqlActivityRepository(db)
	taskRepo := adapters.NewMysqlTaskRepository(db)
	userRepo := adapters.NewMysqlUserRepository(db)

	ctx := context.Background()
	u, tk := seedUserAndTask(t, ctx, userRepo, taskRepo)

	a1, _ := activity.New(tk.UUID(), u.UUID(), activity.TypeStatusChanged, map[string]any{"from": "todo", "to": "in_progress"})
	a2, _ := activity.New(tk.UUID(), u.UUID(), activity.TypeAssigned, map[string]any{"user_id": "user-123"})
	require.NoError(t, repo.Add(ctx, a1))
	require.NoError(t, repo.Add(ctx, a2))

	activities, err := repo.ActivityForTask(ctx, tk.UUID(), 10, 0)
	require.NoError(t, err)
	require.Len(t, activities, 2)

	uuids := make(map[string]bool)
	for _, a := range activities {
		uuids[a.UUID] = true
	}
	require.True(t, uuids[a1.UUID()])
	require.True(t, uuids[a2.UUID()])
}

func TestMysqlActivityRepository_ActivityForTask_Pagination(t *testing.T) {
	t.Parallel()
	db, err := adapters.NewMySQLConnection()
	require.NoError(t, err)
	repo := adapters.NewMysqlActivityRepository(db)
	taskRepo := adapters.NewMysqlTaskRepository(db)
	userRepo := adapters.NewMysqlUserRepository(db)

	ctx := context.Background()
	u, tk := seedUserAndTask(t, ctx, userRepo, taskRepo)

	for i := 0; i < 5; i++ {
		a, _ := activity.New(tk.UUID(), u.UUID(), activity.TypeCommentAdded, map[string]any{"index": i})
		require.NoError(t, repo.Add(ctx, a))
	}

	page1, err := repo.ActivityForTask(ctx, tk.UUID(), 2, 0)
	require.NoError(t, err)
	require.Len(t, page1, 2)

	page2, err := repo.ActivityForTask(ctx, tk.UUID(), 2, 2)
	require.NoError(t, err)
	require.Len(t, page2, 2)

	empty, err := repo.ActivityForTask(ctx, tk.UUID(), 2, 10)
	require.NoError(t, err)
	require.Len(t, empty, 0)
}

func TestMysqlActivityRepository_ActivityForTask_ActorName(t *testing.T) {
	t.Parallel()
	db, err := adapters.NewMySQLConnection()
	require.NoError(t, err)
	repo := adapters.NewMysqlActivityRepository(db)
	taskRepo := adapters.NewMysqlTaskRepository(db)
	userRepo := adapters.NewMysqlUserRepository(db)

	ctx := context.Background()
	u, tk := seedUserAndTask(t, ctx, userRepo, taskRepo)

	a, _ := activity.New(tk.UUID(), u.UUID(), activity.TypeStatusChanged, map[string]any{"from": "todo", "to": "in_progress"})
	require.NoError(t, repo.Add(ctx, a))

	activities, err := repo.ActivityForTask(ctx, tk.UUID(), 10, 0)
	require.NoError(t, err)
	require.Len(t, activities, 1)
	require.Equal(t, u.Name(), activities[0].ActorName)
}
