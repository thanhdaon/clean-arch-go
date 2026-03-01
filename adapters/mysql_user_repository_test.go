package adapters_test

import (
	"clean-arch-go/adapters"
	"clean-arch-go/core/errors"
	"clean-arch-go/domain/errkind"
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

func TestMysqlUserRepository_UpdateByID(t *testing.T) {
	t.Parallel()
	userRepository := newMysqlUserRepository(t)

	t.Run("update_role", func(t *testing.T) {
		t.Parallel()
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
	})

	t.Run("not_found", func(t *testing.T) {
		t.Parallel()

		err := userRepository.UpdateByID(context.Background(), "non-existent-uuid", func(ctx context.Context, found user.User) (user.User, error) {
			return found, nil
		})

		require.Error(t, err)
		require.True(t, errors.Is(errkind.NotExist, err))
	})
}

func TestMysqlUserRepository_FindById(t *testing.T) {
	t.Parallel()
	userRepository := newMysqlUserRepository(t)

	existingUser := newEmployeeUser(t)
	err := userRepository.Add(context.Background(), existingUser)
	require.NoError(t, err)

	deletedUser := newEmployerUser(t)
	err = userRepository.Add(context.Background(), deletedUser)
	require.NoError(t, err)
	err = userRepository.DeleteByID(context.Background(), deletedUser.UUID())
	require.NoError(t, err)
	deletedUserUUID := deletedUser.UUID()

	testCases := []struct {
		Name        string
		UUID        string
		ShouldExist bool
	}{
		{
			Name:        "found",
			UUID:        existingUser.UUID(),
			ShouldExist: true,
		},
		{
			Name:        "not_found",
			UUID:        "non-existent-uuid",
			ShouldExist: false,
		},
		{
			Name:        "deleted_user",
			UUID:        deletedUserUUID,
			ShouldExist: false,
		},
	}

	for _, c := range testCases {
		t.Run(c.Name, func(t *testing.T) {
			t.Parallel()

			found, err := userRepository.FindById(context.Background(), c.UUID)

			if c.ShouldExist {
				require.NoError(t, err)
				assertUsersEquals(t, existingUser, found)
			} else {
				require.Error(t, err)
				require.True(t, errors.Is(errkind.NotExist, err))
			}
		})
	}
}

func TestMysqlUserRepository_FindByEmail(t *testing.T) {
	t.Parallel()
	userRepository := newMysqlUserRepository(t)

	existingUser := newEmployeeUser(t)
	err := userRepository.Add(context.Background(), existingUser)
	require.NoError(t, err)

	testCases := []struct {
		Name        string
		Email       string
		ShouldExist bool
	}{
		{
			Name:        "found",
			Email:       existingUser.Email(),
			ShouldExist: true,
		},
		{
			Name:        "not_found",
			Email:       "nonexistent@example.com",
			ShouldExist: false,
		},
	}

	for _, c := range testCases {
		t.Run(c.Name, func(t *testing.T) {
			t.Parallel()

			found, err := userRepository.FindByEmail(context.Background(), c.Email)

			if c.ShouldExist {
				require.NoError(t, err)
				require.Equal(t, existingUser.Email(), found.Email())
			} else {
				require.Error(t, err)
				require.True(t, errors.Is(errkind.NotExist, err))
			}
		})
	}
}

func TestMysqlUserRepository_FindAll(t *testing.T) {
	userRepository := newMysqlUserRepository(t)
	require.NoError(t, userRepository.RemoveAll(context.Background()))

	t.Run("empty", func(t *testing.T) {
		users, err := userRepository.FindAll(context.Background())

		require.NoError(t, err)
		require.Empty(t, users)
	})

	t.Run("multiple_users", func(t *testing.T) {
		user1 := newEmployeeUser(t)
		user2 := newEmployerUser(t)
		require.NoError(t, userRepository.Add(context.Background(), user1))
		require.NoError(t, userRepository.Add(context.Background(), user2))

		users, err := userRepository.FindAll(context.Background())

		require.NoError(t, err)
		require.Len(t, users, 2)
	})
}

func newMysqlUserRepository(t *testing.T) adapters.MysqlUserRepository {
	t.Helper()
	db, err := adapters.NewMySQLConnection()
	require.NoError(t, err)
	return adapters.NewMysqlUserRepository(db)
}

func newEmployeeUser(t *testing.T) user.User {
	t.Helper()
	created, err := user.NewUser(adapters.NewID().New(), user.RoleEmployee, "Employee Name", "employee@example.com")
	require.NoError(t, err)
	return created
}

func newEmployerUser(t *testing.T) user.User {
	t.Helper()
	created, err := user.NewUser(adapters.NewID().New(), user.RoleEmployer, "Employer Name", "employer@example.com")
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
