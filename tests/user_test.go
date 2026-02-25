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
	require.True(t, ok, "response should contain 'users' key; got: %s", string(body))
	require.NotNil(t, users)

	// Assert the user we just created appears in the list
	userList, ok := users.([]any)
	require.True(t, ok, "users should be a JSON array; got: %T", users)

	found := false
	for _, u := range userList {
		userMap, ok := u.(map[string]any)
		if !ok {
			continue
		}
		if userMap["email"] == email {
			found = true
			break
		}
	}
	assert.True(t, found, "created user with email %q should appear in list response", email)
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

func testLogin(t *testing.T, f *TestFixtures) {
	t.Helper()

	email := fmt.Sprintf("login-%d@example.com", time.Now().UnixNano())
	resp, body := postUser(t, map[string]any{
		"role":     "employee",
		"name":     "Login User",
		"email":    email,
		"password": "password123",
	})
	require.Equal(t, http.StatusOK, resp.StatusCode, "setup failed: %s", string(body))

	resp, body = login(t, map[string]any{
		"email":    email,
		"password": "password123",
	})
	assert.Equal(t, http.StatusOK, resp.StatusCode, "unexpected status: %s", string(body))

	var result map[string]any
	err := json.Unmarshal(body, &result)
	require.NoError(t, err)
	assert.NotEmpty(t, result["token"], "response should contain a token")
	assert.NotNil(t, result["user"], "response should contain a user object")
}

func testUpdateUserProfile(t *testing.T, f *TestFixtures) {
	t.Helper()

	email := fmt.Sprintf("update-profile-%d@example.com", time.Now().UnixNano())
	resp, body := postUser(t, map[string]any{
		"role":     "employee",
		"name":     "Original Name",
		"email":    email,
		"password": "password123",
	})
	require.Equal(t, http.StatusOK, resp.StatusCode, "setup failed: %s", string(body))

	var userID string
	err := f.DB.Get(&userID, "SELECT id FROM users WHERE email = ? AND deleted_at IS NULL", email)
	require.NoError(t, err)

	newEmail := fmt.Sprintf("updated-%d@example.com", time.Now().UnixNano())
	resp, body = patchUser(t, f.AuthToken, userID, map[string]any{
		"name":  "Updated Name",
		"email": newEmail,
	})
	assert.Equal(t, http.StatusOK, resp.StatusCode, "unexpected status: %s", string(body))

	var name, storedEmail string
	err = f.DB.QueryRow(
		"SELECT name, email FROM users WHERE id = ? AND deleted_at IS NULL",
		userID,
	).Scan(&name, &storedEmail)
	require.NoError(t, err)
	assert.Equal(t, "Updated Name", name)
	assert.Equal(t, newEmail, storedEmail)
}

func testUpdateUserRole(t *testing.T, f *TestFixtures) {
	t.Helper()

	email := fmt.Sprintf("update-role-%d@example.com", time.Now().UnixNano())
	resp, body := postUser(t, map[string]any{
		"role":     "employee",
		"name":     "Role User",
		"email":    email,
		"password": "password123",
	})
	require.Equal(t, http.StatusOK, resp.StatusCode, "setup failed: %s", string(body))

	var userID string
	err := f.DB.Get(&userID, "SELECT id FROM users WHERE email = ? AND deleted_at IS NULL", email)
	require.NoError(t, err)

	resp, body = putUserRole(t, f.AuthToken, userID, map[string]any{
		"role": "employer",
	})
	assert.Equal(t, http.StatusOK, resp.StatusCode, "unexpected status: %s", string(body))

	var role string
	err = f.DB.Get(&role, "SELECT role FROM users WHERE id = ?", userID)
	require.NoError(t, err)
	assert.Equal(t, "employer", role)
}

func testDeleteUser(t *testing.T, f *TestFixtures) {
	t.Helper()

	email := fmt.Sprintf("delete-user-%d@example.com", time.Now().UnixNano())
	resp, body := postUser(t, map[string]any{
		"role":     "employee",
		"name":     "Delete User",
		"email":    email,
		"password": "password123",
	})
	require.Equal(t, http.StatusOK, resp.StatusCode, "setup failed: %s", string(body))

	var userID string
	err := f.DB.Get(&userID, "SELECT id FROM users WHERE email = ? AND deleted_at IS NULL", email)
	require.NoError(t, err)

	resp, body = deleteUser(t, f.AuthToken, userID)
	assert.Equal(t, http.StatusOK, resp.StatusCode, "unexpected status: %s", string(body))

	assertUserDeletedInDB(t, f.DB, userID)
}
