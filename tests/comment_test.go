package tests_test

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func testAddComment(t *testing.T, f *TestFixtures) {
	t.Helper()

	taskID := createTaskAndGetID(t, f.DB, f.AuthToken, "")

	resp, body := addComment(t, f.AuthToken, taskID, map[string]any{
		"content": "hello world",
	})
	assert.Equal(t, http.StatusCreated, resp.StatusCode, "unexpected status: %s", string(body))

	assertCommentExistsInDB(t, f.DB, getCommentIDByContent(t, f.DB, taskID, "hello world"))
}

func testUpdateComment(t *testing.T, f *TestFixtures) {
	t.Helper()

	taskID := createTaskAndGetID(t, f.DB, f.AuthToken, "")

	resp, body := addComment(t, f.AuthToken, taskID, map[string]any{
		"content": "original content",
	})
	require.Equal(t, http.StatusCreated, resp.StatusCode, "setup failed: %s", string(body))

	commentID := getCommentIDByContent(t, f.DB, taskID, "original content")

	resp, body = updateComment(t, f.AuthToken, taskID, commentID, map[string]any{
		"content": "edited content",
	})
	assert.Equal(t, http.StatusOK, resp.StatusCode, "unexpected status: %s", string(body))

	assertCommentExistsInDB(t, f.DB, getCommentIDByContent(t, f.DB, taskID, "edited content"))
}

func testDeleteComment(t *testing.T, f *TestFixtures) {
	t.Helper()

	taskID := createTaskAndGetID(t, f.DB, f.AuthToken, "")

	resp, body := addComment(t, f.AuthToken, taskID, map[string]any{
		"content": "to be deleted",
	})
	require.Equal(t, http.StatusCreated, resp.StatusCode, "setup failed: %s", string(body))

	commentID := getCommentIDByContent(t, f.DB, taskID, "to be deleted")

	resp, body = deleteComment(t, f.AuthToken, taskID, commentID)
	assert.Equal(t, http.StatusOK, resp.StatusCode, "unexpected status: %s", string(body))

	assertCommentDeletedInDB(t, f.DB, commentID)
}

func testGetTaskActivity(t *testing.T, f *TestFixtures) {
	t.Helper()

	taskID := createTaskAndGetID(t, f.DB, f.AuthToken, "")

	resp, body := addComment(t, f.AuthToken, taskID, map[string]any{
		"content": "activity trigger",
	})
	require.Equal(t, http.StatusCreated, resp.StatusCode, "setup failed: %s", string(body))

	resp, body = getTaskActivity(t, f.AuthToken, taskID)
	assert.Equal(t, http.StatusOK, resp.StatusCode, "unexpected status: %s", string(body))

	var result map[string]any
	err := json.Unmarshal(body, &result)
	require.NoError(t, err, "response should be JSON: %s", string(body))

	activities, ok := result["activities"].([]any)
	require.True(t, ok, "response should contain 'activities' array: %s", string(body))
	assert.NotEmpty(t, activities, "activity log should contain at least one entry")

	first, ok := activities[0].(map[string]any)
	require.True(t, ok, "activity entry should be an object")
	_, hasType := first["type"]
	assert.True(t, hasType, "activity entry should have a 'type' field")
}
