package tests_test

import (
	"testing"
)

func TestComponent(t *testing.T) {
	fixtures := SetupComponentTest(t)
	defer fixtures.Cancel()

	t.Run("add_user", func(t *testing.T) {
		testAddUser(t, fixtures)
	})

	t.Run("list_users", func(t *testing.T) {
		testListUsers(t, fixtures)
	})

	t.Run("get_user", func(t *testing.T) {
		testGetUser(t, fixtures)
	})

	t.Run("login", func(t *testing.T) {
		testLogin(t, fixtures)
	})

	t.Run("update_user_profile", func(t *testing.T) {
		testUpdateUserProfile(t, fixtures)
	})

	t.Run("update_user_role", func(t *testing.T) {
		testUpdateUserRole(t, fixtures)
	})

	t.Run("delete_user", func(t *testing.T) {
		testDeleteUser(t, fixtures)
	})

	t.Run("create_task", func(t *testing.T) {
		testCreateTask(t, fixtures)
	})

	t.Run("get_tasks", func(t *testing.T) {
		testGetTasks(t, fixtures)
	})

	t.Run("update_task_title", func(t *testing.T) {
		testUpdateTaskTitle(t, fixtures)
	})

	t.Run("change_task_status", func(t *testing.T) {
		testChangeTaskStatus(t, fixtures)
	})

	t.Run("set_task_priority", func(t *testing.T) {
		testSetTaskPriority(t, fixtures)
	})

	t.Run("set_task_due_date", func(t *testing.T) {
		testSetTaskDueDate(t, fixtures)
	})

	t.Run("set_task_description", func(t *testing.T) {
		testSetTaskDescription(t, fixtures)
	})

	t.Run("add_task_tag", func(t *testing.T) {
		testAddTaskTag(t, fixtures)
	})

	t.Run("remove_task_tag", func(t *testing.T) {
		testRemoveTaskTag(t, fixtures)
	})

	t.Run("assign_task", func(t *testing.T) {
		testAssignTask(t, fixtures)
	})

	t.Run("unassign_task", func(t *testing.T) {
		testUnassignTask(t, fixtures)
	})

	t.Run("reopen_task", func(t *testing.T) {
		testReopenTask(t, fixtures)
	})

	t.Run("archive_task", func(t *testing.T) {
		testArchiveTask(t, fixtures)
	})

	t.Run("delete_task", func(t *testing.T) {
		testDeleteTask(t, fixtures)
	})

	t.Run("add_comment", func(t *testing.T) {
		testAddComment(t, fixtures)
	})

	t.Run("update_comment", func(t *testing.T) {
		testUpdateComment(t, fixtures)
	})

	t.Run("delete_comment", func(t *testing.T) {
		testDeleteComment(t, fixtures)
	})

	t.Run("get_task_activity", func(t *testing.T) {
		testGetTaskActivity(t, fixtures)
	})
}
