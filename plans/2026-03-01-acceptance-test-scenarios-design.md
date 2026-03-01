# Acceptance Test Scenarios Design

## Overview

This document defines the structure and scenarios for acceptance test `.txt` files following the framework defined in `plans/permanent/acceptance-test-framework.md`.

## Decisions

| Decision | Choice |
|----------|--------|
| Coverage | Comprehensive (all edge cases, invalid inputs, boundary conditions) |
| File organization | Single file per entity |
| Location | `/acceptanceTests/` at project root |
| Internal structure | Grouped by endpoint |
| Task permissions | Merged into `task.txt` |

## File Structure

```
acceptanceTests/
├── user.txt        # Auth, user CRUD, role management
├── task.txt        # Task lifecycle + permissions
├── comment.txt     # Comment CRUD, author restrictions
├── tag.txt         # Tag add/remove
└── activity.txt    # Audit trail verification
```

## Scenario Count Summary

| File | Scenarios |
|------|-----------|
| user.txt | 35 |
| task.txt | 55 |
| comment.txt | 15 |
| tag.txt | 10 |
| activity.txt | 12 |
| **Total** | **127** |

---

## 1. user.txt — 35 scenarios

### POST /api/auth/login (6 scenarios)

1. Login with valid credentials → JWT token returned
2. Login with invalid password → 401 error
3. Login with non-existent email → 401 error
4. Login with missing email → 400 error
5. Login with missing password → 400 error
6. Login with empty credentials → 400 error

### POST /api/users (8 scenarios)

1. Admin creates user with role admin → succeeds
2. Admin creates user with role employer → succeeds
3. Admin creates user with role employee → succeeds
4. Create user with duplicate email → 409 conflict
5. Create user with invalid email format → 400 error
6. Create user with missing required fields → 400 error
7. Create user with empty name → 400 error
8. Create user with weak password → 400 error

### GET /api/users (3 scenarios)

1. Admin lists all users → receives list
2. Non-admin lists users → forbidden
3. Empty database → empty list

### GET /api/users/{id} (3 scenarios)

1. Admin gets any user by ID → receives user
2. User gets own profile → succeeds
3. User gets another user's profile → forbidden

### PATCH /api/users/{id} (4 scenarios)

1. User updates own profile → succeeds
2. Admin updates another user's profile → succeeds
3. Non-admin updates another user's profile → forbidden
4. Update with empty name → 400 error

### DELETE /api/users/{id} (3 scenarios)

1. Admin deletes another user → succeeds (soft delete)
2. User deletes own account → succeeds
3. Non-admin deletes another user → forbidden

### PUT /api/users/{id}/role (8 scenarios)

1. Admin changes user role to employer → succeeds
2. Admin changes user role to employee → succeeds
3. Admin changes user role to admin → succeeds
4. Admin tries to change own role → forbidden
5. Non-admin changes role → forbidden
6. Change role of non-existent user → 404 error
7. Change role to invalid value → 400 error
8. Change role to same value → succeeds (idempotent)

---

## 2. task.txt — 55 scenarios

### POST /api/tasks (6 scenarios)

1. Admin creates task → succeeds
2. Employer creates task → succeeds
3. Employee creates task → forbidden
4. Create task with empty title → 400 error
5. Create task with missing title → 400 error
6. Create task with very long title → succeeds (or 400 if bounded)

### GET /api/tasks (8 scenarios)

1. Admin lists all tasks → sees all
2. Employer lists tasks → sees all
3. Employee lists tasks → sees only assigned tasks
4. Employee with no assignments → empty list
5. List tasks with pagination → correct offset/limit
6. List tasks includes archived → based on query param
7. List tasks filtered by status → correct results
8. List tasks sorted by due date → correct order

### PUT /api/tasks/{id}/status (12 scenarios)

1. Valid transition: todo → in_progress → succeeds
2. Valid transition: todo → pending → succeeds
3. Valid transition: pending → todo → succeeds
4. Valid transition: pending → in_progress → succeeds
5. Valid transition: in_progress → completed → succeeds
6. Valid transition: in_progress → pending → succeeds
7. Invalid transition: todo → completed → 422 error
8. Invalid transition: completed → any → 422 error (terminal state)
9. Assigned user changes status → succeeds
10. Unassigned employee changes status → forbidden
11. Non-admin/non-employer changes unassigned task → forbidden
12. Change status of non-existent task → 404 error

### PATCH /api/tasks/{id} (title) (6 scenarios)

1. Admin updates title → succeeds
2. Employer updates title → succeeds
3. Assigned user updates title → succeeds
4. Unassigned employee updates title → forbidden
5. Update title to empty → 400 error
6. Update non-existent task → 404 error

### PUT /api/tasks/{id}/priority (5 scenarios)

1. Admin sets priority → succeeds
2. Employer sets priority → succeeds
3. Employee sets priority → forbidden
4. Set invalid priority value → 400 error
5. Set priority on non-existent task → 404 error

### PATCH /api/tasks/{id}/description (5 scenarios)

1. Admin updates description → succeeds
2. Employer updates description → succeeds
3. Assigned user updates description → succeeds
4. Unassigned employee updates description → forbidden
5. Update description on non-existent task → 404 error

### PUT /api/tasks/{id}/due-date (3 scenarios)

1. Admin sets due date → succeeds
2. Employer sets due date → succeeds
3. Employee sets due date → forbidden

### PUT /api/tasks/{id}/assign/{userId} (5 scenarios)

1. Admin assigns task to employer → succeeds
2. Employer assigns task → succeeds
3. Admin assigns task to employee → fails (assignee must be employer)
4. Employee assigns task → forbidden
5. Assign to non-existent user → 404 error

### DELETE /api/tasks/{id}/assign (3 scenarios)

1. Admin unassigns task → succeeds
2. Employer unassigns task → succeeds
3. Employee unassigns task → forbidden

### PUT /api/tasks/{id}/archive (4 scenarios)

1. Admin archives task → succeeds (irreversible)
2. Employer archives task → succeeds
3. Employee archives task → forbidden
4. Archive already archived task → idempotent success

### PUT /api/tasks/{id}/reopen (3 scenarios)

1. Admin reopens archived task → succeeds
2. Employer reopens archived task → succeeds
3. Employee reopens task → forbidden

### DELETE /api/tasks/{id} (4 scenarios)

1. Admin deletes task → succeeds (soft delete)
2. Employer (creator) deletes own task → succeeds
3. Employer deletes another's task → forbidden
4. Employee deletes task → forbidden

---

## 3. comment.txt — 15 scenarios

### POST /api/tasks/{taskId}/comments (6 scenarios)

1. Admin adds comment → succeeds
2. Employer adds comment → succeeds
3. Assigned user adds comment → succeeds
4. Unassigned user adds comment → forbidden
5. Add comment with empty content → 400 error
6. Add comment to non-existent task → 404 error

### PATCH /api/tasks/{taskId}/comments/{id} (5 scenarios)

1. Author updates own comment → succeeds
2. Non-author updates comment → forbidden
3. Admin updates any comment → succeeds
4. Update comment with empty content → 400 error
5. Update comment on non-existent task → 404 error

### DELETE /api/tasks/{taskId}/comments/{id} (4 scenarios)

1. Author deletes own comment → succeeds
2. Non-author deletes comment → forbidden
3. Admin deletes any comment → succeeds
4. Delete non-existent comment → 404 error

---

## 4. tag.txt — 10 scenarios

### POST /api/tasks/{taskId}/tags (5 scenarios)

1. Admin adds tag → succeeds
2. Employer adds tag → succeeds
3. Employee adds tag → forbidden
4. Add tag with empty name → 400 error
5. Add tag to non-existent task → 404 error

### DELETE /api/tasks/{taskId}/tags/{tagId} (5 scenarios)

1. Admin removes tag → succeeds
2. Employer removes tag → succeeds
3. Employee removes tag → forbidden
4. Remove non-existent tag → 404 error
5. Remove tag from non-existent task → 404 error

---

## 5. activity.txt — 12 scenarios

### GET /api/tasks/{taskId}/activity (4 scenarios)

1. Admin gets activity log → succeeds
2. Employer gets activity log → succeeds
3. Employee gets activity of assigned task → succeeds
4. Employee gets activity of unassigned task → forbidden

### Activity entries created for mutations (8 scenarios)

1. Status change creates status_changed activity
2. Task assignment creates assigned activity
3. Task unassignment creates unassigned activity
4. Title update creates title_updated activity
5. Priority change creates priority_changed activity
6. Due date set creates due_date_set activity
7. Description set creates description_set activity
8. Comment add/update/delete creates corresponding activity

---

## Authorization Reference

| Operation | admin | employer | employee |
|-----------|:-----:|:--------:|:--------:|
| Create task | ✓ | ✓ | ✗ |
| Assign task | ✓ | ✓ | ✗ |
| Change task status | ✓ | ✓ | ✓ (if assigned) |
| Update task title | ✓ | ✓ | ✓ (if assigned) |
| Delete task | ✓ | ✓ (own) | ✗ |
| Archive task | ✓ | ✓ | ✗ |
| Reopen task | ✓ | ✓ | ✗ |
| Set priority | ✓ | ✓ | ✗ |
| Set description | ✓ | ✓ | ✓ (if assigned) |
| Add comment | ✓ | ✓ | ✓ (if assigned) |
| Edit own comment | ✓ | ✓ | ✓ |
| Delete own comment | ✓ | ✓ | ✓ |
| Change user role | ✓ | ✗ | ✗ |

## Status Transitions

```
todo        → [pending, in_progress]
pending     → [todo, in_progress]
in_progress → [completed, pending]
completed   → [] (terminal)
```

## Priority Values

- `low`
- `medium`
- `high`
- `urgent`

## Activity Types

- `status_changed`
- `assigned`
- `unassigned`
- `title_updated`
- `priority_changed`
- `due_date_set`
- `description_set`
- `archived`
- `reopened`
- `comment_added`
- `comment_updated`
- `comment_deleted`
