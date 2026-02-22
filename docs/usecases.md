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