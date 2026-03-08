package task_test

import (
	"clean-arch-go/domain/task"
	"clean-arch-go/domain/user"
	"database/sql"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func newEmployer(t *testing.T, id string) user.User {
	t.Helper()
	u, err := user.NewUser(id, user.RoleEmployer, "Test Employer", id+"@example.com")
	require.NoError(t, err)
	return u
}

func newEmployee(t *testing.T, id string) user.User {
	t.Helper()
	u, err := user.NewUser(id, user.RoleEmployee, "Test Employee", id+"@example.com")
	require.NoError(t, err)
	return u
}

func newAdmin(t *testing.T, id string) user.User {
	t.Helper()
	u, err := user.NewUser(id, user.RoleAdmin, "Admin User", id+"@example.com")
	require.NoError(t, err)
	return u
}

func newTask(t *testing.T, creator user.User) task.Task {
	t.Helper()
	tk, err := task.NewTask(creator, "task-uuid", "Initial Title")
	require.NoError(t, err)
	return tk
}

func newTaskWithAssignee(t *testing.T, creator, assignee user.User) task.Task {
	t.Helper()
	tk := newTask(t, creator)
	err := tk.AssignTo(creator, assignee)
	require.NoError(t, err)
	return tk
}

func completeTask(t *testing.T, tk task.Task, u user.User) {
	t.Helper()
	err := tk.ChangeStatus(u, task.StatusPending)
	require.NoError(t, err)
	err = tk.ChangeStatus(u, task.StatusInProgress)
	require.NoError(t, err)
	err = tk.ChangeStatus(u, task.StatusCompleted)
	require.NoError(t, err)
}

func TestNewTask(t *testing.T) {
	t.Parallel()
	creator := newEmployer(t, "123")

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
	creator := newEmployee(t, "123")

	tk, err := task.NewTask(creator, "task-uuid", "Initial Title")
	require.Error(t, err)
	require.Nil(t, tk)
	require.EqualError(t, err, "only employer or admin can create task")
}

func TestUpdateTitle(t *testing.T) {
	t.Parallel()
	creator := newEmployer(t, "123")
	tk := newTask(t, creator)

	err := tk.UpdateTitle(creator, "Updated Title")
	require.NoError(t, err)
	require.Equal(t, "Updated Title", tk.Title())
}

func TestUpdateTitle_InvalidUser(t *testing.T) {
	t.Parallel()
	creator := newEmployer(t, "123")
	otherUser := newEmployee(t, "456")
	tk := newTask(t, creator)

	err := tk.UpdateTitle(otherUser, "Updated Title")
	require.Error(t, err)
	require.EqualError(t, err, "user is not allowed to update this task title")
}

func TestChangeStatus(t *testing.T) {
	t.Parallel()
	creator := newEmployer(t, "123")
	tk := newTask(t, creator)

	err := tk.ChangeStatus(creator, task.StatusInProgress)
	require.NoError(t, err)
	require.Equal(t, task.StatusInProgress, tk.Status())
}

func TestChangeStatus_InvalidUser(t *testing.T) {
	t.Parallel()
	creator := newEmployer(t, "123")
	otherUser := newEmployee(t, "456")
	tk := newTask(t, creator)

	err := tk.ChangeStatus(otherUser, task.StatusInProgress)
	require.Error(t, err)
	require.EqualError(t, err, "user is not allowed to update status of this task")
}

func TestAssignTo(t *testing.T) {
	t.Parallel()
	assigner := newEmployer(t, "123")
	assignee := newEmployer(t, "456")
	tk := newTask(t, assigner)

	err := tk.AssignTo(assigner, assignee)
	require.NoError(t, err)
	require.Equal(t, assignee.UUID(), tk.AssignedTo())
}

func TestAssignTo_InvalidAssignee(t *testing.T) {
	t.Parallel()
	assigner := newEmployer(t, "123")
	assignee := newEmployee(t, "456")
	tk := newTask(t, assigner)

	err := tk.AssignTo(assigner, assignee)
	require.Error(t, err)
	require.EqualError(t, err, "only employer role can assign task")
}

func TestChangeStatus_ValidTransitions(t *testing.T) {
	t.Parallel()
	creator := newEmployer(t, "123")
	tk := newTask(t, creator)

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
	creator := newEmployer(t, "123")
	tk := newTask(t, creator)

	err := tk.ChangeStatus(creator, task.StatusCompleted)
	require.Error(t, err)
	require.Contains(t, err.Error(), "cannot transition")
}

func TestReopen(t *testing.T) {
	t.Parallel()
	creator := newEmployer(t, "123")
	tk := newTask(t, creator)
	completeTask(t, tk, creator)

	err := tk.Reopen(creator)
	require.NoError(t, err)
	require.Equal(t, task.StatusInProgress, tk.Status())
}

func TestReopen_NotCompleted(t *testing.T) {
	t.Parallel()
	creator := newEmployer(t, "123")
	tk := newTask(t, creator)

	err := tk.Reopen(creator)
	require.Error(t, err)
	require.EqualError(t, err, "only completed tasks can be reopened")
}

func TestReopen_InvalidUser(t *testing.T) {
	t.Parallel()
	creator := newEmployer(t, "123")
	otherUser := newEmployee(t, "456")
	tk := newTask(t, creator)
	completeTask(t, tk, creator)

	err := tk.Reopen(otherUser)
	require.Error(t, err)
	require.EqualError(t, err, "user is not allowed to reopen this task")
}

func TestSetPriority(t *testing.T) {
	t.Parallel()
	creator := newEmployer(t, "123")
	tk := newTask(t, creator)

	err := tk.SetPriority(creator, task.PriorityHigh)
	require.NoError(t, err)
	require.Equal(t, task.PriorityHigh, tk.Priority())
}

func TestSetPriority_InvalidUser(t *testing.T) {
	t.Parallel()
	creator := newEmployer(t, "123")
	employee := newEmployee(t, "456")
	tk := newTask(t, creator)

	err := tk.SetPriority(employee, task.PriorityHigh)
	require.Error(t, err)
	require.EqualError(t, err, "user is not allowed to update this task")
}

func TestSetDueDate(t *testing.T) {
	t.Parallel()
	creator := newEmployer(t, "123")
	tk := newTask(t, creator)
	due := time.Date(2026, 3, 1, 10, 0, 0, 0, time.UTC)

	err := tk.SetDueDate(creator, due)
	require.NoError(t, err)
	require.Equal(t, due, tk.DueDate())
}

func TestSetDueDate_InvalidUser(t *testing.T) {
	t.Parallel()
	creator := newEmployer(t, "123")
	employee := newEmployee(t, "456")
	tk := newTask(t, creator)
	due := time.Date(2026, 3, 1, 10, 0, 0, 0, time.UTC)

	err := tk.SetDueDate(employee, due)
	require.Error(t, err)
	require.EqualError(t, err, "user is not allowed to update this task")
}

func TestSetDescription(t *testing.T) {
	t.Parallel()
	creator := newEmployer(t, "123")
	tk := newTask(t, creator)

	err := tk.SetDescription(creator, "This is a description")
	require.NoError(t, err)
	require.Equal(t, "This is a description", tk.Description())
}

func TestSetDescription_InvalidUser(t *testing.T) {
	t.Parallel()
	creator := newEmployer(t, "123")
	employee := newEmployee(t, "456")
	tk := newTask(t, creator)

	err := tk.SetDescription(employee, "desc")
	require.Error(t, err)
	require.EqualError(t, err, "user is not allowed to update this task")
}

func TestSetDescription_AssignedEmployeeAllowed(t *testing.T) {
	t.Parallel()
	creator := newEmployer(t, "123")
	assignee := newEmployer(t, "456")
	tk := newTaskWithAssignee(t, creator, assignee)

	err := tk.SetDescription(assignee, "desc")
	require.NoError(t, err)
}

func TestNewTask_DefaultPriority(t *testing.T) {
	t.Parallel()
	creator := newEmployer(t, "123")
	tk := newTask(t, creator)

	require.Equal(t, task.PriorityMedium, tk.Priority())
}

func TestNewTask_AdminCanCreate(t *testing.T) {
	t.Parallel()
	admin := newAdmin(t, "123")
	tk, err := task.NewTask(admin, "task-uuid", "Initial Title")

	require.NoError(t, err)
	require.NotNil(t, tk)
	require.Equal(t, admin.UUID(), tk.CreatedBy())
}

func TestNewTask_EmptyUUID(t *testing.T) {
	t.Parallel()
	creator := newEmployer(t, "123")
	tk, err := task.NewTask(creator, "", "Initial Title")

	require.Error(t, err)
	require.Nil(t, tk)
	require.EqualError(t, err, "empty task uuid")
}

func TestNewTask_EmptyTitle(t *testing.T) {
	t.Parallel()
	creator := newEmployer(t, "123")
	tk, err := task.NewTask(creator, "task-uuid", "")

	require.Error(t, err)
	require.Nil(t, tk)
	require.EqualError(t, err, "empty task title")
}

func TestUpdateTitle_AssignedUserCanUpdate(t *testing.T) {
	t.Parallel()
	creator := newEmployer(t, "123")
	assignee := newEmployer(t, "456")
	tk := newTaskWithAssignee(t, creator, assignee)

	err := tk.UpdateTitle(assignee, "Updated Title")

	require.NoError(t, err)
	require.Equal(t, "Updated Title", tk.Title())
}

func TestUpdateTitle_EmptyTitle(t *testing.T) {
	t.Parallel()
	creator := newEmployer(t, "123")
	tk := newTask(t, creator)

	err := tk.UpdateTitle(creator, "")

	require.Error(t, err)
	require.EqualError(t, err, "Cannot update task title to empty")
}

func TestChangeStatus_AssignedUserCanChange(t *testing.T) {
	t.Parallel()
	creator := newEmployer(t, "123")
	assignee := newEmployer(t, "456")
	tk := newTaskWithAssignee(t, creator, assignee)

	err := tk.ChangeStatus(assignee, task.StatusPending)

	require.NoError(t, err)
	require.Equal(t, task.StatusPending, tk.Status())
}

func TestChangeStatus_AdminCanChange(t *testing.T) {
	t.Parallel()
	creator := newEmployer(t, "123")
	admin := newAdmin(t, "456")
	tk := newTask(t, creator)

	err := tk.ChangeStatus(admin, task.StatusPending)

	require.NoError(t, err)
	require.Equal(t, task.StatusPending, tk.Status())
}

func TestChangeStatus_ZeroStatus(t *testing.T) {
	t.Parallel()
	creator := newEmployer(t, "123")
	tk := newTask(t, creator)

	err := tk.ChangeStatus(creator, task.StatusUnknown)

	require.Error(t, err)
	require.EqualError(t, err, "cannot update status of task to empty")
}

func TestAssignTo_Reassign(t *testing.T) {
	t.Parallel()
	creator := newEmployer(t, "123")
	assignee1 := newEmployer(t, "456")
	assignee2 := newEmployer(t, "789")
	tk := newTaskWithAssignee(t, creator, assignee1)

	err := tk.AssignTo(creator, assignee2)

	require.NoError(t, err)
	require.Equal(t, assignee2.UUID(), tk.AssignedTo())
}

func TestAssignTo_AdminAsAssigner(t *testing.T) {
	t.Parallel()
	admin := newAdmin(t, "123")
	assignee := newEmployer(t, "456")
	tk := newTask(t, admin)

	err := tk.AssignTo(admin, assignee)

	require.NoError(t, err)
	require.Equal(t, assignee.UUID(), tk.AssignedTo())
}

func TestReopen_AssignedUserCanReopen(t *testing.T) {
	t.Parallel()
	creator := newEmployer(t, "123")
	assignee := newEmployer(t, "456")
	tk := newTaskWithAssignee(t, creator, assignee)
	completeTask(t, tk, assignee)

	err := tk.Reopen(assignee)

	require.NoError(t, err)
	require.Equal(t, task.StatusInProgress, tk.Status())
}

func TestUnassign(t *testing.T) {
	t.Parallel()
	creator := newEmployer(t, "123")
	assignee := newEmployer(t, "456")
	tk := newTaskWithAssignee(t, creator, assignee)

	err := tk.Unassign(creator)

	require.NoError(t, err)
	require.Equal(t, "", tk.AssignedTo())
}

func TestUnassign_AdminCanUnassign(t *testing.T) {
	t.Parallel()
	creator := newEmployer(t, "123")
	admin := newAdmin(t, "456")
	assignee := newEmployer(t, "789")
	tk := newTaskWithAssignee(t, creator, assignee)

	err := tk.Unassign(admin)

	require.NoError(t, err)
	require.Equal(t, "", tk.AssignedTo())
}

func TestUnassign_EmployeeCannotUnassign(t *testing.T) {
	t.Parallel()
	creator := newEmployer(t, "123")
	employee := newEmployee(t, "456")
	assignee := newEmployer(t, "789")
	tk := newTaskWithAssignee(t, creator, assignee)

	err := tk.Unassign(employee)

	require.Error(t, err)
	require.EqualError(t, err, "only employer can unassign task")
}

func TestDelete(t *testing.T) {
	t.Parallel()
	creator := newEmployer(t, "123")
	tk := newTask(t, creator)

	err := tk.Delete(creator)

	require.NoError(t, err)
	require.True(t, tk.IsDeleted())
	require.False(t, tk.DeletedAt().IsZero())
}

func TestDelete_AdminCanDelete(t *testing.T) {
	t.Parallel()
	creator := newEmployer(t, "123")
	admin := newAdmin(t, "456")
	tk := newTask(t, creator)

	err := tk.Delete(admin)

	require.NoError(t, err)
	require.True(t, tk.IsDeleted())
}

func TestDelete_NonCreatorCannotDelete(t *testing.T) {
	t.Parallel()
	creator := newEmployer(t, "123")
	otherEmployer := newEmployer(t, "456")
	tk := newTask(t, creator)

	err := tk.Delete(otherEmployer)

	require.Error(t, err)
	require.EqualError(t, err, "only task creator can delete task")
	require.False(t, tk.IsDeleted())
}

func TestDelete_EmployeeCannotDelete(t *testing.T) {
	t.Parallel()
	creator := newEmployer(t, "123")
	employee := newEmployee(t, "456")
	tk := newTask(t, creator)

	err := tk.Delete(employee)

	require.Error(t, err)
	require.EqualError(t, err, "only employer can delete task")
	require.False(t, tk.IsDeleted())
}

func TestArchive(t *testing.T) {
	t.Parallel()
	creator := newEmployer(t, "123")
	tk := newTask(t, creator)

	err := tk.Archive(creator)

	require.NoError(t, err)
	require.True(t, tk.IsArchived())
	require.False(t, tk.ArchivedAt().IsZero())
}

func TestArchive_AdminCanArchive(t *testing.T) {
	t.Parallel()
	creator := newEmployer(t, "123")
	admin := newAdmin(t, "456")
	tk := newTask(t, creator)

	err := tk.Archive(admin)

	require.NoError(t, err)
	require.True(t, tk.IsArchived())
}

func TestArchive_EmployeeCannotArchive(t *testing.T) {
	t.Parallel()
	creator := newEmployer(t, "123")
	employee := newEmployee(t, "456")
	tk := newTask(t, creator)

	err := tk.Archive(employee)

	require.Error(t, err)
	require.EqualError(t, err, "only employer can archive task")
	require.False(t, tk.IsArchived())
}

func TestFrom(t *testing.T) {
	t.Parallel()

	createdAt := time.Date(2026, 1, 1, 10, 0, 0, 0, time.UTC)
	updatedAt := time.Date(2026, 1, 2, 10, 0, 0, 0, time.UTC)
	deletedAt := sql.NullTime{Time: time.Date(2026, 1, 3, 10, 0, 0, 0, time.UTC), Valid: true}
	archivedAt := sql.NullTime{Time: time.Date(2026, 1, 4, 10, 0, 0, 0, time.UTC), Valid: true}
	dueDate := sql.NullTime{Time: time.Date(2026, 2, 1, 10, 0, 0, 0, time.UTC), Valid: true}
	description := sql.NullString{String: "Test description", Valid: true}

	tk, err := task.From(
		"task-uuid",
		"Task Title",
		"in_progress",
		"creator-uuid",
		"assignee-uuid",
		createdAt,
		updatedAt,
		deletedAt,
		archivedAt,
		int(task.PriorityHigh),
		dueDate,
		description,
	)

	require.NoError(t, err)
	require.Equal(t, "task-uuid", tk.UUID())
	require.Equal(t, "Task Title", tk.Title())
	require.Equal(t, task.StatusInProgress, tk.Status())
	require.Equal(t, "creator-uuid", tk.CreatedBy())
	require.Equal(t, "assignee-uuid", tk.AssignedTo())
	require.Equal(t, createdAt, tk.CreatedAt())
	require.Equal(t, updatedAt, tk.UpdatedAt())
	require.True(t, tk.IsDeleted())
	require.Equal(t, deletedAt.Time, tk.DeletedAt())
	require.True(t, tk.IsArchived())
	require.Equal(t, archivedAt.Time, tk.ArchivedAt())
	require.Equal(t, task.PriorityHigh, tk.Priority())
	require.Equal(t, dueDate.Time, tk.DueDate())
	require.Equal(t, "Test description", tk.Description())
}

func TestFrom_InvalidStatus(t *testing.T) {
	t.Parallel()

	tk, err := task.From(
		"task-uuid",
		"Task Title",
		"invalid_status",
		"creator-uuid",
		"assignee-uuid",
		time.Now(),
		time.Now(),
		sql.NullTime{},
		sql.NullTime{},
		int(task.PriorityMedium),
		sql.NullTime{},
		sql.NullString{},
	)

	require.Error(t, err)
	require.Nil(t, tk)
}

func TestSetPriority_AssignedUserCanSet(t *testing.T) {
	t.Parallel()
	creator := newEmployer(t, "123")
	assignee := newEmployer(t, "456")
	tk := newTaskWithAssignee(t, creator, assignee)

	err := tk.SetPriority(assignee, task.PriorityHigh)

	require.NoError(t, err)
	require.Equal(t, task.PriorityHigh, tk.Priority())
}

func TestSetDueDate_AssignedUserCanSet(t *testing.T) {
	t.Parallel()
	creator := newEmployer(t, "123")
	assignee := newEmployer(t, "456")
	tk := newTaskWithAssignee(t, creator, assignee)
	due := time.Date(2026, 3, 1, 10, 0, 0, 0, time.UTC)

	err := tk.SetDueDate(assignee, due)

	require.NoError(t, err)
	require.Equal(t, due, tk.DueDate())
}

func TestIsDeleted_NotDeletedByDefault(t *testing.T) {
	t.Parallel()
	creator := newEmployer(t, "123")
	tk := newTask(t, creator)

	require.False(t, tk.IsDeleted())
	require.True(t, tk.DeletedAt().IsZero())
}

func TestIsArchived_NotArchivedByDefault(t *testing.T) {
	t.Parallel()
	creator := newEmployer(t, "123")
	tk := newTask(t, creator)

	require.False(t, tk.IsArchived())
	require.True(t, tk.ArchivedAt().IsZero())
}
