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
}
