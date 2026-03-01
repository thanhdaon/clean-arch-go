# MySQL User Repository Test Coverage Design

## Summary

Expand test coverage for `MysqlUserRepository` from 4 test cases to 18 test cases, covering all 7 repository methods with happy path and error scenarios.

## Current State

| Method | Test Coverage |
|--------|---------------|
| `Add` | 2 cases (employee, employer) |
| `UpdateByID` | 1 case (update role) |
| `FindById` | Not tested directly |
| `FindByEmail` | Not tested |
| `FindAll` | Not tested |
| `DeleteByID` | Not tested |
| `RemoveAll` | Not tested |

## Proposed Test Cases

### TestMysqlUserRepository_Add (existing + 1 new)

| Case | Description |
|------|-------------|
| `employee_user` | Add user with employee role |
| `employer_user` | Add user with employer role |
| `duplicate_id` | Add same user twice → error |

### TestMysqlUserRepository_FindById (new - 3 cases)

| Case | Description |
|------|-------------|
| `found` | Find existing user, verify all fields |
| `not_found` | Find non-existent ID → `errkind.NotExist` |
| `deleted_user` | Find soft-deleted user → `errkind.NotExist` |

### TestMysqlUserRepository_FindByEmail (new - 2 cases)

| Case | Description |
|------|-------------|
| `found` | Find by email, verify user |
| `not_found` | Non-existent email → `errkind.NotExist` |

### TestMysqlUserRepository_FindAll (new - 2 cases)

| Case | Description |
|------|-------------|
| `empty` | No users → empty slice |
| `multiple_users` | Multiple users persisted → all returned |

### TestMysqlUserRepository_UpdateByID (existing + 2 new)

| Case | Description |
|------|-------------|
| `update_role` | Update user role (employer→employee) |
| `not_found` | Update non-existent ID → `errkind.NotExist` |
| `updateFn_error` | updateFn returns error → transaction rollback |

### TestMysqlUserRepository_DeleteByID (new - 3 cases)

| Case | Description |
|------|-------------|
| `success` | Delete existing user, verify FindById fails |
| `not_found` | Delete non-existent ID → `errkind.NotExist` |
| `already_deleted` | Delete twice → second fails |

### TestMysqlUserRepository_RemoveAll (new - 1 case)

| Case | Description |
|------|-------------|
| `success` | Add users, RemoveAll, verify empty |

## Total Coverage

- **Existing:** 4 test cases
- **New:** 14 test cases
- **Total:** 18 test cases

## Approach

Method-based test functions with table-driven sub-tests, matching existing code style.
