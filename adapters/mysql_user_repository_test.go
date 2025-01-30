package adapters_test

import (
	"clean-arch-go/adapters"
	"clean-arch-go/domain/user"
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMysqlUserRepository_Add(t *testing.T) {
	t.Parallel()
	userRepository := newMysqlUserRepository(t)

	testCases := []struct {
		Name            string
		UserConstructor func(t *testing.T) user.User
	}{
		{
			Name:            "employee_user",
			UserConstructor: newEmployeeUser,
		},
		{
			Name:            "employer_user",
			UserConstructor: newEmployerUser,
		},
	}

	for _, c := range testCases {
		t.Run(c.Name, func(t *testing.T) {
			t.Parallel()

			expectedUser := c.UserConstructor(t)
			err := userRepository.Add(context.Background(), expectedUser)
			require.NoError(t, err)
			assertPersistedUserEquals(t, userRepository, expectedUser)
		})
	}

	require.NotNil(t, userRepository)
}

func TestMysqlUserRepository_Update(t *testing.T) {
	t.Parallel()

	userRepository := newMysqlUserRepository(t)
	expectedUser := newEmployerUser(t)

	err := userRepository.Add(context.Background(), expectedUser)
	require.NoError(t, err)

	var updatedUser user.User

	err = userRepository.UpdateByID(context.TODO(), expectedUser.UUID(), func(ctx context.Context, found user.User) (user.User, error) {
		assertUsersEquals(t, expectedUser, found)

		found.ChangeRole(user.RoleEmployee)

		updatedUser = found
		return found, nil
	})
	require.NoError(t, err)

	assertPersistedUserEquals(t, userRepository, updatedUser)
}

func newMysqlUserRepository(t *testing.T) adapters.MysqlUserRepository {
	t.Helper()
	db, err := adapters.NewMySQLConnection()
	require.NoError(t, err)
	return adapters.NewMysqlUserRepository(db)
}

func newEmployeeUser(t *testing.T) user.User {
	t.Helper()
	created, err := user.NewUser(adapters.NewID().New(), user.RoleEmployee)
	require.NoError(t, err)
	return created
}

func newEmployerUser(t *testing.T) user.User {
	t.Helper()
	created, err := user.NewUser(adapters.NewID().New(), user.RoleEmployer)
	require.NoError(t, err)
	return created
}

func assertPersistedUserEquals(t *testing.T, repo adapters.MysqlUserRepository, target user.User) {
	t.Helper()

	persistedUser, err := repo.FindById(context.Background(), target.UUID())
	require.NoError(t, err)
	require.NotNil(t, persistedUser)

	assertUsersEquals(t, target, persistedUser)
}

func assertUsersEquals(t *testing.T, user1, user2 user.User) {
	t.Helper()

	require.Equal(t, user1.UUID(), user2.UUID())
	require.Equal(t, user1.Role().String(), user2.Role().String())
}
