package tests_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"testing"

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
