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

func testAddUser(t *testing.T, f *TestFixtures) {
	t.Helper()

	email := fmt.Sprintf("test-%d@example.com", time.Now().UnixNano())

	before := f.VideoService.CallCountSnapshot()

	resp, body := postUser(t, map[string]any{
		"role":     "employer",
		"name":     "Test User",
		"email":    email,
		"password": "password123",
	})

	assert.Equal(t, http.StatusOK, resp.StatusCode, "unexpected status: %s", string(body))

	var userID string
	err := f.DB.Get(&userID, "SELECT id FROM users WHERE email = ? AND deleted_at IS NULL", email)
	require.NoError(t, err, "user should be stored in DB")
	assert.NotEmpty(t, userID)

	after := f.VideoService.CallCount()
	assert.Equal(t, 1, after-before, "adding a user should trigger exactly one VideoService call")
}

func testListUsers(t *testing.T, f *TestFixtures) {
	t.Helper()

	email := fmt.Sprintf("list-user-%d@example.com", time.Now().UnixNano())
	resp, body := postUser(t, map[string]any{
		"role":     "employee",
		"name":     "List User",
		"email":    email,
		"password": "password123",
	})
	require.Equal(t, http.StatusOK, resp.StatusCode, "setup failed: %s", string(body))

	resp, body = getUsers(t, f.AuthToken)
	assert.Equal(t, http.StatusOK, resp.StatusCode, "unexpected status: %s", string(body))

	var result map[string]any
	err := json.Unmarshal(body, &result)
	require.NoError(t, err)

	users, ok := result["users"]
	assert.True(t, ok, "response should contain 'users' key; got: %s", string(body))
	assert.NotNil(t, users)
}

func testGetUser(t *testing.T, f *TestFixtures) {
	t.Helper()

	email := fmt.Sprintf("get-user-%d@example.com", time.Now().UnixNano())
	resp, body := postUser(t, map[string]any{
		"role":     "employee",
		"name":     "Get User",
		"email":    email,
		"password": "password123",
	})
	require.Equal(t, http.StatusOK, resp.StatusCode, "setup failed: %s", string(body))

	var userID string
	err := f.DB.Get(&userID, "SELECT id FROM users WHERE email = ? AND deleted_at IS NULL", email)
	require.NoError(t, err)

	resp, body = getUser(t, f.AuthToken, userID)
	assert.Equal(t, http.StatusOK, resp.StatusCode, "unexpected status: %s", string(body))

	var result map[string]any
	err = json.Unmarshal(body, &result)
	require.NoError(t, err)
	assert.NotNil(t, result["user"], "response should contain 'user' key; got: %s", string(body))

	resp, body = getUser(t, f.AuthToken, "00000000-nonexistent-uuid")
	assert.Equal(t, http.StatusOK, resp.StatusCode, "unexpected status: %s", string(body))
	var errorResult map[string]any
	err = json.Unmarshal(body, &errorResult)
	require.NoError(t, err)
	assert.Contains(t, errorResult["error"], "item does not exist")
}
