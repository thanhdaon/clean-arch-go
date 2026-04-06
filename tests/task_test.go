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

	title := fmt.Sprintf("Create-Task-%d", time.Now().UnixNano())
	resp, body := createTask(t, f.AuthToken, map[string]any{
		"title": title,
	})
	assert.Equal(t, http.StatusOK, resp.StatusCode, "unexpected status: %s", string(body))

	taskID := getTaskIDByTitle(t, f.DB, title)
	assert.NotEmpty(t, taskID)
}

func testGetTasks(t *testing.T, f *TestFixtures) {
	t.Helper()

	title := fmt.Sprintf("GetTasks-Task-%d", time.Now().UnixNano())

	resp, body := createTask(t, f.AuthToken, map[string]any{
		"title": title,
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

	assertTaskFieldEquals(t, f.DB, taskID, "title", newTitle)
}

func testChangeTaskStatus(t *testing.T, f *TestFixtures) {
	t.Helper()

	creatorID := createUserAndGetID(t, f.DB)
	taskID := createTaskAndGetID(t, f.DB, f.AuthToken, creatorID)

	resp, body := putTaskStatus(t, f.AuthToken, taskID, map[string]any{
		"status": "in_progress",
	})
	assert.Equal(t, http.StatusOK, resp.StatusCode, "unexpected status: %s", string(body))

	assertTaskFieldEquals(t, f.DB, taskID, "status", "in_progress")
}

func testSetTaskPriority(t *testing.T, f *TestFixtures) {
	t.Helper()

	creatorID := createUserAndGetID(t, f.DB)
	taskID := createTaskAndGetID(t, f.DB, f.AuthToken, creatorID)

	resp, body := putTaskPriority(t, f.AuthToken, taskID, map[string]any{
		"priority": "high",
	})
	assert.Equal(t, http.StatusOK, resp.StatusCode, "unexpected status: %s", string(body))

	priority := getTaskField[int](t, f.DB, taskID, "priority")
	assert.Greater(t, priority, 0, "priority should be set to a non-zero value")
}

func testSetTaskDueDate(t *testing.T, f *TestFixtures) {
	t.Helper()

	creatorID := createUserAndGetID(t, f.DB)
	taskID := createTaskAndGetID(t, f.DB, f.AuthToken, creatorID)

	expectedDueDate := "2027-03-01T10:00:00Z"
	resp, body := putTaskDueDate(t, f.AuthToken, taskID, map[string]any{
		"due_date": expectedDueDate,
	})
	assert.Equal(t, http.StatusOK, resp.StatusCode, "unexpected status: %s", string(body))

	expectedTime, err := time.Parse(time.RFC3339, expectedDueDate)
	require.NoError(t, err, "failed to parse expected due date")
	assertTaskFieldEquals(t, f.DB, taskID, "due_date", expectedTime)
}

func testSetTaskDescription(t *testing.T, f *TestFixtures) {
	t.Helper()

	creatorID := createUserAndGetID(t, f.DB)
	taskID := createTaskAndGetID(t, f.DB, f.AuthToken, creatorID)

	resp, body := patchTaskDescription(t, f.AuthToken, taskID, map[string]any{
		"description": "This is a test description",
	})
	assert.Equal(t, http.StatusOK, resp.StatusCode, "unexpected status: %s", string(body))

	assertTaskFieldEquals(t, f.DB, taskID, "description", "This is a test description")
}

func testAddTaskTag(t *testing.T, f *TestFixtures) {
	t.Helper()

	creatorID := createUserAndGetID(t, f.DB)
	taskID := createTaskAndGetID(t, f.DB, f.AuthToken, creatorID)

	resp, body := addTaskTag(t, f.AuthToken, taskID, map[string]any{
		"name": "bug",
	})
	assert.Equal(t, http.StatusOK, resp.StatusCode, "unexpected status: %s", string(body))

	count := assertTaskTagCount(t, f.DB, "task_id = ? AND name = ?", taskID, "bug")
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

	tagID := getTaskTagIDByName(t, f.DB, taskID, "to-remove")

	resp, body = removeTaskTag(t, f.AuthToken, taskID, tagID)
	assert.Equal(t, http.StatusOK, resp.StatusCode, "unexpected status: %s", string(body))

	count := assertTaskTagCount(t, f.DB, "id = ?", tagID)
	assert.Equal(t, 0, count, "tag %s should be removed from DB", tagID)
}

func testAssignTask(t *testing.T, f *TestFixtures) {
	t.Helper()

	creatorID := createUserAndGetID(t, f.DB)
	taskID := createTaskAndGetID(t, f.DB, f.AuthToken, creatorID)
	assigneeID := createUserAndGetID(t, f.DB)

	resp, body := assignTask(t, f.AuthToken, taskID, assigneeID)
	assert.Equal(t, http.StatusOK, resp.StatusCode, "unexpected status: %s", string(body))

	assertTaskFieldEquals(t, f.DB, taskID, "assigned_to", assigneeID)
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

	assertTaskFieldEmpty[string](t, f.DB, taskID, "assigned_to")
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

	status := getTaskField[string](t, f.DB, taskID, "status")
	assert.NotEqual(t, "done", status, "task should not be in 'done' status after reopen")
}

func testArchiveTask(t *testing.T, f *TestFixtures) {
	t.Helper()

	creatorID := createUserAndGetID(t, f.DB)
	taskID := createTaskAndGetID(t, f.DB, f.AuthToken, creatorID)

	resp, body := archiveTask(t, f.AuthToken, taskID)
	assert.Equal(t, http.StatusOK, resp.StatusCode, "unexpected status: %s", string(body))

	assertTaskArchived(t, f.DB, taskID)
}

func testDeleteTask(t *testing.T, f *TestFixtures) {
	t.Helper()

	t.Run("admin_deletes_task", func(t *testing.T) {
		t.Parallel()
		creatorID := createUserAndGetID(t, f.DB)
		taskID := createTaskAndGetID(t, f.DB, f.AuthToken, creatorID)

		resp, body := deleteTask(t, f.AuthToken, taskID)
		assert.Equal(t, http.StatusOK, resp.StatusCode, "unexpected status: %s", string(body))

		assertTaskDeletedInDB(t, f.DB, taskID)
	})

	t.Run("employer_deletes_own_task", func(t *testing.T) {
		t.Parallel()
		employerID, employerToken := createUserWithRoleAndGetToken(t, f.DB, "employer")
		taskID := createTaskAndGetID(t, f.DB, employerToken, employerID)

		resp, body := deleteTask(t, employerToken, taskID)
		assert.Equal(t, http.StatusOK, resp.StatusCode, "unexpected status: %s", string(body))

		assertTaskDeletedInDB(t, f.DB, taskID)
	})

	t.Run("employer_deletes_another_task_returns_forbidden", func(t *testing.T) {
		t.Parallel()
		_, otherEmployerToken := createUserWithRoleAndGetToken(t, f.DB, "employer")
		creatorID := createUserAndGetID(t, f.DB)
		taskID := createTaskAndGetID(t, f.DB, f.AuthToken, creatorID)

		resp, body := deleteTask(t, otherEmployerToken, taskID)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assertResponseStatus(t, body, http.StatusForbidden)

		assertTaskNotDeletedInDB(t, f.DB, taskID)
	})

	t.Run("employee_deletes_task_returns_forbidden", func(t *testing.T) {
		t.Parallel()
		_, employeeToken := createUserWithRoleAndGetToken(t, f.DB, "employee")
		creatorID := createUserAndGetID(t, f.DB)
		taskID := createTaskAndGetID(t, f.DB, f.AuthToken, creatorID)

		resp, body := deleteTask(t, employeeToken, taskID)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assertResponseStatus(t, body, http.StatusForbidden)

		assertTaskNotDeletedInDB(t, f.DB, taskID)
	})
}
