# Comments & Activity Log Design

## Overview

Add commenting, user mentions, task references, and activity logging to the Task Management API.

## Use Cases

| Use Case | Endpoint | Description |
|----------|----------|-------------|
| AddComment | `POST /api/tasks/{taskId}/comments` | Add comments with mentions |
| UpdateComment | `PATCH /api/tasks/{taskId}/comments/{commentId}` | Edit own comment |
| DeleteComment | `DELETE /api/tasks/{taskId}/comments/{commentId}` | Delete own comment |
| GetTaskActivityLog | `GET /api/tasks/{taskId}/activity` | View task history |
| SearchUsers | `GET /api/search/users` | Autocomplete for user mentions |
| SearchTasks | `GET /api/search/tasks` | Autocomplete for task references |

## Design Decisions

### Activity Log Scope
Captures all task mutations: status changes, assignments, title/description/priority/due date updates, archive/delete/reopen, and comment operations.

### Comment Authorization
Only comment author can update or delete their comment.

### Storage Strategy
- Single `task_activities` table with polymorphic events (type field)
- Separate `comments` and `comment_references` tables

### Mention Format
`@[type:id]` tokens stored in raw comment body:
- `@[user:uuid-123]` - user mention
- `@[task:uuid-456]` - task reference

---

## Domain Layer

### Comment Entity (`domain/comment/comment.go`)

```go
type Comment interface {
    UUID() string
    TaskID() string
    AuthorID() string
    Content() string
    References() []Reference
    CreatedAt() time.Time
    UpdatedAt() time.Time
    IsDeleted() bool
    
    Update(editor User, content string) error
    Delete(deleter User) error
}

type Reference struct {
    Type  ReferenceType
    ID    string
}

type ReferenceType string

const (
    ReferenceTypeUser ReferenceType = "user"
    ReferenceTypeTask ReferenceType = "task"
)
```

**Behavior:**
- `NewComment(author, taskID, content)` parses `@\[(\w+):([^\]]+)\]` tokens
- `Update`/`Delete` only allowed by comment author
- References validated on create/update (dropped if invalid)

### Activity Entity (`domain/activity/activity.go`)

```go
type Activity interface {
    UUID() string
    TaskID() string
    ActorID() string
    Type() ActivityType
    Payload() map[string]any
    CreatedAt() time.Time
}

type ActivityType string

const (
    ActivityTypeStatusChanged   ActivityType = "status_changed"
    ActivityTypeAssigned        ActivityType = "assigned"
    ActivityTypeUnassigned      ActivityType = "unassigned"
    ActivityTypeTitleUpdated    ActivityType = "title_updated"
    ActivityTypePriorityChanged ActivityType = "priority_changed"
    ActivityTypeDueDateSet      ActivityType = "due_date_set"
    ActivityTypeDescriptionSet  ActivityType = "description_set"
    ActivityTypeArchived        ActivityType = "archived"
    ActivityTypeReopened        ActivityType = "reopened"
    ActivityTypeCommentAdded    ActivityType = "comment_added"
    ActivityTypeCommentUpdated  ActivityType = "comment_updated"
    ActivityTypeCommentDeleted  ActivityType = "comment_deleted"
)
```

---

## Application Layer

### Commands (`app/command/`)

| File | Handler | Description |
|------|---------|-------------|
| `add_comment.go` | `AddCommentHandler` | Creates Comment + Activity |
| `update_comment.go` | `UpdateCommentHandler` | Updates Comment + Activity |
| `delete_comment.go` | `DeleteCommentHandler` | Soft deletes Comment + Activity |

### Queries (`app/query/`)

| File | Handler | Description |
|------|---------|-------------|
| `task_activities.go` | `TaskActivitiesHandler` | Returns activities with actor names |
| `search_users.go` | `SearchUsersHandler` | Ranked user autocomplete |
| `search_tasks.go` | `SearchTasksHandler` | Ranked task autocomplete |

### Repository Interfaces (`app/command/adapters.go`)

```go
type CommentRepository interface {
    Add(ctx context.Context, c comment.Comment) error
    UpdateByID(ctx context.Context, uuid string, updateFn CommentUpdater) error
}

type CommentUpdater func(context.Context, comment.Comment) (comment.Comment, error)

type ActivityRepository interface {
    Add(ctx context.Context, a activity.Activity) error
    FindByTaskID(ctx context.Context, taskID string, limit, offset int) ([]activity.Activity, error)
}
```

### Activity Recording

- Comment commands create activities directly
- Existing commands inject `ActivityRepository` and record after successful mutations

---

## Adapters Layer

### Database Schema

```sql
CREATE TABLE comments (
    id VARCHAR(36) PRIMARY KEY,
    task_id VARCHAR(36) NOT NULL,
    author_id VARCHAR(36) NOT NULL,
    content TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL,
    deleted_at TIMESTAMP NULL,
    INDEX idx_comments_task_id (task_id),
    INDEX idx_comments_author_id (author_id)
);

CREATE TABLE comment_references (
    id VARCHAR(36) PRIMARY KEY,
    comment_id VARCHAR(36) NOT NULL,
    reference_type VARCHAR(20) NOT NULL,
    reference_id VARCHAR(100) NOT NULL,
    created_at TIMESTAMP NOT NULL,
    INDEX idx_refs_comment_id (comment_id),
    INDEX idx_refs_type_id (reference_type, reference_id)
);

CREATE TABLE task_activities (
    id VARCHAR(36) PRIMARY KEY,
    task_id VARCHAR(36) NOT NULL,
    actor_id VARCHAR(36) NOT NULL,
    type VARCHAR(50) NOT NULL,
    payload JSON,
    created_at TIMESTAMP NOT NULL,
    INDEX idx_activities_task_id (task_id, created_at DESC)
);
```

### Repository Implementations

| File | Description |
|------|-------------|
| `mysql_comment_repository.go` | Add, UpdateByID |
| `mysql_activity_repository.go` | Add, FindByTaskID |

---

## Ports Layer (OpenAPI)

### New Endpoints

```yaml
/tasks/{taskId}/comments:
  post:
    operationId: addComment
    parameters: [taskId]
    requestBody:
      content:
        application/json:
          schema:
            $ref: '#/components/schemas/PostComment'
    responses: [201, 400, 401, 404]

/tasks/{taskId}/comments/{commentId}:
  patch:
    operationId: updateComment
    parameters: [taskId, commentId]
    requestBody:
      content:
        application/json:
          schema:
            $ref: '#/components/schemas/PatchComment'
    responses: [200, 400, 401, 403, 404]
  delete:
    operationId: deleteComment
    parameters: [taskId, commentId]
    responses: [200, 401, 403, 404]

/tasks/{taskId}/activity:
  get:
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
    responses: [200, 401, 404]

/search/users:
  get:
    operationId: searchUsers
    parameters:
      - name: q
        in: query
        required: true
        schema: { type: string }
      - name: taskId
        in: query
        schema: { type: string }
    responses: [200, 401]

/search/tasks:
  get:
    operationId: searchTasks
    parameters:
      - name: q
        in: query
        required: true
        schema: { type: string }
    responses: [200, 401]
```

### Schemas

```yaml
PostComment:
  type: object
  required: [content]
  properties:
    content:
      type: string
      example: "Hey @[user:uuid-123] can you review @[task:uuid-456]?"

PatchComment:
  type: object
  required: [content]
  properties:
    content:
      type: string

Comment:
  type: object
  properties:
    id:
      type: string
      format: uuid
    taskId:
      type: string
      format: uuid
    authorId:
      type: string
      format: uuid
    authorName:
      type: string
    content:
      type: string
    references:
      type: array
      items:
        $ref: '#/components/schemas/Reference'
    createdAt:
      type: string
      format: date-time
    updatedAt:
      type: string
      format: date-time

Reference:
  type: object
  properties:
    type:
      type: string
      enum: [user, task]
    id:
      type: string
    displayName:
      type: string

Activity:
  type: object
  properties:
    id:
      type: string
      format: uuid
    taskId:
      type: string
      format: uuid
    actorId:
      type: string
      format: uuid
    actorName:
      type: string
    type:
      type: string
      enum: [status_changed, assigned, unassigned, title_updated, 
             priority_changed, due_date_set, description_set, 
             archived, reopened, comment_added, comment_updated, 
             comment_deleted]
    payload:
      type: object
    createdAt:
      type: string
      format: date-time

UserSearchResult:
  type: object
  properties:
    id:
      type: string
      format: uuid
    name:
      type: string
    email:
      type: string

TaskSearchResult:
  type: object
  properties:
    id:
      type: string
      format: uuid
    title:
      type: string
    status:
      type: string
```

---

## Autocomplete Ranking

### Users
1. Task assignee
2. Prior commenters on task
3. All other users (alphabetical)

### Tasks
1. Recently updated
2. Open (todo/in_progress) before done/archived

---

## Implementation Order

1. Domain: `activity` entity + tests
2. Domain: `comment` entity + tests
3. Adapters: `mysql_activity_repository` + tests
4. Adapters: `mysql_comment_repository` + tests
5. App: activity queries + tests
6. App: comment commands + tests
7. Ports: OpenAPI spec update
8. Ports: HTTP handlers + tests
9. Integration: wire dependencies in `cmd/`
