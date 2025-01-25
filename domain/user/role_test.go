package user_test

import (
	"clean-arch-go/domain/user"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestUserRole_String(t *testing.T) {
	tests := []struct {
		role     user.Role
		expected string
	}{
		{user.RoleEmployer, "employer"},
		{user.RoleEmployee, "employee"},
		{user.RoleUnknow, ""},
	}

	for _, tt := range tests {
		require.Equal(t, tt.expected, tt.role.String())
	}
}

func TestUserRole_IsZero(t *testing.T) {
	t.Parallel()

	tests := []struct {
		role     user.Role
		expected bool
	}{
		{user.RoleEmployer, false},
		{user.RoleEmployee, false},
		{user.RoleUnknow, true},
	}

	for _, tt := range tests {
		require.Equal(t, tt.expected, tt.role.IsZero())
	}
}

func TestUserRole_FromString(t *testing.T) {
	t.Parallel()

	tests := []struct {
		input    string
		expected user.Role
		wantErr  bool
	}{
		{"employer", user.RoleEmployer, false},
		{"employee", user.RoleEmployee, false},
		{"manager", user.RoleUnknow, true},
		{"", user.RoleUnknow, true},
	}

	for _, tt := range tests {
		result, err := user.UserRoleFromString(tt.input)
		if tt.wantErr {
			require.Error(t, err)
		} else {
			require.NoError(t, err)
		}
		require.Equal(t, tt.expected, result)
	}
}
