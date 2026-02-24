package task_test

import (
	"clean-arch-go/domain/task"
	"clean-arch-go/domain/user"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestNewTask(t *testing.T) {
	t.Parallel()

	creator, err := user.NewUser("123", user.RoleEmployer, "Test User 1", "test1@example.com")
	require.NoError(t, err)

	tk, err := task.NewTask(creator, "task-uuid", "Initial Title")
	require.NoError(t, err)
	require.NotNil(t, tk)
	require.Equal(t, "task-uuid", tk.UUID())
	require.Equal(t, "Initial Title", tk.Title())
	require.Equal(t, creator.UUID(), tk.CreatedBy())
	require.Equal(t, task.StatusTodo, tk.Status())
}

func TestNewTask_InvalidRole(t *testing.T) {
	t.Parallel()

	creator, err := user.NewUser("123", user.RoleEmployee, "Test User 1", "test1@example.com")
	require.NoError(t, err)

	tk, err := task.NewTask(creator, "task-uuid", "Initial Title")
	require.Error(t, err)
	require.Nil(t, tk)
	require.EqualError(t, err, "only employ can create task")
}

func TestUpdateTitle(t *testing.T) {
	t.Parallel()

	creator, _ := user.NewUser("123", user.RoleEmployer, "Test User 1", "test1@example.com")
	tk, _ := task.NewTask(creator, "task-uuid", "Initial Title")

	err := tk.UpdateTitle(creator, "Updated Title")
	require.NoError(t, err)
	require.Equal(t, "Updated Title", tk.Title())
}

func TestUpdateTitle_InvalidUser(t *testing.T) {
	t.Parallel()

	creator, _ := user.NewUser("123", user.RoleEmployer, "Test User 1", "test1@example.com")
	otherUser, _ := user.NewUser("456", user.RoleEmployee, "Test User 2", "test2@example.com")
	tk, _ := task.NewTask(creator, "task-uuid", "Initial Title")

	err := tk.UpdateTitle(otherUser, "Updated Title")
	require.Error(t, err)
	require.EqualError(t, err, "user is not allow to update this task title")
}

func TestChangeStatus(t *testing.T) {
	t.Parallel()

	creator, _ := user.NewUser("123", user.RoleEmployer, "Test User 1", "test1@example.com")
	tk, _ := task.NewTask(creator, "task-uuid", "Initial Title")

	err := tk.ChangeStatus(creator, task.StatusInProgress)
	require.NoError(t, err)
	require.Equal(t, task.StatusInProgress, tk.Status())
}

func TestChangeStatus_InvalidUser(t *testing.T) {
	t.Parallel()

	creator, _ := user.NewUser("123", user.RoleEmployer, "Test User 1", "test1@example.com")
	otherUser, _ := user.NewUser("456", user.RoleEmployee, "Test User 2", "test2@example.com")
	tk, _ := task.NewTask(creator, "task-uuid", "Initial Title")

	err := tk.ChangeStatus(otherUser, task.StatusInProgress)
	require.Error(t, err)
	require.EqualError(t, err, "user is not allow to update status of this task")
}

func TestAssignTo(t *testing.T) {
	t.Parallel()

	assigner, _ := user.NewUser("123", user.RoleEmployer, "Test User 1", "test1@example.com")
	assignee, _ := user.NewUser("456", user.RoleEmployer, "Test User 2", "test2@example.com")
	tk, _ := task.NewTask(assigner, "task-uuid", "Initial Title")

	err := tk.AssignTo(assigner, assignee)
	require.NoError(t, err)
	require.Equal(t, assignee.UUID(), tk.AssignedTo())
}

func TestAssignTo_InvalidAssignee(t *testing.T) {
	t.Parallel()

	assigner, _ := user.NewUser("123", user.RoleEmployer, "Test User 1", "test1@example.com")
	assignee, _ := user.NewUser("456", user.RoleEmployee, "Test User 2", "test2@example.com")
	tk, _ := task.NewTask(assigner, "task-uuid", "Initial Title")

	err := tk.AssignTo(assigner, assignee)
	require.Error(t, err)
	require.EqualError(t, err, "only employer role can assign task")
}

func TestChangeStatus_ValidTransitions(t *testing.T) {
	t.Parallel()

	creator, _ := user.NewUser("123", user.RoleEmployer, "Test User 1", "test1@example.com")
	tk, _ := task.NewTask(creator, "task-uuid", "Initial Title")

	err := tk.ChangeStatus(creator, task.StatusPending)
	require.NoError(t, err)
	require.Equal(t, task.StatusPending, tk.Status())

	err = tk.ChangeStatus(creator, task.StatusInProgress)
	require.NoError(t, err)
	require.Equal(t, task.StatusInProgress, tk.Status())

	err = tk.ChangeStatus(creator, task.StatusCompleted)
	require.NoError(t, err)
	require.Equal(t, task.StatusCompleted, tk.Status())
}

func TestChangeStatus_InvalidTransition(t *testing.T) {
	t.Parallel()

	creator, _ := user.NewUser("123", user.RoleEmployer, "Test User 1", "test1@example.com")
	tk, _ := task.NewTask(creator, "task-uuid", "Initial Title")

	err := tk.ChangeStatus(creator, task.StatusCompleted)
	require.Error(t, err)
	require.Contains(t, err.Error(), "cannot transition")
}

func TestReopen(t *testing.T) {
	t.Parallel()

	creator, _ := user.NewUser("123", user.RoleEmployer, "Test User 1", "test1@example.com")
	tk, _ := task.NewTask(creator, "task-uuid", "Initial Title")

	tk.ChangeStatus(creator, task.StatusPending)
	tk.ChangeStatus(creator, task.StatusInProgress)
	tk.ChangeStatus(creator, task.StatusCompleted)
	require.Equal(t, task.StatusCompleted, tk.Status())

	err := tk.Reopen(creator)
	require.NoError(t, err)
	require.Equal(t, task.StatusInProgress, tk.Status())
}

func TestReopen_NotCompleted(t *testing.T) {
	t.Parallel()

	creator, _ := user.NewUser("123", user.RoleEmployer, "Test User 1", "test1@example.com")
	tk, _ := task.NewTask(creator, "task-uuid", "Initial Title")

	err := tk.Reopen(creator)
	require.Error(t, err)
	require.EqualError(t, err, "only completed tasks can be reopened")
}

func TestReopen_InvalidUser(t *testing.T) {
	t.Parallel()

	creator, _ := user.NewUser("123", user.RoleEmployer, "Test User 1", "test1@example.com")
	otherUser, _ := user.NewUser("456", user.RoleEmployee, "Test User 2", "test2@example.com")
	tk, _ := task.NewTask(creator, "task-uuid", "Initial Title")

	tk.ChangeStatus(creator, task.StatusPending)
	tk.ChangeStatus(creator, task.StatusInProgress)
	tk.ChangeStatus(creator, task.StatusCompleted)

	err := tk.Reopen(otherUser)
	require.Error(t, err)
	require.EqualError(t, err, "user is not allowed to reopen this task")
}

func TestSetPriority(t *testing.T) {
	t.Parallel()

	creator, _ := user.NewUser("123", user.RoleEmployer, "Test User 1", "test1@example.com")
	tk, _ := task.NewTask(creator, "task-uuid", "Initial Title")

	err := tk.SetPriority(creator, task.PriorityHigh)
	require.NoError(t, err)
	require.Equal(t, task.PriorityHigh, tk.Priority())
}

func TestSetPriority_InvalidUser(t *testing.T) {
	t.Parallel()

	creator, _ := user.NewUser("123", user.RoleEmployer, "Test User 1", "test1@example.com")
	employee, _ := user.NewUser("456", user.RoleEmployee, "Test User 2", "test2@example.com")
	tk, _ := task.NewTask(creator, "task-uuid", "Initial Title")

	err := tk.SetPriority(employee, task.PriorityHigh)
	require.Error(t, err)
	require.EqualError(t, err, "user is not allowed to update this task")
}

func TestSetDueDate(t *testing.T) {
	t.Parallel()

	creator, _ := user.NewUser("123", user.RoleEmployer, "Test User 1", "test1@example.com")
	tk, _ := task.NewTask(creator, "task-uuid", "Initial Title")
	due := time.Date(2026, 3, 1, 10, 0, 0, 0, time.UTC)

	err := tk.SetDueDate(creator, due)
	require.NoError(t, err)
	require.Equal(t, due, tk.DueDate())
}

func TestSetDueDate_InvalidUser(t *testing.T) {
	t.Parallel()

	creator, _ := user.NewUser("123", user.RoleEmployer, "Test User 1", "test1@example.com")
	employee, _ := user.NewUser("456", user.RoleEmployee, "Test User 2", "test2@example.com")
	tk, _ := task.NewTask(creator, "task-uuid", "Initial Title")
	due := time.Date(2026, 3, 1, 10, 0, 0, 0, time.UTC)

	err := tk.SetDueDate(employee, due)
	require.Error(t, err)
	require.EqualError(t, err, "user is not allowed to update this task")
}

func TestSetDescription(t *testing.T) {
	t.Parallel()

	creator, _ := user.NewUser("123", user.RoleEmployer, "Test User 1", "test1@example.com")
	tk, _ := task.NewTask(creator, "task-uuid", "Initial Title")

	err := tk.SetDescription(creator, "This is a description")
	require.NoError(t, err)
	require.Equal(t, "This is a description", tk.Description())
}

func TestSetDescription_InvalidUser(t *testing.T) {
	t.Parallel()

	creator, _ := user.NewUser("123", user.RoleEmployer, "Test User 1", "test1@example.com")
	employee, _ := user.NewUser("456", user.RoleEmployee, "Test User 2", "test2@example.com")
	tk, _ := task.NewTask(creator, "task-uuid", "Initial Title")

	err := tk.SetDescription(employee, "desc")
	require.Error(t, err)
	require.EqualError(t, err, "user is not allowed to update this task")
}

func TestSetDescription_AssignedEmployeeAllowed(t *testing.T) {
	t.Parallel()

	creator, _ := user.NewUser("123", user.RoleEmployer, "Test User 1", "test1@example.com")
	assignee, _ := user.NewUser("456", user.RoleEmployer, "Test User 2", "test2@example.com")
	tk, _ := task.NewTask(creator, "task-uuid", "Initial Title")
	tk.AssignTo(creator, assignee)

	err := tk.SetDescription(assignee, "desc")
	require.NoError(t, err)
}

func TestNewTask_DefaultPriority(t *testing.T) {
	t.Parallel()

	creator, _ := user.NewUser("123", user.RoleEmployer, "Test User 1", "test1@example.com")
	tk, _ := task.NewTask(creator, "task-uuid", "Initial Title")

	require.Equal(t, task.PriorityMedium, tk.Priority())
}
