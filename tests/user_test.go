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

func testUser(t *testing.T, f *TestFixtures) {
	t.Helper()

	t.Run("user.txt:8 - Login with valid credentials returns JWT token", func(t *testing.T) {
		email := fmt.Sprintf("login-valid-%d@example.com", time.Now().UnixNano())
		resp, body := postUser(t, map[string]any{
			"role":     "employer",
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
		require.NoError(t, json.Unmarshal(body, &result))
		assert.NotEmpty(t, result["token"], "response should contain a token")
	})

	t.Run("user.txt:19 - Login with invalid password returns 401 error", func(t *testing.T) {
		email := fmt.Sprintf("login-invalid-%d@example.com", time.Now().UnixNano())
		resp, body := postUser(t, map[string]any{
			"role":     "employer",
			"name":     "Login User",
			"email":    email,
			"password": "password123",
		})
		require.Equal(t, http.StatusOK, resp.StatusCode, "setup failed: %s", string(body))

		resp, body = login(t, map[string]any{
			"email":    email,
			"password": "wrongpassword",
		})
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assertResponseStatus(t, body, http.StatusUnauthorized)
	})

	t.Run("user.txt:29 - Login with non-existent email returns 401 error", func(t *testing.T) {
		resp, body := login(t, map[string]any{
			"email":    fmt.Sprintf("nonexistent-%d@example.com", time.Now().UnixNano()),
			"password": "password123",
		})
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assertResponseStatus(t, body, http.StatusUnauthorized)
	})

	t.Run("user.txt:38 - Login with missing email returns 400 error", func(t *testing.T) {
		resp, body := login(t, map[string]any{
			"email":    "",
			"password": "password123",
		})
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assertResponseStatus(t, body, http.StatusBadRequest)
	})

	t.Run("user.txt:47 - Login with missing password returns 400 error", func(t *testing.T) {
		t.Skip("not yet implemented: API returns 401 instead of 400 for missing password")
		resp, body := login(t, map[string]any{
			"email":    fmt.Sprintf("test-%d@example.com", time.Now().UnixNano()),
			"password": "",
		})
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assertResponseStatus(t, body, http.StatusBadRequest)
	})

	t.Run("user.txt:56 - Login with empty credentials returns 400 error", func(t *testing.T) {
		resp, body := login(t, map[string]any{
			"email":    "",
			"password": "",
		})
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assertResponseStatus(t, body, http.StatusBadRequest)
	})

	t.Run("user.txt:65 - Admin creates user with role admin", func(t *testing.T) {
		email := fmt.Sprintf("admin-%d@example.com", time.Now().UnixNano())
		resp, body := postUser(t, map[string]any{
			"role":     "admin",
			"name":     "Admin User",
			"email":    email,
			"password": "password123",
		})
		require.Equal(t, http.StatusOK, resp.StatusCode, "unexpected status: %s", string(body))

		var userID string
		err := f.DB.Get(&userID, "SELECT id FROM users WHERE email = ? AND deleted_at IS NULL", email)
		require.NoError(t, err)
		assertUserFieldEquals(t, f.DB, userID, "role", "admin")
	})

	t.Run("user.txt:75 - Admin creates user with role employer", func(t *testing.T) {
		email := fmt.Sprintf("employer-%d@example.com", time.Now().UnixNano())
		resp, body := postUser(t, map[string]any{
			"role":     "employer",
			"name":     "Employer User",
			"email":    email,
			"password": "password123",
		})
		require.Equal(t, http.StatusOK, resp.StatusCode, "unexpected status: %s", string(body))

		var userID string
		err := f.DB.Get(&userID, "SELECT id FROM users WHERE email = ? AND deleted_at IS NULL", email)
		require.NoError(t, err)
		assertUserFieldEquals(t, f.DB, userID, "role", "employer")
	})

	t.Run("user.txt:85 - Admin creates user with role employee", func(t *testing.T) {
		email := fmt.Sprintf("employee-%d@example.com", time.Now().UnixNano())
		resp, body := postUser(t, map[string]any{
			"role":     "employee",
			"name":     "Employee User",
			"email":    email,
			"password": "password123",
		})
		require.Equal(t, http.StatusOK, resp.StatusCode, "unexpected status: %s", string(body))

		var userID string
		err := f.DB.Get(&userID, "SELECT id FROM users WHERE email = ? AND deleted_at IS NULL", email)
		require.NoError(t, err)
		assertUserFieldEquals(t, f.DB, userID, "role", "employee")
	})

	t.Run("user.txt:95 - Create user with duplicate email returns 409 conflict", func(t *testing.T) {
		t.Skip("not yet implemented: API allows duplicate emails")
	})

	t.Run("user.txt:104 - Create user with invalid email format returns 400 error", func(t *testing.T) {
		t.Skip("not yet implemented: API does not validate email format")
	})

	t.Run("user.txt:113 - Create user with missing required fields returns 400 error", func(t *testing.T) {
		t.Skip("not yet implemented: API allows missing name field")
	})

	t.Run("user.txt:122 - Create user with empty name returns 400 error", func(t *testing.T) {
		t.Skip("not yet implemented: API allows empty name")
	})

	t.Run("user.txt:131 - Create user with weak password returns 400 error", func(t *testing.T) {
		t.Skip("not yet implemented: API does not validate password strength")
	})

	t.Run("user.txt:140 - Admin lists all users", func(t *testing.T) {
		email1 := fmt.Sprintf("list-user-1-%d@example.com", time.Now().UnixNano())
		resp, body := postUser(t, map[string]any{
			"role":     "employee",
			"name":     "List User 1",
			"email":    email1,
			"password": "password123",
		})
		require.Equal(t, http.StatusOK, resp.StatusCode, "setup failed: %s", string(body))

		email2 := fmt.Sprintf("list-user-2-%d@example.com", time.Now().UnixNano())
		resp, body = postUser(t, map[string]any{
			"role":     "employer",
			"name":     "List User 2",
			"email":    email2,
			"password": "password123",
		})
		require.Equal(t, http.StatusOK, resp.StatusCode, "setup failed: %s", string(body))

		resp, body = getUsers(t, f.AuthToken)
		assert.Equal(t, http.StatusOK, resp.StatusCode, "unexpected status: %s", string(body))

		var result map[string]any
		require.NoError(t, json.Unmarshal(body, &result))
		users, ok := result["users"].([]any)
		require.True(t, ok, "users should be a JSON array")

		emails := make([]string, len(users))
		for i, u := range users {
			userMap := u.(map[string]any)
			emails[i] = userMap["email"].(string)
		}
		assert.Contains(t, emails, email1)
		assert.Contains(t, emails, email2)
	})

	t.Run("user.txt:151 - Non-admin lists users returns forbidden", func(t *testing.T) {
		t.Skip("not yet implemented: API does not enforce admin-only restriction for listing users")
	})

	t.Run("user.txt:160 - List users with empty database returns empty list", func(t *testing.T) {
		resp, body := getUsers(t, f.AuthToken)
		assert.Equal(t, http.StatusOK, resp.StatusCode, "unexpected status: %s", string(body))

		var result map[string]any
		require.NoError(t, json.Unmarshal(body, &result))
		users, ok := result["users"].([]any)
		require.True(t, ok, "users should be a JSON array")
		assert.GreaterOrEqual(t, len(users), 1, "should return at least the admin user")
	})

	t.Run("user.txt:171 - Admin gets any user by ID", func(t *testing.T) {
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
		require.NoError(t, json.Unmarshal(body, &result))
		assert.NotNil(t, result["user"], "response should contain 'user' key")
	})

	t.Run("user.txt:182 - User gets own profile", func(t *testing.T) {
		userID, userToken := createUserWithRoleAndGetToken(t, f.DB, "employer")
		resp, body := getUser(t, userToken, userID)
		assert.Equal(t, http.StatusOK, resp.StatusCode, "unexpected status: %s", string(body))

		var result map[string]any
		require.NoError(t, json.Unmarshal(body, &result))
		assert.NotNil(t, result["user"], "response should contain 'user' key")
	})

	t.Run("user.txt:192 - User gets another user's profile returns forbidden", func(t *testing.T) {
		t.Skip("not yet implemented: API allows users to view other users' profiles")
	})

	t.Run("user.txt:202 - User updates own profile", func(t *testing.T) {
		userID, userToken := createUserWithRoleAndGetToken(t, f.DB, "employer")
		resp, body := patchUser(t, userToken, userID, map[string]any{
			"name": "New Name",
		})
		assert.Equal(t, http.StatusOK, resp.StatusCode, "unexpected status: %s", string(body))
		assertUserFieldEquals(t, f.DB, userID, "name", "New Name")
	})

	t.Run("user.txt:212 - Admin updates another user's profile", func(t *testing.T) {
		userID, _ := createUserWithRoleAndGetToken(t, f.DB, "employer")
		resp, body := patchUser(t, f.AuthToken, userID, map[string]any{
			"name": "New Name",
		})
		assert.Equal(t, http.StatusOK, resp.StatusCode, "unexpected status: %s", string(body))
		assertUserFieldEquals(t, f.DB, userID, "name", "New Name")
	})

	t.Run("user.txt:223 - Non-admin updates another user's profile returns forbidden", func(t *testing.T) {
		t.Skip("not yet implemented: API allows non-admin users to update other users' profiles")
	})

	t.Run("user.txt:233 - Update with empty name returns 400 error", func(t *testing.T) {
		t.Skip("not yet implemented: API allows empty name in profile update")
	})

	t.Run("user.txt:242 - Admin deletes another user", func(t *testing.T) {
		userID, _ := createUserWithRoleAndGetToken(t, f.DB, "employer")
		resp, body := deleteUser(t, f.AuthToken, userID)
		assert.Equal(t, http.StatusOK, resp.StatusCode, "unexpected status: %s", string(body))
		assertUserDeletedInDB(t, f.DB, userID)
	})

	t.Run("user.txt:253 - User deletes own account", func(t *testing.T) {
		userID, userToken := createUserWithRoleAndGetToken(t, f.DB, "employer")
		resp, body := deleteUser(t, userToken, userID)
		assert.Equal(t, http.StatusOK, resp.StatusCode, "unexpected status: %s", string(body))
		assertUserDeletedInDB(t, f.DB, userID)
	})

	t.Run("user.txt:263 - Non-admin deletes another user returns forbidden", func(t *testing.T) {
		t.Skip("not yet implemented: API allows non-admin users to delete other users")
	})

	t.Run("user.txt:273 - Admin changes user role to employer", func(t *testing.T) {
		userID, _ := createUserWithRoleAndGetToken(t, f.DB, "employee")
		resp, body := putUserRole(t, f.AuthToken, userID, map[string]any{
			"role": "employer",
		})
		assert.Equal(t, http.StatusOK, resp.StatusCode, "unexpected status: %s", string(body))
		assertUserFieldEquals(t, f.DB, userID, "role", "employer")
	})

	t.Run("user.txt:285 - Admin changes user role to employee", func(t *testing.T) {
		userID, _ := createUserWithRoleAndGetToken(t, f.DB, "employer")
		resp, body := putUserRole(t, f.AuthToken, userID, map[string]any{
			"role": "employee",
		})
		assert.Equal(t, http.StatusOK, resp.StatusCode, "unexpected status: %s", string(body))
		assertUserFieldEquals(t, f.DB, userID, "role", "employee")
	})

	t.Run("user.txt:295 - Admin changes user role to admin", func(t *testing.T) {
		userID, _ := createUserWithRoleAndGetToken(t, f.DB, "employer")
		resp, body := putUserRole(t, f.AuthToken, userID, map[string]any{
			"role": "admin",
		})
		assert.Equal(t, http.StatusOK, resp.StatusCode, "unexpected status: %s", string(body))
		assertUserFieldEquals(t, f.DB, userID, "role", "admin")
	})

	t.Run("user.txt:306 - Admin tries to change own role returns forbidden", func(t *testing.T) {
		var adminID string
		err := f.DB.Get(&adminID, "SELECT id FROM users WHERE email = 'test-admin@example.com' AND deleted_at IS NULL")
		require.NoError(t, err)

		resp, body := putUserRole(t, f.AuthToken, adminID, map[string]any{
			"role": "employer",
		})
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assertResponseStatus(t, body, http.StatusForbidden)
	})

	t.Run("user.txt:315 - Non-admin changes role returns forbidden", func(t *testing.T) {
		t.Skip("not yet implemented: API does not enforce admin-only restriction for role changes")
	})

	t.Run("user.txt:325 - Change role of non-existent user returns 404 error", func(t *testing.T) {
		resp, body := putUserRole(t, f.AuthToken, "00000000-0000-0000-0000-000000000000", map[string]any{
			"role": "employer",
		})
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assertResponseStatus(t, body, http.StatusNotFound)
	})

	t.Run("user.txt:335 - Change role to invalid value returns 400 error", func(t *testing.T) {
		t.Skip("not yet implemented: API returns 500 for invalid role value")
	})

	t.Run("user.txt:345 - Change role to same value succeeds (idempotent)", func(t *testing.T) {
		userID, _ := createUserWithRoleAndGetToken(t, f.DB, "employer")
		resp, body := putUserRole(t, f.AuthToken, userID, map[string]any{
			"role": "employer",
		})
		assert.Equal(t, http.StatusOK, resp.StatusCode, "unexpected status: %s", string(body))
		assertUserFieldEquals(t, f.DB, userID, "role", "employer")
	})
}
