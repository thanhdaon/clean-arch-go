package tests_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/require"
)

func baseURL() string {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	return "http://localhost:" + port
}

func postUser(t *testing.T, body map[string]any) (*http.Response, []byte) {
	t.Helper()

	payload, err := json.Marshal(body)
	require.NoError(t, err)

	req, err := http.NewRequest(http.MethodPost, baseURL()+"/api/users", bytes.NewBuffer(payload))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)

	respBody, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	resp.Body.Close()

	return resp, respBody
}

func getUsers(t *testing.T, token string) (*http.Response, []byte) {
	t.Helper()

	req, err := http.NewRequest(http.MethodGet, baseURL()+"/api/users", nil)
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)

	respBody, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	resp.Body.Close()

	return resp, respBody
}

func getUser(t *testing.T, token, userID string) (*http.Response, []byte) {
	t.Helper()

	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/api/users/%s", baseURL(), userID), nil)
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)

	respBody, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	resp.Body.Close()

	return resp, respBody
}

func assertUserExistsInDB(t *testing.T, db *sqlx.DB, userID string) {
	t.Helper()

	var count int
	err := db.Get(&count, "SELECT COUNT(*) FROM users WHERE id = ? AND deleted_at IS NULL", userID)
	require.NoError(t, err)
	require.Equal(t, 1, count, "user %s should exist in DB", userID)
}

func login(t *testing.T, body map[string]any) (*http.Response, []byte) {
	t.Helper()

	payload, err := json.Marshal(body)
	require.NoError(t, err)

	req, err := http.NewRequest(http.MethodPost, baseURL()+"/api/auth/login", bytes.NewBuffer(payload))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)

	respBody, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	resp.Body.Close()

	return resp, respBody
}

func deleteUser(t *testing.T, token, userID string) (*http.Response, []byte) {
	t.Helper()

	req, err := http.NewRequest(http.MethodDelete, fmt.Sprintf("%s/api/users/%s", baseURL(), userID), nil)
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)

	respBody, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	resp.Body.Close()

	return resp, respBody
}

func patchUser(t *testing.T, token, userID string, body map[string]any) (*http.Response, []byte) {
	t.Helper()

	payload, err := json.Marshal(body)
	require.NoError(t, err)

	req, err := http.NewRequest(http.MethodPatch, fmt.Sprintf("%s/api/users/%s", baseURL(), userID), bytes.NewBuffer(payload))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)

	respBody, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	resp.Body.Close()

	return resp, respBody
}

func putUserRole(t *testing.T, token, userID string, body map[string]any) (*http.Response, []byte) {
	t.Helper()

	payload, err := json.Marshal(body)
	require.NoError(t, err)

	req, err := http.NewRequest(http.MethodPut, fmt.Sprintf("%s/api/users/%s/role", baseURL(), userID), bytes.NewBuffer(payload))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)

	respBody, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	resp.Body.Close()

	return resp, respBody
}

func assertUserDeletedInDB(t *testing.T, db *sqlx.DB, userID string) {
	t.Helper()

	var count int
	err := db.Get(&count, "SELECT COUNT(*) FROM users WHERE id = ? AND deleted_at IS NOT NULL", userID)
	require.NoError(t, err)
	require.Equal(t, 1, count, "user %s should be soft-deleted in DB", userID)
}

func createTask(t *testing.T, token string, body map[string]any) (*http.Response, []byte) {
	t.Helper()

	payload, err := json.Marshal(body)
	require.NoError(t, err)

	req, err := http.NewRequest(http.MethodPost, baseURL()+"/api/tasks", bytes.NewBuffer(payload))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)

	respBody, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	resp.Body.Close()

	return resp, respBody
}

func getTasks(t *testing.T, token string) (*http.Response, []byte) {
	t.Helper()

	req, err := http.NewRequest(http.MethodGet, baseURL()+"/api/tasks", nil)
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)

	respBody, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	resp.Body.Close()

	return resp, respBody
}

func deleteTask(t *testing.T, token, taskID string) (*http.Response, []byte) {
	t.Helper()

	req, err := http.NewRequest(http.MethodDelete, fmt.Sprintf("%s/api/tasks/%s", baseURL(), taskID), nil)
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)

	respBody, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	resp.Body.Close()

	return resp, respBody
}

func patchTaskTitle(t *testing.T, token, taskID string, body map[string]any) (*http.Response, []byte) {
	t.Helper()

	payload, err := json.Marshal(body)
	require.NoError(t, err)

	req, err := http.NewRequest(http.MethodPatch, fmt.Sprintf("%s/api/tasks/%s", baseURL(), taskID), bytes.NewBuffer(payload))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)

	respBody, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	resp.Body.Close()

	return resp, respBody
}

func putTaskStatus(t *testing.T, token, taskID string, body map[string]any) (*http.Response, []byte) {
	t.Helper()

	payload, err := json.Marshal(body)
	require.NoError(t, err)

	req, err := http.NewRequest(http.MethodPut, fmt.Sprintf("%s/api/tasks/%s/status", baseURL(), taskID), bytes.NewBuffer(payload))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)

	respBody, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	resp.Body.Close()

	return resp, respBody
}

func putTaskPriority(t *testing.T, token, taskID string, body map[string]any) (*http.Response, []byte) {
	t.Helper()

	payload, err := json.Marshal(body)
	require.NoError(t, err)

	req, err := http.NewRequest(http.MethodPut, fmt.Sprintf("%s/api/tasks/%s/priority", baseURL(), taskID), bytes.NewBuffer(payload))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)

	respBody, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	resp.Body.Close()

	return resp, respBody
}

func putTaskDueDate(t *testing.T, token, taskID string, body map[string]any) (*http.Response, []byte) {
	t.Helper()

	payload, err := json.Marshal(body)
	require.NoError(t, err)

	req, err := http.NewRequest(http.MethodPut, fmt.Sprintf("%s/api/tasks/%s/due-date", baseURL(), taskID), bytes.NewBuffer(payload))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)

	respBody, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	resp.Body.Close()

	return resp, respBody
}

func patchTaskDescription(t *testing.T, token, taskID string, body map[string]any) (*http.Response, []byte) {
	t.Helper()

	payload, err := json.Marshal(body)
	require.NoError(t, err)

	req, err := http.NewRequest(http.MethodPatch, fmt.Sprintf("%s/api/tasks/%s/description", baseURL(), taskID), bytes.NewBuffer(payload))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)

	respBody, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	resp.Body.Close()

	return resp, respBody
}

func addTaskTag(t *testing.T, token, taskID string, body map[string]any) (*http.Response, []byte) {
	t.Helper()

	payload, err := json.Marshal(body)
	require.NoError(t, err)

	req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("%s/api/tasks/%s/tags", baseURL(), taskID), bytes.NewBuffer(payload))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)

	respBody, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	resp.Body.Close()

	return resp, respBody
}

func removeTaskTag(t *testing.T, token, taskID, tagID string) (*http.Response, []byte) {
	t.Helper()

	req, err := http.NewRequest(http.MethodDelete, fmt.Sprintf("%s/api/tasks/%s/tags/%s", baseURL(), taskID, tagID), nil)
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)

	respBody, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	resp.Body.Close()

	return resp, respBody
}

func assignTask(t *testing.T, token, taskID, assigneeID string) (*http.Response, []byte) {
	t.Helper()

	req, err := http.NewRequest(http.MethodPut, fmt.Sprintf("%s/api/tasks/%s/assign/%s", baseURL(), taskID, assigneeID), nil)
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)

	respBody, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	resp.Body.Close()

	return resp, respBody
}

func unassignTask(t *testing.T, token, taskID string) (*http.Response, []byte) {
	t.Helper()

	req, err := http.NewRequest(http.MethodDelete, fmt.Sprintf("%s/api/tasks/%s/assign", baseURL(), taskID), nil)
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)

	respBody, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	resp.Body.Close()

	return resp, respBody
}

func reopenTask(t *testing.T, token, taskID string) (*http.Response, []byte) {
	t.Helper()

	req, err := http.NewRequest(http.MethodPut, fmt.Sprintf("%s/api/tasks/%s/reopen", baseURL(), taskID), nil)
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)

	respBody, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	resp.Body.Close()

	return resp, respBody
}

func archiveTask(t *testing.T, token, taskID string) (*http.Response, []byte) {
	t.Helper()

	req, err := http.NewRequest(http.MethodPut, fmt.Sprintf("%s/api/tasks/%s/archive", baseURL(), taskID), nil)
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)

	respBody, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	resp.Body.Close()

	return resp, respBody
}

func assertTaskExistsInDB(t *testing.T, db *sqlx.DB, taskID string) {
	t.Helper()

	var count int
	err := db.Get(&count, "SELECT COUNT(*) FROM tasks WHERE id = ? AND deleted_at IS NULL", taskID)
	require.NoError(t, err)
	require.Equal(t, 1, count, "task %s should exist in DB", taskID)
}

func assertTaskDeletedInDB(t *testing.T, db *sqlx.DB, taskID string) {
	t.Helper()

	var count int
	err := db.Get(&count, "SELECT COUNT(*) FROM tasks WHERE id = ? AND deleted_at IS NOT NULL", taskID)
	require.NoError(t, err)
	require.Equal(t, 1, count, "task %s should be soft-deleted in DB", taskID)
}

func createUserAndGetID(t *testing.T, db *sqlx.DB) string {
	t.Helper()

	email := fmt.Sprintf("task-creator-%d@example.com", time.Now().UnixNano())
	resp, body := postUser(t, map[string]any{
		"role":     "employer",
		"name":     "Task Creator",
		"email":    email,
		"password": "password123",
	})
	require.Equal(t, http.StatusOK, resp.StatusCode, "setup failed: %s", string(body))

	var userID string
	err := db.Get(&userID, "SELECT id FROM users WHERE email = ? AND deleted_at IS NULL", email)
	require.NoError(t, err)
	return userID
}

func createTaskAndGetID(t *testing.T, db *sqlx.DB, token, creatorID string) string {
	t.Helper()

	title := fmt.Sprintf("Task-%d", time.Now().UnixNano())
	resp, body := createTask(t, token, map[string]any{
		"title": title,
	})
	require.Equal(t, http.StatusOK, resp.StatusCode, "setup failed: %s", string(body))

	var taskID string
	err := db.Get(&taskID, "SELECT id FROM tasks WHERE title = ? AND deleted_at IS NULL", title)
	require.NoError(t, err)
	return taskID
}

func getTaskField[T any](t *testing.T, db *sqlx.DB, taskID, field string) T {
	t.Helper()
	var result T
	err := db.Get(&result, fmt.Sprintf("SELECT %s FROM tasks WHERE id = ?", field), taskID)
	require.NoError(t, err)
	return result
}

func getTaskIDByTitle(t *testing.T, db *sqlx.DB, title string) string {
	t.Helper()
	var taskID string
	err := db.Get(&taskID, "SELECT id FROM tasks WHERE title = ? AND deleted_at IS NULL", title)
	require.NoError(t, err, "task should be stored in DB")
	return taskID
}

func getTaskTagIDByName(t *testing.T, db *sqlx.DB, taskID, name string) string {
	t.Helper()
	var tagID string
	err := db.Get(&tagID, "SELECT id FROM task_tags WHERE task_id = ? AND name = ?", taskID, name)
	require.NoError(t, err)
	return tagID
}

func countTableWhere(t *testing.T, db *sqlx.DB, table, whereClause string, args ...any) int {
	t.Helper()
	var count int
	query := fmt.Sprintf("SELECT COUNT(*) FROM %s WHERE %s", table, whereClause)
	err := db.Get(&count, query, args...)
	require.NoError(t, err)
	return count
}

func assertTaskFieldEquals[T comparable](t *testing.T, db *sqlx.DB, taskID, field string, expected T) {
	t.Helper()
	actual := getTaskField[T](t, db, taskID, field)
	require.Equal(t, expected, actual, "%s should match", field)
}

func assertTaskFieldEmpty[T comparable](t *testing.T, db *sqlx.DB, taskID, field string) {
	t.Helper()
	var zero T
	actual := getTaskField[T](t, db, taskID, field)
	require.Equal(t, zero, actual, "%s should be empty", field)
}

func assertTaskArchived(t *testing.T, db *sqlx.DB, taskID string) {
	t.Helper()
	count := countTableWhere(t, db, "tasks", "id = ? AND archived_at IS NOT NULL", taskID)
	require.Equal(t, 1, count, "task %s should be archived in DB", taskID)
}

func assertTaskHasDueDate(t *testing.T, db *sqlx.DB, taskID string) {
	t.Helper()
	count := countTableWhere(t, db, "tasks", "id = ? AND due_date IS NOT NULL", taskID)
	require.Equal(t, 1, count, "task due_date should be set in DB")
}

func assertTaskTagCount(t *testing.T, db *sqlx.DB, whereClause string, args ...any) int {
	t.Helper()
	return countTableWhere(t, db, "task_tags", whereClause, args...)
}

func addComment(t *testing.T, token, taskID string, body map[string]any) (*http.Response, []byte) {
	t.Helper()

	payload, err := json.Marshal(body)
	require.NoError(t, err)

	req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("%s/api/tasks/%s/comments", baseURL(), taskID), bytes.NewBuffer(payload))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)

	respBody, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	resp.Body.Close()

	return resp, respBody
}

func updateComment(t *testing.T, token, taskID, commentID string, body map[string]any) (*http.Response, []byte) {
	t.Helper()

	payload, err := json.Marshal(body)
	require.NoError(t, err)

	req, err := http.NewRequest(http.MethodPatch, fmt.Sprintf("%s/api/tasks/%s/comments/%s", baseURL(), taskID, commentID), bytes.NewBuffer(payload))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)

	respBody, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	resp.Body.Close()

	return resp, respBody
}

func deleteComment(t *testing.T, token, taskID, commentID string) (*http.Response, []byte) {
	t.Helper()

	req, err := http.NewRequest(http.MethodDelete, fmt.Sprintf("%s/api/tasks/%s/comments/%s", baseURL(), taskID, commentID), nil)
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)

	respBody, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	resp.Body.Close()

	return resp, respBody
}

func getTaskActivity(t *testing.T, token, taskID string) (*http.Response, []byte) {
	t.Helper()

	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/api/tasks/%s/activity", baseURL(), taskID), nil)
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)

	respBody, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	resp.Body.Close()

	return resp, respBody
}

func getCommentIDByContent(t *testing.T, db *sqlx.DB, taskID, content string) string {
	t.Helper()
	var id string
	err := db.Get(&id, "SELECT id FROM comments WHERE task_id = ? AND content = ? AND deleted_at IS NULL", taskID, content)
	require.NoError(t, err, "comment with content %q should exist for task %s", content, taskID)
	return id
}

func assertCommentExistsInDB(t *testing.T, db *sqlx.DB, commentID string) {
	t.Helper()
	count := countTableWhere(t, db, "comments", "id = ? AND deleted_at IS NULL", commentID)
	require.Equal(t, 1, count, "comment %s should exist in DB", commentID)
}

func assertCommentDeletedInDB(t *testing.T, db *sqlx.DB, commentID string) {
	t.Helper()
	count := countTableWhere(t, db, "comments", "id = ? AND deleted_at IS NOT NULL", commentID)
	require.Equal(t, 1, count, "comment %s should be soft-deleted in DB", commentID)
}
