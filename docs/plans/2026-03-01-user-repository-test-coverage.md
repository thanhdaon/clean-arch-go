# User Repository Test Coverage Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Expand `MysqlUserRepository` test coverage from 4 to 18 test cases covering all methods.

**Architecture:** Table-driven tests with sub-tests per method, following existing patterns in `adapters/mysql_user_repository_test.go`.

**Tech Stack:** Go, stretchr/testify, MySQL test database

---

## Task 1: Add FindById Tests

**Files:**
- Modify: `adapters/mysql_user_repository_test.go`

**Step 1: Write the failing test for FindById**

Add after `TestMysqlUserRepository_Update`:

```go
func TestMysqlUserRepository_FindById(t *testing.T) {
	t.Parallel()
	userRepository := newMysqlUserRepository(t)

	existingUser := newEmployeeUser(t)
	err := userRepository.Add(context.Background(), existingUser)
	require.NoError(t, err)

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
```

Add import: `"clean-arch-go/core/errors"` and `"clean-arch-go/domain/errkind"`

**Step 2: Run test to verify it passes**

Run: `make test`
Expected: PASS (existing implementation should work)

**Step 3: Commit**

```bash
git add adapters/mysql_user_repository_test.go
git commit -m "test: add FindById test cases"
```

---

## Task 2: Add deleted_user Test Case to FindById

**Files:**
- Modify: `adapters/mysql_user_repository_test.go`

**Step 1: Add deleted_user case to FindById test**

Add to `testCases` in `TestMysqlUserRepository_FindById`:

```go
		{
			Name:        "deleted_user",
			UUID:        deletedUserUUID,
			ShouldExist: false,
		},
```

Add before `testCases` definition:

```go
	deletedUser := newEmployerUser(t)
	err = userRepository.Add(context.Background(), deletedUser)
	require.NoError(t, err)
	err = userRepository.DeleteByID(context.Background(), deletedUser.UUID())
	require.NoError(t, err)
	deletedUserUUID := deletedUser.UUID()
```

**Step 2: Run test to verify it passes**

Run: `make test`
Expected: PASS

**Step 3: Commit**

```bash
git add adapters/mysql_user_repository_test.go
git commit -m "test: add deleted_user case to FindById"
```

---

## Task 3: Add FindByEmail Tests

**Files:**
- Modify: `adapters/mysql_user_repository_test.go`

**Step 1: Write the test for FindByEmail**

Add after `TestMysqlUserRepository_FindById`:

```go
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
```

**Step 2: Run test to verify it passes**

Run: `make test`
Expected: PASS

**Step 3: Commit**

```bash
git add adapters/mysql_user_repository_test.go
git commit -m "test: add FindByEmail test cases"
```

---

## Task 4: Add FindAll Tests

**Files:**
- Modify: `adapters/mysql_user_repository_test.go`

**Step 1: Write the test for FindAll**

Add after `TestMysqlUserRepository_FindByEmail`:

```go
func TestMysqlUserRepository_FindAll(t *testing.T) {
	t.Parallel()

	t.Run("empty", func(t *testing.T) {
		t.Parallel()
		userRepository := newMysqlUserRepository(t)

		users, err := userRepository.FindAll(context.Background())

		require.NoError(t, err)
		require.Empty(t, users)
	})

	t.Run("multiple_users", func(t *testing.T) {
		t.Parallel()
		userRepository := newMysqlUserRepository(t)

		user1 := newEmployeeUser(t)
		user2 := newEmployerUser(t)
		require.NoError(t, userRepository.Add(context.Background(), user1))
		require.NoError(t, userRepository.Add(context.Background(), user2))

		users, err := userRepository.FindAll(context.Background())

		require.NoError(t, err)
		require.Len(t, users, 2)
	})
}
```

**Step 2: Run test to verify it passes**

Run: `make test`
Expected: PASS

**Step 3: Commit**

```bash
git add adapters/mysql_user_repository_test.go
git commit -m "test: add FindAll test cases"
```

---

## Task 5: Add UpdateByID not_found Test

**Files:**
- Modify: `adapters/mysql_user_repository_test.go`

**Step 1: Add not_found case to UpdateByID**

Refactor `TestMysqlUserRepository_Update` to table-driven and add `not_found` case:

```go
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
```

**Step 2: Run test to verify it passes**

Run: `make test`
Expected: PASS

**Step 3: Commit**

```bash
git add adapters/mysql_user_repository_test.go
git commit -m "test: add not_found case to UpdateByID"
```

---

## Task 6: Add UpdateByID updateFn_error Test

**Files:**
- Modify: `adapters/mysql_user_repository_test.go`

**Step 1: Add updateFn_error case**

Add to `TestMysqlUserRepository_UpdateByID`:

```go
	t.Run("updateFn_error", func(t *testing.T) {
		t.Parallel()
		existingUser := newEmployeeUser(t)
		require.NoError(t, userRepository.Add(context.Background(), existingUser))

		expectedErr := errors.New("update failed")
		err := userRepository.UpdateByID(context.Background(), existingUser.UUID(), func(ctx context.Context, found user.User) (user.User, error) {
			return found, expectedErr
		})

		require.Error(t, err)
		require.True(t, errors.Is(err, expectedErr))
		assertPersistedUserEquals(t, userRepository, existingUser)
	})
```

Add import: `"errors"`

**Step 2: Run test to verify it passes**

Run: `make test`
Expected: PASS

**Step 3: Commit**

```bash
git add adapters/mysql_user_repository_test.go
git commit -m "test: add updateFn_error case to UpdateByID"
```

---

## Task 7: Add DeleteByID Tests

**Files:**
- Modify: `adapters/mysql_user_repository_test.go`

**Step 1: Write the test for DeleteByID**

Add after `TestMysqlUserRepository_UpdateByID`:

```go
func TestMysqlUserRepository_DeleteByID(t *testing.T) {
	t.Parallel()
	userRepository := newMysqlUserRepository(t)

	t.Run("success", func(t *testing.T) {
		t.Parallel()
		existingUser := newEmployeeUser(t)
		require.NoError(t, userRepository.Add(context.Background(), existingUser))

		err := userRepository.DeleteByID(context.Background(), existingUser.UUID())

		require.NoError(t, err)
		_, err = userRepository.FindById(context.Background(), existingUser.UUID())
		require.True(t, errors.Is(errkind.NotExist, err))
	})

	t.Run("not_found", func(t *testing.T) {
		t.Parallel()

		err := userRepository.DeleteByID(context.Background(), "non-existent-uuid")

		require.Error(t, err)
		require.True(t, errors.Is(errkind.NotExist, err))
	})

	t.Run("already_deleted", func(t *testing.T) {
		t.Parallel()
		existingUser := newEmployeeUser(t)
		require.NoError(t, userRepository.Add(context.Background(), existingUser))
		require.NoError(t, userRepository.DeleteByID(context.Background(), existingUser.UUID()))

		err := userRepository.DeleteByID(context.Background(), existingUser.UUID())

		require.Error(t, err)
		require.True(t, errors.Is(errkind.NotExist, err))
	})
}
```

**Step 2: Run test to verify it passes**

Run: `make test`
Expected: PASS

**Step 3: Commit**

```bash
git add adapters/mysql_user_repository_test.go
git commit -m "test: add DeleteByID test cases"
```

---

## Task 8: Add RemoveAll Test

**Files:**
- Modify: `adapters/mysql_user_repository_test.go`

**Step 1: Write the test for RemoveAll**

Add after `TestMysqlUserRepository_DeleteByID`:

```go
func TestMysqlUserRepository_RemoveAll(t *testing.T) {
	t.Parallel()
	userRepository := newMysqlUserRepository(t)

	user1 := newEmployeeUser(t)
	user2 := newEmployerUser(t)
	require.NoError(t, userRepository.Add(context.Background(), user1))
	require.NoError(t, userRepository.Add(context.Background(), user2))

	err := userRepository.RemoveAll(context.Background())

	require.NoError(t, err)
	users, err := userRepository.FindAll(context.Background())
	require.NoError(t, err)
	require.Empty(t, users)
}
```

**Step 2: Run test to verify it passes**

Run: `make test`
Expected: PASS

**Step 3: Commit**

```bash
git add adapters/mysql_user_repository_test.go
git commit -m "test: add RemoveAll test case"
```

---

## Task 9: Add duplicate_id Test to Add

**Files:**
- Modify: `adapters/mysql_user_repository_test.go`

**Step 1: Add duplicate_id case to Add test**

Refactor `TestMysqlUserRepository_Add` to include error case:

```go
func TestMysqlUserRepository_Add(t *testing.T) {
	t.Parallel()

	t.Run("employee_user", func(t *testing.T) {
		t.Parallel()
		userRepository := newMysqlUserRepository(t)

		expectedUser := newEmployeeUser(t)
		err := userRepository.Add(context.Background(), expectedUser)
		require.NoError(t, err)
		assertPersistedUserEquals(t, userRepository, expectedUser)
	})

	t.Run("employer_user", func(t *testing.T) {
		t.Parallel()
		userRepository := newMysqlUserRepository(t)

		expectedUser := newEmployerUser(t)
		err := userRepository.Add(context.Background(), expectedUser)
		require.NoError(t, err)
		assertPersistedUserEquals(t, userRepository, expectedUser)
	})

	t.Run("duplicate_id", func(t *testing.T) {
		t.Parallel()
		userRepository := newMysqlUserRepository(t)

		existingUser := newEmployeeUser(t)
		require.NoError(t, userRepository.Add(context.Background(), existingUser))

		err := userRepository.Add(context.Background(), existingUser)

		require.Error(t, err)
	})
}
```

**Step 2: Run test to verify it passes**

Run: `make test`
Expected: PASS

**Step 3: Commit**

```bash
git add adapters/mysql_user_repository_test.go
git commit -m "test: add duplicate_id case to Add"
```

---

## Summary

| Task | Method | Cases Added |
|------|--------|-------------|
| 1-2 | FindById | found, not_found, deleted_user |
| 3 | FindByEmail | found, not_found |
| 4 | FindAll | empty, multiple_users |
| 5-6 | UpdateByID | not_found, updateFn_error |
| 7 | DeleteByID | success, not_found, already_deleted |
| 8 | RemoveAll | success |
| 9 | Add | duplicate_id |

**Total: 14 new test cases**
