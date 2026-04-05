# Acceptance Test Framework for clean-arch-go

## Overview

**When to read this file:** When writing or translating acceptance test scenarios from `.feature` files into Go component tests in `tests/`. Also read when debugging a test that doesn't match the intent of its `.feature` source.

---

## Principle

Two-artifact approach — `.feature` files are the source of truth, component tests are the runnable form:

```
.feature (Gherkin spec) → component_test.go (tests/) → go test
```

Gherkin Feature files describe behaviour in domain language using standard `Given/When/Then` syntax. Component tests at the HTTP layer are written to mirror those scenarios exactly. The source filename and first `Given` line number are embedded in each test name for traceability.

This fulfills the BDD/TDD workflow described in `AGENTS.md`:
> Write a custom parser/runner as glue code connecting scenario language to production code.

No parser or generator tooling is required — the `.txt` files serve as living specification that developers (or AI) translate directly into component tests.

---

## Where Tests Fit

| Layer | Package | Test Type | Focus |
|---|---|---|---|
| Domain | `domain/` | Unit | Business rules, value objects |
| Application | `app/command/` | Unit | Command handler logic |
| Adapters | `adapters/` | Integration | MySQL persistence |
| HTTP | `tests/` | **Component** | Behaviour (GIVEN/WHEN/THEN) |

Component tests sit at the **HTTP layer** — they exercise the full stack through real HTTP requests and responses, providing end-to-end confidence without needing a deployed environment.

---

## Directory Structure

```
features/                # Gherkin .feature spec files (never auto-modified)
  task.feature
  user.feature
  comment.feature
  activity.feature
  tag.feature
tests/
  helpers_test.go         # HTTP helpers and DB assertion functions
  setup_test.go           # SetupComponentTest and TestFixtures
  component_test.go       # Main test runner with t.Run blocks
  task_test.go            # Task-related test functions
  user_test.go            # User-related test functions
  comment_test.go         # Comment-related test functions
```

---

## Design Principle

Prefer tests based on **observable behaviour** — assert what the HTTP API returns and what is persisted in the database, not what the command handler does internally.

For THEN assertions, prefer:
- HTTP response status codes
- Database state via the existing DB assertion helpers in `tests/helpers_test.go`
- HTTP response body fields for read operations

Avoid asserting on internal domain objects or anything not visible through the HTTP surface.

---

## Execution

No parser or generator tooling exists. The `.txt` file is the specification; tests in `tests/` are written by hand (or by AI) to mirror each scenario exactly.

Each test uses `SetupComponentTest(t)` which returns a `*TestFixtures` containing `DB`, `AuthToken`, `VideoService`, and `Cancel`. The server runs on `localhost:8080` (or `$PORT`). Every test function is isolated by scoping its own users, tasks, and other data via API calls — never share state between tests.

**Test name convention:**
```
"<source-file>:<first-Given-line> - <description>"
```
Example: `"task.feature:8 - Admin creates task"`

---

## File Format (Gherkin)

Gherkin files use standard `.feature` extension with the following structure:

```gherkin
Feature: Task lifecycle and permissions

  # Description of this test case.
  Scenario: Admin creates task
    Given an admin user is authenticated.
    Given a task exists with status todo.
    Given the task is assigned to the user.

    When the admin creates a task with title "New Task".

    Then no error is returned.
    And the task exists in the database with status "todo".
```

### Structure Rules

- `Feature:` is the top-level keyword — one per file, describes the domain
- `Scenario:` marks each independent test case
- `Given`, `When`, `Then` are case-sensitive keywords (Title Case)
- `And` continues the preceding step type (Given/When/Then)
- Lines starting with `#` are comments
- Blank lines separate scenarios for readability
- Each `Scenario:` block becomes one `t.Run` subtest

---

## GIVEN Directives

GIVEN directives set up state via API calls before the action under test. Use the admin token (`f.AuthToken` from `SetupComponentTest`) for all setup calls unless the test is specifically about role restriction during setup.

### Users

```
GIVEN a user exists with role employee.
GIVEN a user exists with role employer.
GIVEN a user exists with role admin.
```

```go
email := fmt.Sprintf("user-%s@test.com", uuid.New())
resp, body := postUser(t, map[string]any{
    "role": "employee", "name": "Test User",
    "email": email, "password": "password123",
})
require.Equal(t, http.StatusOK, resp.StatusCode, "setup failed: %s", string(body))

loginResp, loginBody := login(t, map[string]any{"email": email, "password": "password123"})
require.Equal(t, http.StatusOK, loginResp.StatusCode, "login failed: %s", string(loginBody))

var loginResult struct {
    Token string `json:"token"`
}
require.NoError(t, json.Unmarshal(loginBody, &loginResult))
userToken := loginResult.Token

var userID string
require.NoError(t, f.DB.Get(&userID, "SELECT id FROM users WHERE email = ? AND deleted_at IS NULL", email))
```

Use `createUserAndGetID(t, f.DB)` for quick setup when the user token is not needed.

Use `createUserWithRoleAndGetToken(t, f.DB, "employer")` when you need both userID and token — returns `(userID, token string)`.

### Tasks

```
GIVEN a task exists with status todo.
```

```go
creatorID := createUserAndGetID(t, f.DB)
taskID := createTaskAndGetID(t, f.DB, f.AuthToken, creatorID)
```

Tasks are always created with `status = todo`. To reach a different status, advance through valid transitions:

```
GIVEN a task exists with status in_progress.
```

```go
creatorID := createUserAndGetID(t, f.DB)
taskID := createTaskAndGetID(t, f.DB, f.AuthToken, creatorID)
resp, body := putTaskStatus(t, f.AuthToken, taskID, map[string]any{"status": "in_progress"})
require.Equal(t, http.StatusOK, resp.StatusCode, "setup failed: %s", string(body))
```

Valid multi-hop setup for `completed`:
```go
// todo → in_progress → completed
creatorID := createUserAndGetID(t, f.DB)
taskID := createTaskAndGetID(t, f.DB, f.AuthToken, creatorID)
resp, _ := putTaskStatus(t, f.AuthToken, taskID, map[string]any{"status": "in_progress"})
require.Equal(t, http.StatusOK, resp.StatusCode)
resp, _ = putTaskStatus(t, f.AuthToken, taskID, map[string]any{"status": "completed"})
require.Equal(t, http.StatusOK, resp.StatusCode)
```

`completed` is a terminal state — no further transitions are valid from it.

### Assignment

```
GIVEN the task is assigned to the user.
```

```go
resp, body := assignTask(t, f.AuthToken, taskID, userID)
require.Equal(t, http.StatusOK, resp.StatusCode, "setup failed: %s", string(body))
```

**Note:** The assignee must have `employer` role. If a `.txt` scenario describes "assigning to an employee", the user should still be created with `employer` role — the domain enforces this constraint.

### Tags

```
GIVEN the task has tag "backend".
```

```go
resp, body := addTaskTag(t, f.AuthToken, taskID, map[string]any{"name": "backend"})
require.Equal(t, http.StatusOK, resp.StatusCode, "setup failed: %s", string(body))
tagID := getTaskTagIDByName(t, f.DB, taskID, "backend")
```

### Comments

```
GIVEN a comment exists on the task.
```

```go
resp, body := addComment(t, userToken, taskID, map[string]any{"content": "test comment"})
require.Equal(t, http.StatusOK, resp.StatusCode, "setup failed: %s", string(body))
commentID := getCommentIDByContent(t, f.DB, taskID, "test comment")
```

### Priority

```
GIVEN the task has priority high.
```

```go
resp, body := putTaskPriority(t, f.AuthToken, taskID, map[string]any{"priority": "high"})
require.Equal(t, http.StatusOK, resp.StatusCode, "setup failed: %s", string(body))
```

Valid values: `"low"`, `"medium"`, `"high"`, `"urgent"`.

### Description

```
GIVEN the task has a description "Fix the login bug".
```

```go
resp, body := patchTaskDescription(t, f.AuthToken, taskID, map[string]any{"description": "Fix the login bug"})
require.Equal(t, http.StatusOK, resp.StatusCode, "setup failed: %s", string(body))
```

---

## WHEN Directives

WHEN directives perform the action under test. Capture the HTTP response — THEN assertions use it. All helper functions return `(*http.Response, []byte)`.

### Task Status Change

```
WHEN the employee changes the task status to in-progress.
```

```go
resp, body := putTaskStatus(t, userToken, taskID, map[string]any{"status": "in_progress"})
```

Status string values: `"todo"`, `"pending"`, `"in_progress"`, `"completed"`.

### Task Creation

```
WHEN the employer creates a task with title "Fix the login bug".
```

```go
resp, body := createTask(t, empToken, map[string]any{"title": "Fix the login bug"})
newTaskID := getTaskIDByTitle(t, f.DB, "Fix the login bug")
```

### Task Assignment

```
WHEN the employer assigns the task to the employee.
```

```go
resp, body := assignTask(t, empToken, taskID, userID)
```

```
WHEN the employee unassigns themselves from the task.
```

```go
resp, body := unassignTask(t, userToken, taskID)
```

### Comments

```
WHEN the user adds a comment "I fixed this.".
```

```go
resp, body := addComment(t, userToken, taskID, map[string]any{"content": "I fixed this."})
commentID := getCommentIDByContent(t, f.DB, taskID, "I fixed this.")
```

```
WHEN the user updates the comment to "Actually I didn't.".
```

```go
resp, body := updateComment(t, userToken, taskID, commentID, map[string]any{"content": "Actually I didn't."})
```

```
WHEN the user deletes the comment.
```

```go
resp, body := deleteComment(t, userToken, taskID, commentID)
```

### Archive, Reopen, Delete

```
WHEN the employer archives the task.
WHEN the employer reopens the task.
WHEN the employer deletes the task.
```

```go
resp, body := archiveTask(t, empToken, taskID)
resp, body := reopenTask(t, empToken, taskID)
resp, body := deleteTask(t, empToken, taskID)
```

### Priority and Description

```
WHEN the employer sets the task priority to high.
```

```go
resp, body := putTaskPriority(t, empToken, taskID, map[string]any{"priority": "high"})
```

```
WHEN the employer updates the task description to "New description".
```

```go
resp, body := patchTaskDescription(t, empToken, taskID, map[string]any{"description": "New description"})
```

### User Role

```
WHEN the admin changes the user's role to employer.
```

```go
resp, body := putUserRole(t, f.AuthToken, userID, map[string]any{"role": "employer"})
```

---

## THEN Directives

### HTTP Status Assertions

**Important:** This API returns HTTP 200 with the actual status code in the JSON response body:
```json
{"status":401,"error":"http.Login: authorization error: ...","trace_id":"..."}
```

Use `assertResponseStatus` to check error status codes from the response body, not `resp.StatusCode`.

```
THEN no error is returned.
THEN the request succeeds.
```

```go
require.Equal(t, http.StatusOK, resp.StatusCode, "unexpected status: %s", string(body))
```

```
THEN an error is returned.
```

```go
assert.Equal(t, http.StatusOK, resp.StatusCode)
assert.GreaterOrEqual(t, int(getResponseStatus(body)), 400)
```

```
THEN a forbidden error is returned.
THEN a not found error is returned.
THEN a bad request error is returned.
THEN an unauthorized error is returned.
```

```go
assert.Equal(t, http.StatusOK, resp.StatusCode)
assertResponseStatus(t, body, http.StatusForbidden)
assertResponseStatus(t, body, http.StatusNotFound)
assertResponseStatus(t, body, http.StatusBadRequest)
assertResponseStatus(t, body, http.StatusUnauthorized)
```

### Task State Assertions

```
THEN the task has status in_progress.
THEN the task still has status todo.
```

```go
assertTaskFieldEquals[string](t, f.DB, taskID, "status", "in_progress")
```

```
THEN the task is assigned to the user.
```

```go
assertTaskFieldEquals[string](t, f.DB, taskID, "assigned_to", userID)
```

```
THEN the task is not assigned.
```

```go
assertTaskFieldEmpty[string](t, f.DB, taskID, "assigned_to")
```

```
THEN the task is archived.
THEN the task is deleted.
THEN the task exists.
```

```go
assertTaskArchived(t, f.DB, taskID)
assertTaskDeletedInDB(t, f.DB, taskID)
assertTaskExistsInDB(t, f.DB, taskID)
```

### Activity Assertions

```
THEN the task activity log includes a status_changed event.
```

```go
resp, body := getTaskActivity(t, f.AuthToken, taskID)
var result struct {
    Activities []struct {
        Type string `json:"type"`
    } `json:"activities"`
}
require.NoError(t, json.Unmarshal(body, &result))
types := make([]string, len(result.Activities))
for i, a := range result.Activities {
    types[i] = a.Type
}
assert.Contains(t, types, "status_changed")
```

Activity type values: `status_changed`, `assigned`, `unassigned`, `title_updated`, `priority_changed`, `due_date_set`, `description_set`, `archived`, `reopened`, `comment_added`, `comment_updated`, `comment_deleted`.

### Comment Assertions

```
THEN the comment exists on the task.
THEN the comment is deleted.
```

```go
assertCommentExistsInDB(t, f.DB, commentID)
assertCommentDeletedInDB(t, f.DB, commentID)
```

### Response Body Assertions

For read operations that return a list:

```
THEN the task list contains the task.
```

```go
resp, body := getTasks(t, userToken)
var result struct {
    Tasks []struct{ UUID string `json:"uuid"` } `json:"tasks"`
}
require.NoError(t, json.Unmarshal(body, &result))
ids := make([]string, len(result.Tasks))
for i, tsk := range result.Tasks { ids[i] = tsk.UUID }
assert.Contains(t, ids, taskID)
```

---

## Authorization Reference

| Operation | admin | employer | employee |
|-----------|:-----:|:--------:|:--------:|
| Create task | ✓ | ✓ | ✗ |
| Assign task | ✓ | ✓ | ✗ |
| Change task status | ✓ | ✓ | ✓ (if assigned) |
| Update task title | ✓ | ✓ | ✓ (if assigned) |
| Delete task | ✓ | ✓ | ✗ |
| Archive task | ✓ | ✓ | ✗ |
| Reopen task | ✓ | ✓ | ✗ |
| Set priority | ✓ | ✓ | ✗ |
| Set description | ✓ | ✓ | ✓ (if assigned) |
| Add comment | ✓ | ✓ | ✓ (if assigned) |
| Edit own comment | ✓ | ✓ | ✓ |
| Delete own comment | ✓ | ✓ | ✓ |
| Change user role | ✓ | ✗ | ✗ |

**Employee visibility:** Employees only see tasks assigned to them. `GET /api/tasks` returns an empty list for an employee with no assigned tasks.

---

## Status Transition Reference

```
todo        → [pending, in_progress]
pending     → [todo, in_progress]
in_progress → [completed, pending]
completed   → [] (terminal)
```

Invalid transitions return an error response. When writing `GIVEN a task exists with status X`, use the minimum number of valid hops from `todo` to reach the required status.

---

## Common Pitfalls

### Employee Visibility Restriction

Employees only see tasks assigned to them. A test that creates a task and reads it as an employee (without assigning it first) will see an empty list — not a 404.

**Fix:** Assign the task before the employee accesses it, or use an employer/admin token for reads that are not the subject of the test.

### Assignee Role Restriction

`AssignTask` requires the assignee to have `employer` role. If a `.txt` scenario says "assign to an employee", create the user with `employer` role regardless.

### Multi-Hop Status Setup

Ensure each intermediate transition is valid:

```go
// ✓ valid: todo → in_progress
creatorID := createUserAndGetID(t, f.DB)
taskID := createTaskAndGetID(t, f.DB, f.AuthToken, creatorID)
resp, _ := putTaskStatus(t, f.AuthToken, taskID, map[string]any{"status": "in_progress"})
require.Equal(t, http.StatusOK, resp.StatusCode)

// ✓ valid: todo → in_progress → completed
creatorID := createUserAndGetID(t, f.DB)
taskID := createTaskAndGetID(t, f.DB, f.AuthToken, creatorID)
resp, _ := putTaskStatus(t, f.AuthToken, taskID, map[string]any{"status": "in_progress"})
require.Equal(t, http.StatusOK, resp.StatusCode)
resp, _ = putTaskStatus(t, f.AuthToken, taskID, map[string]any{"status": "completed"})
require.Equal(t, http.StatusOK, resp.StatusCode)

// ✗ invalid: todo → completed (not a direct transition — returns 422)
resp, _ := putTaskStatus(t, f.AuthToken, taskID, map[string]any{"status": "completed"})
// resp.StatusCode will be 422 or 400
```

### GIVEN Always Uses Admin Token

Setup GIVEN calls use `f.AuthToken` (admin) to avoid role restriction failures during setup. Only the WHEN directive uses the actor's token.

### Body Decoding

Helpers return `[]byte` — decode it with `json.Unmarshal`:

```go
resp, body := createTask(t, token, payload)
var result map[string]any
require.NoError(t, json.Unmarshal(body, &result))
require.Equal(t, http.StatusOK, resp.StatusCode, "unexpected status: %s", string(body))
assert.Equal(t, "expected", result["field"])
```

### "and" Continuations

Lines starting with `and` are additional THEN assertions in the same `t.Run` block:

```
THEN an error is returned.
and the task still has status todo.
```

```go
assert.GreaterOrEqual(t, resp.StatusCode, 400)
assertTaskFieldEquals[string](t, f.DB, taskID, "status", "todo")
```

### Unique Identifiers — Don't Depend on Cleanup

Write tests so they don't depend on cleanup running successfully. Cleanup may fail because tests are killed, crash mid-run, or timeout.

**Use unique identifiers** for test data instead of shared state:

```go
// ✗ DON'T - Depends on cleanup working
email := "test-user@example.com"

// ✓ DO - Use unique IDs (timestamp or UUID)
email := fmt.Sprintf("test-user-%d@example.com", time.Now().UnixNano())
email := fmt.Sprintf("test-user-%s@example.com", adapters.NewID().New())
```

If cleanup fails, next run gets a new unique ID and doesn't conflict.

**Avoid `t.Parallel()` with TRUNCATE** — tests that truncate tables cannot run in parallel:

```go
// ✗ DON'T - TRUNCATE interferes with parallel tests
func TestRemoveAll(t *testing.T) {
    t.Parallel()  // This will break other tests
    repo.RemoveAll(ctx)
}

// ✓ DO - Remove t.Parallel() for TRUNCATE tests
func TestRemoveAll(t *testing.T) {
    repo.RemoveAll(ctx)
}
```

---

## HTTP Test Helpers (`tests/helpers_test.go`)

Wraps `net/http` to give tests a clean, readable API. All helper functions return `(*http.Response, []byte)` for easy status and body inspection.

```go
package tests_test

import (
    "net/http"
    "testing"
    "github.com/jmoiron/sqlx"
    "github.com/stretchr/testify/require"
)

func postUser(t *testing.T, body map[string]any) (*http.Response, []byte)
func login(t *testing.T, body map[string]any) (*http.Response, []byte)
func createTask(t *testing.T, token string, body map[string]any) (*http.Response, []byte)
func getTasks(t *testing.T, token string) (*http.Response, []byte)
func putTaskStatus(t *testing.T, token, taskID string, body map[string]any) (*http.Response, []byte)
func putTaskPriority(t *testing.T, token, taskID string, body map[string]any) (*http.Response, []byte)
func patchTaskDescription(t *testing.T, token, taskID string, body map[string]any) (*http.Response, []byte)
func assignTask(t *testing.T, token, taskID, assigneeID string) (*http.Response, []byte)
func unassignTask(t *testing.T, token, taskID string) (*http.Response, []byte)
func archiveTask(t *testing.T, token, taskID string) (*http.Response, []byte)
func reopenTask(t *testing.T, token, taskID string) (*http.Response, []byte)
func deleteTask(t *testing.T, token, taskID string) (*http.Response, []byte)
func addTaskTag(t *testing.T, token, taskID string, body map[string]any) (*http.Response, []byte)
func removeTaskTag(t *testing.T, token, taskID, tagID string) (*http.Response, []byte)
func addComment(t *testing.T, token, taskID string, body map[string]any) (*http.Response, []byte)
func updateComment(t *testing.T, token, taskID, commentID string, body map[string]any) (*http.Response, []byte)
func deleteComment(t *testing.T, token, taskID, commentID string) (*http.Response, []byte)
func getTaskActivity(t *testing.T, token, taskID string) (*http.Response, []byte)
func putUserRole(t *testing.T, token, userID string, body map[string]any) (*http.Response, []byte)
func deleteUser(t *testing.T, token, userID string) (*http.Response, []byte)
func patchUser(t *testing.T, token, userID string, body map[string]any) (*http.Response, []byte)
func getUser(t *testing.T, token, userID string) (*http.Response, []byte)
func getUsers(t *testing.T, token string) (*http.Response, []byte)
```

DB assertion helpers:

```go
func assertTaskExistsInDB(t *testing.T, db *sqlx.DB, taskID string)
func assertTaskDeletedInDB(t *testing.T, db *sqlx.DB, taskID string)
func assertTaskArchived(t *testing.T, db *sqlx.DB, taskID string)
func assertTaskFieldEquals[T comparable](t *testing.T, db *sqlx.DB, taskID, field string, expected T)
func assertTaskFieldEmpty[T comparable](t *testing.T, db *sqlx.DB, taskID, field string)
func assertUserExistsInDB(t *testing.T, db *sqlx.DB, userID string)
func assertUserDeletedInDB(t *testing.T, db *sqlx.DB, userID string)
func assertUserFieldEquals[T comparable](t *testing.T, db *sqlx.DB, userID, field string, expected T)
func assertCommentExistsInDB(t *testing.T, db *sqlx.DB, commentID string)
func assertCommentDeletedInDB(t *testing.T, db *sqlx.DB, commentID string)
```

Response status helpers (for APIs that return status in JSON body):

```go
func getResponseStatus(body []byte) float64
func assertResponseStatus(t *testing.T, body []byte, expectedStatus int)
```

Setup helpers:

```go
func createUserAndGetID(t *testing.T, db *sqlx.DB) string
func createUserWithRoleAndGetToken(t *testing.T, db *sqlx.DB, role string) (userID, token string)
func createTaskAndGetID(t *testing.T, db *sqlx.DB, token, creatorID string) string
func getTaskIDByTitle(t *testing.T, db *sqlx.DB, title string) string
func getTaskTagIDByName(t *testing.T, db *sqlx.DB, taskID, name string) string
func getCommentIDByContent(t *testing.T, db *sqlx.DB, taskID, content string) string
func getTaskField[T any](t *testing.T, db *sqlx.DB, taskID, field string) T
func getUserField[T any](t *testing.T, db *sqlx.DB, userID, field string) T
```

---

## Running Tests

```bash
make test
```

Or directly:

```bash
go test ./tests/... -v
```

To run a single test function:

```bash
go test ./tests/... -run TestComponent/testCreateTask -v
```

---

## Makefile Integration

```makefile
test:
    go test ./... -v
```

The tests run against a live server started by `SetupComponentTest`.

---

## Rules

- Never modify a `.feature` acceptance test without explicit permission.
- The test name **must** embed the `.feature` filename and first `Given` line number — this is the traceability link.
- Every `Given`, `When`, and `Then` step must map to a real helper call — no no-op stubs.
- `require` for setup (Given/When failures stop the test), `assert` for Then assertions.
- Each test function is isolated with its own data via API calls — never share state between tests.
- Every test must be seen to fail before it passes (per `AGENTS.md`).
- If a `.feature` scenario cannot be translated, write the test as a failing stub with a `t.Skip("not yet implemented: ...")` and report it.
- `.feature` files are committed; tests are committed; no intermediate generated files exist.

---

## Full Example

`.feature` source (`task.feature`):

```gherkin
Feature: Task lifecycle and permissions

  # Admin creates task
  Scenario: Admin creates task
    Given an admin user is authenticated.
    When the admin creates a task with title "New Task".
    Then no error is returned.
    And the task exists in the database with status "todo".

  # Employee creates task returns forbidden
  Scenario: Employee creates task returns forbidden
    Given an employee user is authenticated.
    When the employee creates a task with title "New Task".
    Then a forbidden error is returned.
```

Component test (`tests/task_test.go`):

```go
// Scenarios from task.feature — keep in sync with that file.
func testCreateTask(t *testing.T, f *TestFixtures) {
    t.Helper()

    // task.feature:8 - Admin creates task
    t.Run("task.feature:8 - Admin creates task", func(t *testing.T) {
        title := fmt.Sprintf("New-Task-%d", time.Now().UnixNano())
        resp, body := createTask(t, f.AuthToken, map[string]any{"title": title})
        require.Equal(t, http.StatusOK, resp.StatusCode, "unexpected status: %s", string(body))

        taskID := getTaskIDByTitle(t, f.DB, title)
        assert.NotEmpty(t, taskID)
        assertTaskFieldEquals[string](t, f.DB, taskID, "status", "todo")
    })

    // task.feature:20 - Employee creates task returns forbidden
    t.Run("task.feature:20 - Employee creates task returns forbidden", func(t *testing.T) {
        email := fmt.Sprintf("emp-%d@test.com", time.Now().UnixNano())
        resp, body := postUser(t, map[string]any{"role": "employee", "name": "Emp", "email": email, "password": "pass"})
        require.Equal(t, http.StatusOK, resp.StatusCode, "setup failed: %s", string(body))

        loginResp, loginBody := login(t, map[string]any{"email": email, "password": "pass"})
        require.Equal(t, http.StatusOK, loginResp.StatusCode, "login failed: %s", string(loginBody))

        var loginResult struct {
            Token string `json:"token"`
        }
        require.NoError(t, json.Unmarshal(loginBody, &loginResult))
        empToken := loginResult.Token

        resp, body = createTask(t, empToken, map[string]any{"title": "New Task"})
        assert.Equal(t, http.StatusForbidden, resp.StatusCode, "unexpected status: %s", string(body))
    })
}
```

---

## Complete Directive → Helper Mapping

| Directive | Helper / Code |
|-----------|---------------|
| `GIVEN user with role X` | `postUser` + `login` → parse token, query DB for `userID` |
| `GIVEN user with role X (quick)` | `createUserWithRoleAndGetToken(t, db, "employer")` → returns `(userID, token)` |
| `GIVEN task with status todo` | `createUserAndGetID` + `createTaskAndGetID` |
| `GIVEN task with status in_progress` | create + `putTaskStatus("in_progress")` |
| `GIVEN task with status completed` | create + putTaskStatus("in_progress") + putTaskStatus("completed") |
| `GIVEN task assigned to user` | `assignTask(t, f.AuthToken, taskID, userID)` |
| `GIVEN task has tag X` | `addTaskTag` + `getTaskTagIDByName` |
| `GIVEN comment on task` | `addComment` + `getCommentIDByContent` |
| `GIVEN task has priority X` | `putTaskPriority(t, f.AuthToken, taskID, ...)` |
| `GIVEN task has description X` | `patchTaskDescription(t, f.AuthToken, taskID, ...)` |
| `WHEN user changes status to X` | `putTaskStatus(t, token, taskID, ...)` |
| `WHEN employer creates task` | `createTask(t, token, ...)` + `getTaskIDByTitle` |
| `WHEN employer assigns task` | `assignTask(t, token, taskID, userID)` |
| `WHEN user unassigns` | `unassignTask(t, token, taskID)` |
| `WHEN user adds comment` | `addComment(t, token, taskID, ...)` + `getCommentIDByContent` |
| `WHEN user updates comment` | `updateComment(t, token, taskID, commentID, ...)` |
| `WHEN user deletes comment` | `deleteComment(t, token, taskID, commentID)` |
| `WHEN employer archives task` | `archiveTask(t, token, taskID)` |
| `WHEN employer reopens task` | `reopenTask(t, token, taskID)` |
| `WHEN employer deletes task` | `deleteTask(t, token, taskID)` |
| `WHEN employer sets priority X` | `putTaskPriority(t, token, taskID, ...)` |
| `WHEN admin changes role to X` | `putUserRole(t, token, userID, ...)` |
| `THEN no error / succeeds` | `require.Equal(t, 200, resp.StatusCode, string(body))` |
| `THEN an error is returned` | `assert.GreaterOrEqual(t, int(getResponseStatus(body)), 400)` |
| `THEN forbidden error` | `assertResponseStatus(t, body, 403)` |
| `THEN not found error` | `assertResponseStatus(t, body, 404)` |
| `THEN bad request error` | `assertResponseStatus(t, body, 400)` |
| `THEN unauthorized error` | `assertResponseStatus(t, body, 401)` |
| `THEN task has status X` | `assertTaskFieldEquals[string](t, db, taskID, "status", X)` |
| `THEN task assigned to user` | `assertTaskFieldEquals[string](t, db, taskID, "assigned_to", userID)` |
| `THEN task not assigned` | `assertTaskFieldEmpty[string](t, db, taskID, "assigned_to")` |
| `THEN task is archived` | `assertTaskArchived(t, db, taskID)` |
| `THEN task is deleted` | `assertTaskDeletedInDB(t, db, taskID)` |
| `THEN task exists` | `assertTaskExistsInDB(t, db, taskID)` |
| `THEN user has role X` | `assertUserFieldEquals[string](t, db, userID, "role", X)` |
| `THEN user has name X` | `assertUserFieldEquals[string](t, db, userID, "name", X)` |
| `THEN user is deleted` | `assertUserDeletedInDB(t, db, userID)` |
| `THEN comment exists` | `assertCommentExistsInDB(t, db, commentID)` |
| `THEN comment is deleted` | `assertCommentDeletedInDB(t, db, commentID)` |
| `THEN activity log includes X` | `getTaskActivity` + `json.Unmarshal` + `assert.Contains(t, types, X)` |

