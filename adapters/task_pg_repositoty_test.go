package adapters_test

import (
	"clean-arch-go/adapters"
	"clean-arch-go/domain/task"
	"clean-arch-go/domain/user"
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestTaskPgRepository_Add(t *testing.T) {
	t.Parallel()
	taskRepository := newTaskPgRepository(t)

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
		c := c
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

func newTaskPgRepository(t *testing.T) adapters.TaskPgRepository {
	t.Helper()
	return adapters.NewTaskPgRepository()
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

func assertPersistedTaskEquals(t *testing.T, repo adapters.TaskPgRepository, target task.Task) {
	t.Helper()

	persistedTask, err := repo.FindById(context.Background(), target.UUID())
	require.NoError(t, err)

	assertTasksEquals(t, target, persistedTask)
}

func assertTasksEquals(t *testing.T, task1, task2 task.Task) {
	t.Helper()

	require.Equal(t, task1.UUID(), task2.UUID())
	require.Equal(t, task1.Title(), task2.Title())
	require.Equal(t, task1.Status().String(), task2.Status().String())
	require.Equal(t, task1.CreatedBy(), task2.CreatedBy())
	require.Equal(t, task1.AssignedTo(), task2.AssignedTo())
	require.True(t, task1.CreatedAt().Equal(task2.CreatedAt()))
	require.True(t, task1.UpdatedAt().Equal(task2.UpdatedAt()))
}
