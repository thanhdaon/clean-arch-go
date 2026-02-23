# Task Lifecycle Use Cases - Implementation Plan

This document outlines the implementation plan for all Task Lifecycle use cases.

## Overview

| Use Case | Endpoint | Priority | Complexity |
|----------|----------|----------|------------|
| 1. UpdateTaskTitle | `PATCH /api/tasks/{taskId}` | High | Low |
| 2. UnassignTask | `DELETE /api/tasks/{taskId}/assign` | Medium | Low |
| 3. ReopenTask | `PUT /api/tasks/{taskId}/reopen` | Medium | Medium |
| 4. DeleteTask | `DELETE /api/tasks/{taskId}` | High | Medium |
| 5. ArchiveTask | `PUT /api/tasks/{taskId}/archive` | Low | Medium |

## Decisions

- **Delete type:** Soft delete (recoverable via `deleted_at` column)
- **Status transitions:** Enforce valid state transitions
- **Archived tasks:** Exclude from all queries by default

## Status Transition Rules

```
todo       → pending, inprogress
pending    → todo, inprogress
inprogress → completed, pending
completed  → (none - use Reopen command)
```

Transition diagram:
```
todo ←→ pending ←→ inprogress → completed
                  ↑_____________|
```

---

## Phase 1: UpdateTaskTitle

Domain method `task.UpdateTitle` already exists. Only command and endpoint needed.

### Files

| File | Action | Description |
|------|--------|-------------|
| `app/command/update_task_title.go` | CREATE | Command handler |
| `ports/openapi.yml` | MODIFY | Add `PATCH /tasks/{taskId}` |
| `ports/openapi_handler.go` | MODIFY | Add `UpdateTaskTitle` handler |

### Command Structure

```go
type UpdateTaskTitle struct {
    TaskId  string
    Title   string
    Updater user.User
}
```

### Authorization

Employer or task assignee (reuse existing `allowToUpdateTitle` domain method).

### OpenAPI Endpoint

```yaml
/tasks/{taskId}:
  patch:
    operationId: updateTaskTitle
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
            $ref: "#/components/schemas/PatchTaskTitle"
    responses:
      "200":
        description: Task title updated
```

---

## Phase 2: UnassignTask

### Files

| File | Action | Description |
|------|--------|-------------|
| `domain/task/task.go` | MODIFY | Add `Unassign(user.User) error` method |
| `app/command/unassign_task.go` | CREATE | Command handler |
| `ports/openapi.yml` | MODIFY | Add `DELETE /tasks/{taskId}/assign` |
| `ports/openapi_handler.go` | MODIFY | Add `UnassignTask` handler |

### Domain Method

```go
func (t *task) Unassign(remover user.User) error {
    if remover.Role() != user.RoleEmployer {
        return errors.New("only employer can unassign task")
    }
    t.assignedTo = ""
    t.updatedAt = time.Now()
    return nil
}
```

### Command Structure

```go
type UnassignTask struct {
    TaskId string
    Caller user.User
}
```

### Authorization

Employer only.

### OpenAPI Endpoint

```yaml
/tasks/{taskId}/assign:
  delete:
    operationId: unassignTask
    parameters:
      - name: taskId
        in: path
        required: true
        schema:
          type: string
    responses:
      "200":
        description: Task unassigned
```

---

## Phase 3: ReopenTask + Status Transitions

### Files

| File | Action | Description |
|------|--------|-------------|
| `domain/task/status.go` | MODIFY | Add transition validation map |
| `domain/task/task.go` | MODIFY | Add `Reopen(user.User) error`, update `ChangeStatus` to validate transitions |
| `domain/task/task_test.go` | MODIFY | Add tests for transitions |
| `app/command/reopen_task.go` | CREATE | Command handler |
| `ports/openapi.yml` | MODIFY | Add `PUT /tasks/{taskId}/reopen` |
| `ports/openapi_handler.go` | MODIFY | Add `ReopenTask` handler |

### Status Transitions

Add to `domain/task/status.go`:

```go
var validTransitions = map[Status][]Status{
    StatusTodo:       {StatusPending, StatusInProgress},
    StatusPending:    {StatusTodo, StatusInProgress},
    StatusInProgress: {StatusCompleted, StatusPending},
    StatusCompleted:  {},
}

func (s Status) CanTransitionTo(target Status) bool {
    allowed, exists := validTransitions[s]
    if !exists {
        return false
    }
    for _, t := range allowed {
        if t == target {
            return true
        }
    }
    return false
}
```

### Domain Method

```go
func (t *task) Reopen(opener user.User) error {
    if t.status != StatusCompleted {
        return errors.New("only completed tasks can be reopened")
    }
    if !t.allowToChangeStatus(opener) {
        return errors.New("user is not allowed to reopen this task")
    }
    t.status = StatusInProgress
    t.updatedAt = time.Now()
    return nil
}
```

### Update ChangeStatus

```go
func (t *task) ChangeStatus(updater user.User, s Status) error {
    if s.IsZero() {
        return errors.New("cannot update status of task to empty")
    }
    if !t.status.CanTransitionTo(s) {
        return fmt.Errorf("cannot transition from %s to %s", t.status, s)
    }
    if allow := t.allowToChangeStatus(updater); !allow {
        return errors.New("user is not allow to update status of this task")
    }
    t.status = s
    t.updatedAt = time.Now()
    return nil
}
```

### Command Structure

```go
type ReopenTask struct {
    TaskId string
    Caller user.User
}
```

### Authorization

Employer or task assignee.

### OpenAPI Endpoint

```yaml
/tasks/{taskId}/reopen:
  put:
    operationId: reopenTask
    parameters:
      - name: taskId
        in: path
        required: true
        schema:
          type: string
    responses:
      "200":
        description: Task reopened
```

---

## Phase 4: DeleteTask (Soft Delete)

### Files

| File | Action | Description |
|------|--------|-------------|
| `migrations/3-add-deleted-at-to-tasks.sql` | CREATE | Add `deleted_at` column |
| `domain/task/task.go` | MODIFY | Add `deletedAt` field, `Delete(user.User) error`, `IsDeleted()` |
| `adapters/mysql_task_repository.go` | MODIFY | Add `deleted_at` to queries, filter deleted in read methods |
| `app/command/delete_task.go` | CREATE | Command handler |
| `ports/openapi.yml` | MODIFY | Add `DELETE /tasks/{taskId}` |
| `ports/openapi_handler.go` | MODIFY | Add `DeleteTask` handler |

### Migration

```sql
-- +migrate Up
ALTER TABLE tasks ADD COLUMN deleted_at DATETIME DEFAULT NULL;

-- +migrate Down
ALTER TABLE tasks DROP COLUMN deleted_at;
```

### Domain Changes

Add to `task` struct:
```go
type task struct {
    // ... existing fields
    deletedAt sql.NullTime
}

func (t *task) IsDeleted() bool {
    return t.deletedAt.Valid
}

func (t *task) DeletedAt() time.Time {
    if t.deletedAt.Valid {
        return t.deletedAt.Time
    }
    return time.Time{}
}

func (t *task) Delete(deleter user.User) error {
    if deleter.Role() != user.RoleEmployer {
        return errors.New("only employer can delete task")
    }
    if deleter.UUID() != t.createdBy {
        return errors.New("only task creator can delete task")
    }
    t.deletedAt = sql.NullTime{Time: time.Now(), Valid: true}
    t.updatedAt = time.Now()
    return nil
}
```

Update `From()` to include `deletedAt`:
```go
func From(id, title, statusString, createdBy, assignedTo string, createdAt, updatedAt time.Time, deletedAt sql.NullTime) (Task, error)
```

### Repository Changes

Update all SELECT queries to filter deleted:
```sql
SELECT * FROM tasks WHERE deleted_at IS NULL
SELECT * FROM tasks WHERE id = ? AND deleted_at IS NULL
```

### Command Structure

```go
type DeleteTask struct {
    TaskId string
    Caller user.User
}
```

### Authorization

Only the task creator (employer who created it).

### OpenAPI Endpoint

```yaml
/tasks/{taskId}:
  delete:
    operationId: deleteTask
    parameters:
      - name: taskId
        in: path
        required: true
        schema:
          type: string
    responses:
      "200":
        description: Task deleted
```

---

## Phase 5: ArchiveTask

### Files

| File | Action | Description |
|------|--------|-------------|
| `migrations/4-add-archived-at-to-tasks.sql` | CREATE | Add `archived_at` column |
| `domain/task/task.go` | MODIFY | Add `archivedAt` field, `Archive(user.User) error`, `IsArchived()` |
| `adapters/mysql_task_repository.go` | MODIFY | Add `archived_at` to queries, filter archived in read methods |
| `app/command/archive_task.go` | CREATE | Command handler |
| `ports/openapi.yml` | MODIFY | Add `PUT /tasks/{taskId}/archive` |
| `ports/openapi_handler.go` | MODIFY | Add `ArchiveTask` handler |

### Migration

```sql
-- +migrate Up
ALTER TABLE tasks ADD COLUMN archived_at DATETIME DEFAULT NULL;

-- +migrate Down
ALTER TABLE tasks DROP COLUMN archived_at;
```

### Domain Changes

Add to `task` struct:
```go
type task struct {
    // ... existing fields
    archivedAt sql.NullTime
}

func (t *task) IsArchived() bool {
    return t.archivedAt.Valid
}

func (t *task) ArchivedAt() time.Time {
    if t.archivedAt.Valid {
        return t.archivedAt.Time
    }
    return time.Time{}
}

func (t *task) Archive(archiver user.User) error {
    if archiver.Role() != user.RoleEmployer {
        return errors.New("only employer can archive task")
    }
    if t.status != StatusCompleted {
        return errors.New("only completed tasks can be archived")
    }
    t.archivedAt = sql.NullTime{Time: time.Now(), Valid: true}
    t.updatedAt = time.Now()
    return nil
}
```

### Repository Changes

Update all SELECT queries to filter archived:
```sql
SELECT * FROM tasks WHERE deleted_at IS NULL AND archived_at IS NULL
SELECT * FROM tasks WHERE id = ? AND deleted_at IS NULL AND archived_at IS NULL
```

### Command Structure

```go
type ArchiveTask struct {
    TaskId string
    Caller user.User
}
```

### Authorization

Employer only. Task must be `completed` status.

### OpenAPI Endpoint

```yaml
/tasks/{taskId}/archive:
  put:
    operationId: archiveTask
    parameters:
      - name: taskId
        in: path
        required: true
        schema:
          type: string
    responses:
      "200":
        description: Task archived
```

---

## Summary

### New Files (7)

| File | Phase |
|------|-------|
| `app/command/update_task_title.go` | 1 |
| `app/command/unassign_task.go` | 2 |
| `app/command/reopen_task.go` | 3 |
| `app/command/delete_task.go` | 4 |
| `app/command/archive_task.go` | 5 |
| `migrations/3-add-deleted-at-to-tasks.sql` | 4 |
| `migrations/4-add-archived-at-to-tasks.sql` | 5 |

### Modified Files

| File | Phases |
|------|--------|
| `domain/task/task.go` | 2, 3, 4, 5 |
| `domain/task/status.go` | 3 |
| `domain/task/task_test.go` | 3 |
| `adapters/mysql_task_repository.go` | 4, 5 |
| `ports/openapi.yml` | 1, 2, 3, 4, 5 |
| `ports/openapi_handler.go` | 1, 2, 3, 4, 5 |

### OpenAPI Schema Additions

```yaml
components:
  schemas:
    PatchTaskTitle:
      type: object
      required: [title]
      properties:
        title:
          type: string
          example: "Updated task title"
```

---

## Implementation Checklist

- [x] Phase 1: UpdateTaskTitle
  - [x] Create `app/command/update_task_title.go`
  - [x] Update `ports/openapi.yml`
  - [x] Update `ports/openapi_handler.go`
  - [x] Update `app/app.go` to register command

- [x] Phase 2: UnassignTask
  - [x] Add `Unassign` method to `domain/task/task.go`
  - [x] Create `app/command/unassign_task.go`
  - [x] Update `ports/openapi.yml`
  - [x] Update `ports/openapi_handler.go`
  - [x] Update `app/app.go` to register command

- [x] Phase 3: ReopenTask + Status Transitions
  - [x] Add transition validation to `domain/task/status.go`
  - [x] Add `Reopen` method to `domain/task/task.go`
  - [x] Update `ChangeStatus` to validate transitions
  - [x] Create `app/command/reopen_task.go`
  - [x] Update `ports/openapi.yml`
  - [x] Update `ports/openapi_handler.go`
  - [x] Update `app/app.go` to register command

- [x] Phase 4: DeleteTask
  - [x] Create migration `migrations/3-add-deleted-at-to-tasks.sql`
  - [x] Add `deletedAt` field and methods to `domain/task/task.go`
  - [x] Update `adapters/mysql_task_repository.go` to filter deleted
  - [x] Create `app/command/delete_task.go`
  - [x] Update `ports/openapi.yml`
  - [x] Update `ports/openapi_handler.go`
  - [x] Update `app/app.go` to register command

- [x] Phase 5: ArchiveTask
  - [x] Create migration `migrations/4-add-archived-at-to-tasks.sql`
  - [x] Add `archivedAt` field and methods to `domain/task/task.go`
  - [x] Update `adapters/mysql_task_repository.go` to filter archived
  - [x] Create `app/command/archive_task.go`
  - [x] Update `ports/openapi.yml`
  - [x] Update `ports/openapi_handler.go`
  - [x] Update `app/app.go` to register command
