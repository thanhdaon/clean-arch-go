# Acceptance Test Scenarios Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Create 5 acceptance test `.txt` files in `/acceptanceTests/` with 127 comprehensive scenarios following the GIVEN/WHEN/THEN format.

**Architecture:** Each `.txt` file is organized by endpoint, with scenarios separated by `===` delimiters. Scenarios use the framework defined in `plans/permanent/acceptance-test-framework.md`.

**Tech Stack:** Plain text files with GIVEN/WHEN/THEN syntax; no code generation.

---

## Task 1: Create directory and user.txt (35 scenarios)

**Files:**
- Create: `acceptanceTests/user.txt`

**Step 1: Create acceptanceTests directory**

```bash
mkdir -p acceptanceTests
```

**Step 2: Write user.txt with all 35 scenarios**

Create `acceptanceTests/user.txt`:

```
;===============================================================
; user.txt - Auth, user CRUD, role management
;===============================================================

;===============================================================
; POST /api/auth/login - Login scenarios
;===============================================================

;===============================================================
; Login with valid credentials returns JWT token
;===============================================================
GIVEN a user exists with role employer.
GIVEN the user has email "test@example.com" and password "password123".

WHEN the user logs in with email "test@example.com" and password "password123".

THEN no error is returned.
and the response contains a JWT token.

;===============================================================
; Login with invalid password returns 401 error
;===============================================================
GIVEN a user exists with role employer.
GIVEN the user has email "test@example.com" and password "password123".

WHEN the user logs in with email "test@example.com" and password "wrongpassword".

THEN an unauthorized error is returned.

;===============================================================
; Login with non-existent email returns 401 error
;===============================================================
GIVEN no user exists with email "nonexistent@example.com".

WHEN the user logs in with email "nonexistent@example.com" and password "password123".

THEN an unauthorized error is returned.

;===============================================================
; Login with missing email returns 400 error
;===============================================================
GIVEN no preconditions.

WHEN the user logs in with empty email and password "password123".

THEN a bad request error is returned.

;===============================================================
; Login with missing password returns 400 error
;===============================================================
GIVEN no preconditions.

WHEN the user logs in with email "test@example.com" and empty password.

THEN a bad request error is returned.

;===============================================================
; Login with empty credentials returns 400 error
;===============================================================
GIVEN no preconditions.

WHEN the user logs in with empty email and empty password.

THEN a bad request error is returned.

;===============================================================
; POST /api/users - User creation scenarios
;===============================================================

;===============================================================
; Admin creates user with role admin
;===============================================================
GIVEN an admin user is authenticated.

WHEN the admin creates a user with role "admin", name "Admin User", email "admin@example.com", and password "password123".

THEN no error is returned.
and the user exists in the database with role "admin".

;===============================================================
; Admin creates user with role employer
;===============================================================
GIVEN an admin user is authenticated.

WHEN the admin creates a user with role "employer", name "Employer User", email "employer@example.com", and password "password123".

THEN no error is returned.
and the user exists in the database with role "employer".

;===============================================================
; Admin creates user with role employee
;===============================================================
GIVEN an admin user is authenticated.

WHEN the admin creates a user with role "employee", name "Employee User", email "employee@example.com", and password "password123".

THEN no error is returned.
and the user exists in the database with role "employee".

;===============================================================
; Create user with duplicate email returns 409 conflict
;===============================================================
GIVEN a user exists with email "existing@example.com".

WHEN the admin creates a user with email "existing@example.com".

THEN a conflict error is returned.

;===============================================================
; Create user with invalid email format returns 400 error
;===============================================================
GIVEN an admin user is authenticated.

WHEN the admin creates a user with email "invalid-email".

THEN a bad request error is returned.

;===============================================================
; Create user with missing required fields returns 400 error
;===============================================================
GIVEN an admin user is authenticated.

WHEN the admin creates a user with empty name.

THEN a bad request error is returned.

;===============================================================
; Create user with empty name returns 400 error
;===============================================================
GIVEN an admin user is authenticated.

WHEN the admin creates a user with name "".

THEN a bad request error is returned.

;===============================================================
; Create user with weak password returns 400 error
;===============================================================
GIVEN an admin user is authenticated.

WHEN the admin creates a user with password "123".

THEN a bad request error is returned.

;===============================================================
; GET /api/users - List users scenarios
;===============================================================

;===============================================================
; Admin lists all users
;===============================================================
GIVEN an admin user is authenticated.
GIVEN multiple users exist in the database.

WHEN the admin requests the user list.

THEN no error is returned.
and the response contains all users.

;===============================================================
; Non-admin lists users returns forbidden
;===============================================================
GIVEN an employer user is authenticated.

WHEN the employer requests the user list.

THEN a forbidden error is returned.

;===============================================================
; List users with empty database returns empty list
;===============================================================
GIVEN an admin user is authenticated.
GIVEN no users exist except the admin.

WHEN the admin requests the user list.

THEN no error is returned.
and the response contains an empty list.

;===============================================================
; GET /api/users/{id} - Get user scenarios
;===============================================================

;===============================================================
; Admin gets any user by ID
;===============================================================
GIVEN an admin user is authenticated.
GIVEN a user exists with ID "user-123".

WHEN the admin requests user "user-123".

THEN no error is returned.
and the response contains the user details.

;===============================================================
; User gets own profile
;===============================================================
GIVEN a user is authenticated with ID "user-123".

WHEN the user requests their own profile "user-123".

THEN no error is returned.
and the response contains the user details.

;===============================================================
; User gets another user's profile returns forbidden
;===============================================================
GIVEN a user is authenticated with ID "user-123".
GIVEN another user exists with ID "user-456".

WHEN the user requests profile "user-456".

THEN a forbidden error is returned.

;===============================================================
; PATCH /api/users/{id} - Update profile scenarios
;===============================================================

;===============================================================
; User updates own profile
;===============================================================
GIVEN a user is authenticated with ID "user-123".

WHEN the user updates their profile with name "New Name".

THEN no error is returned.
and the user has name "New Name" in the database.

;===============================================================
; Admin updates another user's profile
;===============================================================
GIVEN an admin user is authenticated.
GIVEN a user exists with ID "user-123".

WHEN the admin updates user "user-123" with name "New Name".

THEN no error is returned.
and the user has name "New Name" in the database.

;===============================================================
; Non-admin updates another user's profile returns forbidden
;===============================================================
GIVEN an employer user is authenticated with ID "user-123".
GIVEN another user exists with ID "user-456".

WHEN the employer updates user "user-456" with name "New Name".

THEN a forbidden error is returned.

;===============================================================
; Update with empty name returns 400 error
;===============================================================
GIVEN a user is authenticated with ID "user-123".

WHEN the user updates their profile with name "".

THEN a bad request error is returned.

;===============================================================
; DELETE /api/users/{id} - Delete user scenarios
;===============================================================

;===============================================================
; Admin deletes another user
;===============================================================
GIVEN an admin user is authenticated.
GIVEN a user exists with ID "user-123".

WHEN the admin deletes user "user-123".

THEN no error is returned.
and the user is soft-deleted in the database.

;===============================================================
; User deletes own account
;===============================================================
GIVEN a user is authenticated with ID "user-123".

WHEN the user deletes their own account "user-123".

THEN no error is returned.
and the user is soft-deleted in the database.

;===============================================================
; Non-admin deletes another user returns forbidden
;===============================================================
GIVEN an employer user is authenticated with ID "user-123".
GIVEN another user exists with ID "user-456".

WHEN the employer deletes user "user-456".

THEN a forbidden error is returned.

;===============================================================
; PUT /api/users/{id}/role - Role management scenarios
;===============================================================

;===============================================================
; Admin changes user role to employer
;===============================================================
GIVEN an admin user is authenticated.
GIVEN a user exists with ID "user-123" and role "employee".

WHEN the admin changes user "user-123" role to "employer".

THEN no error is returned.
and the user has role "employer" in the database.

;===============================================================
; Admin changes user role to employee
;===============================================================
GIVEN an admin user is authenticated.
GIVEN a user exists with ID "user-123" and role "employer".

WHEN the admin changes user "user-123" role to "employee".

THEN no error is returned.
and the user has role "employee" in the database.

;===============================================================
; Admin changes user role to admin
;===============================================================
GIVEN an admin user is authenticated.
GIVEN a user exists with ID "user-123" and role "employer".

WHEN the admin changes user "user-123" role to "admin".

THEN no error is returned.
and the user has role "admin" in the database.

;===============================================================
; Admin tries to change own role returns forbidden
;===============================================================
GIVEN an admin user is authenticated with ID "admin-123".

WHEN the admin changes their own role to "employer".

THEN a forbidden error is returned.

;===============================================================
; Non-admin changes role returns forbidden
;===============================================================
GIVEN an employer user is authenticated.
GIVEN a user exists with ID "user-123".

WHEN the employer changes user "user-123" role to "admin".

THEN a forbidden error is returned.

;===============================================================
; Change role of non-existent user returns 404 error
;===============================================================
GIVEN an admin user is authenticated.
GIVEN no user exists with ID "nonexistent-123".

WHEN the admin changes user "nonexistent-123" role to "employer".

THEN a not found error is returned.

;===============================================================
; Change role to invalid value returns 400 error
;===============================================================
GIVEN an admin user is authenticated.
GIVEN a user exists with ID "user-123".

WHEN the admin changes user "user-123" role to "invalid_role".

THEN a bad request error is returned.

;===============================================================
; Change role to same value succeeds (idempotent)
;===============================================================
GIVEN an admin user is authenticated.
GIVEN a user exists with ID "user-123" and role "employer".

WHEN the admin changes user "user-123" role to "employer".

THEN no error is returned.
and the user has role "employer" in the database.
```

**Step 3: Commit**

```bash
git add acceptanceTests/user.txt
git commit -m "feat: add user.txt acceptance test scenarios (35 scenarios)"
```

---

## Task 2: Create task.txt (55 scenarios)

**Files:**
- Create: `acceptanceTests/task.txt`

**Step 1: Write task.txt with all 55 scenarios**

Create `acceptanceTests/task.txt`:

```
;===============================================================
; task.txt - Task lifecycle and permissions
;===============================================================

;===============================================================
; POST /api/tasks - Task creation scenarios
;===============================================================

;===============================================================
; Admin creates task
;===============================================================
GIVEN an admin user is authenticated.

WHEN the admin creates a task with title "New Task".

THEN no error is returned.
and the task exists in the database with status "todo".

;===============================================================
; Employer creates task
;===============================================================
GIVEN an employer user is authenticated.

WHEN the employer creates a task with title "New Task".

THEN no error is returned.
and the task exists in the database with status "todo".

;===============================================================
; Employee creates task returns forbidden
;===============================================================
GIVEN an employee user is authenticated.

WHEN the employee creates a task with title "New Task".

THEN a forbidden error is returned.

;===============================================================
; Create task with empty title returns 400 error
;===============================================================
GIVEN an admin user is authenticated.

WHEN the admin creates a task with title "".

THEN a bad request error is returned.

;===============================================================
; Create task with missing title returns 400 error
;===============================================================
GIVEN an admin user is authenticated.

WHEN the admin creates a task without a title.

THEN a bad request error is returned.

;===============================================================
; Create task with very long title succeeds
;===============================================================
GIVEN an admin user is authenticated.

WHEN the admin creates a task with a 500-character title.

THEN no error is returned.

;===============================================================
; GET /api/tasks - List tasks scenarios
;===============================================================

;===============================================================
; Admin lists all tasks
;===============================================================
GIVEN an admin user is authenticated.
GIVEN multiple tasks exist in the database.

WHEN the admin requests the task list.

THEN no error is returned.
and the response contains all tasks.

;===============================================================
; Employer lists tasks
;===============================================================
GIVEN an employer user is authenticated.
GIVEN multiple tasks exist in the database.

WHEN the employer requests the task list.

THEN no error is returned.
and the response contains all tasks.

;===============================================================
; Employee lists only assigned tasks
;===============================================================
GIVEN an employee user is authenticated with ID "emp-123".
GIVEN a task exists with status todo.
GIVEN the task is assigned to the employee.
GIVEN another task exists without assignment.

WHEN the employee requests the task list.

THEN no error is returned.
and the response contains only the assigned task.

;===============================================================
; Employee with no assignments gets empty list
;===============================================================
GIVEN an employee user is authenticated.
GIVEN tasks exist but none are assigned to the employee.

WHEN the employee requests the task list.

THEN no error is returned.
and the response contains an empty list.

;===============================================================
; List tasks with pagination
;===============================================================
GIVEN an admin user is authenticated.
GIVEN 50 tasks exist in the database.

WHEN the admin requests the task list with limit 10 and offset 20.

THEN no error is returned.
and the response contains 10 tasks.
and the tasks are from offset 20.

;===============================================================
; List tasks includes archived
;===============================================================
GIVEN an admin user is authenticated.
GIVEN a task exists with status todo.
GIVEN the task is archived.

WHEN the admin requests the task list including archived.

THEN no error is returned.
and the response contains the archived task.

;===============================================================
; List tasks filtered by status
;===============================================================
GIVEN an admin user is authenticated.
GIVEN a task exists with status "todo".
GIVEN a task exists with status "in_progress".

WHEN the admin requests the task list filtered by status "todo".

THEN no error is returned.
and the response contains only tasks with status "todo".

;===============================================================
; List tasks sorted by due date
;===============================================================
GIVEN an admin user is authenticated.
GIVEN multiple tasks exist with different due dates.

WHEN the admin requests the task list sorted by due date ascending.

THEN no error is returned.
and the tasks are ordered by due date ascending.

;===============================================================
; PUT /api/tasks/{id}/status - Status change scenarios
;===============================================================

;===============================================================
; Valid transition: todo to in_progress
;===============================================================
GIVEN a task exists with status todo.

WHEN the admin changes the task status to "in_progress".

THEN no error is returned.
and the task has status "in_progress" in the database.

;===============================================================
; Valid transition: todo to pending
;===============================================================
GIVEN a task exists with status todo.

WHEN the admin changes the task status to "pending".

THEN no error is returned.
and the task has status "pending" in the database.

;===============================================================
; Valid transition: pending to todo
;===============================================================
GIVEN a task exists with status pending.

WHEN the admin changes the task status to "todo".

THEN no error is returned.
and the task has status "todo" in the database.

;===============================================================
; Valid transition: pending to in_progress
;===============================================================
GIVEN a task exists with status pending.

WHEN the admin changes the task status to "in_progress".

THEN no error is returned.
and the task has status "in_progress" in the database.

;===============================================================
; Valid transition: in_progress to completed
;===============================================================
GIVEN a task exists with status in_progress.

WHEN the admin changes the task status to "completed".

THEN no error is returned.
and the task has status "completed" in the database.

;===============================================================
; Valid transition: in_progress to pending
;===============================================================
GIVEN a task exists with status in_progress.

WHEN the admin changes the task status to "pending".

THEN no error is returned.
and the task has status "pending" in the database.

;===============================================================
; Invalid transition: todo to completed
;===============================================================
GIVEN a task exists with status todo.

WHEN the admin changes the task status to "completed".

THEN an error is returned.
and the task still has status "todo" in the database.

;===============================================================
; Invalid transition: completed to any
;===============================================================
GIVEN a task exists with status completed.

WHEN the admin changes the task status to "todo".

THEN an error is returned.
and the task still has status "completed" in the database.

;===============================================================
; Assigned user changes status
;===============================================================
GIVEN a task exists with status todo.
GIVEN a user exists with role employer.
GIVEN the task is assigned to the user.

WHEN the assigned user changes the task status to "in_progress".

THEN no error is returned.
and the task has status "in_progress" in the database.

;===============================================================
; Unassigned employee changes status returns forbidden
;===============================================================
GIVEN a task exists with status todo.
GIVEN an employee user is authenticated.
GIVEN the task is not assigned to the employee.

WHEN the employee changes the task status to "in_progress".

THEN a forbidden error is returned.
and the task still has status "todo" in the database.

;===============================================================
; Non-admin/non-employer changes unassigned task returns forbidden
;===============================================================
GIVEN a task exists with status todo.
GIVEN an employee user is authenticated.
GIVEN the task is not assigned to the employee.

WHEN the employee changes the task status to "in_progress".

THEN a forbidden error is returned.

;===============================================================
; Change status of non-existent task returns 404 error
;===============================================================
GIVEN an admin user is authenticated.
GIVEN no task exists with ID "nonexistent-123".

WHEN the admin changes task "nonexistent-123" status to "in_progress".

THEN a not found error is returned.

;===============================================================
; PATCH /api/tasks/{id} - Title update scenarios
;===============================================================

;===============================================================
; Admin updates title
;===============================================================
GIVEN a task exists with title "Old Title".

WHEN the admin updates the task title to "New Title".

THEN no error is returned.
and the task has title "New Title" in the database.

;===============================================================
; Employer updates title
;===============================================================
GIVEN an employer user is authenticated.
GIVEN a task exists with title "Old Title".

WHEN the employer updates the task title to "New Title".

THEN no error is returned.
and the task has title "New Title" in the database.

;===============================================================
; Assigned user updates title
;===============================================================
GIVEN a task exists with title "Old Title".
GIVEN a user exists with role employer.
GIVEN the task is assigned to the user.

WHEN the assigned user updates the task title to "New Title".

THEN no error is returned.
and the task has title "New Title" in the database.

;===============================================================
; Unassigned employee updates title returns forbidden
;===============================================================
GIVEN a task exists with title "Old Title".
GIVEN an employee user is authenticated.
GIVEN the task is not assigned to the employee.

WHEN the employee updates the task title to "New Title".

THEN a forbidden error is returned.
and the task still has title "Old Title" in the database.

;===============================================================
; Update title to empty returns 400 error
;===============================================================
GIVEN a task exists with title "Old Title".

WHEN the admin updates the task title to "".

THEN a bad request error is returned.

;===============================================================
; Update non-existent task returns 404 error
;===============================================================
GIVEN an admin user is authenticated.
GIVEN no task exists with ID "nonexistent-123".

WHEN the admin updates task "nonexistent-123" title to "New Title".

THEN a not found error is returned.

;===============================================================
; PUT /api/tasks/{id}/priority - Priority scenarios
;===============================================================

;===============================================================
; Admin sets priority
;===============================================================
GIVEN a task exists with priority "medium".

WHEN the admin sets the task priority to "high".

THEN no error is returned.
and the task has priority "high" in the database.

;===============================================================
; Employer sets priority
;===============================================================
GIVEN an employer user is authenticated.
GIVEN a task exists with priority "medium".

WHEN the employer sets the task priority to "high".

THEN no error is returned.
and the task has priority "high" in the database.

;===============================================================
; Employee sets priority returns forbidden
;===============================================================
GIVEN an employee user is authenticated.
GIVEN a task exists with priority "medium".

WHEN the employee sets the task priority to "high".

THEN a forbidden error is returned.

;===============================================================
; Set invalid priority value returns 400 error
;===============================================================
GIVEN a task exists.

WHEN the admin sets the task priority to "invalid".

THEN a bad request error is returned.

;===============================================================
; Set priority on non-existent task returns 404 error
;===============================================================
GIVEN an admin user is authenticated.
GIVEN no task exists with ID "nonexistent-123".

WHEN the admin sets task "nonexistent-123" priority to "high".

THEN a not found error is returned.

;===============================================================
; PATCH /api/tasks/{id}/description - Description scenarios
;===============================================================

;===============================================================
; Admin updates description
;===============================================================
GIVEN a task exists with description "Old description".

WHEN the admin updates the task description to "New description".

THEN no error is returned.
and the task has description "New description" in the database.

;===============================================================
; Employer updates description
;===============================================================
GIVEN an employer user is authenticated.
GIVEN a task exists with description "Old description".

WHEN the employer updates the task description to "New description".

THEN no error is returned.
and the task has description "New description" in the database.

;===============================================================
; Assigned user updates description
;===============================================================
GIVEN a task exists with description "Old description".
GIVEN a user exists with role employer.
GIVEN the task is assigned to the user.

WHEN the assigned user updates the task description to "New description".

THEN no error is returned.
and the task has description "New description" in the database.

;===============================================================
; Unassigned employee updates description returns forbidden
;===============================================================
GIVEN a task exists with description "Old description".
GIVEN an employee user is authenticated.
GIVEN the task is not assigned to the employee.

WHEN the employee updates the task description to "New description".

THEN a forbidden error is returned.

;===============================================================
; Update description on non-existent task returns 404 error
;===============================================================
GIVEN an admin user is authenticated.
GIVEN no task exists with ID "nonexistent-123".

WHEN the admin updates task "nonexistent-123" description to "New description".

THEN a not found error is returned.

;===============================================================
; PUT /api/tasks/{id}/due-date - Due date scenarios
;===============================================================

;===============================================================
; Admin sets due date
;===============================================================
GIVEN a task exists without a due date.

WHEN the admin sets the task due date to "2027-03-01T10:00:00Z".

THEN no error is returned.
and the task has due date "2027-03-01T10:00:00Z" in the database.

;===============================================================
; Employer sets due date
;===============================================================
GIVEN an employer user is authenticated.
GIVEN a task exists without a due date.

WHEN the employer sets the task due date to "2027-03-01T10:00:00Z".

THEN no error is returned.
and the task has due date "2027-03-01T10:00:00Z" in the database.

;===============================================================
; Employee sets due date returns forbidden
;===============================================================
GIVEN an employee user is authenticated.
GIVEN a task exists without a due date.

WHEN the employee sets the task due date to "2027-03-01T10:00:00Z".

THEN a forbidden error is returned.

;===============================================================
; PUT /api/tasks/{id}/assign/{userId} - Assignment scenarios
;===============================================================

;===============================================================
; Admin assigns task to employer
;===============================================================
GIVEN a task exists with status todo.
GIVEN a user exists with role employer.

WHEN the admin assigns the task to the employer.

THEN no error is returned.
and the task is assigned to the employer in the database.

;===============================================================
; Employer assigns task
;===============================================================
GIVEN an employer user is authenticated.
GIVEN a task exists with status todo.
GIVEN another user exists with role employer.

WHEN the employer assigns the task to the other employer.

THEN no error is returned.
and the task is assigned to the other employer in the database.

;===============================================================
; Admin assigns task to employee fails
;===============================================================
GIVEN a task exists with status todo.
GIVEN a user exists with role employee.

WHEN the admin assigns the task to the employee.

THEN an error is returned.
and the task is not assigned in the database.

;===============================================================
; Employee assigns task returns forbidden
;===============================================================
GIVEN an employee user is authenticated.
GIVEN a task exists with status todo.
GIVEN a user exists with role employer.

WHEN the employee assigns the task to the employer.

THEN a forbidden error is returned.

;===============================================================
; Assign to non-existent user returns 404 error
;===============================================================
GIVEN an admin user is authenticated.
GIVEN a task exists with status todo.
GIVEN no user exists with ID "nonexistent-123".

WHEN the admin assigns the task to user "nonexistent-123".

THEN a not found error is returned.

;===============================================================
; DELETE /api/tasks/{id}/assign - Unassignment scenarios
;===============================================================

;===============================================================
; Admin unassigns task
;===============================================================
GIVEN a task exists with status todo.
GIVEN the task is assigned to a user.

WHEN the admin unassigns the task.

THEN no error is returned.
and the task is not assigned in the database.

;===============================================================
; Employer unassigns task
;===============================================================
GIVEN an employer user is authenticated.
GIVEN a task exists with status todo.
GIVEN the task is assigned to a user.

WHEN the employer unassigns the task.

THEN no error is returned.
and the task is not assigned in the database.

;===============================================================
; Employee unassigns task returns forbidden
;===============================================================
GIVEN an employee user is authenticated.
GIVEN a task exists with status todo.
GIVEN the task is assigned to a user.

WHEN the employee unassigns the task.

THEN a forbidden error is returned.

;===============================================================
; PUT /api/tasks/{id}/archive - Archive scenarios
;===============================================================

;===============================================================
; Admin archives task
;===============================================================
GIVEN a task exists with status todo.

WHEN the admin archives the task.

THEN no error is returned.
and the task is archived in the database.

;===============================================================
; Employer archives task
;===============================================================
GIVEN an employer user is authenticated.
GIVEN a task exists with status todo.

WHEN the employer archives the task.

THEN no error is returned.
and the task is archived in the database.

;===============================================================
; Employee archives task returns forbidden
;===============================================================
GIVEN an employee user is authenticated.
GIVEN a task exists with status todo.

WHEN the employee archives the task.

THEN a forbidden error is returned.

;===============================================================
; Archive already archived task is idempotent
;===============================================================
GIVEN a task exists with status todo.
GIVEN the task is archived.

WHEN the admin archives the task again.

THEN no error is returned.

;===============================================================
; PUT /api/tasks/{id}/reopen - Reopen scenarios
;===============================================================

;===============================================================
; Admin reopens archived task
;===============================================================
GIVEN a task exists with status todo.
GIVEN the task is archived.

WHEN the admin reopens the task.

THEN no error is returned.
and the task is not archived in the database.

;===============================================================
; Employer reopens archived task
;===============================================================
GIVEN an employer user is authenticated.
GIVEN a task exists with status todo.
GIVEN the task is archived.

WHEN the employer reopens the task.

THEN no error is returned.
and the task is not archived in the database.

;===============================================================
; Employee reopens task returns forbidden
;===============================================================
GIVEN an employee user is authenticated.
GIVEN a task exists with status todo.
GIVEN the task is archived.

WHEN the employee reopens the task.

THEN a forbidden error is returned.

;===============================================================
; DELETE /api/tasks/{id} - Delete scenarios
;===============================================================

;===============================================================
; Admin deletes task
;===============================================================
GIVEN a task exists with status todo.

WHEN the admin deletes the task.

THEN no error is returned.
and the task is soft-deleted in the database.

;===============================================================
; Employer (creator) deletes own task
;===============================================================
GIVEN an employer user is authenticated.
GIVEN a task exists created by the employer.

WHEN the employer deletes the task.

THEN no error is returned.
and the task is soft-deleted in the database.

;===============================================================
; Employer deletes another's task returns forbidden
;===============================================================
GIVEN an employer user is authenticated.
GIVEN a task exists created by another user.

WHEN the employer deletes the task.

THEN a forbidden error is returned.

;===============================================================
; Employee deletes task returns forbidden
;===============================================================
GIVEN an employee user is authenticated.
GIVEN a task exists with status todo.

WHEN the employee deletes the task.

THEN a forbidden error is returned.
```

**Step 2: Commit**

```bash
git add acceptanceTests/task.txt
git commit -m "feat: add task.txt acceptance test scenarios (55 scenarios)"
```

---

## Task 3: Create comment.txt (15 scenarios)

**Files:**
- Create: `acceptanceTests/comment.txt`

**Step 1: Write comment.txt with all 15 scenarios**

Create `acceptanceTests/comment.txt`:

```
;===============================================================
; comment.txt - Comment CRUD and author restrictions
;===============================================================

;===============================================================
; POST /api/tasks/{taskId}/comments - Add comment scenarios
;===============================================================

;===============================================================
; Admin adds comment
;===============================================================
GIVEN a task exists with status todo.
GIVEN an admin user is authenticated.

WHEN the admin adds a comment "This is a comment" to the task.

THEN no error is returned.
and the comment exists on the task.

;===============================================================
; Employer adds comment
;===============================================================
GIVEN a task exists with status todo.
GIVEN an employer user is authenticated.

WHEN the employer adds a comment "This is a comment" to the task.

THEN no error is returned.
and the comment exists on the task.

;===============================================================
; Assigned user adds comment
;===============================================================
GIVEN a task exists with status todo.
GIVEN a user exists with role employer.
GIVEN the task is assigned to the user.

WHEN the assigned user adds a comment "This is a comment" to the task.

THEN no error is returned.
and the comment exists on the task.

;===============================================================
; Unassigned user adds comment returns forbidden
;===============================================================
GIVEN a task exists with status todo.
GIVEN an employee user is authenticated.
GIVEN the task is not assigned to the employee.

WHEN the employee adds a comment "This is a comment" to the task.

THEN a forbidden error is returned.

;===============================================================
; Add comment with empty content returns 400 error
;===============================================================
GIVEN a task exists with status todo.
GIVEN an admin user is authenticated.

WHEN the admin adds a comment "" to the task.

THEN a bad request error is returned.

;===============================================================
; Add comment to non-existent task returns 404 error
;===============================================================
GIVEN an admin user is authenticated.
GIVEN no task exists with ID "nonexistent-123".

WHEN the admin adds a comment "This is a comment" to task "nonexistent-123".

THEN a not found error is returned.

;===============================================================
; PATCH /api/tasks/{taskId}/comments/{id} - Update comment scenarios
;===============================================================

;===============================================================
; Author updates own comment
;===============================================================
GIVEN a task exists with status todo.
GIVEN a comment exists on the task by user "user-123".
GIVEN user "user-123" is authenticated.

WHEN the author updates the comment to "Updated content".

THEN no error is returned.
and the comment has content "Updated content" in the database.

;===============================================================
; Non-author updates comment returns forbidden
;===============================================================
GIVEN a task exists with status todo.
GIVEN a comment exists on the task by user "user-123".
GIVEN another user is authenticated with ID "user-456".

WHEN the non-author updates the comment to "Updated content".

THEN a forbidden error is returned.

;===============================================================
; Admin updates any comment
;===============================================================
GIVEN a task exists with status todo.
GIVEN a comment exists on the task by user "user-123".
GIVEN an admin user is authenticated.

WHEN the admin updates the comment to "Updated by admin".

THEN no error is returned.
and the comment has content "Updated by admin" in the database.

;===============================================================
; Update comment with empty content returns 400 error
;===============================================================
GIVEN a task exists with status todo.
GIVEN a comment exists on the task.
GIVEN the comment author is authenticated.

WHEN the author updates the comment to "".

THEN a bad request error is returned.

;===============================================================
; Update comment on non-existent task returns 404 error
;===============================================================
GIVEN an admin user is authenticated.
GIVEN no task exists with ID "nonexistent-123".
GIVEN a comment ID "comment-123" is provided.

WHEN the admin updates comment "comment-123" on task "nonexistent-123".

THEN a not found error is returned.

;===============================================================
; DELETE /api/tasks/{taskId}/comments/{id} - Delete comment scenarios
;===============================================================

;===============================================================
; Author deletes own comment
;===============================================================
GIVEN a task exists with status todo.
GIVEN a comment exists on the task by user "user-123".
GIVEN user "user-123" is authenticated.

WHEN the author deletes the comment.

THEN no error is returned.
and the comment is deleted in the database.

;===============================================================
; Non-author deletes comment returns forbidden
;===============================================================
GIVEN a task exists with status todo.
GIVEN a comment exists on the task by user "user-123".
GIVEN another user is authenticated with ID "user-456".

WHEN the non-author deletes the comment.

THEN a forbidden error is returned.

;===============================================================
; Admin deletes any comment
;===============================================================
GIVEN a task exists with status todo.
GIVEN a comment exists on the task by user "user-123".
GIVEN an admin user is authenticated.

WHEN the admin deletes the comment.

THEN no error is returned.
and the comment is deleted in the database.

;===============================================================
; Delete non-existent comment returns 404 error
;===============================================================
GIVEN a task exists with status todo.
GIVEN an admin user is authenticated.
GIVEN no comment exists with ID "nonexistent-123".

WHEN the admin deletes comment "nonexistent-123".

THEN a not found error is returned.
```

**Step 2: Commit**

```bash
git add acceptanceTests/comment.txt
git commit -m "feat: add comment.txt acceptance test scenarios (15 scenarios)"
```

---

## Task 4: Create tag.txt (10 scenarios)

**Files:**
- Create: `acceptanceTests/tag.txt`

**Step 1: Write tag.txt with all 10 scenarios**

Create `acceptanceTests/tag.txt`:

```
;===============================================================
; tag.txt - Tag add/remove operations
;===============================================================

;===============================================================
; POST /api/tasks/{taskId}/tags - Add tag scenarios
;===============================================================

;===============================================================
; Admin adds tag
;===============================================================
GIVEN a task exists with status todo.
GIVEN an admin user is authenticated.

WHEN the admin adds a tag "backend" to the task.

THEN no error is returned.
and the tag "backend" exists on the task.

;===============================================================
; Employer adds tag
;===============================================================
GIVEN a task exists with status todo.
GIVEN an employer user is authenticated.

WHEN the employer adds a tag "backend" to the task.

THEN no error is returned.
and the tag "backend" exists on the task.

;===============================================================
; Employee adds tag returns forbidden
;===============================================================
GIVEN a task exists with status todo.
GIVEN an employee user is authenticated.

WHEN the employee adds a tag "backend" to the task.

THEN a forbidden error is returned.

;===============================================================
; Add tag with empty name returns 400 error
;===============================================================
GIVEN a task exists with status todo.
GIVEN an admin user is authenticated.

WHEN the admin adds a tag "" to the task.

THEN a bad request error is returned.

;===============================================================
; Add tag to non-existent task returns 404 error
;===============================================================
GIVEN an admin user is authenticated.
GIVEN no task exists with ID "nonexistent-123".

WHEN the admin adds a tag "backend" to task "nonexistent-123".

THEN a not found error is returned.

;===============================================================
; DELETE /api/tasks/{taskId}/tags/{tagId} - Remove tag scenarios
;===============================================================

;===============================================================
; Admin removes tag
;===============================================================
GIVEN a task exists with status todo.
GIVEN the task has tag "backend".
GIVEN an admin user is authenticated.

WHEN the admin removes the tag from the task.

THEN no error is returned.
and the tag no longer exists on the task.

;===============================================================
; Employer removes tag
;===============================================================
GIVEN a task exists with status todo.
GIVEN the task has tag "backend".
GIVEN an employer user is authenticated.

WHEN the employer removes the tag from the task.

THEN no error is returned.
and the tag no longer exists on the task.

;===============================================================
; Employee removes tag returns forbidden
;===============================================================
GIVEN a task exists with status todo.
GIVEN the task has tag "backend".
GIVEN an employee user is authenticated.

WHEN the employee removes the tag from the task.

THEN a forbidden error is returned.

;===============================================================
; Remove non-existent tag returns 404 error
;===============================================================
GIVEN a task exists with status todo.
GIVEN an admin user is authenticated.
GIVEN no tag exists with ID "nonexistent-123".

WHEN the admin removes tag "nonexistent-123" from the task.

THEN a not found error is returned.

;===============================================================
; Remove tag from non-existent task returns 404 error
;===============================================================
GIVEN an admin user is authenticated.
GIVEN no task exists with ID "nonexistent-123".

WHEN the admin removes a tag from task "nonexistent-123".

THEN a not found error is returned.
```

**Step 2: Commit**

```bash
git add acceptanceTests/tag.txt
git commit -m "feat: add tag.txt acceptance test scenarios (10 scenarios)"
```

---

## Task 5: Create activity.txt (12 scenarios)

**Files:**
- Create: `acceptanceTests/activity.txt`

**Step 1: Write activity.txt with all 12 scenarios**

Create `acceptanceTests/activity.txt`:

```
;===============================================================
; activity.txt - Audit trail verification
;===============================================================

;===============================================================
; GET /api/tasks/{taskId}/activity - Activity log access scenarios
;===============================================================

;===============================================================
; Admin gets activity log
;===============================================================
GIVEN a task exists with status todo.
GIVEN an admin user is authenticated.

WHEN the admin requests the activity log for the task.

THEN no error is returned.
and the response contains the activity log.

;===============================================================
; Employer gets activity log
;===============================================================
GIVEN a task exists with status todo.
GIVEN an employer user is authenticated.

WHEN the employer requests the activity log for the task.

THEN no error is returned.
and the response contains the activity log.

;===============================================================
; Employee gets activity of assigned task
;===============================================================
GIVEN a task exists with status todo.
GIVEN a user exists with role employer.
GIVEN the task is assigned to the user.

WHEN the assigned user requests the activity log for the task.

THEN no error is returned.
and the response contains the activity log.

;===============================================================
; Employee gets activity of unassigned task returns forbidden
;===============================================================
GIVEN a task exists with status todo.
GIVEN an employee user is authenticated.
GIVEN the task is not assigned to the employee.

WHEN the employee requests the activity log for the task.

THEN a forbidden error is returned.

;===============================================================
; Activity entries created for mutations
;===============================================================

;===============================================================
; Status change creates status_changed activity
;===============================================================
GIVEN a task exists with status todo.
GIVEN an admin user is authenticated.

WHEN the admin changes the task status to "in_progress".

THEN no error is returned.
and the task activity log includes a "status_changed" event.

;===============================================================
; Task assignment creates assigned activity
;===============================================================
GIVEN a task exists with status todo.
GIVEN a user exists with role employer.
GIVEN an admin user is authenticated.

WHEN the admin assigns the task to the employer.

THEN no error is returned.
and the task activity log includes an "assigned" event.

;===============================================================
; Task unassignment creates unassigned activity
;===============================================================
GIVEN a task exists with status todo.
GIVEN the task is assigned to a user.
GIVEN an admin user is authenticated.

WHEN the admin unassigns the task.

THEN no error is returned.
and the task activity log includes an "unassigned" event.

;===============================================================
; Title update creates title_updated activity
;===============================================================
GIVEN a task exists with title "Old Title".
GIVEN an admin user is authenticated.

WHEN the admin updates the task title to "New Title".

THEN no error is returned.
and the task activity log includes a "title_updated" event.

;===============================================================
; Priority change creates priority_changed activity
;===============================================================
GIVEN a task exists with priority "medium".
GIVEN an admin user is authenticated.

WHEN the admin sets the task priority to "high".

THEN no error is returned.
and the task activity log includes a "priority_changed" event.

;===============================================================
; Due date set creates due_date_set activity
;===============================================================
GIVEN a task exists without a due date.
GIVEN an admin user is authenticated.

WHEN the admin sets the task due date to "2027-03-01T10:00:00Z".

THEN no error is returned.
and the task activity log includes a "due_date_set" event.

;===============================================================
; Description set creates description_set activity
;===============================================================
GIVEN a task exists without a description.
GIVEN an admin user is authenticated.

WHEN the admin updates the task description to "New description".

THEN no error is returned.
and the task activity log includes a "description_set" event.

;===============================================================
; Comment add/update/delete creates corresponding activity
;===============================================================
GIVEN a task exists with status todo.
GIVEN an admin user is authenticated.

WHEN the admin adds a comment "Test comment" to the task.

THEN no error is returned.
and the task activity log includes a "comment_added" event.
```

**Step 2: Commit**

```bash
git add acceptanceTests/activity.txt
git commit -m "feat: add activity.txt acceptance test scenarios (12 scenarios)"
```

---

## Task 6: Final verification

**Step 1: Verify all files exist**

```bash
ls -la acceptanceTests/
wc -l acceptanceTests/*.txt
```

Expected output: 5 files with line counts matching scenario counts.

**Step 2: Verify scenario count**

```bash
grep -c "^;==="$ acceptanceTests/*.txt
```

This counts the number of test case separators in each file.

**Step 3: Final commit (if any changes)**

```bash
git status
git add -A
git commit -m "feat: add acceptance test scenario files (127 scenarios total)"
```

---

## Summary

| File | Scenarios | Status |
|------|-----------|--------|
| user.txt | 35 | Pending |
| task.txt | 55 | Pending |
| comment.txt | 15 | Pending |
| tag.txt | 10 | Pending |
| activity.txt | 12 | Pending |
| **Total** | **127** | - |
