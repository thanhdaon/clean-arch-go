package user_test

import (
	"clean-arch-go/domain/user"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewUser(t *testing.T) {
	t.Parallel()

	tests := []struct {
		uuid    string
		role    user.Role
		wantErr bool
	}{
		{"1234", user.RoleEmployer, false},
		{"5678", user.RoleEmployee, false},
		{"1234", user.RoleUnknow, true},
		{"", user.RoleEmployer, true},
	}

	for _, tt := range tests {
		usr, err := user.NewUser(tt.uuid, tt.role)
		if tt.wantErr {
			require.Error(t, err)
			require.Nil(t, usr)
		} else {
			require.NoError(t, err)
			require.NotNil(t, usr)
			require.Equal(t, tt.uuid, usr.UUID())
			require.Equal(t, tt.role, usr.Role())
		}
	}
}
