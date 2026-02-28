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
