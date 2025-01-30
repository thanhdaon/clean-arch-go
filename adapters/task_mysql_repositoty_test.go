package adapters_test

import (
	"clean-arch-go/adapters"
	"clean-arch-go/app/query"
	"clean-arch-go/domain/task"
	"clean-arch-go/domain/user"
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestMsqlTaskRepository_Add(t *testing.T) {
	t.Parallel()
	taskRepository := newMysqlTaskRepository(t)

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

func TestMysqlTaskRepository_AllTasks(t *testing.T) {
	t.Parallel()

	taskRepository := newMysqlTaskRepository(t)

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

func TestMysqlTaskRepository_Update(t *testing.T) {
	t.Parallel()

	taskRepository := newMysqlTaskRepository(t)

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

func newMysqlTaskRepository(t *testing.T) adapters.MysqlTaskRepository {
	t.Helper()
	db, err := adapters.NewMySQLConnection()
	require.NoError(t, err)
	return adapters.NewMysqlTaskRepository(db)
}

func newExampleEmployer(t *testing.T) user.User {
	t.Helper()
	employer, err := user.NewUser(adapters.NewID().New(), user.RoleEmployer)
	require.NoError(t, err)
	return employer
}

func newExampleTask(t *testing.T, creator user.User) task.Task {
	task, err := task.NewTask(creator, adapters.NewID().New(), "example task")
	require.NoError(t, err)
	return task
}

func newUpdatedTask(t *testing.T, creator user.User) task.Task {
	newTask, err := task.NewTask(creator, adapters.NewID().New(), "updated task")
	require.NoError(t, err)
	err = newTask.ChangeStatus(creator, task.StatusCompleted)
	require.NoError(t, err)
	return newTask
}

func assertQueryTasksIncludes(t *testing.T, allTasks []query.Task, tasks []task.Task) {
	t.Helper()

	for _, t1 := range tasks {
		for _, t2 := range allTasks {
			if t1.UUID() == t2.UUID {
				require.Equal(t, t1.Title(), t2.Title)
				require.Equal(t, t1.Status().String(), t2.Status)
				require.Equal(t, t1.CreatedBy(), t2.CreatedBy)
				require.Equal(t, t1.AssignedTo(), t2.AssignedTo)
				compareTimesIgnoringNanoseconds(t, t1.CreatedAt(), t2.CreatedAt)
				compareTimesIgnoringNanoseconds(t, t1.UpdatedAt(), t2.UpdatedAt)
				return
			}
		}

		t.Fatalf("task %s not found in query result", t1.UUID())
	}
}

func assertPersistedTaskEquals(t *testing.T, repo adapters.MysqlTaskRepository, target task.Task) {
	t.Helper()

	persistedTask, err := repo.FindById(context.Background(), target.UUID())
	require.NoError(t, err)
	require.NotNil(t, persistedTask)

	assertTasksEquals(t, target, persistedTask)
}

func assertTasksEquals(t *testing.T, task1, task2 task.Task) {
	t.Helper()

	require.Equal(t, task1.UUID(), task2.UUID())
	require.Equal(t, task1.Title(), task2.Title())
	require.Equal(t, task1.Status().String(), task2.Status().String())
	require.Equal(t, task1.CreatedBy(), task2.CreatedBy())
	require.Equal(t, task1.AssignedTo(), task2.AssignedTo())
	compareTimesIgnoringNanoseconds(t, task1.CreatedAt(), task2.CreatedAt())
	compareTimesIgnoringNanoseconds(t, task1.UpdatedAt(), task2.UpdatedAt())
}

func compareTimesIgnoringNanoseconds(t *testing.T, t1, t2 time.Time) {
	t.Helper()

	require.Equal(t, t1.Truncate(time.Minute).UTC(), t2.Truncate(time.Minute).UTC())
}
