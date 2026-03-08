package user_test

import (
	"testing"

	"clean-arch-go/domain/user"
	"github.com/stretchr/testify/require"
)

func TestNewUser(t *testing.T) {
	t.Parallel()

	u, err := user.NewUser("uuid-123", user.RoleEmployer, "John Doe", "john@example.com")
	require.NoError(t, err)
	require.NotNil(t, u)
	require.Equal(t, "uuid-123", u.UUID())
	require.Equal(t, user.RoleEmployer, u.Role())
	require.Equal(t, "John Doe", u.Name())
	require.Equal(t, "john@example.com", u.Email())
}

func TestNewUser_MissingFields(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		uuid      string
		role      user.Role
		nameStr   string
		email     string
		wantError string
	}{
		{"missing uuid", "", user.RoleEmployer, "John Doe", "john@example.com", "missing user uuid"},
		{"missing role", "uuid-123", user.RoleUnknown, "John Doe", "john@example.com", "missing user role"},
		{"missing name", "uuid-123", user.RoleEmployer, "", "john@example.com", "missing user name"},
		{"missing email", "uuid-123", user.RoleEmployer, "John Doe", "", "missing user email"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			_, err := user.NewUser(tt.uuid, tt.role, tt.nameStr, tt.email)
			require.Error(t, err)
			require.Contains(t, err.Error(), tt.wantError)
		})
	}
}

func TestChangeRole(t *testing.T) {
	t.Parallel()

	u, err := user.NewUser("uuid-123", user.RoleEmployer, "John Doe", "john@example.com")
	require.NoError(t, err)
	require.Equal(t, user.RoleEmployer, u.Role())

	u.ChangeRole(user.RoleEmployee)
	require.Equal(t, user.RoleEmployee, u.Role())
}

func TestUpdateProfile(t *testing.T) {
	t.Parallel()

	u, err := user.NewUser("uuid-123", user.RoleEmployer, "John Doe", "john@example.com")
	require.NoError(t, err)
	require.Equal(t, "John Doe", u.Name())
	require.Equal(t, "john@example.com", u.Email())

	u.UpdateProfile("Jane Doe", "jane@example.com")
	require.Equal(t, "Jane Doe", u.Name())
	require.Equal(t, "jane@example.com", u.Email())
}

func TestSetPasswordHash(t *testing.T) {
	t.Parallel()

	u, err := user.NewUser("uuid-123", user.RoleEmployer, "John Doe", "john@example.com")
	require.NoError(t, err)
	require.Empty(t, u.PasswordHash())

	u.SetPasswordHash("hashed-password-123")
	require.Equal(t, "hashed-password-123", u.PasswordHash())
}

func TestCanChangeRoleOf_AdminCanChangeOthers(t *testing.T) {
	t.Parallel()

	admin, err := user.NewUser("admin-uuid", user.RoleAdmin, "Admin", "admin@example.com")
	require.NoError(t, err)

	employer, err := user.NewUser("employer-uuid", user.RoleEmployer, "Employer", "employer@example.com")
	require.NoError(t, err)

	require.True(t, admin.CanChangeRoleOf(employer))
}

func TestCanChangeRoleOf_CannotChangeSelf(t *testing.T) {
	t.Parallel()

	admin, err := user.NewUser("admin-uuid", user.RoleAdmin, "Admin", "admin@example.com")
	require.NoError(t, err)

	require.False(t, admin.CanChangeRoleOf(admin))
}

func TestCanChangeRoleOf_NonAdminCannotChange(t *testing.T) {
	t.Parallel()

	employer, err := user.NewUser("employer-uuid", user.RoleEmployer, "Employer", "employer@example.com")
	require.NoError(t, err)

	employee, err := user.NewUser("employee-uuid", user.RoleEmployee, "Employee", "employee@example.com")
	require.NoError(t, err)

	require.False(t, employer.CanChangeRoleOf(employee))
}

func TestHashPassword(t *testing.T) {
	t.Parallel()

	hash, err := user.HashPassword("password123")
	require.NoError(t, err)
	require.NotEmpty(t, hash)
	require.NotEqual(t, "password123", hash)
}

func TestVerifyPassword(t *testing.T) {
	t.Parallel()

	hash, err := user.HashPassword("password123")
	require.NoError(t, err)

	require.True(t, user.VerifyPassword("password123", hash))
	require.False(t, user.VerifyPassword("wrong-password", hash))
}
