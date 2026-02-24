# Task Details Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Implement 5 Task Details use cases (SetTaskPriority, SetTaskDueDate, SetTaskDescription, AddTaskTag, RemoveTaskTag) following Clean Architecture patterns already established in this codebase.

**Architecture:** Domain-first with TDD. Start from inner layers (domain) and work outward (adapters → app → ports). Each layer only imports inward — domain knows nothing about adapters or HTTP. The Tag aggregate lives in its own `domain/tag/` package since it is its own entity with independent lifecycle.

**Tech Stack:** Go, MySQL (sqlx), chi router, testify/require, logrus, otel

---

## Reference: How the codebase works

Before starting, understand these patterns from existing code:

**Domain method pattern** (`domain/task/task.go`):
```go
func (t *task) UpdateTitle(updater user.User, title string) error {
    if title == "" {
        return errors.New("Cannot update task title to empty")
    }
    if allow := t.allowToUpdateTitle(updater); !allow {
        return errors.New("user is not allow to update this task title")
    }
    t.title = title
    t.updatedAt = time.Now()
    return nil
}

func (t *task) allowToUpdateTitle(updater user.User) bool {
    if updater.Role() == user.RoleEmployer { return true }
    if updater.UUID() == t.assignedTo { return true }
    return false
}
```

**Command handler pattern** (`app/command/update_task_title.go`):
```go
type UpdateTaskTitle struct {
    TaskId  string
    Title   string
    Updater user.User
}

type UpdateTaskTitleHandler decorator.CommandHandler[UpdateTaskTitle]

type updateTaskTitleHandler struct {
    tasks TaskRepository
}

func NewUpdateTaskTitleHandler(taskRepository TaskRepository, logger *logrus.Entry) UpdateTaskTitleHandler {
    handler := updateTaskTitleHandler{tasks: taskRepository}
    return decorator.ApplyCommandDecorators(handler, logger)
}

func (h updateTaskTitleHandler) Handle(ctx context.Context, cmd UpdateTaskTitle) error {
    op := errors.Op("cmd.UpdateTaskTitle")
    updateFn := func(ctx context.Context, t task.Task) (task.Task, error) {
        if err := t.UpdateTitle(cmd.Updater, cmd.Title); err != nil {
            return nil, errors.E(errors.E("cmd.UpdateTaskTitle.updateFn"), err)
        }
        return t, nil
    }
    if err := h.tasks.UpdateByID(ctx, cmd.TaskId, updateFn); err != nil {
        return errors.E(op, err)
    }
    return nil
}
```

**HTTP handler pattern** (`ports/openapi_handler.go`):
```go
func (h HttpHandler) UpdateTaskTitle(w http.ResponseWriter, r *http.Request, taskId string) {
    op := errors.Op("http.UpdateTaskTitle")
    updater, err := userFromCtx(r.Context())
    if err != nil {
        unauthorised(r.Context(), errors.E(op, err), w, r)
        return
    }
    body := PatchTaskTitle{}
    if err := render.Decode(r, &body); err != nil {
        badRequest(r.Context(), err, w, r)
        return
    }
    err = h.app.Commands.UpdateTaskTitle.Handle(r.Context(), command.UpdateTaskTitle{
        TaskId:  taskId,
        Title:   body.Title,
        Updater: updater,
    })
    if err != nil {
        badRequest(r.Context(), errors.E(op, err), w, r)
        return
    }
    responseSuccess(r.Context(), w, r)
}
```

**Error pattern** (`core/errors`): `errors.E(op, err)` wraps errors with operation name. Use `errors.New(...)` for simple domain errors (the stdlib `errors` package is used in domain layer — check existing imports).

**Domain test pattern** (`domain/task/task_test.go`):
```go
func TestUpdateTitle(t *testing.T) {
    t.Parallel()
    creator, _ := user.NewUser("123", user.RoleEmployer, "Test User 1", "test1@example.com")
    tk, _ := task.NewTask(creator, "task-uuid", "Initial Title")
    err := tk.UpdateTitle(creator, "Updated Title")
    require.NoError(t, err)
    require.Equal(t, "Updated Title", tk.Title())
}
```

**Adapter test pattern** (`adapters/mysql_task_repositoty_test.go`): Integration tests, requires live MySQL. Run with `make test`. Uses `adapters.NewMySQLConnection()` and `adapters.NewID().New()`.

---

## Task 1: Database Migration

**Files:**
- Create: `migrations/6-add-task-details.sql`

**Step 1: Create the migration file**

```sql
-- +migrate Up
ALTER TABLE tasks ADD COLUMN priority INT DEFAULT 2;
ALTER TABLE tasks ADD COLUMN due_date DATETIME NULL;
ALTER TABLE tasks ADD COLUMN description TEXT NULL;

CREATE TABLE task_tags (
    id VARCHAR(255) NOT NULL PRIMARY KEY,
    task_id VARCHAR(255) NOT NULL,
    name VARCHAR(100) NOT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (task_id) REFERENCES tasks(id) ON DELETE CASCADE
);

-- +migrate Down
DROP TABLE task_tags;
ALTER TABLE tasks DROP COLUMN priority;
ALTER TABLE tasks DROP COLUMN due_date;
ALTER TABLE tasks DROP COLUMN description;
```

Note: `priority INT DEFAULT 2` corresponds to `PriorityMedium = 2` (iota starting at 0: Zero=0, Low=1, Medium=2, High=3, Urgent=4).

**Step 2: Apply the migration**

```bash
make migration-up
```

Expected: Migration applies without error.

**Step 3: Commit**

```bash
git add migrations/6-add-task-details.sql
git commit -m "feat: add migration for task details (priority, due_date, description, tags)"
```

---

## Task 2: Priority Type

**Files:**
- Create: `domain/task/priority.go`
- Create: `domain/task/priority_test.go`

**Step 1: Write the failing test**

Create `domain/task/priority_test.go`:

```go
package task_test

import (
    "clean-arch-go/domain/task"
    "testing"

    "github.com/stretchr/testify/require"
)

func TestPriorityString(t *testing.T) {
    t.Parallel()
    require.Equal(t, "low", task.PriorityLow.String())
    require.Equal(t, "medium", task.PriorityMedium.String())
    require.Equal(t, "high", task.PriorityHigh.String())
    require.Equal(t, "urgent", task.PriorityUrgent.String())
}

func TestPriorityFromString(t *testing.T) {
    t.Parallel()

    tests := []struct {
        input    string
        expected task.Priority
    }{
        {"low", task.PriorityLow},
        {"medium", task.PriorityMedium},
        {"high", task.PriorityHigh},
        {"urgent", task.PriorityUrgent},
    }

    for _, tt := range tests {
        t.Run(tt.input, func(t *testing.T) {
            t.Parallel()
            p, err := task.PriorityFromString(tt.input)
            require.NoError(t, err)
            require.Equal(t, tt.expected, p)
        })
    }
}

func TestPriorityFromString_Invalid(t *testing.T) {
    t.Parallel()
    _, err := task.PriorityFromString("unknown")
    require.Error(t, err)
}

func TestPriorityIsZero(t *testing.T) {
    t.Parallel()
    require.True(t, task.PriorityZero.IsZero())
    require.False(t, task.PriorityLow.IsZero())
    require.False(t, task.PriorityMedium.IsZero())
}
```

**Step 2: Run test to verify it fails**

```bash
make test
```

Expected: FAIL — `task.PriorityLow`, `PriorityFromString` etc. undefined.

**Step 3: Implement Priority type**

Create `domain/task/priority.go`:

```go
package task

import "errors"

type Priority int

const (
    PriorityZero   Priority = iota // 0
    PriorityLow                    // 1
    PriorityMedium                 // 2 — default (matches migration DEFAULT 2)
    PriorityHigh                   // 3
    PriorityUrgent                 // 4
)

func (p Priority) String() string {
    switch p {
    case PriorityLow:
        return "low"
    case PriorityMedium:
        return "medium"
    case PriorityHigh:
        return "high"
    case PriorityUrgent:
        return "urgent"
    default:
        return ""
    }
}

func PriorityFromString(s string) (Priority, error) {
    switch s {
    case "low":
        return PriorityLow, nil
    case "medium":
        return PriorityMedium, nil
    case "high":
        return PriorityHigh, nil
    case "urgent":
        return PriorityUrgent, nil
    default:
        return PriorityZero, errors.New("unknown priority: " + s)
    }
}

func (p Priority) IsZero() bool {
    return p == PriorityZero
}
```

**Step 4: Run tests to verify they pass**

```bash
make test
```

Expected: PASS for all priority tests.

**Step 5: Commit**

```bash
git add domain/task/priority.go domain/task/priority_test.go
git commit -m "feat: add Priority type to task domain"
```

---

## Task 3: Task Domain — Priority, DueDate, Description

**Files:**
- Modify: `domain/task/task.go`
- Modify: `domain/task/task_test.go`

**Step 1: Write the failing tests**

Append to `domain/task/task_test.go`:

```go
func TestSetPriority(t *testing.T) {
    t.Parallel()

    creator, _ := user.NewUser("123", user.RoleEmployer, "Test User 1", "test1@example.com")
    tk, _ := task.NewTask(creator, "task-uuid", "Initial Title")

    err := tk.SetPriority(creator, task.PriorityHigh)
    require.NoError(t, err)
    require.Equal(t, task.PriorityHigh, tk.Priority())
}

func TestSetPriority_InvalidUser(t *testing.T) {
    t.Parallel()

    creator, _ := user.NewUser("123", user.RoleEmployer, "Test User 1", "test1@example.com")
    employee, _ := user.NewUser("456", user.RoleEmployee, "Test User 2", "test2@example.com")
    tk, _ := task.NewTask(creator, "task-uuid", "Initial Title")

    err := tk.SetPriority(employee, task.PriorityHigh)
    require.Error(t, err)
    require.EqualError(t, err, "user is not allowed to update this task")
}

func TestSetDueDate(t *testing.T) {
    t.Parallel()

    creator, _ := user.NewUser("123", user.RoleEmployer, "Test User 1", "test1@example.com")
    tk, _ := task.NewTask(creator, "task-uuid", "Initial Title")
    due := time.Date(2026, 3, 1, 10, 0, 0, 0, time.UTC)

    err := tk.SetDueDate(creator, due)
    require.NoError(t, err)
    require.Equal(t, due, tk.DueDate())
}

func TestSetDueDate_InvalidUser(t *testing.T) {
    t.Parallel()

    creator, _ := user.NewUser("123", user.RoleEmployer, "Test User 1", "test1@example.com")
    employee, _ := user.NewUser("456", user.RoleEmployee, "Test User 2", "test2@example.com")
    tk, _ := task.NewTask(creator, "task-uuid", "Initial Title")
    due := time.Date(2026, 3, 1, 10, 0, 0, 0, time.UTC)

    err := tk.SetDueDate(employee, due)
    require.Error(t, err)
    require.EqualError(t, err, "user is not allowed to update this task")
}

func TestSetDescription(t *testing.T) {
    t.Parallel()

    creator, _ := user.NewUser("123", user.RoleEmployer, "Test User 1", "test1@example.com")
    tk, _ := task.NewTask(creator, "task-uuid", "Initial Title")

    err := tk.SetDescription(creator, "This is a description")
    require.NoError(t, err)
    require.Equal(t, "This is a description", tk.Description())
}

func TestSetDescription_InvalidUser(t *testing.T) {
    t.Parallel()

    creator, _ := user.NewUser("123", user.RoleEmployer, "Test User 1", "test1@example.com")
    employee, _ := user.NewUser("456", user.RoleEmployee, "Test User 2", "test2@example.com")
    tk, _ := task.NewTask(creator, "task-uuid", "Initial Title")

    err := tk.SetDescription(employee, "desc")
    require.Error(t, err)
    require.EqualError(t, err, "user is not allowed to update this task")
}

func TestSetDescription_AssignedEmployeeAllowed(t *testing.T) {
    t.Parallel()

    creator, _ := user.NewUser("123", user.RoleEmployer, "Test User 1", "test1@example.com")
    assignee, _ := user.NewUser("456", user.RoleEmployer, "Test User 2", "test2@example.com")
    tk, _ := task.NewTask(creator, "task-uuid", "Initial Title")
    tk.AssignTo(creator, assignee)

    err := tk.SetDescription(assignee, "desc")
    require.NoError(t, err)
}

func TestNewTask_DefaultPriority(t *testing.T) {
    t.Parallel()

    creator, _ := user.NewUser("123", user.RoleEmployer, "Test User 1", "test1@example.com")
    tk, _ := task.NewTask(creator, "task-uuid", "Initial Title")

    require.Equal(t, task.PriorityMedium, tk.Priority())
}
```

**Step 2: Run test to verify it fails**

```bash
make test
```

Expected: FAIL — `tk.Priority()`, `tk.SetPriority()` etc. undefined.

**Step 3: Update task.go — interface, struct, methods**

In `domain/task/task.go`, update the `Task` interface to add after `Archive`:

```go
    Priority() Priority
    DueDate() time.Time
    Description() string
    SetPriority(updater user.User, p Priority) error
    SetDueDate(updater user.User, d time.Time) error
    SetDescription(updater user.User, desc string) error
```

Update the `task` struct to add after `archivedAt`:

```go
    priority    Priority
    dueDate     sql.NullTime
    description sql.NullString
```

Add getter methods after the existing `ArchivedAt()` method:

```go
func (t *task) Priority() Priority {
    return t.priority
}

func (t *task) DueDate() time.Time {
    if t.dueDate.Valid {
        return t.dueDate.Time
    }
    return time.Time{}
}

func (t *task) Description() string {
    if t.description.Valid {
        return t.description.String
    }
    return ""
}
```

Add permission check helper (reuse the same pattern as `allowToUpdateTitle`):

```go
func (t *task) allowToUpdate(updater user.User) bool {
    if updater.Role() == user.RoleEmployer {
        return true
    }
    if updater.UUID() == t.assignedTo {
        return true
    }
    return false
}
```

Add setter methods:

```go
func (t *task) SetPriority(updater user.User, p Priority) error {
    if !t.allowToUpdate(updater) {
        return errors.New("user is not allowed to update this task")
    }
    t.priority = p
    t.updatedAt = time.Now()
    return nil
}

func (t *task) SetDueDate(updater user.User, d time.Time) error {
    if !t.allowToUpdate(updater) {
        return errors.New("user is not allowed to update this task")
    }
    t.dueDate = sql.NullTime{Time: d, Valid: true}
    t.updatedAt = time.Now()
    return nil
}

func (t *task) SetDescription(updater user.User, desc string) error {
    if !t.allowToUpdate(updater) {
        return errors.New("user is not allowed to update this task")
    }
    t.description = sql.NullString{String: desc, Valid: desc != ""}
    t.updatedAt = time.Now()
    return nil
}
```

Update `NewTask` to set default priority. Find the `return &task{` in `NewTask` and add `priority: PriorityMedium,` to the struct literal.

Update the `From` function signature to accept `priority int, dueDate sql.NullTime, description sql.NullString`:

```go
func From(id, title, statusString, createdBy, assignedTo string, createdAt, updatedAt time.Time, deletedAt, archivedAt sql.NullTime, priority int, dueDate sql.NullTime, description sql.NullString) (Task, error) {
    status, err := StatusFromString(statusString)
    if err != nil {
        return nil, err
    }

    return &task{
        uuid:        id,
        title:       title,
        status:      status,
        createdBy:   createdBy,
        assignedTo:  assignedTo,
        createdAt:   createdAt,
        updatedAt:   updatedAt,
        deletedAt:   deletedAt,
        archivedAt:  archivedAt,
        priority:    Priority(priority),
        dueDate:     dueDate,
        description: description,
    }, nil
}
```

**Step 4: Run tests to verify they pass**

```bash
make test
```

Expected: PASS for all task domain tests. Note: `adapters/mysql_task_repository.go` will now fail to compile because `task.From()` signature changed — that's expected; we'll fix it in Task 5.

**Step 5: Commit**

```bash
git add domain/task/task.go domain/task/task_test.go
git commit -m "feat: add Priority, DueDate, Description to task domain"
```

---

## Task 4: Tag Aggregate

**Files:**
- Create: `domain/tag/tag.go`
- Create: `domain/tag/tag_test.go`

**Step 1: Write the failing tests**

Create `domain/tag/tag_test.go`:

```go
package tag_test

import (
    "clean-arch-go/domain/tag"
    "testing"

    "github.com/stretchr/testify/require"
)

func TestNewTag(t *testing.T) {
    t.Parallel()

    tg, err := tag.NewTag("task-id-1", "bug")
    require.NoError(t, err)
    require.NotNil(t, tg)
    require.NotEmpty(t, tg.UUID())
    require.Equal(t, "task-id-1", tg.TaskID())
    require.Equal(t, "bug", tg.Name())
}

func TestNewTag_EmptyTaskID(t *testing.T) {
    t.Parallel()

    _, err := tag.NewTag("", "bug")
    require.Error(t, err)
    require.EqualError(t, err, "empty task id")
}

func TestNewTag_EmptyName(t *testing.T) {
    t.Parallel()

    _, err := tag.NewTag("task-id-1", "")
    require.Error(t, err)
    require.EqualError(t, err, "empty tag name")
}

func TestFrom(t *testing.T) {
    t.Parallel()

    tg, err := tag.From("tag-id-1", "task-id-1", "feature")
    require.NoError(t, err)
    require.Equal(t, "tag-id-1", tg.UUID())
    require.Equal(t, "task-id-1", tg.TaskID())
    require.Equal(t, "feature", tg.Name())
}

func TestFrom_EmptyID(t *testing.T) {
    t.Parallel()

    _, err := tag.From("", "task-id-1", "feature")
    require.Error(t, err)
    require.EqualError(t, err, "empty tag id")
}
```

Note: `NewTag` generates a UUID internally. We need a simple UUID generator. The domain layer must not import external packages (the `adapters.NewID()` is in the adapters layer). Use `github.com/google/uuid` which is already available (check `go.mod` — if not present, use a simple random string).

**Step 2: Run test to verify it fails**

```bash
make test
```

Expected: FAIL — `tag` package does not exist.

**Step 3: Check go.mod for uuid dependency**

```bash
grep uuid go.mod
```

If `github.com/google/uuid` is present, use it. Otherwise use `fmt.Sprintf("%d", time.Now().UnixNano())` as a fallback (not suitable for production but keeps domain free of new deps). Check with the command above before writing the code.

**Step 4: Implement Tag aggregate**

Create `domain/tag/tag.go`. If `github.com/google/uuid` exists in go.mod:

```go
package tag

import (
    "errors"

    "github.com/google/uuid"
)

type Tag interface {
    UUID() string
    Name() string
    TaskID() string
}

type tag struct {
    uuid   string
    name   string
    taskID string
}

func (t *tag) UUID() string   { return t.uuid }
func (t *tag) Name() string   { return t.name }
func (t *tag) TaskID() string { return t.taskID }

func NewTag(taskID, name string) (Tag, error) {
    if taskID == "" {
        return nil, errors.New("empty task id")
    }
    if name == "" {
        return nil, errors.New("empty tag name")
    }
    return &tag{
        uuid:   uuid.New().String(),
        name:   name,
        taskID: taskID,
    }, nil
}

func From(id, taskID, name string) (Tag, error) {
    if id == "" {
        return nil, errors.New("empty tag id")
    }
    if taskID == "" {
        return nil, errors.New("empty task id")
    }
    if name == "" {
        return nil, errors.New("empty tag name")
    }
    return &tag{
        uuid:   id,
        name:   name,
        taskID: taskID,
    }, nil
}
```

If `github.com/google/uuid` is NOT in go.mod, replace the uuid import and generation with:
```go
import "fmt"
import "time"
// ...
uuid: fmt.Sprintf("tag-%d", time.Now().UnixNano()),
```

**Step 5: Run tests to verify they pass**

```bash
make test
```

Expected: PASS for all tag tests.

**Step 6: Commit**

```bash
git add domain/tag/tag.go domain/tag/tag_test.go
git commit -m "feat: add Tag aggregate to domain"
```

---

## Task 5: Update MySQL Task Repository

**Files:**
- Modify: `adapters/mysql_task_repository.go`
- Modify: `adapters/mysql_task_repositoty_test.go`

This fixes the compilation break from Task 3 (`task.From()` signature change) and adds persistence of new fields.

**Step 1: Update MysqlTask struct and repository**

In `adapters/mysql_task_repository.go`, update `MysqlTask` struct — add after `ArchivedAt`:

```go
    Priority    int            `db:"priority"`
    DueDate     sql.NullTime   `db:"due_date"`
    Description sql.NullString `db:"description"`
```

Update the `Add()` method — update the `added` struct literal and the INSERT query to include new fields:

```go
func (r MysqlTaskRepository) Add(ctx context.Context, t task.Task) error {
    op := errors.Op("MysqlTaskRepository.Add")

    added := MysqlTask{
        ID:          t.UUID(),
        Title:       t.Title(),
        Status:      t.Status().String(),
        CreatedBy:   t.CreatedBy(),
        AssignedTo:  t.AssignedTo(),
        CreatedAt:   t.CreatedAt(),
        UpdatedAt:   timeToNullTime(t.UpdatedAt()),
        Priority:    int(t.Priority()),
        DueDate:     timeToNullTime(t.DueDate()),
        Description: stringToNullString(t.Description()),
    }

    query := `
        INSERT INTO tasks
            (id, title, status, created_by, assigned_to, created_at, updated_at, priority, due_date, description) 
        VALUES 
            (:id, :title, :status, :created_by, :assigned_to, :created_at, :updated_at, :priority, :due_date, :description)
    `

    if _, err := r.db.NamedExecContext(ctx, query, added); err != nil {
        return errors.E(op, err)
    }

    return nil
}
```

Update `UpdateByID()` — update the `updated` struct literal and the UPDATE query:

```go
    updated := MysqlTask{
        ID:          updatedTask.UUID(),
        Title:       updatedTask.Title(),
        Status:      updatedTask.Status().String(),
        CreatedBy:   updatedTask.CreatedBy(),
        AssignedTo:  updatedTask.AssignedTo(),
        CreatedAt:   updatedTask.CreatedAt(),
        UpdatedAt:   timeToNullTime(updatedTask.UpdatedAt()),
        Priority:    int(updatedTask.Priority()),
        DueDate:     timeToNullTime(updatedTask.DueDate()),
        Description: stringToNullString(updatedTask.Description()),
    }

    query := `
        UPDATE tasks 
        SET 
            title = :title,
            status = :status,
            created_by = :created_by,
            assigned_to = :assigned_to,
            created_at = :created_at,
            updated_at = :updated_at,
            priority = :priority,
            due_date = :due_date,
            description = :description
        WHERE id = :id;
    `
```

Update `FindById()` — fix the `task.From()` call at the bottom to pass new fields:

```go
    domainTask, err := task.From(
        data.ID, data.Title, data.Status, data.CreatedBy, data.AssignedTo,
        data.CreatedAt, nullTimeToTime(data.UpdatedAt), data.DeletedAt, data.ArchivedAt,
        int(data.Priority), data.DueDate, data.Description,
    )
```

Add a helper function `stringToNullString` at the bottom of the file (look for existing helpers like `timeToNullTime` to understand where to add it):

```go
func stringToNullString(s string) sql.NullString {
    if s == "" {
        return sql.NullString{}
    }
    return sql.NullString{String: s, Valid: true}
}
```

**Step 2: Run tests to verify compilation is fixed**

```bash
make test
```

Expected: Compilation errors gone. Integration tests pass (requires DB running — skip with `-short` if no DB).

**Step 3: Update adapter test to assert new fields**

In `adapters/mysql_task_repositoty_test.go`, update `assertTasksEquals` to include the new fields:

```go
func assertTasksEquals(t *testing.T, task1, task2 task.Task) {
    t.Helper()

    require.Equal(t, task1.UUID(), task2.UUID())
    require.Equal(t, task1.Title(), task2.Title())
    require.Equal(t, task1.Status().String(), task2.Status().String())
    require.Equal(t, task1.CreatedBy(), task2.CreatedBy())
    require.Equal(t, task1.AssignedTo(), task2.AssignedTo())
    require.Equal(t, task1.Priority(), task2.Priority())
    require.Equal(t, task1.Description(), task2.Description())
    compareTimesIgnoringNanoseconds(t, task1.CreatedAt(), task2.CreatedAt())
    compareTimesIgnoringNanoseconds(t, task1.UpdatedAt(), task2.UpdatedAt())
}
```

**Step 4: Run tests**

```bash
make test
```

Expected: PASS.

**Step 5: Commit**

```bash
git add adapters/mysql_task_repository.go adapters/mysql_task_repositoty_test.go
git commit -m "feat: update MySQL task repository to persist priority, due_date, description"
```

---

## Task 6: MySQL Tag Repository

**Files:**
- Create: `adapters/mysql_tag_repository.go`
- Create: `adapters/mysql_tag_repository_test.go`

**Step 1: Write the failing tests**

Create `adapters/mysql_tag_repository_test.go`:

```go
package adapters_test

import (
    "clean-arch-go/adapters"
    "clean-arch-go/domain/tag"
    "context"
    "testing"

    "github.com/stretchr/testify/require"
)

func TestMysqlTagRepository_Add(t *testing.T) {
    t.Parallel()

    taskRepo := newMysqlTaskRepository(t)
    tagRepo := newMysqlTagRepository(t)

    creator := newExampleEmployer(t)
    tk := newExampleTask(t, creator)
    err := taskRepo.Add(context.Background(), tk)
    require.NoError(t, err)

    tg, err := tag.NewTag(tk.UUID(), "bug")
    require.NoError(t, err)

    err = tagRepo.Add(context.Background(), tg)
    require.NoError(t, err)

    found, err := tagRepo.FindById(context.Background(), tg.UUID())
    require.NoError(t, err)
    require.Equal(t, tg.UUID(), found.UUID())
    require.Equal(t, tg.Name(), found.Name())
    require.Equal(t, tg.TaskID(), found.TaskID())
}

func TestMysqlTagRepository_Remove(t *testing.T) {
    t.Parallel()

    taskRepo := newMysqlTaskRepository(t)
    tagRepo := newMysqlTagRepository(t)

    creator := newExampleEmployer(t)
    tk := newExampleTask(t, creator)
    err := taskRepo.Add(context.Background(), tk)
    require.NoError(t, err)

    tg, err := tag.NewTag(tk.UUID(), "feature")
    require.NoError(t, err)

    err = tagRepo.Add(context.Background(), tg)
    require.NoError(t, err)

    err = tagRepo.Remove(context.Background(), tg.UUID())
    require.NoError(t, err)

    _, err = tagRepo.FindById(context.Background(), tg.UUID())
    require.Error(t, err)
}

func TestMysqlTagRepository_FindByTaskId(t *testing.T) {
    t.Parallel()

    taskRepo := newMysqlTaskRepository(t)
    tagRepo := newMysqlTagRepository(t)

    creator := newExampleEmployer(t)
    tk := newExampleTask(t, creator)
    err := taskRepo.Add(context.Background(), tk)
    require.NoError(t, err)

    tag1, _ := tag.NewTag(tk.UUID(), "bug")
    tag2, _ := tag.NewTag(tk.UUID(), "feature")

    err = tagRepo.Add(context.Background(), tag1)
    require.NoError(t, err)
    err = tagRepo.Add(context.Background(), tag2)
    require.NoError(t, err)

    tags, err := tagRepo.FindByTaskId(context.Background(), tk.UUID())
    require.NoError(t, err)
    require.Len(t, tags, 2)
}

func newMysqlTagRepository(t *testing.T) adapters.MysqlTagRepository {
    t.Helper()
    db, err := adapters.NewMySQLConnection()
    require.NoError(t, err)
    return adapters.NewMysqlTagRepository(db)
}
```

**Step 2: Run test to verify it fails**

```bash
make test
```

Expected: FAIL — `adapters.MysqlTagRepository` and `adapters.NewMysqlTagRepository` undefined.

**Step 3: Implement MySQL tag repository**

Create `adapters/mysql_tag_repository.go`:

```go
package adapters

import (
    "clean-arch-go/core/errors"
    "clean-arch-go/domain/errkind"
    "clean-arch-go/domain/tag"
    "context"
    "database/sql"
    "time"

    "github.com/jmoiron/sqlx"
)

type MysqlTag struct {
    ID        string    `db:"id"`
    TaskID    string    `db:"task_id"`
    Name      string    `db:"name"`
    CreatedAt time.Time `db:"created_at"`
}

type MysqlTagRepository struct {
    db *sqlx.DB
}

func NewMysqlTagRepository(db *sqlx.DB) MysqlTagRepository {
    return MysqlTagRepository{db: db}
}

func (r MysqlTagRepository) Add(ctx context.Context, t tag.Tag) error {
    op := errors.Op("MysqlTagRepository.Add")

    row := MysqlTag{
        ID:        t.UUID(),
        TaskID:    t.TaskID(),
        Name:      t.Name(),
        CreatedAt: time.Now(),
    }

    query := `
        INSERT INTO task_tags (id, task_id, name, created_at)
        VALUES (:id, :task_id, :name, :created_at)
    `

    if _, err := r.db.NamedExecContext(ctx, query, row); err != nil {
        return errors.E(op, err)
    }

    return nil
}

func (r MysqlTagRepository) Remove(ctx context.Context, tagID string) error {
    op := errors.Op("MysqlTagRepository.Remove")

    query := `DELETE FROM task_tags WHERE id = ?`

    result, err := r.db.ExecContext(ctx, query, tagID)
    if err != nil {
        return errors.E(op, err)
    }

    rowsAffected, err := result.RowsAffected()
    if err != nil {
        return errors.E(op, err)
    }

    if rowsAffected == 0 {
        return errors.E(op, errkind.NotExist, errors.Str("tag not found"))
    }

    return nil
}

func (r MysqlTagRepository) FindById(ctx context.Context, uuid string) (tag.Tag, error) {
    op := errors.Op("MysqlTagRepository.FindById")

    data := MysqlTag{}
    query := `SELECT * FROM task_tags WHERE id = ?`

    if err := r.db.GetContext(ctx, &data, query, uuid); err != nil {
        if err == sql.ErrNoRows {
            return nil, errors.E(op, errkind.NotExist, err)
        }
        return nil, errors.E(op, err)
    }

    return tag.From(data.ID, data.TaskID, data.Name)
}

func (r MysqlTagRepository) FindByTaskId(ctx context.Context, taskID string) ([]tag.Tag, error) {
    op := errors.Op("MysqlTagRepository.FindByTaskId")

    data := []MysqlTag{}
    query := `SELECT * FROM task_tags WHERE task_id = ?`

    if err := r.db.SelectContext(ctx, &data, query, taskID); err != nil {
        return nil, errors.E(op, err)
    }

    tags := make([]tag.Tag, 0, len(data))
    for _, row := range data {
        t, err := tag.From(row.ID, row.TaskID, row.Name)
        if err != nil {
            return nil, errors.E(op, err)
        }
        tags = append(tags, t)
    }

    return tags, nil
}
```

Note: `errors.Str` is a helper in `clean-arch-go/core/errors` — check the errors package; if it doesn't exist use `fmt.Errorf("tag not found")` instead.

**Step 4: Check errors.Str exists**

```bash
grep -r "func Str" core/errors/
```

Use accordingly.

**Step 5: Run tests**

```bash
make test
```

Expected: PASS.

**Step 6: Commit**

```bash
git add adapters/mysql_tag_repository.go adapters/mysql_tag_repository_test.go
git commit -m "feat: add MySQL tag repository"
```

---

## Task 7: Update Repository Interfaces (adapters.go)

**Files:**
- Modify: `app/command/adapters.go`

**Step 1: Add Tag repository interfaces**

In `app/command/adapters.go`, add imports for `tag` and update `TaskRepository`, plus add a new `TagRepository`:

```go
package command

import (
    "clean-arch-go/domain/tag"
    "clean-arch-go/domain/task"
    "clean-arch-go/domain/user"
    "context"
)

type ID interface {
    New() string
    IsValid(id string) bool
}

type TaskUpdater func(context.Context, task.Task) (task.Task, error)

type TaskRepository interface {
    Add(context.Context, task.Task) error
    UpdateByID(ctx context.Context, uuid string, updateFn TaskUpdater) error
    AddTag(ctx context.Context, t tag.Tag) error
    RemoveTag(ctx context.Context, tagID string) error
}

type TagRepository interface {
    FindById(ctx context.Context, uuid string) (tag.Tag, error)
}

type UserRepository interface {
    Add(context.Context, user.User) error
    FindById(ctx context.Context, uuid string) (user.User, error)
    FindByEmail(ctx context.Context, email string) (user.User, error)
    FindAll(ctx context.Context) ([]user.User, error)
    UpdateByID(ctx context.Context, uuid string, updateFn UserUpdater) error
    DeleteByID(ctx context.Context, uuid string) error
}

type UserUpdater func(context.Context, user.User) (user.User, error)

type AuthService interface {
    CreateIDToken(data map[string]any) (string, error)
}

type VideoService interface {
    GetAll(context.Context) error
}
```

**Step 2: Add AddTag and RemoveTag to MysqlTaskRepository**

The `MysqlTaskRepository` now needs to implement `TaskRepository.AddTag` and `TaskRepository.RemoveTag`. These delegate to the tag repository. However, looking at the design, `TaskRepository` owns the tag operations (not a separate adapter). Add these methods to `adapters/mysql_task_repository.go`:

```go
func (r MysqlTaskRepository) AddTag(ctx context.Context, t tag.Tag) error {
    op := errors.Op("MysqlTaskRepository.AddTag")

    row := MysqlTag{
        ID:        t.UUID(),
        TaskID:    t.TaskID(),
        Name:      t.Name(),
        CreatedAt: time.Now(),
    }

    query := `
        INSERT INTO task_tags (id, task_id, name, created_at)
        VALUES (:id, :task_id, :name, :created_at)
    `

    if _, err := r.db.NamedExecContext(ctx, query, row); err != nil {
        return errors.E(op, err)
    }

    return nil
}

func (r MysqlTaskRepository) RemoveTag(ctx context.Context, tagID string) error {
    op := errors.Op("MysqlTaskRepository.RemoveTag")

    query := `DELETE FROM task_tags WHERE id = ?`

    result, err := r.db.ExecContext(ctx, query, tagID)
    if err != nil {
        return errors.E(op, err)
    }

    rowsAffected, err := result.RowsAffected()
    if err != nil {
        return errors.E(op, err)
    }

    if rowsAffected == 0 {
        return errors.E(op, errkind.NotExist, fmt.Errorf("tag with id %s not found", tagID))
    }

    return nil
}
```

Also add the `tag` import to `mysql_task_repository.go`.

**Step 3: Run tests**

```bash
make test
```

Expected: PASS (or compile errors in command handlers we haven't written yet — that's fine at this point if `adapters.go` changes don't break existing commands).

**Step 4: Commit**

```bash
git add app/command/adapters.go adapters/mysql_task_repository.go
git commit -m "feat: extend TaskRepository interface with tag operations"
```

---

## Task 8: Command Handlers for Priority, DueDate, Description

**Files:**
- Create: `app/command/set_task_priority.go`
- Create: `app/command/set_task_due_date.go`
- Create: `app/command/set_task_description.go`

Note: No test files for command handlers exist in this codebase — that is intentional (they follow the pattern already established). The domain layer has unit tests, and adapters have integration tests.

**Step 1: Create SetTaskPriority handler**

Create `app/command/set_task_priority.go`:

```go
package command

import (
    "clean-arch-go/core/decorator"
    "clean-arch-go/core/errors"
    "clean-arch-go/domain/task"
    "clean-arch-go/domain/user"
    "context"
    "log"

    "github.com/sirupsen/logrus"
)

type SetTaskPriority struct {
    TaskId   string
    Priority task.Priority
    Updater  user.User
}

type SetTaskPriorityHandler decorator.CommandHandler[SetTaskPriority]

type setTaskPriorityHandler struct {
    tasks TaskRepository
}

func NewSetTaskPriorityHandler(taskRepository TaskRepository, logger *logrus.Entry) SetTaskPriorityHandler {
    if taskRepository == nil {
        log.Fatalln("nil taskRepository")
    }
    handler := setTaskPriorityHandler{tasks: taskRepository}
    return decorator.ApplyCommandDecorators(handler, logger)
}

func (h setTaskPriorityHandler) Handle(ctx context.Context, cmd SetTaskPriority) error {
    op := errors.Op("cmd.SetTaskPriority")

    updateFn := func(ctx context.Context, t task.Task) (task.Task, error) {
        if err := t.SetPriority(cmd.Updater, cmd.Priority); err != nil {
            return nil, errors.E(errors.Op("cmd.SetTaskPriority.updateFn"), err)
        }
        return t, nil
    }

    if err := h.tasks.UpdateByID(ctx, cmd.TaskId, updateFn); err != nil {
        return errors.E(op, err)
    }
    return nil
}
```

**Step 2: Create SetTaskDueDate handler**

Create `app/command/set_task_due_date.go`:

```go
package command

import (
    "clean-arch-go/core/decorator"
    "clean-arch-go/core/errors"
    "clean-arch-go/domain/task"
    "clean-arch-go/domain/user"
    "context"
    "log"
    "time"

    "github.com/sirupsen/logrus"
)

type SetTaskDueDate struct {
    TaskId  string
    DueDate time.Time
    Updater user.User
}

type SetTaskDueDateHandler decorator.CommandHandler[SetTaskDueDate]

type setTaskDueDateHandler struct {
    tasks TaskRepository
}

func NewSetTaskDueDateHandler(taskRepository TaskRepository, logger *logrus.Entry) SetTaskDueDateHandler {
    if taskRepository == nil {
        log.Fatalln("nil taskRepository")
    }
    handler := setTaskDueDateHandler{tasks: taskRepository}
    return decorator.ApplyCommandDecorators(handler, logger)
}

func (h setTaskDueDateHandler) Handle(ctx context.Context, cmd SetTaskDueDate) error {
    op := errors.Op("cmd.SetTaskDueDate")

    updateFn := func(ctx context.Context, t task.Task) (task.Task, error) {
        if err := t.SetDueDate(cmd.Updater, cmd.DueDate); err != nil {
            return nil, errors.E(errors.Op("cmd.SetTaskDueDate.updateFn"), err)
        }
        return t, nil
    }

    if err := h.tasks.UpdateByID(ctx, cmd.TaskId, updateFn); err != nil {
        return errors.E(op, err)
    }
    return nil
}
```

**Step 3: Create SetTaskDescription handler**

Create `app/command/set_task_description.go`:

```go
package command

import (
    "clean-arch-go/core/decorator"
    "clean-arch-go/core/errors"
    "clean-arch-go/domain/task"
    "clean-arch-go/domain/user"
    "context"
    "log"

    "github.com/sirupsen/logrus"
)

type SetTaskDescription struct {
    TaskId      string
    Description string
    Updater     user.User
}

type SetTaskDescriptionHandler decorator.CommandHandler[SetTaskDescription]

type setTaskDescriptionHandler struct {
    tasks TaskRepository
}

func NewSetTaskDescriptionHandler(taskRepository TaskRepository, logger *logrus.Entry) SetTaskDescriptionHandler {
    if taskRepository == nil {
        log.Fatalln("nil taskRepository")
    }
    handler := setTaskDescriptionHandler{tasks: taskRepository}
    return decorator.ApplyCommandDecorators(handler, logger)
}

func (h setTaskDescriptionHandler) Handle(ctx context.Context, cmd SetTaskDescription) error {
    op := errors.Op("cmd.SetTaskDescription")

    updateFn := func(ctx context.Context, t task.Task) (task.Task, error) {
        if err := t.SetDescription(cmd.Updater, cmd.Description); err != nil {
            return nil, errors.E(errors.Op("cmd.SetTaskDescription.updateFn"), err)
        }
        return t, nil
    }

    if err := h.tasks.UpdateByID(ctx, cmd.TaskId, updateFn); err != nil {
        return errors.E(op, err)
    }
    return nil
}
```

**Step 4: Run tests**

```bash
make test
```

Expected: PASS.

**Step 5: Commit**

```bash
git add app/command/set_task_priority.go app/command/set_task_due_date.go app/command/set_task_description.go
git commit -m "feat: add SetTaskPriority, SetTaskDueDate, SetTaskDescription command handlers"
```

---

## Task 9: Command Handlers for Tags

**Files:**
- Create: `app/command/add_task_tag.go`
- Create: `app/command/remove_task_tag.go`

**Step 1: Create AddTaskTag handler**

Create `app/command/add_task_tag.go`:

```go
package command

import (
    "clean-arch-go/core/decorator"
    "clean-arch-go/core/errors"
    "clean-arch-go/domain/tag"
    "context"
    "log"

    "github.com/sirupsen/logrus"
)

type AddTaskTag struct {
    TaskId  string
    TagName string
}

type AddTaskTagHandler decorator.CommandHandler[AddTaskTag]

type addTaskTagHandler struct {
    tasks TaskRepository
}

func NewAddTaskTagHandler(taskRepository TaskRepository, logger *logrus.Entry) AddTaskTagHandler {
    if taskRepository == nil {
        log.Fatalln("nil taskRepository")
    }
    handler := addTaskTagHandler{tasks: taskRepository}
    return decorator.ApplyCommandDecorators(handler, logger)
}

func (h addTaskTagHandler) Handle(ctx context.Context, cmd AddTaskTag) error {
    op := errors.Op("cmd.AddTaskTag")

    newTag, err := tag.NewTag(cmd.TaskId, cmd.TagName)
    if err != nil {
        return errors.E(op, err)
    }

    if err := h.tasks.AddTag(ctx, newTag); err != nil {
        return errors.E(op, err)
    }
    return nil
}
```

**Step 2: Create RemoveTaskTag handler**

Create `app/command/remove_task_tag.go`:

```go
package command

import (
    "clean-arch-go/core/decorator"
    "clean-arch-go/core/errors"
    "context"
    "log"

    "github.com/sirupsen/logrus"
)

type RemoveTaskTag struct {
    TagId string
}

type RemoveTaskTagHandler decorator.CommandHandler[RemoveTaskTag]

type removeTaskTagHandler struct {
    tasks TaskRepository
}

func NewRemoveTaskTagHandler(taskRepository TaskRepository, logger *logrus.Entry) RemoveTaskTagHandler {
    if taskRepository == nil {
        log.Fatalln("nil taskRepository")
    }
    handler := removeTaskTagHandler{tasks: taskRepository}
    return decorator.ApplyCommandDecorators(handler, logger)
}

func (h removeTaskTagHandler) Handle(ctx context.Context, cmd RemoveTaskTag) error {
    op := errors.Op("cmd.RemoveTaskTag")

    if err := h.tasks.RemoveTag(ctx, cmd.TagId); err != nil {
        return errors.E(op, err)
    }
    return nil
}
```

**Step 3: Run tests**

```bash
make test
```

Expected: PASS.

**Step 4: Commit**

```bash
git add app/command/add_task_tag.go app/command/remove_task_tag.go
git commit -m "feat: add AddTaskTag, RemoveTaskTag command handlers"
```

---

## Task 10: Wire up Application

**Files:**
- Modify: `app/app.go`
- Modify: `cmd/start/main.go`

**Step 1: Update app.go**

In `app/app.go`, add the 5 new command handlers to `Commands`:

```go
type Commands struct {
    AddUser           command.AddUserHandler
    UpdateUserRole    command.UpdateUserRoleHandler
    DeleteUser        command.DeleteUserHandler
    UpdateUserProfile command.UpdateUserProfileHandler
    CreateTask        command.CreateTaskHandler
    ChangeTaskStatus  command.ChangeTaskStatusHandler
    AssignTask        command.AssignTaskHandler
    UpdateTaskTitle   command.UpdateTaskTitleHandler
    UnassignTask      command.UnassignTaskHandler
    ReopenTask        command.ReopenTaskHandler
    DeleteTask        command.DeleteTaskHandler
    ArchiveTask       command.ArchiveTaskHandler
    SetTaskPriority   command.SetTaskPriorityHandler
    SetTaskDueDate    command.SetTaskDueDateHandler
    SetTaskDescription command.SetTaskDescriptionHandler
    AddTaskTag        command.AddTaskTagHandler
    RemoveTaskTag     command.RemoveTaskTagHandler
}
```

**Step 2: Update main.go**

In `cmd/start/main.go`, add the 5 new handlers to the `app.Commands` struct:

```go
SetTaskPriority:    command.NewSetTaskPriorityHandler(taskRepository, logger),
SetTaskDueDate:     command.NewSetTaskDueDateHandler(taskRepository, logger),
SetTaskDescription: command.NewSetTaskDescriptionHandler(taskRepository, logger),
AddTaskTag:         command.NewAddTaskTagHandler(taskRepository, logger),
RemoveTaskTag:      command.NewRemoveTaskTagHandler(taskRepository, logger),
```

**Step 3: Run tests**

```bash
make test
```

Expected: PASS.

**Step 4: Commit**

```bash
git add app/app.go cmd/start/main.go
git commit -m "feat: wire up new task detail command handlers in application"
```

---

## Task 11: OpenAPI Types

**Files:**
- Modify: `ports/openapi_types.gen.go`

**Step 1: Inspect the current openapi_types.gen.go**

Read the file to see what's already there and find the right place to add new types:

```bash
# In your editor, open ports/openapi_types.gen.go
# Look for existing types like PutTaskStatus, PatchTaskTitle
```

**Step 2: Add new request types**

Find `PatchTaskTitle` struct in `ports/openapi_types.gen.go`. Add after it (or in alphabetical order with existing types):

```go
type PriorityEnum string

const (
    PriorityEnumLow    PriorityEnum = "low"
    PriorityEnumMedium PriorityEnum = "medium"
    PriorityEnumHigh   PriorityEnum = "high"
    PriorityEnumUrgent PriorityEnum = "urgent"
)

type PutTaskPriority struct {
    Priority PriorityEnum `json:"priority"`
}

type PutTaskDueDate struct {
    DueDate time.Time `json:"due_date"`
}

type PatchTaskDescription struct {
    Description *string `json:"description"`
}

type PostTaskTag struct {
    Name string `json:"name"`
}
```

Make sure `time` is imported in `openapi_types.gen.go`. If it's not already there, add `"time"` to the imports.

**Step 3: Run tests**

```bash
make test
```

Expected: PASS.

**Step 4: Commit**

```bash
git add ports/openapi_types.gen.go
git commit -m "feat: add OpenAPI request types for task detail endpoints"
```

---

## Task 12: HTTP Handlers

**Files:**
- Modify: `ports/openapi_handler.go`

**Step 1: Add 5 HTTP handler methods**

Append to `ports/openapi_handler.go`. Add the following 5 methods. Note the imports needed: `"clean-arch-go/domain/task"` for `task.PriorityFromString`.

```go
func (h HttpHandler) SetTaskPriority(w http.ResponseWriter, r *http.Request, taskId string) {
    op := errors.Op("http.SetTaskPriority")

    updater, err := userFromCtx(r.Context())
    if err != nil {
        unauthorised(r.Context(), errors.E(op, err), w, r)
        return
    }

    body := PutTaskPriority{}
    if err := render.Decode(r, &body); err != nil {
        badRequest(r.Context(), err, w, r)
        return
    }

    priority, err := task.PriorityFromString(string(body.Priority))
    if err != nil {
        badRequest(r.Context(), errors.E(op, err), w, r)
        return
    }

    err = h.app.Commands.SetTaskPriority.Handle(r.Context(), command.SetTaskPriority{
        TaskId:   taskId,
        Priority: priority,
        Updater:  updater,
    })

    if err != nil {
        badRequest(r.Context(), errors.E(op, err), w, r)
        return
    }

    responseSuccess(r.Context(), w, r)
}

func (h HttpHandler) SetTaskDueDate(w http.ResponseWriter, r *http.Request, taskId string) {
    op := errors.Op("http.SetTaskDueDate")

    updater, err := userFromCtx(r.Context())
    if err != nil {
        unauthorised(r.Context(), errors.E(op, err), w, r)
        return
    }

    body := PutTaskDueDate{}
    if err := render.Decode(r, &body); err != nil {
        badRequest(r.Context(), err, w, r)
        return
    }

    err = h.app.Commands.SetTaskDueDate.Handle(r.Context(), command.SetTaskDueDate{
        TaskId:  taskId,
        DueDate: body.DueDate,
        Updater: updater,
    })

    if err != nil {
        badRequest(r.Context(), errors.E(op, err), w, r)
        return
    }

    responseSuccess(r.Context(), w, r)
}

func (h HttpHandler) SetTaskDescription(w http.ResponseWriter, r *http.Request, taskId string) {
    op := errors.Op("http.SetTaskDescription")

    updater, err := userFromCtx(r.Context())
    if err != nil {
        unauthorised(r.Context(), errors.E(op, err), w, r)
        return
    }

    body := PatchTaskDescription{}
    if err := render.Decode(r, &body); err != nil {
        badRequest(r.Context(), err, w, r)
        return
    }

    desc := ""
    if body.Description != nil {
        desc = *body.Description
    }

    err = h.app.Commands.SetTaskDescription.Handle(r.Context(), command.SetTaskDescription{
        TaskId:      taskId,
        Description: desc,
        Updater:     updater,
    })

    if err != nil {
        badRequest(r.Context(), errors.E(op, err), w, r)
        return
    }

    responseSuccess(r.Context(), w, r)
}

func (h HttpHandler) AddTaskTag(w http.ResponseWriter, r *http.Request, taskId string) {
    op := errors.Op("http.AddTaskTag")

    body := PostTaskTag{}
    if err := render.Decode(r, &body); err != nil {
        badRequest(r.Context(), err, w, r)
        return
    }

    err := h.app.Commands.AddTaskTag.Handle(r.Context(), command.AddTaskTag{
        TaskId:  taskId,
        TagName: body.Name,
    })

    if err != nil {
        badRequest(r.Context(), errors.E(op, err), w, r)
        return
    }

    responseSuccess(r.Context(), w, r)
}

func (h HttpHandler) RemoveTaskTag(w http.ResponseWriter, r *http.Request, taskId, tagId string) {
    op := errors.Op("http.RemoveTaskTag")

    err := h.app.Commands.RemoveTaskTag.Handle(r.Context(), command.RemoveTaskTag{
        TagId: tagId,
    })

    if err != nil {
        badRequest(r.Context(), errors.E(op, err), w, r)
        return
    }

    responseSuccess(r.Context(), w, r)
}
```

Also add `"clean-arch-go/domain/task"` to the imports in `openapi_handler.go`.

**Step 2: Run tests**

```bash
make test
```

Expected: PASS. There may be compile errors if the router doesn't have routes registered yet — that's fine if `HandlerFromMux` is generated from OpenAPI spec. Check if routes need registering.

**Step 3: Check if routes need manual registration**

```bash
# Look at how routes are registered
grep -n "SetTaskPriority\|PutTaskPriority\|/priority" ports/
```

If routes are auto-generated by `make openapi`, you'll need to update the OpenAPI spec first. If routes are manually registered, add them now.

**Step 4: Run the full test suite**

```bash
make test
```

Expected: PASS all tests.

**Step 5: Commit**

```bash
git add ports/openapi_handler.go
git commit -m "feat: add HTTP handlers for task detail endpoints"
```

---

## Task 13: OpenAPI Spec Update (if needed)

**Files:**
- Possibly: `ports/openapi.gen.go` or `api/openapi.yaml` (check what `make openapi` generates from)

**Step 1: Check how OpenAPI is generated**

```bash
cat Makefile | grep -A5 openapi
```

```bash
ls ports/
```

Look for `openapi.yaml`, `openapi.json`, or a spec file that `make openapi` reads.

**Step 2: Add routes to the spec**

If there is an `api/openapi.yaml` or similar, add the 5 new endpoints following the existing pattern:

```yaml
/api/tasks/{taskId}/priority:
  put:
    operationId: SetTaskPriority
    parameters:
      - name: taskId
        in: path
        required: true
        schema:
          type: string
    requestBody:
      required: true
      content:
        application/json:
          schema:
            $ref: '#/components/schemas/PutTaskPriority'
    responses:
      '200':
        description: Success

/api/tasks/{taskId}/due-date:
  put:
    operationId: SetTaskDueDate
    # ... similar pattern

/api/tasks/{taskId}:
  patch:
    operationId: SetTaskDescription
    # ... similar pattern

/api/tasks/{taskId}/tags:
  post:
    operationId: AddTaskTag
    # ... similar pattern

/api/tasks/{taskId}/tags/{tagId}:
  delete:
    operationId: RemoveTaskTag
    parameters:
      - name: taskId
        in: path
        required: true
        schema:
          type: string
      - name: tagId
        in: path
        required: true
        schema:
          type: string
    responses:
      '200':
        description: Success
```

**Step 3: Regenerate**

```bash
make openapi
```

**Step 4: Run tests**

```bash
make test
```

Expected: PASS.

**Step 5: Commit**

```bash
git add .
git commit -m "feat: update OpenAPI spec and regenerate for task detail endpoints"
```

---

## Final Verification

```bash
make test
make start  # Verify it starts without errors, then Ctrl+C
```

Expected: All tests pass, server starts clean.
