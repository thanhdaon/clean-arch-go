# Component Tests: Comments & Activity Endpoints

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Add component (HTTP end-to-end) tests for the four new endpoints: add comment, update comment, delete comment, and get task activity.

**Architecture:** Tests live in `tests/` package (`tests_test`). A single shared `TestFixtures` (real MySQL + real HTTP server) is used across all sub-tests. New scenario functions go in `tests/comment_test.go`; new HTTP helpers and DB helpers go in `tests/helpers_test.go`. The four sub-tests are registered in `TestComponent` in `tests/component_test.go`.

**Tech Stack:** Go stdlib `net/http`, `testify/assert`, `testify/require`, `jmoiron/sqlx` for direct DB assertions.

---

### Task 1: Add HTTP helpers for comment and activity endpoints

**Files:**
- Modify: `tests/helpers_test.go`

**Step 1: Append the following functions to `tests/helpers_test.go`**

```go
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
	err := db.Get(&id, "SELECT id FROM comments WHERE task_id = ? AND body = ? AND deleted_at IS NULL", taskID, content)
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
```

**Step 2: Verify the file compiles**

```bash
go build ./tests/...
```

Expected: no errors (functions won't be called yet, but they must compile).

---

### Task 2: Write scenario functions in `comment_test.go`

**Files:**
- Create: `tests/comment_test.go`

**Step 1: Write the file with all four scenario functions**

```go
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

	var activities []map[string]any
	err := json.Unmarshal(body, &activities)
	require.NoError(t, err, "response should be a JSON array: %s", string(body))
	assert.NotEmpty(t, activities, "activity log should contain at least one entry")

	_, hasType := activities[0]["type"]
	assert.True(t, hasType, "activity entry should have a 'type' field")
}
```

**Step 2: Build to verify it compiles**

```bash
go build ./tests/...
```

Expected: no errors.

---

### Task 3: Register sub-tests in `TestComponent`

**Files:**
- Modify: `tests/component_test.go`

**Step 1: Add the four new `t.Run` calls** after the existing `delete_task` sub-test:

```go
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
```

**Step 2: Build to verify**

```bash
go build ./tests/...
```

Expected: no errors.

---

### Task 4: Run the component tests and fix any failures

**Step 1: Start the server and DB if not running**

```bash
make start
```

(Or ensure `docker compose up` is running and MySQL is reachable.)

**Step 2: Run only the new sub-tests**

```bash
go test ./tests/... -run TestComponent/add_comment -v
go test ./tests/... -run TestComponent/update_comment -v
go test ./tests/... -run TestComponent/delete_comment -v
go test ./tests/... -run TestComponent/get_task_activity -v
```

Expected: all four PASS.

**Step 3: Run the full suite to confirm no regressions**

```bash
make component-test
```

Expected: all tests PASS.

**Step 4: Commit**

```bash
git add tests/comment_test.go tests/helpers_test.go tests/component_test.go
git commit -m "test: add component tests for comment and activity endpoints"
```

---

### Notes

- `createTaskAndGetID` accepts a `creatorID` parameter — pass `""` (empty string) since the admin token is used and the function's internal logic doesn't require a real creatorID for test purposes. **Check the function signature** in `tests/helpers_test.go:478` first — if it requires a non-empty creatorID, call `createUserAndGetID(t, f.DB)` to obtain one first, as done in other task tests.
- The `comments` table column holding the text is named `body` (verified from `migrations/2_comments_activities.sql`).
- The add comment endpoint returns `201 Created` per the OpenAPI spec — not `200 OK`.
