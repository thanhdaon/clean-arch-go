# User Acceptance Tests Implementation Design

**Date:** 2026-03-06  
**Status:** Approved  
**Related Files:** `acceptanceTests/user.txt`, `tests/user_test.go`

## Overview

Implement 13 skipped acceptance tests in `tests/user_test.go` following Test-Driven Development (TDD). Remove all `t.Skip()` calls and implement test logic to match scenarios defined in `acceptanceTests/user.txt`. After tests are written, fix API endpoints incrementally to make all tests pass.

**Goal:** Achieve 100% acceptance test coverage for user management scenarios defined in user.txt (lines 1-351).

## Current State

The existing test suite in `tests/user_test.go` has 13 skipped tests marked with `t.Skip("not yet implemented: ...")`. These tests correspond to scenarios in `acceptanceTests/user.txt` that verify:

- Authentication edge cases
- Input validation
- Authorization and access control

## Test Categories

### Authentication Tests (2 tests)

1. **user.txt:47** - Login with missing password returns 400 error
   - Currently skipped: "API returns 401 instead of 400 for missing password"
   - Expected: HTTP 400 Bad Request when password is empty

### Input Validation Tests (6 tests)

2. **user.txt:95** - Create user with duplicate email returns 409 conflict
   - Currently skipped: "API allows duplicate emails"
   - Expected: HTTP 409 Conflict when email already exists

3. **user.txt:104** - Create user with invalid email format returns 400 error
   - Currently skipped: "API does not validate email format"
   - Expected: HTTP 400 Bad Request for malformed emails

4. **user.txt:113** - Create user with missing required fields returns 400 error
   - Currently skipped: "API allows missing name field"
   - Expected: HTTP 400 Bad Request when name is missing

5. **user.txt:122** - Create user with empty name returns 400 error
   - Currently skipped: "API allows empty name"
   - Expected: HTTP 400 Bad Request when name is ""

6. **user.txt:131** - Create user with weak password returns 400 error
   - Currently skipped: "API does not validate password strength"
   - Expected: HTTP 400 Bad Request for passwords < 8 characters

7. **user.txt:233** - Update with empty name returns 400 error
   - Currently skipped: "API allows empty name in profile update"
   - Expected: HTTP 400 Bad Request when updating name to ""

### Authorization Tests (5 tests)

8. **user.txt:151** - Non-admin lists users returns forbidden
   - Currently skipped: "API does not enforce admin-only restriction for listing users"
   - Expected: HTTP 403 Forbidden for non-admin users

9. **user.txt:192** - User gets another user's profile returns forbidden
   - Currently skipped: "API allows users to view other users' profiles"
   - Expected: HTTP 403 Forbidden when accessing another user's profile

10. **user.txt:223** - Non-admin updates another user's profile returns forbidden
    - Currently skipped: "API allows non-admin users to update other users' profiles"
    - Expected: HTTP 403 Forbidden for non-admin updating another user

11. **user.txt:263** - Non-admin deletes another user returns forbidden
    - Currently skipped: "API allows non-admin users to delete other users"
    - Expected: HTTP 403 Forbidden for non-admin deleting another user

12. **user.txt:315** - Non-admin changes role returns forbidden
    - Currently skipped: "API does not enforce admin-only restriction for role changes"
    - Expected: HTTP 403 Forbidden for non-admin changing roles

13. **user.txt:335** - Change role to invalid value returns 400 error
    - Currently skipped: "API returns 500 for invalid role value"
    - Expected: HTTP 400 Bad Request for invalid role enum values

## Test Implementation Pattern

Each test will follow this consistent structure:

```go
t.Run("user.txt:XX - Description", func(t *testing.T) {
    t.Parallel()
    
    // Setup - create unique test data
    email := fmt.Sprintf("test-%d@example.com", time.Now().UnixNano())
    
    // Execute action using existing helper functions
    resp, body := someHelper(t, payload)
    
    // Assert HTTP status and response body
    assert.Equal(t, http.StatusOK, resp.StatusCode)
    assertResponseStatus(t, body, http.StatusBadRequest)
})
```

All tests use existing helper functions from `tests/helpers_test.go`:
- `postUser()` - Create users
- `login()` - Authenticate users
- `getUser()` - Fetch user profiles
- `patchUser()` - Update users
- `deleteUser()` - Delete users
- `putUserRole()` - Change user roles
- `getUsers()` - List all users
- `createUserWithRoleAndGetToken()` - Setup authenticated users

New helpers may be added to `helpers_test.go` if needed for duplicate email testing.

## API Fixes Required

After implementing all tests, these API changes will be needed:

### Validation Middleware

**Email Format Validation:**
- Add regex validation for email format
- Return 400 Bad Request for invalid formats
- Location: HTTP handler or domain validation

**Password Strength Validation:**
- Minimum 8 characters
- Return 400 Bad Request for weak passwords
- Location: User creation endpoint

**Required Field Validation:**
- Validate name is not empty or missing
- Return 400 Bad Request for validation failures
- Location: User creation and update endpoints

**Unique Email Constraint:**
- Check email uniqueness before user creation
- Return 409 Conflict for duplicate emails
- Location: Database adapter or application layer

### Authorization Checks

**Admin-Only Endpoints:**
- `GET /api/users` - List all users (admin only)
- `PUT /api/users/:id/role` - Change user role (admin only)

**Self-or-Admin Endpoints:**
- `GET /api/users/:id` - Get user profile (self or admin)
- `PATCH /api/users/:id` - Update user profile (self or admin)
- `DELETE /api/users/:id` - Delete user (self or admin)

**Implementation:**
- Add role checks in HTTP handlers
- Compare authenticated user ID with requested resource ID
- Return 403 Forbidden for unauthorized access

### Role Enum Validation

**Valid Roles:**
- admin
- employer
- employee

**Implementation:**
- Add validation in `PUT /api/users/:id/role` endpoint
- Return 400 Bad Request for invalid role values
- Prevent 500 errors from invalid enum values

### Database Schema

**Unique Constraint:**
```sql
ALTER TABLE users ADD UNIQUE INDEX idx_email (email);
```

**Error Handling:**
- Catch MySQL duplicate entry errors (error code 1062)
- Return 409 Conflict instead of 500 Internal Server Error

## Testing Strategy

### Phase 1: Implement Tests (TDD)

1. Remove all `t.Skip()` calls from 13 tests
2. Implement test logic for each scenario
3. Run `make test` to verify all 13 tests fail as expected
4. Document failure modes for each test

### Phase 2: Fix API Endpoints

**Iteration 1 - Validation (6 tests):**
- Implement email format validation
- Implement password strength validation
- Implement required field validation
- Implement unique email constraint
- Run tests, verify validation tests pass

**Iteration 2 - Authorization (5 tests):**
- Add admin-only checks to list users endpoint
- Add self-or-admin checks to get/update/delete endpoints
- Add admin-only checks to role change endpoint
- Run tests, verify authorization tests pass

**Iteration 3 - Authentication (2 tests):**
- Add empty password validation to login endpoint
- Run tests, verify all tests pass

### Phase 3: Verification

1. Run full test suite: `make test`
2. Verify all 13 previously skipped tests now pass
3. Check test coverage: `go test -cover ./tests`
4. Verify no regressions in existing tests

## Success Criteria

- ✅ All 13 skipped tests implemented and passing
- ✅ No regressions in existing tests
- ✅ API returns correct HTTP status codes:
  - 200 OK for successful operations
  - 400 Bad Request for validation errors
  - 401 Unauthorized for authentication failures
  - 403 Forbidden for authorization failures
  - 404 Not Found for missing resources
  - 409 Conflict for duplicate resources
- ✅ Test coverage remains in high 90s for line and branch
- ✅ All acceptance test scenarios in user.txt are covered

## Implementation Order

1. Implement all 13 tests (no API changes yet)
2. Run tests, document failures
3. Fix validation tests (6 tests)
4. Fix authorization tests (5 tests)
5. Fix authentication tests (2 tests)
6. Final verification and coverage check

## Related Files

- `acceptanceTests/user.txt` - BDD scenarios (source of truth)
- `tests/user_test.go` - Test implementation
- `tests/helpers_test.go` - Test helper functions
- `tests/setup_test.go` - Test fixtures and setup
- `tests/component_test.go` - Test runner
- `ports/api.go` - HTTP handlers (authorization logic)
- `domain/user/user.go` - Domain validation logic
- `adapters/mysql_user_repository.go` - Database operations

## References

- README.md - Testing architecture and BDD/TDD workflow
- AGENTS.md - Development guidelines and code quality standards