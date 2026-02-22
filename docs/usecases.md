# Use Cases

This document describes all application use cases in the system, organized by type (commands and queries).

---

## Commands (Write Operations)

### 1. AddUser

**File:** `app/command/add_user.go`

Creates a new user in the system.

**Input:**

| Field  | Type     | Description                        |
|--------|----------|------------------------------------|
| `Role` | `string` | User role: `"employer"` or `"employee"` |

**Steps:**
1. Generate a new ULID as the user UUID.
2. Parse the role string and construct a domain `user.User` via `user.From()`.
3. Persist the user via `UserRepository.Add()`.
4. Call `VideoService.GetAll()` (demo external HTTP integration).

**Errors:**
- Invalid role string → domain validation error
- Persistence failure → internal error
- `VideoService` failure → external service error

**Decorators:** logging, tracing

---

### 2. CreateTask

**File:** `app/command/create_task.go`

Creates a new task. Only users with the `Employer` role can perform this action.

**Input:**

| Field     | Type        | Description                              |
|-----------|-------------|------------------------------------------|
| `Title`   | `string`    | Title of the task                        |
| `Creator` | `user.User` | Authenticated user (resolved from JWT)   |

**Steps:**
1. Generate a new ULID as the task UUID.
2. Create a domain task via `task.NewTask(creator, uuid, title)` — enforces the Employer-only creation rule and sets initial status to `todo`.
3. Persist the task via `TaskRepository.Add()`.

**Errors:**
- Creator is not an Employer → `"only employ can create task"`
- Empty title → `"empty task title"`
- Persistence failure → internal error

**Decorators:** none

---

### 3. ChangeTaskStatus

**File:** `app/command/change_task_status.go`

Updates the status of an existing task. The authenticated user must be an Employer or the assignee of the task.

**Input:**

| Field     | Type        | Description                                           |
|-----------|-------------|-------------------------------------------------------|
| `TaskId`  | `string`    | ULID of the task to update                            |
| `Status`  | `string`    | New status: `"todo"`, `"pending"`, `"inprogress"`, `"completed"` |
| `Changer` | `user.User` | Authenticated user (resolved from JWT)                |

**Steps:**
1. Parse the status string into a typed `task.Status`.
2. Load and update the task via `TaskRepository.UpdateByID()` (read-modify-write).
3. Call `task.ChangeStatus(changer, status)` inside the update closure, which enforces authorization.

**Authorization rule:** Allowed if the user is an `Employer` or the task's current assignee.

**Errors:**
- Unknown status string → domain error
- Task not found → `NotExist` error
- Unauthorized user → `"user is not allow to update status of this task"`
- Persistence failure → internal error

**Decorators:** none

---

### 4. AssignTask

**File:** `app/command/assign_task.go`

Assigns a task to a user.

**Input:**

| Field        | Type        | Description                              |
|--------------|-------------|------------------------------------------|
| `TaskId`     | `string`    | ULID of the task to assign               |
| `AssigneeId` | `string`    | ULID of the user to assign the task to   |
| `Assigner`   | `user.User` | Authenticated user (resolved from JWT)   |

**Steps:**
1. Resolve the assignee from the database via `UserRepository.FindById()`.
2. Load and update the task via `TaskRepository.UpdateByID()` (read-modify-write).
3. Call `task.AssignTo(assigner, assignee)` inside the update closure.

**Errors:**
- Assignee not found → `NotExist` error
- Task not found → `NotExist` error
- Domain rule violation on `AssignTo` → domain error
- Persistence failure → internal error

**Known issue:** `task.AssignTo` currently rejects non-Employer assignees (`"only employer role can assign task"`), which is logically inverted — Employees should be assignable. The `assigner` parameter is also accepted but never validated.

**Decorators:** none

---

## Queries (Read Operations)

### 5. Tasks

**File:** `app/query/tasks.go`

Retrieves tasks with role-based filtering.

**Input:**

| Field  | Type        | Description                            |
|--------|-------------|----------------------------------------|
| `User` | `user.User` | Authenticated user (resolved from JWT) |

**Output:** `[]query.Task` — flat DTO projection (not domain objects)

```go
type Task struct {
    UUID       string
    Title      string
    Status     string
    CreatedBy  string
    AssignedTo string
    CreatedAt  time.Time
    UpdatedAt  time.Time
}
```

**Filtering logic:**

| Role       | Behaviour                                       |
|------------|-------------------------------------------------|
| `employee` | Returns only tasks assigned to the current user |
| `employer` | Returns all tasks in the system                 |

**Errors:**
- Read model query failure → internal error
- `FindTasksForUser` is currently a stub returning a hardcoded error

---

## Proposed Use Cases

The following use cases are **not yet implemented**. They are organized by category and represent the next logical extensions of the system.

---

### Fixes for Existing Gaps

These use cases are implied by the current domain model but are missing or broken.

#### G1. UpdateTaskTitle

Updates the title of an existing task. The domain method `task.UpdateTitle` is already implemented but has no corresponding command or HTTP endpoint.

**Proposed endpoint:** `PATCH /api/tasks/{taskId}`

**Input:**

| Field     | Type        | Description                            |
|-----------|-------------|----------------------------------------|
| `TaskId`  | `string`    | ULID of the task to update             |
| `Title`   | `string`    | New title                              |
| `Updater` | `user.User` | Authenticated user (resolved from JWT) |

**Authorization:** Employer or the task's current assignee.

**Errors:**
- Empty title → domain error
- Task not found → `NotExist` error
- Unauthorized user → permission error
- Persistence failure → internal error

---

#### G2. Login / Issue JWT

Issues a signed JWT for a known user. Currently there is no mechanism for callers to obtain a token — the API is unusable without one.

**Proposed endpoint:** `POST /api/auth/login`

**Input:**

| Field    | Type     | Description    |
|----------|----------|----------------|
| `UserId` | `string` | ULID of a user |

**Output:** Signed JWT containing `user_uuid` and `user_role` claims.

**Errors:**
- User not found → `NotExist` error
- Token signing failure → internal error

---

#### G3. FindTasksForUser (fix stub)

The `FindTasksForUser` method in `MysqlTaskRepository` always returns a hardcoded `"dump"` error. Employee task listing is entirely broken as a result.

**Fix:** Implement `SELECT * FROM tasks WHERE assigned_to = ?` in the MySQL adapter, and add an index on `tasks.assigned_to`.

---

### New Commands (Write Operations)

#### C1. DeleteTask

Soft-deletes a task. Only the employer who created the task may delete it.

**Proposed endpoint:** `DELETE /api/tasks/{taskId}`

**Input:**

| Field    | Type        | Description                            |
|----------|-------------|----------------------------------------|
| `TaskId` | `string`    | ULID of the task                       |
| `Caller` | `user.User` | Authenticated user (resolved from JWT) |

**Steps:**
1. Load the task via `TaskRepository.UpdateByID()`.
2. Verify the caller is the task creator and has `employer` role.
3. Set `deleted_at = now()` (requires adding `deleted_at` column to the `tasks` table).
4. All read queries must filter out `deleted_at IS NOT NULL` rows.

**Errors:**
- Task not found → `NotExist` error
- Caller is not the task creator → permission error
- Persistence failure → internal error

---

#### C2. UnassignTask

Removes the current assignee from a task, returning it to unassigned state. Useful before reassigning to a different employee.

**Proposed endpoint:** `DELETE /api/tasks/{taskId}/assign`

**Input:**

| Field    | Type        | Description                            |
|----------|-------------|----------------------------------------|
| `TaskId` | `string`    | ULID of the task                       |
| `Caller` | `user.User` | Authenticated user (resolved from JWT) |

**Authorization:** Employer only.

**Errors:**
- Task not found → `NotExist` error
- Caller is not an employer → permission error
- Persistence failure → internal error

---

#### C3. AddComment

Attaches a text comment to a task. Requires a new `task_comments` table.

**Proposed endpoint:** `POST /api/tasks/{taskId}/comments`

**Input:**

| Field    | Type        | Description                            |
|----------|-------------|----------------------------------------|
| `TaskId` | `string`    | ULID of the task                       |
| `Body`   | `string`    | Comment text                           |
| `Author` | `user.User` | Authenticated user (resolved from JWT) |

**Authorization:** Employer or the task's current assignee.

**Steps:**
1. Verify the task exists.
2. Verify the caller is an employer or the assigned employee.
3. Persist the comment via a new `CommentRepository.Add()`.

**Errors:**
- Empty body → validation error
- Task not found → `NotExist` error
- Unauthorized caller → permission error
- Persistence failure → internal error

---

#### C4. SetTaskDueDate

Sets or updates the due date on a task. The due date must be in the future at the time of creation.

**Proposed endpoint:** `PUT /api/tasks/{taskId}/due-date`

**Input:**

| Field     | Type        | Description                            |
|-----------|-------------|----------------------------------------|
| `TaskId`  | `string`    | ULID of the task                       |
| `DueDate` | `time.Time` | Target due date (must be future)       |
| `Caller`  | `user.User` | Authenticated user (resolved from JWT) |

**Authorization:** Employer only.

**Errors:**
- Due date is in the past → domain validation error
- Task not found → `NotExist` error
- Unauthorized caller → permission error
- Persistence failure → internal error

---

#### C5. SetTaskPriority

Assigns a priority level to a task. Enriches filtering and sorting capabilities.

**Proposed endpoint:** `PUT /api/tasks/{taskId}/priority`

**Input:**

| Field      | Type        | Description                                              |
|------------|-------------|----------------------------------------------------------|
| `TaskId`   | `string`    | ULID of the task                                         |
| `Priority` | `string`    | One of: `"low"`, `"medium"`, `"high"`, `"urgent"`       |
| `Caller`   | `user.User` | Authenticated user (resolved from JWT)                   |

**Authorization:** Employer only.

**Errors:**
- Unknown priority string → domain validation error
- Task not found → `NotExist` error
- Unauthorized caller → permission error
- Persistence failure → internal error

---

#### C6. ArchiveTask

Moves a completed task to an archived state so it no longer appears in active task lists. Distinct from deletion — archived tasks are retained for history.

**Proposed endpoint:** `PUT /api/tasks/{taskId}/archive`

**Input:**

| Field    | Type        | Description                            |
|----------|-------------|----------------------------------------|
| `TaskId` | `string`    | ULID of the task                       |
| `Caller` | `user.User` | Authenticated user (resolved from JWT) |

**Authorization:** Employer only. Task must be in `completed` status before archiving.

**Errors:**
- Task not in `completed` status → domain error
- Task not found → `NotExist` error
- Unauthorized caller → permission error
- Persistence failure → internal error

---

#### C7. UpdateUserRole

Changes a user's role. Uses the existing `user.ChangeRole` domain method which is implemented but has no command or endpoint.

**Proposed endpoint:** `PUT /api/users/{userId}/role`

**Input:**

| Field    | Type        | Description                            |
|----------|-------------|----------------------------------------|
| `UserId` | `string`    | ULID of the user to update             |
| `Role`   | `string`    | New role: `"employer"` or `"employee"` |
| `Caller` | `user.User` | Authenticated user (resolved from JWT) |

**Authorization:** Employer only (or a future `admin` role).

**Errors:**
- Unknown role string → domain validation error
- User not found → `NotExist` error
- Unauthorized caller → permission error
- Persistence failure → internal error

---

### New Queries (Read Operations)

#### Q1. GetTask

Retrieves a single task by its ID.

**Proposed endpoint:** `GET /api/tasks/{taskId}`

**Input:**

| Field    | Type        | Description                            |
|----------|-------------|----------------------------------------|
| `TaskId` | `string`    | ULID of the task                       |
| `Caller` | `user.User` | Authenticated user (resolved from JWT) |

**Output:** Single `query.Task` DTO.

**Authorization:** Employer sees any task; Employee sees only tasks assigned to them.

**Errors:**
- Task not found → `NotExist` error
- Unauthorized caller → permission error

---

#### Q2. FilterTasks

Queries tasks with optional filters. Extends the existing `Tasks` query with filter parameters.

**Proposed endpoint:** `GET /api/tasks?status=&assignedTo=&createdBy=&from=&to=`

**Input:**

| Field        | Type        | Description                            |
|--------------|-------------|----------------------------------------|
| `Status`     | `string`    | Optional: filter by task status        |
| `AssignedTo` | `string`    | Optional: filter by assignee UUID      |
| `CreatedBy`  | `string`    | Optional: filter by creator UUID       |
| `From`       | `time.Time` | Optional: created after this timestamp |
| `To`         | `time.Time` | Optional: created before this timestamp |
| `Caller`     | `user.User` | Authenticated user (resolved from JWT) |

**Note:** Requires adding indexes on `tasks.status`, `tasks.assigned_to`, and `tasks.created_by` columns.

---

#### Q3. PaginatedTasks

Adds `limit`/`offset` pagination to task list queries. The current `SELECT *` with no limit is a production risk for large datasets.

**Proposed extension to:** `GET /api/tasks?limit=&offset=`

**Input additions:**

| Field    | Type  | Description                          |
|----------|-------|--------------------------------------|
| `Limit`  | `int` | Maximum number of results to return  |
| `Offset` | `int` | Number of results to skip            |

**Output:** `{ tasks: [], total: int, limit: int, offset: int }`

---

#### Q4. GetUser

Retrieves a single user by ID. Currently users can be created but never read back via the API.

**Proposed endpoint:** `GET /api/users/{userId}`

**Input:**

| Field    | Type        | Description                            |
|----------|-------------|----------------------------------------|
| `UserId` | `string`    | ULID of the user                       |
| `Caller` | `user.User` | Authenticated user (resolved from JWT) |

**Output:** `{ uuid, role }`

---

#### Q5. ListUsers

Lists all users in the system. Useful for employers when selecting assignees.

**Proposed endpoint:** `GET /api/users`

**Input:**

| Field    | Type        | Description                            |
|----------|-------------|----------------------------------------|
| `Caller` | `user.User` | Authenticated user (resolved from JWT) |

**Authorization:** Employer only.

**Output:** `[]{ uuid, role }`

---

#### Q6. TaskStats

Returns aggregate counts of tasks grouped by status, optionally per user. Useful for dashboards and workload visibility.

**Proposed endpoint:** `GET /api/tasks/stats`

**Input:**

| Field    | Type        | Description                                        |
|----------|-------------|----------------------------------------------------|
| `UserId` | `string`    | Optional: scope stats to a specific user           |
| `Caller` | `user.User` | Authenticated user (resolved from JWT)             |

**Output:** `{ todo: int, pending: int, inprogress: int, completed: int }`

**Authorization:** Employer sees system-wide or per-user stats; Employee sees only their own stats.

---

### New Domain Capabilities

These require changes to the domain layer in addition to new commands/queries.

#### D1. Task Status State Machine

Enforce valid status transitions to prevent nonsensical moves (e.g., `completed → todo`).

**Proposed transition graph:**

```
todo → pending → inprogress → completed
                     ↑
               (re-open path)
todo ← pending ← inprogress
```

**Implementation:** Add a `ValidTransitions` map in `domain/task` and check it inside `ChangeStatus`.

---

#### D2. Multi-Assignee Tasks

Allow a task to be assigned to multiple employees simultaneously. Requires changing `assigned_to` from a single UUID string to a list, and a new join table `task_assignees`.

**Domain change:** `AssignedTo() []string` instead of `AssignedTo() string`.

---

#### D3. Audit Log

Record every state-changing command as an immutable log entry. Answers: who changed what, when, and from which value to which.

**Implementation approach:** Implement as a command decorator (`commandAuditDecorator`) that wraps any command handler and writes to a `task_audit_log` table after successful execution. No domain changes required.

**Schema:**

| Column       | Type          | Description                        |
|--------------|---------------|------------------------------------|
| `id`         | `VARCHAR(255)` | ULID                              |
| `entity_id`  | `VARCHAR(255)` | UUID of the affected task/user    |
| `operation`  | `VARCHAR(100)` | Command name (e.g. `ChangeStatus`) |
| `actor_id`   | `VARCHAR(255)` | UUID of the user who acted        |
| `before`     | `JSON`         | State snapshot before the change  |
| `after`      | `JSON`         | State snapshot after the change   |
| `created_at` | `DATETIME`     | Timestamp                         |

---

#### D4. Notification Webhooks

Fire an outgoing HTTP webhook when a task is assigned or its status changes. Introduces a new `WebhookService` port.

**New port interface:**

```go
type WebhookService interface {
    TaskAssigned(ctx context.Context, taskID, assigneeID string) error
    TaskStatusChanged(ctx context.Context, taskID, newStatus string) error
}
```

**Implementation:** HTTP adapter POSTing a JSON payload to a configurable URL (stored in env or DB).