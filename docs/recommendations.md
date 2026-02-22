# Recommended Fixes

This document describes known bugs, design issues, and missing critical infrastructure in the current codebase. Items are ordered by priority — fix higher-priority items first, as some lower-priority items depend on them.

---

## Priority 1 — Blockers (Fix First)

These issues make core functionality broken or the API unusable in practice.

---

### R1. Fix `FindTasksForUser` stub (breaks employee task listing)

**File:** `adapters/mysql_task_repository.go`

**Problem:** The `FindTasksForUser` method always returns a hardcoded `fmt.Errorf("dump")` error. This means `GET /api/tasks` always fails for any employee-role user — the primary read use case for employees is completely broken.

**Fix:**

```go
func (r *MysqlTaskRepository) FindTasksForUser(ctx context.Context, userUUID string) ([]query.Task, error) {
    var tasks []taskRow
    err := r.db.SelectContext(ctx, &tasks, "SELECT * FROM tasks WHERE assigned_to = ?", userUUID)
    if err != nil {
        return nil, errors.E(errors.Op("MysqlTaskRepository.FindTasksForUser"), errkind.Internal, err)
    }
    // map taskRow → query.Task ...
    return result, nil
}
```

**Also required:** Add a database index on `tasks.assigned_to`:

```sql
ALTER TABLE tasks ADD INDEX idx_tasks_assigned_to (assigned_to);
```

---

### R2. Fix inverted role check in `task.AssignTo` (tasks cannot be assigned to employees)

**File:** `domain/task/task.go`

**Problem:** The `AssignTo` method guards on `assignee.Role() != RoleEmployer`, meaning only employers can be task assignees. This is the opposite of the intended behaviour — employees should be the ones receiving task assignments.

**Current code:**

```go
func (t *task) AssignTo(assigner user.User, assignee user.User) error {
    if assignee.Role() != user.RoleEmployer {
        return errors.New("only employer role can assign task")
    }
    // ...
}
```

**Fix:**

```go
func (t *task) AssignTo(assigner user.User, assignee user.User) error {
    if assigner.Role() != user.RoleEmployer {
        return errors.E(errors.Op("task.AssignTo"), errkind.Permission,
            errors.Str("only an employer can assign tasks"))
    }
    if assignee.Role() != user.RoleEmployee {
        return errors.E(errors.Op("task.AssignTo"), errkind.Permission,
            errors.Str("tasks can only be assigned to employees"))
    }
    t.assignedTo = assignee.UUID()
    t.updatedAt = time.Now()
    return nil
}
```

**Note:** Update all tests that currently expect the inverted behaviour.

---

### R3. Fix `CreateTask` HTTP handler ignoring request body (title is always empty)

**File:** `ports/openapi_handler.go`

**Problem:** The `CreateTask` handler never decodes the request body, so `Title` is always passed as an empty string to the command. The domain rejects empty titles, meaning `POST /api/tasks` always returns an error.

**Current code (approx.):**

```go
err = h.app.Commands.CreateTask.Handle(ctx, command.CreateTask{
    Title:   "",   // ← body never read
    Creator: user,
})
```

**Fix:** Decode the request body before calling the command:

```go
var body PostTask
if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
    responseBadRequest(w, r, err)
    return
}
err = h.app.Commands.CreateTask.Handle(ctx, command.CreateTask{
    Title:   body.Title,
    Creator: user,
})
```

---

### R4. Fix `AssignTask` HTTP handler silently ignoring errors

**File:** `ports/openapi_handler.go`

**Problem:** The `AssignTask` handler calls the command, assigns the result to `err`, but then never checks it. Any domain error, not-found error, or persistence failure is silently swallowed and a 200 OK is returned.

**Fix:** Add the standard error check after the command call:

```go
err = h.app.Commands.AssignTask.Handle(ctx, command.AssignTask{...})
if err != nil {
    responseError(w, r, err)
    return
}
responseSuccess(r.Context(), w, r)
```

---

## Priority 2 — Missing Critical Infrastructure

These are absent features that are essential for the API to be usable in any real environment.

---

### R5. Add a Login / Token Issuance endpoint

**Problem:** There is no `POST /api/auth/login` or equivalent endpoint. The JWT middleware expects a valid `Authorization: Bearer <token>` header, but there is no HTTP mechanism to obtain such a token. This makes the entire API inaccessible without out-of-band token generation.

**Recommended approach:**

1. Add a `UserRepository.FindById` call in a new `Login` command (user already exists in DB).
2. Sign a JWT with `user_uuid` and `user_role` claims using `JWT_SECRET`.
3. Add `exp` claim (e.g., 24h expiry).
4. Return the signed token in the response body.

**Proposed endpoint:** `POST /api/auth/login`

**Request body:** `{ "user_id": "<ulid>" }`

**Response:** `{ "token": "<jwt>" }`

---

### R6. Enforce JWT expiry (`exp` claim)

**File:** `core/auth/auth.go`

**Problem:** The `VerifyIDToken` function parses and validates the JWT signature but does not check the `exp` (expiry) claim. Tokens never expire, which is a security issue.

**Fix:** Use `jwt.WithExpirationRequired()` or manually check the `exp` claim after parsing:

```go
token, err := jwt.Parse(tokenStr, keyFunc,
    jwt.WithValidMethods([]string{"HS256"}),
    jwt.WithExpirationRequired(),
)
```

---

### R7. Return 401 for unauthenticated requests (not silent pass-through)

**File:** `ports/openapi_auth.go`

**Problem:** The auth middleware is permissive — if the `Authorization` header is missing or the token is invalid, the request is passed to the next handler without a user in context. Protected handlers then fail with a 500 or similar when they try to read the user from context. This gives misleading error responses.

**Fix:** Return `401 Unauthorized` immediately in the middleware when no valid token is present:

```go
if err != nil || !token.Valid {
    http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
    return
}
```

Only unauthenticated routes (e.g., `POST /api/auth/login`, `POST /api/users`) should be excluded from this check.

---

### R8. Guard `POST /api/users` (open user registration)

**File:** `ports/openapi_handler.go`

**Problem:** `POST /api/users` requires no authentication. Any anonymous caller can register a user with any role, including `employer`. There is no access control or invitation mechanism.

**Recommended approaches (pick one):**

- **Option A (simple):** Require a valid employer JWT to create any user.
- **Option B (admin role):** Introduce a third role `admin` that is the only role allowed to create users.
- **Option C (invitation flow):** Require a pre-generated invitation token.

---

## Priority 3 — Correctness & Performance

These issues do not block functionality today but will cause problems under real load or when queried at scale.

---

### R9. Add `LIMIT` / `OFFSET` pagination to `AllTasks`

**File:** `adapters/mysql_task_repository.go`

**Problem:** `AllTasks` executes `SELECT * FROM tasks` with no limit. As the task table grows, this will return unbounded result sets, consuming excessive memory and degrading response times.

**Fix:** Accept `limit` and `offset` parameters and propagate them from the query handler through to the SQL:

```sql
SELECT * FROM tasks ORDER BY created_at DESC LIMIT ? OFFSET ?
```

---

### R10. Add database indexes for common filter columns

**Problem:** There are no indexes on `tasks.status`, `tasks.assigned_to`, `tasks.created_by`, or `users.role`. Any filtering query on these columns will result in a full table scan.

**Fix:** Add a migration:

```sql
ALTER TABLE tasks ADD INDEX idx_tasks_status (status);
ALTER TABLE tasks ADD INDEX idx_tasks_assigned_to (assigned_to);
ALTER TABLE tasks ADD INDEX idx_tasks_created_by (created_by);
ALTER TABLE users ADD INDEX idx_users_role (role);
```

---

## Priority 4 — Domain Model Quality

These are bugs or design weaknesses in the domain layer itself.

---

### R11. Fix typo in error message in `task.NewTask`

**File:** `domain/task/task.go`

**Problem:** The error message reads `"only employ can create task"` — "employ" should be "employer".

**Fix:**

```go
return errors.E(errors.Op("task.NewTask"), errkind.Permission,
    errors.Str("only employer can create task"))
```

---

### R12. Fix typo in `errkind.Connection` slug

**File:** `domain/errkind/errkind.go`

**Problem:** The `Connection` kind has the slug `"pconnection error"` (extra leading `p`). This affects log output and any code that string-matches on error kind values.

**Fix:**

```go
Connection Kind = "connection error"
```

---

### R13. Enforce valid status transitions (state machine)

**File:** `domain/task/task.go`

**Problem:** `ChangeStatus` allows any status → any status transition, including backwards moves like `completed → todo`. This violates typical task lifecycle semantics and can mask application bugs.

**Recommended transition graph:**

```
todo → pending → inprogress → completed
```

With an explicit re-open path if needed:

```
completed → todo  (only if caller is employer)
```

**Implementation:** Add a `validTransitions` map and check it inside `ChangeStatus`:

```go
var validTransitions = map[Status][]Status{
    StatusTodo:       {StatusPending},
    StatusPending:    {StatusInProgress, StatusTodo},
    StatusInProgress: {StatusCompleted, StatusPending},
    StatusCompleted:  {StatusTodo}, // re-open by employer only
}
```

---

### R14. Validate `user.ChangeRole` input

**File:** `domain/user/user.go`

**Problem:** `ChangeRole(role Role)` accepts any `Role` value including `RoleUnknow` (the zero value) with no validation. A caller can silently corrupt a user's role to an invalid state.

**Fix:**

```go
func (u *user) ChangeRole(role Role) error {
    if role.IsZero() {
        return errors.New("cannot change role to unknown")
    }
    u.role = role
    return nil
}
```

**Note:** This changes the method signature — update the `User` interface and all callers.

---

## Fix Sequencing Summary

| # | Fix | Priority | Effort | Depends On |
|---|-----|----------|--------|------------|
| R1 | Fix `FindTasksForUser` stub | P1 | Low | — |
| R2 | Fix inverted `AssignTo` role check | P1 | Low | — |
| R3 | Fix `CreateTask` body not read | P1 | Low | — |
| R4 | Fix `AssignTask` error ignored | P1 | Low | — |
| R5 | Add Login / JWT issuance endpoint | P2 | Medium | — |
| R6 | Enforce JWT `exp` claim | P2 | Low | R5 |
| R7 | Return 401 for unauthenticated requests | P2 | Low | R5 |
| R8 | Guard open user registration | P2 | Low | R5 |
| R9 | Add pagination to `AllTasks` | P3 | Medium | — |
| R10 | Add DB indexes | P3 | Low | — |
| R11 | Fix typo in task error message | P4 | Trivial | — |
| R12 | Fix typo in `errkind.Connection` | P4 | Trivial | — |
| R13 | Enforce status state machine | P4 | Medium | — |
| R14 | Validate `user.ChangeRole` input | P4 | Low | — |
