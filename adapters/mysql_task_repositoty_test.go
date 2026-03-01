package adapters_test

import (
	"clean-arch-go/adapters"
	"clean-arch-go/app/query"
	"clean-arch-go/core/errors"
	"clean-arch-go/domain/errkind"
	"clean-arch-go/domain/tag"
	"clean-arch-go/domain/task"
	"clean-arch-go/domain/user"
	"context"
	"fmt"
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

	assertQueryTasksIncludesExpected(t, allTasks, tasksToAdd)
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

func TestMysqlTaskRepository_FindById_NotFound(t *testing.T) {
	t.Parallel()
	taskRepository := newMysqlTaskRepository(t)

	_, err := taskRepository.FindById(context.Background(), "non-existent-uuid")
	assertErrorIsNotExist(t, err)
}

func TestMysqlTaskRepository_FindById_DeletedTask(t *testing.T) {
	t.Parallel()
	taskRepository := newMysqlTaskRepository(t)
	creator := newExampleEmployer(t)

	deletedTask := newDeletedTask(t, creator)
	err := taskRepository.Add(context.Background(), deletedTask)
	require.NoError(t, err)

	_, err = taskRepository.FindById(context.Background(), deletedTask.UUID())
	assertErrorIsNotExist(t, err)
}

func TestMysqlTaskRepository_FindById_ArchivedTask(t *testing.T) {
	t.Parallel()
	taskRepository := newMysqlTaskRepository(t)
	creator := newExampleEmployer(t)

	archivedTask := newArchivedTask(t, creator)
	err := taskRepository.Add(context.Background(), archivedTask)
	require.NoError(t, err)

	_, err = taskRepository.FindById(context.Background(), archivedTask.UUID())
	assertErrorIsNotExist(t, err)
}

func TestMysqlTaskRepository_UpdateByID_NotFound(t *testing.T) {
	t.Parallel()
	taskRepository := newMysqlTaskRepository(t)

	err := taskRepository.UpdateByID(context.Background(), "non-existent-uuid", func(ctx context.Context, found task.Task) (task.Task, error) {
		return found, nil
	})
	assertErrorIsNotExist(t, err)
}

func TestMysqlTaskRepository_UpdateByID_DeletedTask(t *testing.T) {
	t.Parallel()
	taskRepository := newMysqlTaskRepository(t)
	creator := newExampleEmployer(t)

	deletedTask := newDeletedTask(t, creator)
	err := taskRepository.Add(context.Background(), deletedTask)
	require.NoError(t, err)

	err = taskRepository.UpdateByID(context.Background(), deletedTask.UUID(), func(ctx context.Context, found task.Task) (task.Task, error) {
		return found, nil
	})
	assertErrorIsNotExist(t, err)
}

func TestMysqlTaskRepository_UpdateByID_ArchivedTask(t *testing.T) {
	t.Parallel()
	taskRepository := newMysqlTaskRepository(t)
	creator := newExampleEmployer(t)

	archivedTask := newArchivedTask(t, creator)
	err := taskRepository.Add(context.Background(), archivedTask)
	require.NoError(t, err)

	err = taskRepository.UpdateByID(context.Background(), archivedTask.UUID(), func(ctx context.Context, found task.Task) (task.Task, error) {
		return found, nil
	})
	assertErrorIsNotExist(t, err)
}

func TestMysqlTaskRepository_UpdateByID_UpdateFnError(t *testing.T) {
	t.Parallel()
	taskRepository := newMysqlTaskRepository(t)
	creator := newExampleEmployer(t)

	existingTask := newExampleTask(t, creator)
	err := taskRepository.Add(context.Background(), existingTask)
	require.NoError(t, err)

	expectedErr := errors.Str("update function failed")
	err = taskRepository.UpdateByID(context.Background(), existingTask.UUID(), func(ctx context.Context, found task.Task) (task.Task, error) {
		return nil, expectedErr
	})
	require.Error(t, err)
	require.Contains(t, err.Error(), expectedErr.Error())
}

func TestMysqlTaskRepository_AllTasks_ExcludesDeleted(t *testing.T) {
	t.Parallel()
	taskRepository := newMysqlTaskRepository(t)
	creator := newExampleEmployer(t)

	activeTask := newExampleTask(t, creator)
	err := taskRepository.Add(context.Background(), activeTask)
	require.NoError(t, err)

	deletedTask := newDeletedTask(t, creator)
	err = taskRepository.Add(context.Background(), deletedTask)
	require.NoError(t, err)

	allTasks, err := taskRepository.AllTasks(context.Background())
	require.NoError(t, err)

	assertTaskPresentInQuery(t, allTasks, activeTask)
	assertTaskNotPresentInQuery(t, allTasks, deletedTask)
}

func TestMysqlTaskRepository_AllTasks_ExcludesArchived(t *testing.T) {
	t.Parallel()
	taskRepository := newMysqlTaskRepository(t)
	creator := newExampleEmployer(t)

	activeTask := newExampleTask(t, creator)
	err := taskRepository.Add(context.Background(), activeTask)
	require.NoError(t, err)

	archivedTask := newArchivedTask(t, creator)
	err = taskRepository.Add(context.Background(), archivedTask)
	require.NoError(t, err)

	allTasks, err := taskRepository.AllTasks(context.Background())
	require.NoError(t, err)

	assertTaskPresentInQuery(t, allTasks, activeTask)
	assertTaskNotPresentInQuery(t, allTasks, archivedTask)
}

func TestMysqlTaskRepository_RemoveAllTasks(t *testing.T) {
	taskRepository := newMysqlTaskRepository(t)
	creator := newExampleEmployer(t)

	task1 := newExampleTask(t, creator)
	task1UUID := task1.UUID()
	err := taskRepository.Add(context.Background(), task1)
	require.NoError(t, err)

	task2 := newExampleTask(t, creator)
	task2UUID := task2.UUID()
	err = taskRepository.Add(context.Background(), task2)
	require.NoError(t, err)

	err = taskRepository.RemoveAllTasks(context.Background())
	require.NoError(t, err)

	_, err = taskRepository.FindById(context.Background(), task1UUID)
	assertErrorIsNotExist(t, err)

	_, err = taskRepository.FindById(context.Background(), task2UUID)
	assertErrorIsNotExist(t, err)
}

func TestMysqlTaskRepository_AddTag(t *testing.T) {
	t.Parallel()
	taskRepository := newMysqlTaskRepository(t)
	creator := newExampleEmployer(t)

	existingTask := newExampleTask(t, creator)
	err := taskRepository.Add(context.Background(), existingTask)
	require.NoError(t, err)

	newTag, err := tag.NewTag(existingTask.UUID(), "important")
	require.NoError(t, err)

	err = taskRepository.AddTag(context.Background(), newTag)
	require.NoError(t, err)
}

func TestMysqlTaskRepository_RemoveTag_Success(t *testing.T) {
	t.Parallel()
	taskRepository := newMysqlTaskRepository(t)
	creator := newExampleEmployer(t)

	existingTask := newExampleTask(t, creator)
	err := taskRepository.Add(context.Background(), existingTask)
	require.NoError(t, err)

	newTag, err := tag.NewTag(existingTask.UUID(), "important")
	require.NoError(t, err)

	err = taskRepository.AddTag(context.Background(), newTag)
	require.NoError(t, err)

	err = taskRepository.RemoveTag(context.Background(), newTag.UUID())
	require.NoError(t, err)
}

func TestMysqlTaskRepository_RemoveTag_NotFound(t *testing.T) {
	t.Parallel()
	taskRepository := newMysqlTaskRepository(t)

	err := taskRepository.RemoveTag(context.Background(), "non-existent-tag-id")
	assertErrorIsNotExist(t, err)
}

func newMysqlTaskRepository(t *testing.T) adapters.MysqlTaskRepository {
	t.Helper()
	db, err := adapters.NewMySQLConnection()
	require.NoError(t, err)
	return adapters.NewMysqlTaskRepository(db)
}

func newExampleEmployer(t *testing.T) user.User {
	t.Helper()
	email := fmt.Sprintf("employer-%s@example.com", adapters.NewID().New())
	employer, err := user.NewUser(adapters.NewID().New(), user.RoleEmployer, "Test Employer", email)
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
	err = newTask.ChangeStatus(creator, task.StatusInProgress)
	require.NoError(t, err)
	err = newTask.ChangeStatus(creator, task.StatusCompleted)
	require.NoError(t, err)
	return newTask
}

func assertQueryTasksIncludesExpected(t *testing.T, allTasks []query.Task, expectedTasks []task.Task) {
	t.Helper()

	for _, expected := range expectedTasks {
		found := false
		for _, qt := range allTasks {
			if expected.UUID() == qt.UUID {
				require.Equal(t, expected.Title(), qt.Title)
				require.Equal(t, expected.Status().String(), qt.Status)
				require.Equal(t, expected.CreatedBy(), qt.CreatedBy)
				require.Equal(t, expected.AssignedTo(), qt.AssignedTo)
				compareTimesIgnoringNanoseconds(t, expected.CreatedAt(), qt.CreatedAt)
				compareTimesIgnoringNanoseconds(t, expected.UpdatedAt(), qt.UpdatedAt)
				found = true
				break
			}
		}
		require.True(t, found, "expected task %s not found in query result", expected.UUID())
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
	require.Equal(t, task1.Priority(), task2.Priority())
	require.Equal(t, task1.Description(), task2.Description())
	compareTimesIgnoringNanoseconds(t, task1.CreatedAt(), task2.CreatedAt())
	compareTimesIgnoringNanoseconds(t, task1.UpdatedAt(), task2.UpdatedAt())
}

func compareTimesIgnoringNanoseconds(t *testing.T, t1, t2 time.Time) {
	t.Helper()

	require.Equal(t, t1.Truncate(time.Minute).UTC(), t2.Truncate(time.Minute).UTC())
}

func newDeletedTask(t *testing.T, creator user.User) task.Task {
	t.Helper()
	newTask, err := task.NewTask(creator, adapters.NewID().New(), "deleted task")
	require.NoError(t, err)
	err = newTask.Delete(creator)
	require.NoError(t, err)
	return newTask
}

func newArchivedTask(t *testing.T, creator user.User) task.Task {
	t.Helper()
	newTask, err := task.NewTask(creator, adapters.NewID().New(), "archived task")
	require.NoError(t, err)
	err = newTask.Archive(creator)
	require.NoError(t, err)
	return newTask
}

func assertErrorIsNotExist(t *testing.T, err error) {
	t.Helper()
	require.Error(t, err)
	require.True(t, errors.Is(errkind.NotExist, err), "expected NotExist error, got: %v", err)
}

func assertTaskPresentInQuery(t *testing.T, allTasks []query.Task, target task.Task) {
	t.Helper()
	for _, qt := range allTasks {
		if qt.UUID == target.UUID() {
			return
		}
	}
	t.Fatalf("task %s not found in query result", target.UUID())
}

func assertTaskNotPresentInQuery(t *testing.T, allTasks []query.Task, target task.Task) {
	t.Helper()
	for _, qt := range allTasks {
		if qt.UUID == target.UUID() {
			t.Fatalf("task %s should not be in query result", target.UUID())
		}
	}
}
