package adapters_test

import (
	"clean-arch-go/adapters"
	"clean-arch-go/domain/task"
	"clean-arch-go/domain/user"
	"context"
	"os"
	"testing"

	"cloud.google.com/go/firestore"
	"github.com/stretchr/testify/require"
	"google.golang.org/api/option"
)

func TestTaskFirestoreRepository_Add(t *testing.T) {
	t.Parallel()
	taskRepository := newFirestoreRepository(t)

	testCases := []struct {
		Name            string
		TaskConstructor func(t *testing.T, creator user.User) task.Task
	}{
		{
			Name:            "standard_task",
			TaskConstructor: newExampleTask,
		},
		{
			Name:            "updated_task",
			TaskConstructor: newUpdatedTask,
		},
	}

	creator := newExampleEmployer(t)

	for _, c := range testCases {
		t.Run(c.Name, func(t *testing.T) {
			t.Parallel()

			expectedTask := c.TaskConstructor(t, creator)
			err := taskRepository.Add(context.Background(), expectedTask)
			require.NoError(t, err)
			assertPersistedTaskEquals(t, taskRepository, expectedTask)
		})
	}

	require.NotNil(t, taskRepository)
}

func TestTaskFirestoreRepository_AllTasks(t *testing.T) {
	t.Parallel()

	taskRepository := newFirestoreRepository(t)

	err := taskRepository.RemoveAllTasks(context.Background())
	require.NoError(t, err)

	creator := newExampleEmployer(t)

	tasksToAdd := []task.Task{
		newExampleTask(t, creator),
		newExampleTask(t, creator),
		newExampleTask(t, creator),
		newExampleTask(t, creator),
	}

	for _, tk := range tasksToAdd {
		err := taskRepository.Add(context.Background(), tk)
		require.NoError(t, err)
	}

	allTasks, err := taskRepository.AllTasks(context.Background())
	require.NoError(t, err)

	assertQueryTasksIncludes(t, allTasks, tasksToAdd)
}

func TestTaskFirestoreRepository_Update(t *testing.T) {
	t.Parallel()

	taskRepository := newFirestoreRepository(t)

	creator := newExampleEmployer(t)
	assignee := newExampleEmployer(t)
	expectedTask := newExampleTask(t, creator)

	err := taskRepository.Add(context.Background(), expectedTask)
	require.NoError(t, err)

	var updatedTask task.Task

	err = taskRepository.UpdateByID(context.TODO(), expectedTask.UUID(), func(ctx context.Context, found task.Task) (task.Task, error) {
		assertTasksEquals(t, expectedTask, found)

		err := found.AssignTo(creator, assignee)
		require.NoError(t, err)

		updatedTask = found
		return found, nil
	})
	require.NoError(t, err)

	assertPersistedTaskEquals(t, taskRepository, updatedTask)
}

func newFirestoreRepository(t *testing.T) adapters.FirestoreTaskRepository {
	t.Helper()
	var opts []option.ClientOption
	if file := os.Getenv("SERVICE_ACCOUNT_FILE"); file != "" {
		opts = append(opts, option.WithCredentialsFile(file))
	}

	firestoreClient, err := firestore.NewClient(context.Background(), os.Getenv("GCP_PROJECT"), opts...)
	require.NoError(t, err)

	return adapters.NewFirestoreTaskRepository(firestoreClient)
}
