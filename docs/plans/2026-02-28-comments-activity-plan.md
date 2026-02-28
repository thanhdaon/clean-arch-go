# Comments & Activity Log Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Implement AddComment, UpdateComment, DeleteComment, GetTaskActivityLog, and autocomplete search endpoints with `@[type:id]` mention parsing.

**Architecture:** Domain-Driven approach with a new `comment` domain entity and `activity` domain entity, following the exact same patterns as `tag` and existing commands. Each command records an activity after its mutation. Activity is appended via a separate `ActivityRepository` injected into handlers.

**Tech Stack:** Go, MySQL (sqlx), OpenAPI (oapi-codegen), chi, testify

---

### Task 1: Activity Domain Entity

**Files:**
- Create: `domain/activity/activity.go`
- Create: `domain/activity/activity_test.go`

**Step 1: Write the failing test**

```go
// domain/activity/activity_test.go
package activity_test

import (
    "clean-arch-go/domain/activity"
    "testing"
    "github.com/stretchr/testify/require"
)

func TestNewActivity(t *testing.T) {
    t.Parallel()
    a, err := activity.New("task-id-1", "actor-id-1", activity.TypeStatusChanged, map[string]any{"from": "todo", "to": "in_progress"})
    require.NoError(t, err)
    require.NotEmpty(t, a.UUID())
    require.Equal(t, "task-id-1", a.TaskID())
    require.Equal(t, "actor-id-1", a.ActorID())
    require.Equal(t, activity.TypeStatusChanged, a.Type())
    require.NotNil(t, a.Payload())
    require.False(t, a.CreatedAt().IsZero())
}

func TestNewActivity_MissingTaskID(t *testing.T) {
    t.Parallel()
    _, err := activity.New("", "actor-id-1", activity.TypeStatusChanged, nil)
    require.EqualError(t, err, "empty task id")
}

func TestNewActivity_MissingActorID(t *testing.T) {
    t.Parallel()
    _, err := activity.New("task-id-1", "", activity.TypeStatusChanged, nil)
    require.EqualError(t, err, "empty actor id")
}
```

**Step 2: Run test to verify it fails**

```bash
make test
```
Expected: FAIL — `activity` package does not exist.

**Step 3: Write minimal implementation**

```go
// domain/activity/activity.go
package activity

import (
    "errors"
    "time"
    "github.com/google/uuid"
)

type Type string

const (
    TypeStatusChanged   Type = "status_changed"
    TypeAssigned        Type = "assigned"
    TypeUnassigned      Type = "unassigned"
    TypeTitleUpdated    Type = "title_updated"
    TypePriorityChanged Type = "priority_changed"
    TypeDueDateSet      Type = "due_date_set"
    TypeDescriptionSet  Type = "description_set"
    TypeArchived        Type = "archived"
    TypeReopened        Type = "reopened"
    TypeCommentAdded    Type = "comment_added"
    TypeCommentUpdated  Type = "comment_updated"
    TypeCommentDeleted  Type = "comment_deleted"
)

type Activity interface {
    UUID() string
    TaskID() string
    ActorID() string
    Type() Type
    Payload() map[string]any
    CreatedAt() time.Time
}

type activityRecord struct {
    uuid      string
    taskID    string
    actorID   string
    actType   Type
    payload   map[string]any
    createdAt time.Time
}

func (a *activityRecord) UUID() string           { return a.uuid }
func (a *activityRecord) TaskID() string         { return a.taskID }
func (a *activityRecord) ActorID() string        { return a.actorID }
func (a *activityRecord) Type() Type             { return a.actType }
func (a *activityRecord) Payload() map[string]any { return a.payload }
func (a *activityRecord) CreatedAt() time.Time   { return a.createdAt }

func New(taskID, actorID string, t Type, payload map[string]any) (Activity, error) {
    if taskID == "" {
        return nil, errors.New("empty task id")
    }
    if actorID == "" {
        return nil, errors.New("empty actor id")
    }
    return &activityRecord{
        uuid:      uuid.New().String(),
        taskID:    taskID,
        actorID:   actorID,
        actType:   t,
        payload:   payload,
        createdAt: time.Now(),
    }, nil
}

func From(id, taskID, actorID string, t Type, payload map[string]any, createdAt time.Time) (Activity, error) {
    if id == "" {
        return nil, errors.New("empty id")
    }
    if taskID == "" {
        return nil, errors.New("empty task id")
    }
    if actorID == "" {
        return nil, errors.New("empty actor id")
    }
    return &activityRecord{
        uuid:      id,
        taskID:    taskID,
        actorID:   actorID,
        actType:   t,
        payload:   payload,
        createdAt: createdAt,
    }, nil
}
```

**Step 4: Run tests to verify they pass**

```bash
make test
```
Expected: PASS

**Step 5: Commit**

```bash
git add domain/activity/
git commit -m "feat: add activity domain entity"
```

---

### Task 2: Comment Domain Entity

**Files:**
- Create: `domain/comment/comment.go`
- Create: `domain/comment/comment_test.go`

**Step 1: Write the failing tests**

```go
// domain/comment/comment_test.go
package comment_test

import (
    "clean-arch-go/domain/comment"
    "testing"
    "github.com/stretchr/testify/require"
)

func TestNewComment(t *testing.T) {
    t.Parallel()
    c, err := comment.New("task-id-1", "author-id-1", "Hello world")
    require.NoError(t, err)
    require.NotEmpty(t, c.UUID())
    require.Equal(t, "task-id-1", c.TaskID())
    require.Equal(t, "author-id-1", c.AuthorID())
    require.Equal(t, "Hello world", c.Content())
    require.Empty(t, c.References())
    require.False(t, c.IsDeleted())
}

func TestNewComment_ParsesMentions(t *testing.T) {
    t.Parallel()
    c, err := comment.New("task-id-1", "author-id-1", "Hey @[user:uuid-123] see @[task:uuid-456]")
    require.NoError(t, err)
    refs := c.References()
    require.Len(t, refs, 2)
    require.Equal(t, comment.ReferenceTypeUser, refs[0].Type)
    require.Equal(t, "uuid-123", refs[0].ID)
    require.Equal(t, comment.ReferenceTypeTask, refs[1].Type)
    require.Equal(t, "uuid-456", refs[1].ID)
}

func TestNewComment_EmptyTaskID(t *testing.T) {
    t.Parallel()
    _, err := comment.New("", "author-id-1", "text")
    require.EqualError(t, err, "empty task id")
}

func TestNewComment_EmptyAuthorID(t *testing.T) {
    t.Parallel()
    _, err := comment.New("task-id-1", "", "text")
    require.EqualError(t, err, "empty author id")
}

func TestNewComment_EmptyContent(t *testing.T) {
    t.Parallel()
    _, err := comment.New("task-id-1", "author-id-1", "")
    require.EqualError(t, err, "empty content")
}

func TestComment_Update(t *testing.T) {
    t.Parallel()
    c, _ := comment.New("task-id-1", "author-id-1", "original")
    err := c.Update("author-id-1", "updated content")
    require.NoError(t, err)
    require.Equal(t, "updated content", c.Content())
}

func TestComment_Update_WrongAuthor(t *testing.T) {
    t.Parallel()
    c, _ := comment.New("task-id-1", "author-id-1", "original")
    err := c.Update("other-user-id", "updated content")
    require.EqualError(t, err, "only comment author can edit this comment")
}

func TestComment_Delete(t *testing.T) {
    t.Parallel()
    c, _ := comment.New("task-id-1", "author-id-1", "text")
    err := c.Delete("author-id-1")
    require.NoError(t, err)
    require.True(t, c.IsDeleted())
}

func TestComment_Delete_WrongAuthor(t *testing.T) {
    t.Parallel()
    c, _ := comment.New("task-id-1", "author-id-1", "text")
    err := c.Delete("other-user-id")
    require.EqualError(t, err, "only comment author can delete this comment")
}
```

**Step 2: Run to verify failure**

```bash
make test
```
Expected: FAIL — `comment` package does not exist.

**Step 3: Implement**

```go
// domain/comment/comment.go
package comment

import (
    "errors"
    "regexp"
    "time"
    "github.com/google/uuid"
)

type ReferenceType string

const (
    ReferenceTypeUser ReferenceType = "user"
    ReferenceTypeTask ReferenceType = "task"
)

type Reference struct {
    Type ReferenceType
    ID   string
}

var mentionRe = regexp.MustCompile(`@\[(\w+):([^\]]+)\]`)

type Comment interface {
    UUID() string
    TaskID() string
    AuthorID() string
    Content() string
    References() []Reference
    CreatedAt() time.Time
    UpdatedAt() time.Time
    IsDeleted() bool
    Update(editorID, content string) error
    Delete(deleterID string) error
}

type commentRecord struct {
    uuid       string
    taskID     string
    authorID   string
    content    string
    references []Reference
    createdAt  time.Time
    updatedAt  time.Time
    deletedAt  time.Time
}

func (c *commentRecord) UUID() string           { return c.uuid }
func (c *commentRecord) TaskID() string         { return c.taskID }
func (c *commentRecord) AuthorID() string       { return c.authorID }
func (c *commentRecord) Content() string        { return c.content }
func (c *commentRecord) References() []Reference { return c.references }
func (c *commentRecord) CreatedAt() time.Time   { return c.createdAt }
func (c *commentRecord) UpdatedAt() time.Time   { return c.updatedAt }
func (c *commentRecord) IsDeleted() bool        { return !c.deletedAt.IsZero() }

func (c *commentRecord) Update(editorID, content string) error {
    if editorID != c.authorID {
        return errors.New("only comment author can edit this comment")
    }
    if content == "" {
        return errors.New("empty content")
    }
    c.content = content
    c.references = parseReferences(content)
    c.updatedAt = time.Now()
    return nil
}

func (c *commentRecord) Delete(deleterID string) error {
    if deleterID != c.authorID {
        return errors.New("only comment author can delete this comment")
    }
    c.deletedAt = time.Now()
    return nil
}

func parseReferences(content string) []Reference {
    matches := mentionRe.FindAllStringSubmatch(content, -1)
    refs := make([]Reference, 0, len(matches))
    for _, m := range matches {
        refType := ReferenceType(m[1])
        if refType != ReferenceTypeUser && refType != ReferenceTypeTask {
            continue
        }
        refs = append(refs, Reference{Type: refType, ID: m[2]})
    }
    return refs
}

func New(taskID, authorID, content string) (Comment, error) {
    if taskID == "" {
        return nil, errors.New("empty task id")
    }
    if authorID == "" {
        return nil, errors.New("empty author id")
    }
    if content == "" {
        return nil, errors.New("empty content")
    }
    return &commentRecord{
        uuid:       uuid.New().String(),
        taskID:     taskID,
        authorID:   authorID,
        content:    content,
        references: parseReferences(content),
        createdAt:  time.Now(),
    }, nil
}

func From(id, taskID, authorID, content string, createdAt, updatedAt time.Time, deletedAt *time.Time) (Comment, error) {
    if id == "" {
        return nil, errors.New("empty id")
    }
    if taskID == "" {
        return nil, errors.New("empty task id")
    }
    if authorID == "" {
        return nil, errors.New("empty author id")
    }
    c := &commentRecord{
        uuid:       id,
        taskID:     taskID,
        authorID:   authorID,
        content:    content,
        references: parseReferences(content),
        createdAt:  createdAt,
        updatedAt:  updatedAt,
    }
    if deletedAt != nil {
        c.deletedAt = *deletedAt
    }
    return c, nil
}
```

**Step 4: Run tests**

```bash
make test
```
Expected: PASS

**Step 5: Commit**

```bash
git add domain/comment/
git commit -m "feat: add comment domain entity with mention parsing"
```

---

### Task 3: Database Migration

**Files:**
- Create: `migrations/2_comments_activities.sql`

**Step 1: Create migration file**

```sql
-- migrations/2_comments_activities.sql

-- +migrate Up
CREATE TABLE comments (
    id VARCHAR(255) NOT NULL PRIMARY KEY,
    task_id VARCHAR(255) NOT NULL,
    author_id VARCHAR(255) NOT NULL,
    content TEXT NOT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT NULL,
    deleted_at DATETIME DEFAULT NULL,
    INDEX idx_comments_task_id (task_id),
    INDEX idx_comments_author_id (author_id),
    FOREIGN KEY (task_id) REFERENCES tasks(id) ON DELETE CASCADE
);

CREATE TABLE comment_references (
    id VARCHAR(255) NOT NULL PRIMARY KEY,
    comment_id VARCHAR(255) NOT NULL,
    reference_type VARCHAR(20) NOT NULL,
    reference_id VARCHAR(255) NOT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_refs_comment_id (comment_id),
    INDEX idx_refs_type_id (reference_type, reference_id),
    FOREIGN KEY (comment_id) REFERENCES comments(id) ON DELETE CASCADE
);

CREATE TABLE task_activities (
    id VARCHAR(255) NOT NULL PRIMARY KEY,
    task_id VARCHAR(255) NOT NULL,
    actor_id VARCHAR(255) NOT NULL,
    type VARCHAR(50) NOT NULL,
    payload JSON,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_activities_task_id_created (task_id, created_at DESC)
);

-- +migrate Down
DROP TABLE task_activities;
DROP TABLE comment_references;
DROP TABLE comments;
```

**Step 2: Apply migration**

```bash
make migration-up
```
Expected: migration applied without errors.

**Step 3: Commit**

```bash
git add migrations/2_comments_activities.sql
git commit -m "feat: add comments, comment_references, task_activities migrations"
```

---

### Task 4: Repository Interfaces

**Files:**
- Modify: `app/command/adapters.go`

**Step 1: Add interfaces**

In `app/command/adapters.go`, add after the existing interfaces:

```go
import (
    "clean-arch-go/domain/activity"
    "clean-arch-go/domain/comment"
    // ... existing imports
)

type CommentUpdater func(context.Context, comment.Comment) (comment.Comment, error)

type CommentRepository interface {
    Add(ctx context.Context, c comment.Comment) error
    UpdateByID(ctx context.Context, uuid string, updateFn CommentUpdater) error
}

type ActivityRepository interface {
    Add(ctx context.Context, a activity.Activity) error
}
```

**Step 2: Verify it compiles**

```bash
make test
```
Expected: PASS (no new tests yet, just compile check)

**Step 3: Commit**

```bash
git add app/command/adapters.go
git commit -m "feat: add CommentRepository and ActivityRepository interfaces"
```

---

### Task 5: MySQL Comment Repository

**Files:**
- Create: `adapters/mysql_comment_repository.go`
- Create: `adapters/mysql_comment_repository_test.go`

**Step 1: Write the failing test**

```go
// adapters/mysql_comment_repository_test.go
package adapters_test

import (
    "clean-arch-go/adapters"
    "clean-arch-go/domain/comment"
    "context"
    "testing"
    "github.com/stretchr/testify/require"
)

func TestMysqlCommentRepository_AddAndUpdate(t *testing.T) {
    db := newTestDB(t)
    repo := adapters.NewMysqlCommentRepository(db)
    taskRepo := adapters.NewMysqlTaskRepository(db)
    userRepo := adapters.NewMysqlUserRepository(db)

    ctx := context.Background()
    u, task := seedUserAndTask(t, ctx, userRepo, taskRepo)

    c, err := comment.New(task.UUID(), u.UUID(), "Hello @[user:uuid-123]")
    require.NoError(t, err)

    err = repo.Add(ctx, c)
    require.NoError(t, err)

    // update
    err = repo.UpdateByID(ctx, c.UUID(), func(ctx context.Context, existing comment.Comment) (comment.Comment, error) {
        return existing, existing.Update(u.UUID(), "Updated content")
    })
    require.NoError(t, err)
}
```

Note: `newTestDB` and `seedUserAndTask` are existing test helpers in `adapters/` — check `mysql_task_repositoty_test.go` for the pattern.

**Step 2: Run to verify failure**

```bash
make test
```
Expected: FAIL — `NewMysqlCommentRepository` does not exist.

**Step 3: Implement**

```go
// adapters/mysql_comment_repository.go
package adapters

import (
    "clean-arch-go/app/command"
    "clean-arch-go/core/errors"
    "clean-arch-go/domain/comment"
    "clean-arch-go/domain/errkind"
    "context"
    "database/sql"
    "time"

    "github.com/jmoiron/sqlx"
)

type MysqlComment struct {
    ID        string       `db:"id"`
    TaskID    string       `db:"task_id"`
    AuthorID  string       `db:"author_id"`
    Content   string       `db:"content"`
    CreatedAt time.Time    `db:"created_at"`
    UpdatedAt sql.NullTime `db:"updated_at"`
    DeletedAt sql.NullTime `db:"deleted_at"`
}

type MysqlCommentReference struct {
    ID            string    `db:"id"`
    CommentID     string    `db:"comment_id"`
    ReferenceType string    `db:"reference_type"`
    ReferenceID   string    `db:"reference_id"`
    CreatedAt     time.Time `db:"created_at"`
}

type MysqlCommentRepository struct {
    db *sqlx.DB
}

func NewMysqlCommentRepository(db *sqlx.DB) MysqlCommentRepository {
    return MysqlCommentRepository{db: db}
}

func (r MysqlCommentRepository) Add(ctx context.Context, c comment.Comment) error {
    op := errors.Op("MysqlCommentRepository.Add")

    tx, err := r.db.Beginx()
    if err != nil {
        return errors.E(op, err)
    }
    defer func() { err = finishTx(err, tx) }()

    row := MysqlComment{
        ID:        c.UUID(),
        TaskID:    c.TaskID(),
        AuthorID:  c.AuthorID(),
        Content:   c.Content(),
        CreatedAt: c.CreatedAt(),
    }

    if _, err = tx.NamedExecContext(ctx, `
        INSERT INTO comments (id, task_id, author_id, content, created_at)
        VALUES (:id, :task_id, :author_id, :content, :created_at)
    `, row); err != nil {
        return errors.E(op, err)
    }

    if err = r.insertReferences(ctx, tx, c); err != nil {
        return errors.E(op, err)
    }

    return nil
}

func (r MysqlCommentRepository) UpdateByID(ctx context.Context, uuid string, updateFn command.CommentUpdater) error {
    op := errors.Op("MysqlCommentRepository.UpdateByID")

    tx, err := r.db.Beginx()
    if err != nil {
        return errors.E(op, err)
    }
    defer func() { err = finishTx(err, tx) }()

    existing, err := r.findByID(ctx, uuid)
    if err != nil {
        return errors.E(op, err)
    }

    updated, err := updateFn(ctx, existing)
    if err != nil {
        return errors.E(op, err)
    }

    if _, err = tx.ExecContext(ctx, `
        UPDATE comments SET content = ?, updated_at = ?, deleted_at = ? WHERE id = ?
    `, updated.Content(), timeToNullTime(updated.UpdatedAt()), deletedAtNullTime(updated), uuid); err != nil {
        return errors.E(op, err)
    }

    if _, err = tx.ExecContext(ctx, `DELETE FROM comment_references WHERE comment_id = ?`, uuid); err != nil {
        return errors.E(op, err)
    }

    if err = r.insertReferences(ctx, tx, updated); err != nil {
        return errors.E(op, err)
    }

    return nil
}

func (r MysqlCommentRepository) findByID(ctx context.Context, uuid string) (comment.Comment, error) {
    op := errors.Op("MysqlCommentRepository.findByID")

    data := MysqlComment{}
    if err := r.db.GetContext(ctx, &data, `SELECT * FROM comments WHERE id = ?`, uuid); err != nil {
        if err == sql.ErrNoRows {
            return nil, errors.E(op, errkind.NotExist, err)
        }
        return nil, errors.E(op, err)
    }

    var deletedAt *time.Time
    if data.DeletedAt.Valid {
        deletedAt = &data.DeletedAt.Time
    }

    return comment.From(data.ID, data.TaskID, data.AuthorID, data.Content,
        data.CreatedAt, nullTimeToTime(data.UpdatedAt), deletedAt)
}

func (r MysqlCommentRepository) insertReferences(ctx context.Context, tx *sqlx.Tx, c comment.Comment) error {
    for _, ref := range c.References() {
        row := MysqlCommentReference{
            ID:            newUUID(),
            CommentID:     c.UUID(),
            ReferenceType: string(ref.Type),
            ReferenceID:   ref.ID,
            CreatedAt:     time.Now(),
        }
        if _, err := tx.NamedExecContext(ctx, `
            INSERT INTO comment_references (id, comment_id, reference_type, reference_id, created_at)
            VALUES (:id, :comment_id, :reference_type, :reference_id, :created_at)
        `, row); err != nil {
            return err
        }
    }
    return nil
}

func deletedAtNullTime(c comment.Comment) sql.NullTime {
    if c.IsDeleted() {
        return sql.NullTime{} // UpdateByID is not used for delete in this design; kept for completeness
    }
    return sql.NullTime{}
}
```

Note: Add a `newUUID()` helper to `adapters/mysql.go`:
```go
import "github.com/google/uuid"
func newUUID() string { return uuid.New().String() }
```

And a shared `finishTx` helper (or reuse pattern from `mysql_task_repository.go`):
```go
func finishTx(err error, tx *sqlx.Tx) error {
    if err != nil {
        _ = tx.Rollback()
        return err
    }
    return tx.Commit()
}
```

**Step 4: Run tests**

```bash
make test
```
Expected: PASS

**Step 5: Commit**

```bash
git add adapters/mysql_comment_repository.go adapters/mysql_comment_repository_test.go adapters/mysql.go
git commit -m "feat: add MySQL comment repository"
```

---

### Task 6: MySQL Activity Repository

**Files:**
- Create: `adapters/mysql_activity_repository.go`
- Create: `adapters/mysql_activity_repository_test.go`

**Step 1: Write the failing test**

```go
// adapters/mysql_activity_repository_test.go
package adapters_test

import (
    "clean-arch-go/adapters"
    "clean-arch-go/domain/activity"
    "context"
    "testing"
    "github.com/stretchr/testify/require"
)

func TestMysqlActivityRepository_Add(t *testing.T) {
    db := newTestDB(t)
    repo := adapters.NewMysqlActivityRepository(db)
    taskRepo := adapters.NewMysqlTaskRepository(db)
    userRepo := adapters.NewMysqlUserRepository(db)

    ctx := context.Background()
    u, task := seedUserAndTask(t, ctx, userRepo, taskRepo)

    a, err := activity.New(task.UUID(), u.UUID(), activity.TypeStatusChanged,
        map[string]any{"from": "todo", "to": "in_progress"})
    require.NoError(t, err)

    err = repo.Add(ctx, a)
    require.NoError(t, err)
}
```

**Step 2: Run to verify failure**

```bash
make test
```
Expected: FAIL

**Step 3: Implement**

```go
// adapters/mysql_activity_repository.go
package adapters

import (
    "clean-arch-go/core/errors"
    "clean-arch-go/domain/activity"
    "context"
    "encoding/json"
    "time"

    "github.com/jmoiron/sqlx"
)

type MysqlActivity struct {
    ID        string    `db:"id"`
    TaskID    string    `db:"task_id"`
    ActorID   string    `db:"actor_id"`
    Type      string    `db:"type"`
    Payload   []byte    `db:"payload"`
    CreatedAt time.Time `db:"created_at"`
}

type MysqlActivityRepository struct {
    db *sqlx.DB
}

func NewMysqlActivityRepository(db *sqlx.DB) MysqlActivityRepository {
    return MysqlActivityRepository{db: db}
}

func (r MysqlActivityRepository) Add(ctx context.Context, a activity.Activity) error {
    op := errors.Op("MysqlActivityRepository.Add")

    payload, err := json.Marshal(a.Payload())
    if err != nil {
        return errors.E(op, err)
    }

    row := MysqlActivity{
        ID:        a.UUID(),
        TaskID:    a.TaskID(),
        ActorID:   a.ActorID(),
        Type:      string(a.Type()),
        Payload:   payload,
        CreatedAt: a.CreatedAt(),
    }

    if _, err = r.db.NamedExecContext(ctx, `
        INSERT INTO task_activities (id, task_id, actor_id, type, payload, created_at)
        VALUES (:id, :task_id, :actor_id, :type, :payload, :created_at)
    `, row); err != nil {
        return errors.E(op, err)
    }

    return nil
}
```

**Step 4: Run tests**

```bash
make test
```
Expected: PASS

**Step 5: Commit**

```bash
git add adapters/mysql_activity_repository.go adapters/mysql_activity_repository_test.go
git commit -m "feat: add MySQL activity repository"
```

---

### Task 7: Activity Query ReadModel

**Files:**
- Create: `app/query/task_activities.go`

**Step 1: Write failing test**

```go
// app/query/task_activities_test.go
package query_test

import (
    "clean-arch-go/app/query"
    "context"
    "testing"
    "time"
    "github.com/stretchr/testify/require"
)

type fakeActivityReadModel struct{}

func (f fakeActivityReadModel) ActivityForTask(ctx context.Context, taskID string, limit, offset int) ([]query.ActivityDTO, error) {
    return []query.ActivityDTO{
        {UUID: "a1", TaskID: taskID, ActorID: "u1", ActorName: "Alice", Type: "status_changed",
         Payload: map[string]any{}, CreatedAt: time.Now()},
    }, nil
}

func TestTaskActivitiesHandler(t *testing.T) {
    handler := query.NewTaskActivitiesHandler(fakeActivityReadModel{}, nil)
    result, err := handler.Handle(context.Background(), query.TaskActivities{TaskID: "task-1", Limit: 10, Offset: 0})
    require.NoError(t, err)
    require.Len(t, result, 1)
    require.Equal(t, "task-1", result[0].TaskID)
}
```

**Step 2: Run to verify failure**

```bash
make test
```
Expected: FAIL

**Step 3: Implement**

```go
// app/query/task_activities.go
package query

import (
    "clean-arch-go/core/decorator"
    "context"
    "time"

    "github.com/sirupsen/logrus"
)

type ActivityDTO struct {
    UUID      string         `json:"uuid"`
    TaskID    string         `json:"task_id"`
    ActorID   string         `json:"actor_id"`
    ActorName string         `json:"actor_name"`
    Type      string         `json:"type"`
    Payload   map[string]any `json:"payload"`
    CreatedAt time.Time      `json:"created_at"`
}

type TaskActivities struct {
    TaskID string
    Limit  int
    Offset int
}

type TaskActivitiesHandler decorator.QueryHandler[TaskActivities, []ActivityDTO]

type TaskActivitiesReadModel interface {
    ActivityForTask(ctx context.Context, taskID string, limit, offset int) ([]ActivityDTO, error)
}

type taskActivitiesHandler struct {
    readModel TaskActivitiesReadModel
}

func NewTaskActivitiesHandler(readModel TaskActivitiesReadModel, logger *logrus.Entry) TaskActivitiesHandler {
    return decorator.ApplyQueryDecorators(taskActivitiesHandler{readModel: readModel}, logger)
}

func (h taskActivitiesHandler) Handle(ctx context.Context, q TaskActivities) ([]ActivityDTO, error) {
    limit := q.Limit
    if limit == 0 {
        limit = 20
    }
    return h.readModel.ActivityForTask(ctx, q.TaskID, limit, q.Offset)
}
```

Also add the readmodel implementation to `adapters/mysql_activity_repository.go`:

```go
func (r MysqlActivityRepository) ActivityForTask(ctx context.Context, taskID string, limit, offset int) ([]query.ActivityDTO, error) {
    op := errors.Op("MysqlActivityRepository.ActivityForTask")

    type row struct {
        MysqlActivity
        ActorName string `db:"actor_name"`
    }

    rows := []row{}
    if err := r.db.SelectContext(ctx, &rows, `
        SELECT a.*, u.name as actor_name
        FROM task_activities a
        LEFT JOIN users u ON u.id = a.actor_id
        WHERE a.task_id = ?
        ORDER BY a.created_at DESC
        LIMIT ? OFFSET ?
    `, taskID, limit, offset); err != nil {
        return nil, errors.E(op, err)
    }

    result := make([]query.ActivityDTO, 0, len(rows))
    for _, r := range rows {
        var payload map[string]any
        _ = json.Unmarshal(r.Payload, &payload)
        result = append(result, query.ActivityDTO{
            UUID:      r.ID,
            TaskID:    r.TaskID,
            ActorID:   r.ActorID,
            ActorName: r.ActorName,
            Type:      r.Type,
            Payload:   payload,
            CreatedAt: r.CreatedAt,
        })
    }
    return result, nil
}
```

**Step 4: Run tests**

```bash
make test
```
Expected: PASS

**Step 5: Commit**

```bash
git add app/query/task_activities.go adapters/mysql_activity_repository.go
git commit -m "feat: add task activities query handler and readmodel"
```

---

### Task 8: AddComment Command

**Files:**
- Create: `app/command/add_comment.go`
- Create: `app/command/add_comment_test.go`

**Step 1: Write failing test**

```go
// app/command/add_comment_test.go
package command_test

import (
    "clean-arch-go/app/command"
    "clean-arch-go/domain/comment"
    "clean-arch-go/domain/activity"
    "context"
    "testing"
    "github.com/stretchr/testify/require"
)

type fakeCommentRepo struct{ added comment.Comment }
func (f *fakeCommentRepo) Add(ctx context.Context, c comment.Comment) error { f.added = c; return nil }
func (f *fakeCommentRepo) UpdateByID(ctx context.Context, uuid string, fn command.CommentUpdater) error { return nil }

type fakeActivityRepo struct{ added activity.Activity }
func (f *fakeActivityRepo) Add(ctx context.Context, a activity.Activity) error { f.added = a; return nil }

func TestAddCommentHandler(t *testing.T) {
    commentRepo := &fakeCommentRepo{}
    activityRepo := &fakeActivityRepo{}

    handler := command.NewAddCommentHandler(commentRepo, activityRepo, nil)
    err := handler.Handle(context.Background(), command.AddComment{
        TaskID:   "task-1",
        AuthorID: "user-1",
        Content:  "Hello @[user:uuid-123]",
    })
    require.NoError(t, err)
    require.Equal(t, "task-1", commentRepo.added.TaskID())
    require.Equal(t, activity.TypeCommentAdded, activityRepo.added.Type())
}
```

**Step 2: Run to verify failure**

```bash
make test
```
Expected: FAIL

**Step 3: Implement**

```go
// app/command/add_comment.go
package command

import (
    "clean-arch-go/core/decorator"
    "clean-arch-go/core/errors"
    "clean-arch-go/domain/activity"
    "clean-arch-go/domain/comment"
    "context"
    "log"

    "github.com/sirupsen/logrus"
)

type AddComment struct {
    TaskID   string
    AuthorID string
    Content  string
}

type AddCommentHandler decorator.CommandHandler[AddComment]

type addCommentHandler struct {
    comments   CommentRepository
    activities ActivityRepository
}

func NewAddCommentHandler(comments CommentRepository, activities ActivityRepository, logger *logrus.Entry) AddCommentHandler {
    if comments == nil {
        log.Fatalln("nil comment repository")
    }
    if activities == nil {
        log.Fatalln("nil activity repository")
    }
    return decorator.ApplyCommandDecorators(addCommentHandler{comments: comments, activities: activities}, logger)
}

func (h addCommentHandler) Handle(ctx context.Context, cmd AddComment) error {
    op := errors.Op("cmd.AddComment")

    c, err := comment.New(cmd.TaskID, cmd.AuthorID, cmd.Content)
    if err != nil {
        return errors.E(op, err)
    }

    if err = h.comments.Add(ctx, c); err != nil {
        return errors.E(op, err)
    }

    a, err := activity.New(cmd.TaskID, cmd.AuthorID, activity.TypeCommentAdded,
        map[string]any{"comment_id": c.UUID()})
    if err != nil {
        return errors.E(op, err)
    }

    if err = h.activities.Add(ctx, a); err != nil {
        return errors.E(op, err)
    }

    return nil
}
```

**Step 4: Run tests**

```bash
make test
```
Expected: PASS

**Step 5: Commit**

```bash
git add app/command/add_comment.go app/command/add_comment_test.go
git commit -m "feat: add AddComment command"
```

---

### Task 9: UpdateComment and DeleteComment Commands

**Files:**
- Create: `app/command/update_comment.go`
- Create: `app/command/delete_comment.go`

**Step 1: Write failing tests**

```go
// app/command/update_comment_test.go
package command_test

func TestUpdateCommentHandler(t *testing.T) {
    c, _ := comment.New("task-1", "user-1", "original")
    commentRepo := &fakeCommentRepoWithData{data: c}
    activityRepo := &fakeActivityRepo{}

    handler := command.NewUpdateCommentHandler(commentRepo, activityRepo, nil)
    err := handler.Handle(context.Background(), command.UpdateComment{
        CommentID: c.UUID(),
        EditorID:  "user-1",
        Content:   "updated",
    })
    require.NoError(t, err)
    require.Equal(t, activity.TypeCommentUpdated, activityRepo.added.Type())
}

// app/command/delete_comment_test.go
func TestDeleteCommentHandler(t *testing.T) {
    c, _ := comment.New("task-1", "user-1", "original")
    commentRepo := &fakeCommentRepoWithData{data: c}
    activityRepo := &fakeActivityRepo{}

    handler := command.NewDeleteCommentHandler(commentRepo, activityRepo, nil)
    err := handler.Handle(context.Background(), command.DeleteComment{
        CommentID: c.UUID(),
        DeleterID: "user-1",
        TaskID:    "task-1",
    })
    require.NoError(t, err)
    require.Equal(t, activity.TypeCommentDeleted, activityRepo.added.Type())
}
```

**Step 2: Run to verify failure**

```bash
make test
```
Expected: FAIL

**Step 3: Implement**

```go
// app/command/update_comment.go
package command

import (
    "clean-arch-go/core/decorator"
    "clean-arch-go/core/errors"
    "clean-arch-go/domain/activity"
    "clean-arch-go/domain/comment"
    "context"
    "log"
    "github.com/sirupsen/logrus"
)

type UpdateComment struct {
    CommentID string
    EditorID  string
    Content   string
}

type UpdateCommentHandler decorator.CommandHandler[UpdateComment]

type updateCommentHandler struct {
    comments   CommentRepository
    activities ActivityRepository
}

func NewUpdateCommentHandler(comments CommentRepository, activities ActivityRepository, logger *logrus.Entry) UpdateCommentHandler {
    if comments == nil { log.Fatalln("nil comment repository") }
    if activities == nil { log.Fatalln("nil activity repository") }
    return decorator.ApplyCommandDecorators(updateCommentHandler{comments: comments, activities: activities}, logger)
}

func (h updateCommentHandler) Handle(ctx context.Context, cmd UpdateComment) error {
    op := errors.Op("cmd.UpdateComment")

    var taskID string
    if err := h.comments.UpdateByID(ctx, cmd.CommentID, func(ctx context.Context, c comment.Comment) (comment.Comment, error) {
        taskID = c.TaskID()
        return c, c.Update(cmd.EditorID, cmd.Content)
    }); err != nil {
        return errors.E(op, err)
    }

    a, err := activity.New(taskID, cmd.EditorID, activity.TypeCommentUpdated,
        map[string]any{"comment_id": cmd.CommentID})
    if err != nil {
        return errors.E(op, err)
    }

    return errors.E(op, h.activities.Add(ctx, a))
}
```

```go
// app/command/delete_comment.go
package command

import (
    "clean-arch-go/core/decorator"
    "clean-arch-go/core/errors"
    "clean-arch-go/domain/activity"
    "clean-arch-go/domain/comment"
    "context"
    "log"
    "github.com/sirupsen/logrus"
)

type DeleteComment struct {
    CommentID string
    DeleterID string
    TaskID    string
}

type DeleteCommentHandler decorator.CommandHandler[DeleteComment]

type deleteCommentHandler struct {
    comments   CommentRepository
    activities ActivityRepository
}

func NewDeleteCommentHandler(comments CommentRepository, activities ActivityRepository, logger *logrus.Entry) DeleteCommentHandler {
    if comments == nil { log.Fatalln("nil comment repository") }
    if activities == nil { log.Fatalln("nil activity repository") }
    return decorator.ApplyCommandDecorators(deleteCommentHandler{comments: comments, activities: activities}, logger)
}

func (h deleteCommentHandler) Handle(ctx context.Context, cmd DeleteComment) error {
    op := errors.Op("cmd.DeleteComment")

    if err := h.comments.UpdateByID(ctx, cmd.CommentID, func(ctx context.Context, c comment.Comment) (comment.Comment, error) {
        return c, c.Delete(cmd.DeleterID)
    }); err != nil {
        return errors.E(op, err)
    }

    a, err := activity.New(cmd.TaskID, cmd.DeleterID, activity.TypeCommentDeleted,
        map[string]any{"comment_id": cmd.CommentID})
    if err != nil {
        return errors.E(op, err)
    }

    return errors.E(op, h.activities.Add(ctx, a))
}
```

**Step 4: Run tests**

```bash
make test
```
Expected: PASS

**Step 5: Commit**

```bash
git add app/command/update_comment.go app/command/delete_comment.go
git commit -m "feat: add UpdateComment and DeleteComment commands"
```

---

### Task 10: Wire ActivityRepository into Existing Commands

The following existing commands need to record activities:
`ChangeTaskStatus`, `AssignTask`, `UnassignTask`, `UpdateTaskTitle`, `SetTaskPriority`, `SetTaskDueDate`, `SetTaskDescription`, `ArchiveTask`, `ReopenTask`, `DeleteTask`

For each, the pattern is identical — shown here for `ChangeTaskStatus` as the template:

**Step 1: Modify `app/command/change_task_status.go`**

Add `activities ActivityRepository` field, inject it, and after `h.tasks.UpdateByID` succeeds, call:
```go
a, _ := activity.New(cmd.TaskId, cmd.Changer.UUID(), activity.TypeStatusChanged,
    map[string]any{"status": cmd.Status})
_ = h.activities.Add(ctx, a)
```

Do the same for all other commands with their corresponding `ActivityType` and payload.

**Step 2: Update handler constructors** to accept `ActivityRepository` as a parameter.

**Step 3: Run tests after each command change**

```bash
make test
```

**Step 4: Commit**

```bash
git add app/command/
git commit -m "feat: record activities in all task mutation commands"
```

---

### Task 11: Update app.Application struct

**Files:**
- Modify: `app/app.go`

Add the new handlers:

```go
Commands struct {
    // ... existing ...
    AddComment    command.AddCommentHandler
    UpdateComment command.UpdateCommentHandler
    DeleteComment command.DeleteCommentHandler
}

Queries struct {
    // ... existing ...
    TaskActivities query.TaskActivitiesHandler
}
```

Then update `bootstrap/bootstrap.go` to wire them:

```go
commentRepo  := adapters.NewMysqlCommentRepository(db)
activityRepo := adapters.NewMysqlActivityRepository(db)

// AddComment, UpdateComment, DeleteComment handlers use commentRepo + activityRepo
// Existing task commands get activityRepo injected
// TaskActivities query uses activityRepo
```

**Step: Run tests**

```bash
make test
```
Expected: PASS

**Step: Commit**

```bash
git add app/app.go bootstrap/bootstrap.go
git commit -m "feat: wire comment and activity handlers into application"
```

---

### Task 12: OpenAPI Spec — New Endpoints and Schemas

**Files:**
- Modify: `ports/openapi.yml`

Add the following paths and schemas (insert before `components:` line):

**Paths to add:**

```yaml
  /tasks/{taskId}/comments:
    post:
      tags: [Tasks]
      operationId: addComment
      parameters:
        - name: taskId
          in: path
          required: true
          schema: { type: string }
      requestBody:
        required: true
        content:
          application/json:
            schema: { $ref: '#/components/schemas/PostComment' }
      responses:
        "201":
          description: Comment added
        "400": { $ref: '#/components/responses/BadRequest' }
        "401": { $ref: '#/components/responses/Unauthorized' }
        "404": { $ref: '#/components/responses/NotFound' }
        "500": { $ref: '#/components/responses/InternalError' }

  /tasks/{taskId}/comments/{commentId}:
    patch:
      tags: [Tasks]
      operationId: updateComment
      parameters:
        - name: taskId
          in: path
          required: true
          schema: { type: string }
        - name: commentId
          in: path
          required: true
          schema: { type: string }
      requestBody:
        required: true
        content:
          application/json:
            schema: { $ref: '#/components/schemas/PatchComment' }
      responses:
        "200":
          description: Comment updated
        "400": { $ref: '#/components/responses/BadRequest' }
        "401": { $ref: '#/components/responses/Unauthorized' }
        "403": { $ref: '#/components/responses/Forbidden' }
        "404": { $ref: '#/components/responses/NotFound' }
        "500": { $ref: '#/components/responses/InternalError' }
    delete:
      tags: [Tasks]
      operationId: deleteComment
      parameters:
        - name: taskId
          in: path
          required: true
          schema: { type: string }
        - name: commentId
          in: path
          required: true
          schema: { type: string }
      responses:
        "200":
          description: Comment deleted
        "401": { $ref: '#/components/responses/Unauthorized' }
        "403": { $ref: '#/components/responses/Forbidden' }
        "404": { $ref: '#/components/responses/NotFound' }
        "500": { $ref: '#/components/responses/InternalError' }

  /tasks/{taskId}/activity:
    get:
      tags: [Tasks]
      operationId: getTaskActivity
      parameters:
        - name: taskId
          in: path
          required: true
          schema: { type: string }
        - name: limit
          in: query
          schema: { type: integer, default: 20 }
        - name: offset
          in: query
          schema: { type: integer, default: 0 }
      responses:
        "200":
          description: Activity log retrieved
          content:
            application/json:
              schema:
                type: array
                items: { $ref: '#/components/schemas/Activity' }
        "401": { $ref: '#/components/responses/Unauthorized' }
        "404": { $ref: '#/components/responses/NotFound' }
        "500": { $ref: '#/components/responses/InternalError' }

  /search/users:
    get:
      tags: [Search]
      operationId: searchUsers
      parameters:
        - name: q
          in: query
          required: true
          schema: { type: string }
        - name: taskId
          in: query
          schema: { type: string }
      responses:
        "200":
          description: User search results
          content:
            application/json:
              schema:
                type: array
                items: { $ref: '#/components/schemas/UserSearchResult' }
        "401": { $ref: '#/components/responses/Unauthorized' }

  /search/tasks:
    get:
      tags: [Search]
      operationId: searchTasks
      parameters:
        - name: q
          in: query
          required: true
          schema: { type: string }
      responses:
        "200":
          description: Task search results
          content:
            application/json:
              schema:
                type: array
                items: { $ref: '#/components/schemas/TaskSearchResult' }
        "401": { $ref: '#/components/responses/Unauthorized' }
```

**Schemas to add (under `components.schemas`):**

```yaml
    PostComment:
      type: object
      required: [content]
      properties:
        content:
          type: string
          example: "Hey @[user:uuid-123] see @[task:uuid-456]"

    PatchComment:
      type: object
      required: [content]
      properties:
        content:
          type: string

    Activity:
      type: object
      properties:
        id: { type: string, format: uuid }
        taskId: { type: string, format: uuid }
        actorId: { type: string, format: uuid }
        actorName: { type: string }
        type:
          type: string
          enum: [status_changed, assigned, unassigned, title_updated,
                 priority_changed, due_date_set, description_set,
                 archived, reopened, comment_added, comment_updated, comment_deleted]
        payload: { type: object }
        createdAt: { type: string, format: date-time }

    UserSearchResult:
      type: object
      properties:
        id: { type: string, format: uuid }
        name: { type: string }
        email: { type: string, format: email }

    TaskSearchResult:
      type: object
      properties:
        id: { type: string, format: uuid }
        title: { type: string }
        status: { type: string }
```

**Step: Regenerate OpenAPI**

```bash
make openapi
```
Expected: `ports/openapi_api.gen.go` and `ports/openapi_types.gen.go` updated.

**Step: Commit**

```bash
git add ports/openapi.yml ports/openapi_api.gen.go ports/openapi_types.gen.go
git commit -m "feat: add comment, activity, and search endpoints to OpenAPI spec"
```

---

### Task 13: HTTP Handlers

**Files:**
- Modify: `ports/openapi_handler.go`

Add six new handler methods to `HttpHandler`, following the exact pattern of existing handlers:

```go
func (h HttpHandler) AddComment(w http.ResponseWriter, r *http.Request, taskId string) { ... }
func (h HttpHandler) UpdateComment(w http.ResponseWriter, r *http.Request, taskId, commentId string) { ... }
func (h HttpHandler) DeleteComment(w http.ResponseWriter, r *http.Request, taskId, commentId string) { ... }
func (h HttpHandler) GetTaskActivity(w http.ResponseWriter, r *http.Request, taskId string, params GetTaskActivityParams) { ... }
func (h HttpHandler) SearchUsers(w http.ResponseWriter, r *http.Request, params SearchUsersParams) { ... }
func (h HttpHandler) SearchTasks(w http.ResponseWriter, r *http.Request, params SearchTasksParams) { ... }
```

Each reads the caller from `userFromCtx`, decodes the body (if any), delegates to the app layer, and calls `responseSuccess` or `responseError`.

`SearchUsers` and `SearchTasks` need query handlers — add `SearchUsers` and `SearchTasks` to `app/query/` using the `TasksReadModel` pattern against the user/task repositories with the ranking logic. These are simple `LIKE` queries with `ORDER BY` clauses.

**Step: Run tests**

```bash
make test
```
Expected: PASS — all handler methods satisfy the generated interface.

**Step: Commit**

```bash
git add ports/openapi_handler.go app/query/
git commit -m "feat: implement HTTP handlers for comments, activity, and search"
```

---

### Task 14: Verify Full Build and Test Coverage

**Step 1: Run full test suite**

```bash
make test
```
Expected: all tests PASS.

**Step 2: Start the server and smoke test**

```bash
make start
```

Test `POST /api/tasks/{taskId}/comments` with a valid JWT and body `{"content": "Hey @[user:uuid-123]"}`.

**Step 3: Check coverage**

```bash
go test ./... -cover
```
Expected: high 90s line/branch coverage maintained.

**Step 4: Final commit**

```bash
git add .
git commit -m "feat: complete comments, activity log, and mention system"
```
