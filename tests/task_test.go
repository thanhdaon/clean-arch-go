package tests_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func testCreateTask(t *testing.T, f *TestFixtures) {
	t.Helper()

	creatorID := createUserAndGetID(t, f.DB)

	title := fmt.Sprintf("Create-Task-%d", time.Now().UnixNano())
	resp, body := createTask(t, f.AuthToken, map[string]any{
		"title":   title,
		"creator": creatorID,
	})
	assert.Equal(t, http.StatusOK, resp.StatusCode, "unexpected status: %s", string(body))

	var taskID string
	err := f.DB.Get(&taskID, "SELECT id FROM tasks WHERE title = ? AND deleted_at IS NULL", title)
	require.NoError(t, err, "task should be stored in DB")
	assert.NotEmpty(t, taskID)
}

func testGetTasks(t *testing.T, f *TestFixtures) {
	t.Helper()

	creatorID := createUserAndGetID(t, f.DB)
	title := fmt.Sprintf("GetTasks-Task-%d", time.Now().UnixNano())

	resp, body := createTask(t, f.AuthToken, map[string]any{
		"title":   title,
		"creator": creatorID,
	})
	require.Equal(t, http.StatusOK, resp.StatusCode, "setup failed: %s", string(body))

	resp, body = getTasks(t, f.AuthToken)
	assert.Equal(t, http.StatusOK, resp.StatusCode, "unexpected status: %s", string(body))

	var result map[string]any
	err := json.Unmarshal(body, &result)
	require.NoError(t, err)

	tasks, ok := result["tasks"]
	require.True(t, ok, "response should contain 'tasks' key; got: %s", string(body))

	taskList, ok := tasks.([]any)
	require.True(t, ok, "tasks should be a JSON array")

	found := false
	for _, item := range taskList {
		taskMap, ok := item.(map[string]any)
		if !ok {
			continue
		}
		if taskMap["title"] == title {
			found = true
			break
		}
	}
	assert.True(t, found, "created task with title %q should appear in list", title)
}

func testUpdateTaskTitle(t *testing.T, f *TestFixtures) {
	t.Helper()

	creatorID := createUserAndGetID(t, f.DB)
	taskID := createTaskAndGetID(t, f.DB, f.AuthToken, creatorID)

	newTitle := fmt.Sprintf("Updated-Title-%d", time.Now().UnixNano())
	resp, body := patchTaskTitle(t, f.AuthToken, taskID, map[string]any{
		"title": newTitle,
	})
	assert.Equal(t, http.StatusOK, resp.StatusCode, "unexpected status: %s", string(body))

	var storedTitle string
	err := f.DB.Get(&storedTitle, "SELECT title FROM tasks WHERE id = ?", taskID)
	require.NoError(t, err)
	assert.Equal(t, newTitle, storedTitle)
}

func testChangeTaskStatus(t *testing.T, f *TestFixtures) {
	t.Helper()

	creatorID := createUserAndGetID(t, f.DB)
	taskID := createTaskAndGetID(t, f.DB, f.AuthToken, creatorID)

	resp, body := putTaskStatus(t, f.AuthToken, taskID, map[string]any{
		"status": "in_progress",
	})
	assert.Equal(t, http.StatusOK, resp.StatusCode, "unexpected status: %s", string(body))

	var status string
	err := f.DB.Get(&status, "SELECT status FROM tasks WHERE id = ?", taskID)
	require.NoError(t, err)
	assert.Equal(t, "in_progress", status)
}

func testSetTaskPriority(t *testing.T, f *TestFixtures) {
	t.Helper()

	creatorID := createUserAndGetID(t, f.DB)
	taskID := createTaskAndGetID(t, f.DB, f.AuthToken, creatorID)

	resp, body := putTaskPriority(t, f.AuthToken, taskID, map[string]any{
		"priority": "high",
	})
	assert.Equal(t, http.StatusOK, resp.StatusCode, "unexpected status: %s", string(body))

	var priority int
	err := f.DB.Get(&priority, "SELECT priority FROM tasks WHERE id = ?", taskID)
	require.NoError(t, err)
	assert.Greater(t, priority, 0, "priority should be set to a non-zero value")
}

func testSetTaskDueDate(t *testing.T, f *TestFixtures) {
	t.Helper()

	creatorID := createUserAndGetID(t, f.DB)
	taskID := createTaskAndGetID(t, f.DB, f.AuthToken, creatorID)

	resp, body := putTaskDueDate(t, f.AuthToken, taskID, map[string]any{
		"due_date": "2027-03-01T10:00:00Z",
	})
	assert.Equal(t, http.StatusOK, resp.StatusCode, "unexpected status: %s", string(body))

	var count int
	err := f.DB.Get(&count, "SELECT COUNT(*) FROM tasks WHERE id = ? AND due_date IS NOT NULL", taskID)
	require.NoError(t, err)
	assert.Equal(t, 1, count, "task due_date should be set in DB")
}

func testSetTaskDescription(t *testing.T, f *TestFixtures) {
	t.Helper()

	creatorID := createUserAndGetID(t, f.DB)
	taskID := createTaskAndGetID(t, f.DB, f.AuthToken, creatorID)

	resp, body := patchTaskDescription(t, f.AuthToken, taskID, map[string]any{
		"description": "This is a test description",
	})
	assert.Equal(t, http.StatusOK, resp.StatusCode, "unexpected status: %s", string(body))

	var description string
	err := f.DB.Get(&description, "SELECT description FROM tasks WHERE id = ?", taskID)
	require.NoError(t, err)
	assert.Equal(t, "This is a test description", description)
}

func testAddTaskTag(t *testing.T, f *TestFixtures) {
	t.Helper()

	creatorID := createUserAndGetID(t, f.DB)
	taskID := createTaskAndGetID(t, f.DB, f.AuthToken, creatorID)

	resp, body := addTaskTag(t, f.AuthToken, taskID, map[string]any{
		"name": "bug",
	})
	assert.Equal(t, http.StatusOK, resp.StatusCode, "unexpected status: %s", string(body))

	var count int
	err := f.DB.Get(&count, "SELECT COUNT(*) FROM tags WHERE task_id = ? AND name = ?", taskID, "bug")
	require.NoError(t, err)
	assert.Equal(t, 1, count, "tag 'bug' should exist for task %s", taskID)
}

func testRemoveTaskTag(t *testing.T, f *TestFixtures) {
	t.Helper()

	creatorID := createUserAndGetID(t, f.DB)
	taskID := createTaskAndGetID(t, f.DB, f.AuthToken, creatorID)

	resp, body := addTaskTag(t, f.AuthToken, taskID, map[string]any{
		"name": "to-remove",
	})
	require.Equal(t, http.StatusOK, resp.StatusCode, "setup failed adding tag: %s", string(body))

	var tagID string
	err := f.DB.Get(&tagID, "SELECT id FROM tags WHERE task_id = ? AND name = ?", taskID, "to-remove")
	require.NoError(t, err)

	resp, body = removeTaskTag(t, f.AuthToken, taskID, tagID)
	assert.Equal(t, http.StatusOK, resp.StatusCode, "unexpected status: %s", string(body))

	var count int
	err = f.DB.Get(&count, "SELECT COUNT(*) FROM tags WHERE id = ?", tagID)
	require.NoError(t, err)
	assert.Equal(t, 0, count, "tag %s should be removed from DB", tagID)
}

func testAssignTask(t *testing.T, f *TestFixtures) {
	t.Helper()

	creatorID := createUserAndGetID(t, f.DB)
	taskID := createTaskAndGetID(t, f.DB, f.AuthToken, creatorID)
	assigneeID := createUserAndGetID(t, f.DB)

	resp, body := assignTask(t, f.AuthToken, taskID, assigneeID)
	assert.Equal(t, http.StatusOK, resp.StatusCode, "unexpected status: %s", string(body))

	var assignedTo string
	err := f.DB.Get(&assignedTo, "SELECT assigned_to FROM tasks WHERE id = ?", taskID)
	require.NoError(t, err)
	assert.Equal(t, assigneeID, assignedTo)
}

func testUnassignTask(t *testing.T, f *TestFixtures) {
	t.Helper()

	creatorID := createUserAndGetID(t, f.DB)
	taskID := createTaskAndGetID(t, f.DB, f.AuthToken, creatorID)
	assigneeID := createUserAndGetID(t, f.DB)

	resp, body := assignTask(t, f.AuthToken, taskID, assigneeID)
	require.Equal(t, http.StatusOK, resp.StatusCode, "setup failed assigning task: %s", string(body))

	resp, body = unassignTask(t, f.AuthToken, taskID)
	assert.Equal(t, http.StatusOK, resp.StatusCode, "unexpected status: %s", string(body))

	var assignedTo string
	err := f.DB.Get(&assignedTo, "SELECT assigned_to FROM tasks WHERE id = ?", taskID)
	require.NoError(t, err)
	assert.Empty(t, assignedTo, "task should have no assignee after unassign")
}

func testReopenTask(t *testing.T, f *TestFixtures) {
	t.Helper()

	creatorID := createUserAndGetID(t, f.DB)
	taskID := createTaskAndGetID(t, f.DB, f.AuthToken, creatorID)

	resp, body := putTaskStatus(t, f.AuthToken, taskID, map[string]any{
		"status": "done",
	})
	require.Equal(t, http.StatusOK, resp.StatusCode, "setup failed changing status to done: %s", string(body))

	resp, body = reopenTask(t, f.AuthToken, taskID)
	assert.Equal(t, http.StatusOK, resp.StatusCode, "unexpected status: %s", string(body))

	var status string
	err := f.DB.Get(&status, "SELECT status FROM tasks WHERE id = ?", taskID)
	require.NoError(t, err)
	assert.NotEqual(t, "done", status, "task should not be in 'done' status after reopen")
}

func testArchiveTask(t *testing.T, f *TestFixtures) {
	t.Helper()

	creatorID := createUserAndGetID(t, f.DB)
	taskID := createTaskAndGetID(t, f.DB, f.AuthToken, creatorID)

	resp, body := archiveTask(t, f.AuthToken, taskID)
	assert.Equal(t, http.StatusOK, resp.StatusCode, "unexpected status: %s", string(body))

	var count int
	err := f.DB.Get(&count, "SELECT COUNT(*) FROM tasks WHERE id = ? AND archived_at IS NOT NULL", taskID)
	require.NoError(t, err)
	assert.Equal(t, 1, count, "task %s should be archived in DB", taskID)
}

func testDeleteTask(t *testing.T, f *TestFixtures) {
	t.Helper()

	creatorID := createUserAndGetID(t, f.DB)
	taskID := createTaskAndGetID(t, f.DB, f.AuthToken, creatorID)

	resp, body := deleteTask(t, f.AuthToken, taskID)
	assert.Equal(t, http.StatusOK, resp.StatusCode, "unexpected status: %s", string(body))

	assertTaskDeletedInDB(t, f.DB, taskID)
}
