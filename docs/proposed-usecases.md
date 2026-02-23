# Proposed Use Cases

This document lists all proposed use cases for the Task Management API, organized by category.

---

## Task Lifecycle

| Use Case | Endpoint | Description |
|----------|----------|-------------|
| UpdateTaskTitle | `PATCH /api/tasks/{taskId}` | Update task title |
| DeleteTask | `DELETE /api/tasks/{taskId}` | Soft delete a task |
| UnassignTask | `DELETE /api/tasks/{taskId}/assign` | Remove assignee from task |
| ReopenTask | `PUT /api/tasks/{taskId}/reopen` | Reopen a completed task |
| ArchiveTask | `PUT /api/tasks/{taskId}/archive` | Archive a completed task |

---

## Task Details

| Use Case | Endpoint | Description |
|----------|----------|-------------|
| SetTaskPriority | `PUT /api/tasks/{taskId}/priority` | Set priority level (low, medium, high, urgent) |
| SetTaskDueDate | `PUT /api/tasks/{taskId}/due-date` | Set or update due date |
| AddTaskDescription | `PATCH /api/tasks/{taskId}` | Add or update task description |
| AddTaskTag | `POST /api/tasks/{taskId}/tags` | Add tags/labels to organize tasks |
| RemoveTaskTag | `DELETE /api/tasks/{taskId}/tags/{tagId}` | Remove tags from tasks |

---

## Collaboration

| Use Case | Endpoint | Description |
|----------|----------|-------------|
| AddComment | `POST /api/tasks/{taskId}/comments` | Add comments to a task |
| DeleteComment | `DELETE /api/tasks/{taskId}/comments/{commentId}` | Remove a comment |
| UpdateComment | `PATCH /api/tasks/{taskId}/comments/{commentId}` | Edit an existing comment |
| GetTaskActivityLog | `GET /api/tasks/{taskId}/activity` | View task history and activity |
| MentionUser | N/A | Mention users in comments (part of AddComment) |

---

## User Management

| Use Case | Endpoint | Description |
|----------|----------|-------------|
| UpdateUserRole | `PUT /api/users/{userId}/role` | Promote/demote users (admin ↔ employee) |
| DeleteUser | `DELETE /api/users/{userId}` | Remove a user |
| GetUser | `GET /api/users/{userId}` | Get user details by ID |
| ListUsers | `GET /api/users` | Get all users (for admin to assign tasks) |
| UpdateUserProfile | `PATCH /api/users/{userId}` | Update user information |
| Login | `POST /api/auth/login` | Issue JWT for authentication |

---

## Search & Filtering

| Use Case | Endpoint | Description |
|----------|----------|-------------|
| SearchTasks | `GET /api/tasks?q={keyword}` | Search tasks by keyword/title/description |
| FilterTasksByStatus | `GET /api/tasks?status={status}` | Filter tasks by status |
| FilterTasksByPriority | `GET /api/tasks?priority={priority}` | Filter tasks by priority level |
| FilterTasksByDueDate | `GET /api/tasks?from={date}&to={date}` | Filter tasks by due date range |
| FilterTasksByAssignee | `GET /api/tasks?assignee={userId}` | Filter tasks by assigned user |
| PaginatedTasks | `GET /api/tasks?limit={n}&offset={n}` | Paginated task listing |

---

## Subtasks

| Use Case | Endpoint | Description |
|----------|----------|-------------|
| CreateSubtask | `POST /api/tasks/{taskId}/subtasks` | Create a subtask within a parent task |
| CompleteSubtask | `PUT /api/tasks/{taskId}/subtasks/{subtaskId}/complete` | Mark subtask as complete |
| DeleteSubtask | `DELETE /api/tasks/{taskId}/subtasks/{subtaskId}` | Remove a subtask |
| UpdateSubtask | `PATCH /api/tasks/{taskId}/subtasks/{subtaskId}` | Update subtask details |
| ListSubtasks | `GET /api/tasks/{taskId}/subtasks` | List all subtasks for a task |

---

## Analytics & Reports

| Use Case | Endpoint | Description |
|----------|----------|-------------|
| GetTaskStatistics | `GET /api/tasks/stats` | Get counts by status, priority, assignee |
| GetDashboardSummary | `GET /api/dashboard` | Get overview for user (assigned, overdue, upcoming) |
| GetOverdueTasks | `GET /api/tasks/overdue` | Get tasks past their due date |
| GetUpcomingTasks | `GET /api/tasks/upcoming?days={n}` | Get tasks due within next X days |
| GetTaskCompletionReport | `GET /api/reports/completion` | Generate completion report for date range |
| GetUserPerformance | `GET /api/users/{userId}/performance` | Get performance metrics per user |

---

## Bulk Operations

| Use Case | Endpoint | Description |
|----------|----------|-------------|
| BulkAssignTasks | `POST /api/tasks/bulk/assign` | Assign multiple tasks to a user |
| BulkUpdateStatus | `POST /api/tasks/bulk/status` | Update status of multiple tasks |
| BulkDeleteTasks | `POST /api/tasks/bulk/delete` | Delete multiple tasks at once |
| BulkArchiveTasks | `POST /api/tasks/bulk/archive` | Archive multiple tasks at once |

---

## Watch/Subscribe

| Use Case | Endpoint | Description |
|----------|----------|-------------|
| WatchTask | `POST /api/tasks/{taskId}/watch` | Subscribe to task updates/notifications |
| UnwatchTask | `DELETE /api/tasks/{taskId}/watch` | Unsubscribe from task notifications |
| ListWatchedTasks | `GET /api/tasks/watching` | List tasks user is watching |

---

## Task Attachments

| Use Case | Endpoint | Description |
|----------|----------|-------------|
| UploadAttachment | `POST /api/tasks/{taskId}/attachments` | Upload file attachment to task |
| DeleteAttachment | `DELETE /api/tasks/{taskId}/attachments/{attachmentId}` | Remove attachment |
| ListAttachments | `GET /api/tasks/{taskId}/attachments` | List all attachments for a task |
| DownloadAttachment | `GET /api/tasks/{taskId}/attachments/{attachmentId}` | Download attachment |

---

## Time Tracking

| Use Case | Endpoint | Description |
|----------|----------|-------------|
| StartTimer | `POST /api/tasks/{taskId}/timer/start` | Start time tracking on task |
| StopTimer | `POST /api/tasks/{taskId}/timer/stop` | Stop time tracking |
| GetTimeLogs | `GET /api/tasks/{taskId}/time-logs` | Get time tracking history |
| GetTotalTime | `GET /api/tasks/{taskId}/time-logs/total` | Get total time spent on task |

---

## Notifications

| Use Case | Endpoint | Description |
|----------|----------|-------------|
| ListNotifications | `GET /api/notifications` | List user notifications |
| MarkNotificationRead | `PUT /api/notifications/{notificationId}/read` | Mark notification as read |
| MarkAllNotificationsRead | `PUT /api/notifications/read-all` | Mark all notifications as read |
| GetUnreadCount | `GET /api/notifications/unread-count` | Get count of unread notifications |

---

## Priority Levels

| Priority | Description |
|----------|-------------|
| `low` | Low priority, no urgency |
| `medium` | Normal priority (default) |
| `high` | High priority, needs attention |
| `urgent` | Critical priority, immediate action required |

---

## Status Values

| Status | Description |
|--------|-------------|
| `todo` | Task created, not started |
| `pending` | Task waiting for something |
| `inprogress` | Task being worked on |
| `completed` | Task finished |
| `archived` | Task archived (not deleted) |

---

## Implementation Priority

### P0 - Critical (Must Have)
1. UpdateTaskTitle
2. DeleteTask
3. Login
4. GetUser
5. ListUsers
6. PaginatedTasks

### P1 - High (Should Have)
1. SetTaskDueDate
2. SetTaskPriority
3. UnassignTask
4. ReopenTask
5. SearchTasks / FilterTasks
6. GetTaskStatistics

### P2 - Medium (Nice to Have)
1. ArchiveTask
2. AddComment
3. UpdateUserRole
4. GetDashboardSummary
5. GetOverdueTasks

### P3 - Low (Future)
1. Subtasks
2. Time Tracking
3. Attachments
4. Bulk Operations
5. Notifications
6. Tags/Labels
