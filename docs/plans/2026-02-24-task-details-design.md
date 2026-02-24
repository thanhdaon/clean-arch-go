# Task Details Use Cases Design

## Overview

Implementation of 5 Task Details use cases following Clean Architecture patterns:
- SetTaskPriority
- SetTaskDueDate
- AddTaskDescription
- AddTaskTag
- RemoveTaskTag

## Decisions

| Decision | Choice | Rationale |
|----------|--------|-----------|
| Scope | All 5 use cases | Complete feature set |
| Tags storage | Separate table | Queryable, supports analytics |
| Permissions | Same as title/status | Employers and assigned employees |

## Domain Layer

### Task Entity Updates (`domain/task/task.go`)

Add to Task interface:
```go
type Task interface {
    // Existing methods...
    Priority() Priority
    DueDate() time.Time
    Description() string
    SetPriority(updater user.User, p Priority) error
    SetDueDate(updater user.User, d time.Time) error
    SetDescription(updater user.User, desc string) error
}
```

Add to task struct:
```go
type task struct {
    // Existing fields...
    priority    Priority
    dueDate     sql.NullTime
    description sql.NullString
}
```

### New Priority Type (`domain/task/priority.go`)

```go
type Priority int

const (
    PriorityZero Priority = iota
    PriorityLow
    PriorityMedium  // default
    PriorityHigh
    PriorityUrgent
)

func (p Priority) String() string
func PriorityFromString(s string) (Priority, error)
func (p Priority) IsZero() bool
```

### New Tag Aggregate (`domain/tag/tag.go`)

```go
type Tag interface {
    UUID() string
    Name() string
    TaskID() string
}

func NewTag(taskID, name string) (Tag, error)
func From(id, taskID, name string) (Tag, error)
```

## Application Layer

### Command Handlers (`app/command/`)

| File | Command | Handler |
|------|---------|---------|
| `set_task_priority.go` | `SetTaskPriority{TaskId, Priority, Updater}` | Calls `task.SetPriority()` |
| `set_task_due_date.go` | `SetTaskDueDate{TaskId, DueDate, Updater}` | Calls `task.SetDueDate()` |
| `set_task_description.go` | `SetTaskDescription{TaskId, Description, Updater}` | Calls `task.SetDescription()` |
| `add_task_tag.go` | `AddTaskTag{TaskId, TagName}` | Creates new Tag, persists |
| `remove_task_tag.go` | `RemoveTaskTag{TagId}` | Deletes Tag by ID |

### Repository Interface Updates (`app/command/adapters.go`)

```go
type TaskRepository interface {
    // Existing...
    AddTag(ctx context.Context, tag tag.Tag) error
    RemoveTag(ctx context.Context, tagID string) error
}

type TagRepository interface {
    FindById(ctx context.Context, uuid string) (tag.Tag, error)
}
```

## Adapters Layer

### MySQL Task Repository (`adapters/mysql_task_repository.go`)

Update MysqlTask struct:
```go
type MysqlTask struct {
    // Existing fields...
    Priority    sql.NullInt64  `db:"priority"`
    DueDate     sql.NullTime   `db:"due_date"`
    Description sql.NullString `db:"description"`
}
```

Update `Add()`, `UpdateByID()`, and `From()` methods.

### New MySQL Tag Repository (`adapters/mysql_tag_repository.go`)

```go
type MysqlTag struct {
    ID        string    `db:"id"`
    TaskID    string    `db:"task_id"`
    Name      string    `db:"name"`
    CreatedAt time.Time `db:"created_at"`
}

func NewMysqlTagRepository(db *sqlx.DB) MysqlTagRepository
func (r MysqlTagRepository) Add(ctx context.Context, t tag.Tag) error
func (r MysqlTagRepository) Remove(ctx context.Context, tagID string) error
func (r MysqlTagRepository) FindById(ctx context.Context, uuid string) (tag.Tag, error)
func (r MysqlTagRepository) FindByTaskId(ctx context.Context, taskID string) ([]tag.Tag, error)
```

### Database Migration (`migrations/6-add-task-details.sql`)

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

## Ports Layer

### HTTP Handlers (`ports/openapi_handler.go`)

| Method | Endpoint | Handler | Request Body |
|--------|----------|---------|--------------|
| PUT | `/api/tasks/{taskId}/priority` | `SetTaskPriority` | `{priority: "low|medium|high|urgent"}` |
| PUT | `/api/tasks/{taskId}/due-date` | `SetTaskDueDate` | `{due_date: "2024-01-15T10:00:00Z"}` |
| PATCH | `/api/tasks/{taskId}` | `SetTaskDescription` | `{description: "..."}` |
| POST | `/api/tasks/{taskId}/tags` | `AddTaskTag` | `{name: "bug"}` |
| DELETE | `/api/tasks/{taskId}/tags/{tagId}` | `RemoveTaskTag` | - |

### OpenAPI Types (`ports/openapi_types.gen.go`)

```go
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

type PriorityEnum string

const (
    PriorityLow    PriorityEnum = "low"
    PriorityMedium PriorityEnum = "medium"
    PriorityHigh   PriorityEnum = "high"
    PriorityUrgent PriorityEnum = "urgent"
)
```

## Testing Strategy

### Domain Tests
- `domain/task/task_test.go` - Priority, DueDate, Description methods
- `domain/task/priority_test.go` - Priority type parsing
- `domain/tag/tag_test.go` - Tag creation validation

### Command Handler Tests
- `app/command/set_task_priority_test.go`
- `app/command/set_task_due_date_test.go`
- `app/command/set_task_description_test.go`
- `app/command/add_task_tag_test.go`
- `app/command/remove_task_tag_test.go`

### Repository Tests
- `adapters/mysql_task_repository_test.go` - Update existing tests
- `adapters/mysql_tag_repository_test.go` - New test file

## Implementation Order

1. **Migration** - Create `migrations/6-add-task-details.sql`
2. **Domain** - Priority type, Task methods, Tag aggregate
3. **Adapters** - Repository implementations
4. **App** - Command handlers
5. **Ports** - HTTP handlers and OpenAPI types
6. **Tests** - Each layer

## Files Summary

### New Files
- `domain/task/priority.go`
- `domain/task/priority_test.go`
- `domain/tag/tag.go`
- `domain/tag/tag_test.go`
- `app/command/set_task_priority.go`
- `app/command/set_task_priority_test.go`
- `app/command/set_task_due_date.go`
- `app/command/set_task_due_date_test.go`
- `app/command/set_task_description.go`
- `app/command/set_task_description_test.go`
- `app/command/add_task_tag.go`
- `app/command/add_task_tag_test.go`
- `app/command/remove_task_tag.go`
- `app/command/remove_task_tag_test.go`
- `adapters/mysql_tag_repository.go`
- `adapters/mysql_tag_repository_test.go`
- `migrations/6-add-task-details.sql`

### Modified Files
- `domain/task/task.go`
- `domain/task/task_test.go`
- `app/command/adapters.go`
- `adapters/mysql_task_repository.go`
- `adapters/mysql_task_repository_test.go`
- `ports/openapi_handler.go`
- `ports/openapi_types.gen.go`
- `app/app.go`
