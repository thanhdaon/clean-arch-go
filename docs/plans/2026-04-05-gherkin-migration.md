# Gherkin Migration Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Migrate all 5 feature specification files from custom `GIVEN/WHEN/THEN` `.txt` format to standard Gherkin `.feature` syntax, and update the acceptance-test-framework documentation accordingly.

**Architecture:** Convert each `features/*.txt` file to its Gherkin equivalent `features/*.feature`. The mapping is mechanical: `;===...` blocks become `Scenario:` headers, `GIVEN` → `Given`, `WHEN` → `When`, `THEN` → `Then`, `and` → `And`, `; comments` → `# comments`. No test code changes required — component tests in `tests/` continue to work as-is since they reference line numbers in `.txt` files. After migration, test names will reference `.feature` files instead.

**Tech Stack:** Gherkin syntax (plain text), Go testing (existing)

---

## Migration Mapping Rules

| Current (.txt) | Gherkin (.feature) |
|---|---|
| `;===============================================================` | *(removed — use blank line between scenarios)* |
| `; Description text` | `# Description text` |
| `GIVEN ...` | `Given ...` |
| `WHEN ...` | `When ...` |
| `THEN ...` | `Then ...` |
| `and ...` | `And ...` |
| `; task.txt - ...` header | `Feature: Task lifecycle and permissions` |
| `; comment lines` | `# comment lines` |

**Key differences:**
- `Scenario:` replaces the `;=== ; Description ;===` block pattern
- `Feature:` is a new top-level keyword (one per file)
- Gherkin uses Title Case keywords: `Given`, `When`, `Then`, `And`
- Trailing periods on step text are preserved as-is
- Blank lines separate scenarios (no `===` separators)

**Files NOT changed:**
- `tests/*_test.go` — test code stays as-is (references updated in a later phase)
- `go.mod` — no new dependencies needed
- `Makefile` — no changes needed

---

## Task 1: Convert `features/tag.txt` → `features/tag.feature`

**Why first:** Smallest file (111 lines, 10 scenarios) — lowest risk, validates the mapping pattern.

**Files:**
- Create: `features/tag.feature`
- Delete (after verification): `features/tag.txt`

**Step 1: Create `features/tag.feature`**

```gherkin
Feature: Tag add/remove operations

  # Admin adds tag
  Scenario: Admin adds tag
    Given a task exists with status todo.
    And an admin user is authenticated.
    When the admin adds a tag "backend" to the task.
    Then no error is returned.
    And the tag "backend" exists on the task.

  # Employer adds tag
  Scenario: Employer adds tag
    Given a task exists with status todo.
    And an employer user is authenticated.
    When the employer adds a tag "backend" to the task.
    Then no error is returned.
    And the tag "backend" exists on the task.

  # Employee adds tag returns forbidden
  Scenario: Employee adds tag returns forbidden
    Given a task exists with status todo.
    And an employee user is authenticated.
    When the employee adds a tag "backend" to the task.
    Then a forbidden error is returned.

  # Add tag with empty name returns 400 error
  Scenario: Add tag with empty name returns 400 error
    Given a task exists with status todo.
    And an admin user is authenticated.
    When the admin adds a tag "" to the task.
    Then a bad request error is returned.

  # Add tag to non-existent task returns 404 error
  Scenario: Add tag to non-existent task returns 404 error
    Given an admin user is authenticated.
    And no task exists with ID "nonexistent-123".
    When the admin adds a tag "backend" to task "nonexistent-123".
    Then a not found error is returned.

  # Admin removes tag
  Scenario: Admin removes tag
    Given a task exists with status todo.
    And the task has tag "backend".
    And an admin user is authenticated.
    When the admin removes the tag from the task.
    Then no error is returned.
    And the tag no longer exists on the task.

  # Employer removes tag
  Scenario: Employer removes tag
    Given a task exists with status todo.
    And the task has tag "backend".
    And an employer user is authenticated.
    When the employer removes the tag from the task.
    Then no error is returned.
    And the tag no longer exists on the task.

  # Employee removes tag returns forbidden
  Scenario: Employee removes tag returns forbidden
    Given a task exists with status todo.
    And the task has tag "backend".
    And an employee user is authenticated.
    When the employee removes the tag from the task.
    Then a forbidden error is returned.

  # Remove non-existent tag returns 404 error
  Scenario: Remove non-existent tag returns 404 error
    Given a task exists with status todo.
    And an admin user is authenticated.
    And no tag exists with ID "nonexistent-123".
    When the admin removes tag "nonexistent-123" from the task.
    Then a not found error is returned.

  # Remove tag from non-existent task returns 404 error
  Scenario: Remove tag from non-existent task returns 404 error
    Given an admin user is authenticated.
    And no task exists with ID "nonexistent-123".
    When the admin removes a tag from task "nonexistent-123".
    Then a not found error is returned.
```

**Step 2: Verify the file was created correctly**

Run: `wc -l features/tag.feature`
Expected: ~50 lines (scenarios are more compact without `===` separators)

**Step 3: Commit**

```bash
git add features/tag.feature
git commit -m "feat: add Gherkin feature file for tag scenarios"
```

---

## Task 2: Convert `features/activity.txt` → `features/activity.feature`

**Files:**
- Create: `features/activity.feature`
- Delete (after verification): `features/activity.txt`

**Step 1: Create `features/activity.feature`**

```gherkin
Feature: Audit trail verification

  # Admin gets activity log
  Scenario: Admin gets activity log
    Given a task exists with status todo.
    And an admin user is authenticated.
    When the admin requests the activity log for the task.
    Then no error is returned.
    And the activity log is available.

  # Employer gets activity log
  Scenario: Employer gets activity log
    Given a task exists with status todo.
    And an employer user is authenticated.
    When the employer requests the activity log for the task.
    Then no error is returned.
    And the activity log is available.

  # Employee gets activity of assigned task
  Scenario: Employee gets activity of assigned task
    Given a task exists with status todo.
    And a user exists with role employer.
    And the task is assigned to the user.
    When the assigned user requests the activity log for the task.
    Then no error is returned.
    And the activity log is available.

  # Employee gets activity of unassigned task returns forbidden
  Scenario: Employee gets activity of unassigned task returns forbidden
    Given a task exists with status todo.
    And an employee user is authenticated.
    And the task is not assigned to the employee.
    When the employee requests the activity log for the task.
    Then a forbidden error is returned.

  # Status change creates status_changed activity
  Scenario: Status change creates status_changed activity
    Given a task exists with status todo.
    And an admin user is authenticated.
    When the admin changes the task status to "in_progress".
    Then no error is returned.
    And the task activity log includes a "status_changed" event.

  # Task assignment creates assigned activity
  Scenario: Task assignment creates assigned activity
    Given a task exists with status todo.
    And a user exists with role employer.
    And an admin user is authenticated.
    When the admin assigns the task to the employer.
    Then no error is returned.
    And the task activity log includes an "assigned" event.

  # Task unassignment creates unassigned activity
  Scenario: Task unassignment creates unassigned activity
    Given a task exists with status todo.
    And the task is assigned to a user.
    And an admin user is authenticated.
    When the admin unassigns the task.
    Then no error is returned.
    And the task activity log includes an "unassigned" event.

  # Title update creates title_updated activity
  Scenario: Title update creates title_updated activity
    Given a task exists with title "Old Title".
    And an admin user is authenticated.
    When the admin updates the task title to "New Title".
    Then no error is returned.
    And the task activity log includes a "title_updated" event.

  # Priority change creates priority_changed activity
  Scenario: Priority change creates priority_changed activity
    Given a task exists with priority "medium".
    And an admin user is authenticated.
    When the admin sets the task priority to "high".
    Then no error is returned.
    And the task activity log includes a "priority_changed" event.

  # Due date set creates due_date_set activity
  Scenario: Due date set creates due_date_set activity
    Given a task exists without a due date.
    And an admin user is authenticated.
    When the admin sets the task due date to "2027-03-01T10:00:00Z".
    Then no error is returned.
    And the task activity log includes a "due_date_set" event.

  # Description set creates description_set activity
  Scenario: Description set creates description_set activity
    Given a task exists without a description.
    And an admin user is authenticated.
    When the admin updates the task description to "New description".
    Then no error is returned.
    And the task activity log includes a "description_set" event.

  # Comment add/update/delete creates corresponding activity
  Scenario: Comment add/update/delete creates corresponding activity
    Given a task exists with status todo.
    And an admin user is authenticated.
    When the admin adds a comment "Test comment" to the task.
    Then no error is returned.
    And the task activity log includes a "comment_added" event.
```

**Step 2: Verify**

Run: `wc -l features/activity.feature`
Expected: ~65 lines

**Step 3: Commit**

```bash
git add features/activity.feature
git commit -m "feat: add Gherkin feature file for activity scenarios"
```

---

## Task 3: Convert `features/comment.txt` → `features/comment.feature`

**Files:**
- Create: `features/comment.feature`
- Delete (after verification): `features/comment.txt`

**Step 1: Create `features/comment.feature`**

```gherkin
Feature: Comment CRUD and author restrictions

  # Admin adds comment
  Scenario: Admin adds comment
    Given a task exists with status todo.
    And an admin user is authenticated.
    When the admin adds a comment "This is a comment" to the task.
    Then no error is returned.
    And the comment exists on the task.

  # Employer adds comment
  Scenario: Employer adds comment
    Given a task exists with status todo.
    And an employer user is authenticated.
    When the employer adds a comment "This is a comment" to the task.
    Then no error is returned.
    And the comment exists on the task.

  # Assigned user adds comment
  Scenario: Assigned user adds comment
    Given a task exists with status todo.
    And a user exists with role employer.
    And the task is assigned to the user.
    When the assigned user adds a comment "This is a comment" to the task.
    Then no error is returned.
    And the comment exists on the task.

  # Unassigned user adds comment returns forbidden
  Scenario: Unassigned user adds comment returns forbidden
    Given a task exists with status todo.
    And an employee user is authenticated.
    And the task is not assigned to the employee.
    When the employee adds a comment "This is a comment" to the task.
    Then a forbidden error is returned.

  # Add comment with empty content returns 400 error
  Scenario: Add comment with empty content returns 400 error
    Given a task exists with status todo.
    And an admin user is authenticated.
    When the admin adds a comment "" to the task.
    Then a bad request error is returned.

  # Add comment to non-existent task returns 404 error
  Scenario: Add comment to non-existent task returns 404 error
    Given an admin user is authenticated.
    And no task exists with ID "nonexistent-123".
    When the admin adds a comment "This is a comment" to task "nonexistent-123".
    Then a not found error is returned.

  # Author updates own comment
  Scenario: Author updates own comment
    Given a task exists with status todo.
    And a comment exists on the task by user "user-123".
    And user "user-123" is authenticated.
    When the author updates the comment to "Updated content".
    Then no error is returned.
    And the comment has content "Updated content" in the database.

  # Non-author updates comment returns forbidden
  Scenario: Non-author updates comment returns forbidden
    Given a task exists with status todo.
    And a comment exists on the task by user "user-123".
    And another user is authenticated with ID "user-456".
    When the non-author updates the comment to "Updated content".
    Then a forbidden error is returned.

  # Admin updates any comment
  Scenario: Admin updates any comment
    Given a task exists with status todo.
    And a comment exists on the task by user "user-123".
    And an admin user is authenticated.
    When the admin updates the comment to "Updated by admin".
    Then no error is returned.
    And the comment has content "Updated by admin" in the database.

  # Update comment with empty content returns 400 error
  Scenario: Update comment with empty content returns 400 error
    Given a task exists with status todo.
    And a comment exists on the task.
    And the comment author is authenticated.
    When the author updates the comment to "".
    Then a bad request error is returned.

  # Update comment on non-existent task returns 404 error
  Scenario: Update comment on non-existent task returns 404 error
    Given an admin user is authenticated.
    And no task exists with ID "nonexistent-123".
    And a comment ID "comment-123" is provided.
    When the admin updates comment "comment-123" on task "nonexistent-123".
    Then a not found error is returned.

  # Author deletes own comment
  Scenario: Author deletes own comment
    Given a task exists with status todo.
    And a comment exists on the task by user "user-123".
    And user "user-123" is authenticated.
    When the author deletes the comment.
    Then no error is returned.
    And the comment is deleted in the database.

  # Non-author deletes comment returns forbidden
  Scenario: Non-author deletes comment returns forbidden
    Given a task exists with status todo.
    And a comment exists on the task by user "user-123".
    And another user is authenticated with ID "user-456".
    When the non-author deletes the comment.
    Then a forbidden error is returned.

  # Admin deletes any comment
  Scenario: Admin deletes any comment
    Given a task exists with status todo.
    And a comment exists on the task by user "user-123".
    And an admin user is authenticated.
    When the admin deletes the comment.
    Then no error is returned.
    And the comment is deleted in the database.

  # Delete non-existent comment returns 404 error
  Scenario: Delete non-existent comment returns 404 error
    Given a task exists with status todo.
    And an admin user is authenticated.
    And no comment exists with ID "nonexistent-123".
    When the admin deletes comment "nonexistent-123".
    Then a not found error is returned.
```

**Step 2: Verify**

Run: `wc -l features/comment.feature`
Expected: ~90 lines

**Step 3: Commit**

```bash
git add features/comment.feature
git commit -m "feat: add Gherkin feature file for comment scenarios"
```

---

## Task 4: Convert `features/user.txt` → `features/user.feature`

**Files:**
- Create: `features/user.feature`
- Delete (after verification): `features/user.txt`

**Step 1: Create `features/user.feature` (Part 1: Login scenarios)**

```gherkin
Feature: Auth, user CRUD, role management

  # Login with valid credentials returns JWT token
  Scenario: Login with valid credentials returns JWT token
    Given a user exists with role employer.
    And the user has email "test@example.com" and password "password123".
    When the user logs in with email "test@example.com" and password "password123".
    Then no error is returned.
    And a JWT token is available.

  # Login with invalid password returns 401 error
  Scenario: Login with invalid password returns 401 error
    Given a user exists with role employer.
    And the user has email "test@example.com" and password "password123".
    When the user logs in with email "test@example.com" and password "wrongpassword".
    Then an unauthorized error is returned.

  # Login with non-existent email returns 401 error
  Scenario: Login with non-existent email returns 401 error
    Given no user exists with email "nonexistent@example.com".
    When the user logs in with email "nonexistent@example.com" and password "password123".
    Then an unauthorized error is returned.

  # Login with missing email returns 400 error
  Scenario: Login with missing email returns 400 error
    Given no preconditions.
    When the user logs in with empty email and password "password123".
    Then a bad request error is returned.

  # Login with missing password returns 400 error
  Scenario: Login with missing password returns 400 error
    Given no preconditions.
    When the user logs in with email "test@example.com" and empty password.
    Then a bad request error is returned.

  # Login with empty credentials returns 400 error
  Scenario: Login with empty credentials returns 400 error
    Given no preconditions.
    When the user logs in with empty email and empty password.
    Then a bad request error is returned.
```

**Step 2: Append user creation scenarios**

Append to `features/user.feature`:

```gherkin

  # Admin creates user with role admin
  Scenario: Admin creates user with role admin
    Given an admin user is authenticated.
    When the admin creates a user with role "admin", name "Admin User", email "admin@example.com", and password "password123".
    Then no error is returned.
    And the user exists in the database with role "admin".

  # Admin creates user with role employer
  Scenario: Admin creates user with role employer
    Given an admin user is authenticated.
    When the admin creates a user with role "employer", name "Employer User", email "employer@example.com", and password "password123".
    Then no error is returned.
    And the user exists in the database with role "employer".

  # Admin creates user with role employee
  Scenario: Admin creates user with role employee
    Given an admin user is authenticated.
    When the admin creates a user with role "employee", name "Employee User", email "employee@example.com", and password "password123".
    Then no error is returned.
    And the user exists in the database with role "employee".

  # Create user with duplicate email returns 409 conflict
  Scenario: Create user with duplicate email returns 409 conflict
    Given a user exists with email "existing@example.com".
    When the admin creates a user with email "existing@example.com".
    Then a conflict error is returned.

  # Create user with invalid email format returns 400 error
  Scenario: Create user with invalid email format returns 400 error
    Given an admin user is authenticated.
    When the admin creates a user with email "invalid-email".
    Then a bad request error is returned.

  # Create user with missing required fields returns 400 error
  Scenario: Create user with missing required fields returns 400 error
    Given an admin user is authenticated.
    When the admin creates a user with empty name.
    Then a bad request error is returned.

  # Create user with empty name returns 400 error
  Scenario: Create user with empty name returns 400 error
    Given an admin user is authenticated.
    When the admin creates a user with name "".
    Then a bad request error is returned.

  # Create user with weak password returns 400 error
  Scenario: Create user with weak password returns 400 error
    Given an admin user is authenticated.
    When the admin creates a user with password "123".
    Then a bad request error is returned.
```

**Step 3: Append user list/get scenarios**

Append to `features/user.feature`:

```gherkin

  # Admin lists all users
  Scenario: Admin lists all users
    Given an admin user is authenticated.
    And multiple users exist in the database.
    When the admin requests the user list.
    Then no error is returned.
    And all users are available.

  # Non-admin lists users returns forbidden
  Scenario: Non-admin lists users returns forbidden
    Given an employer user is authenticated.
    When the employer requests the user list.
    Then a forbidden error is returned.

  # List users with empty database returns empty list
  Scenario: List users with empty database returns empty list
    Given an admin user is authenticated.
    And no users exist except the admin.
    When the admin requests the user list.
    Then no error is returned.
    And an empty list is available.

  # Admin gets any user by ID
  Scenario: Admin gets any user by ID
    Given an admin user is authenticated.
    And a user exists with ID "user-123".
    When the admin requests user "user-123".
    Then no error is returned.
    And the user details are available.

  # User gets own profile
  Scenario: User gets own profile
    Given a user is authenticated with ID "user-123".
    When the user requests their own profile "user-123".
    Then no error is returned.
    And the user details are available.

  # User gets another user's profile returns forbidden
  Scenario: User gets another user's profile returns forbidden
    Given a user is authenticated with ID "user-123".
    And another user exists with ID "user-456".
    When the user requests profile "user-456".
    Then a forbidden error is returned.
```

**Step 4: Append user update/delete scenarios**

Append to `features/user.feature`:

```gherkin

  # User updates own profile
  Scenario: User updates own profile
    Given a user is authenticated with ID "user-123".
    When the user updates their profile with name "New Name".
    Then no error is returned.
    And the user has name "New Name" in the database.

  # Admin updates another user's profile
  Scenario: Admin updates another user's profile
    Given an admin user is authenticated.
    And a user exists with ID "user-123".
    When the admin updates user "user-123" with name "New Name".
    Then no error is returned.
    And the user has name "New Name" in the database.

  # Non-admin updates another user's profile returns forbidden
  Scenario: Non-admin updates another user's profile returns forbidden
    Given an employer user is authenticated with ID "user-123".
    And another user exists with ID "user-456".
    When the employer updates user "user-456" with name "New Name".
    Then a forbidden error is returned.

  # Update with empty name returns 400 error
  Scenario: Update with empty name returns 400 error
    Given a user is authenticated with ID "user-123".
    When the user updates their profile with name "".
    Then a bad request error is returned.

  # Admin deletes another user
  Scenario: Admin deletes another user
    Given an admin user is authenticated.
    And a user exists with ID "user-123".
    When the admin deletes user "user-123".
    Then no error is returned.
    And the user is soft-deleted in the database.

  # User deletes own account
  Scenario: User deletes own account
    Given a user is authenticated with ID "user-123".
    When the user deletes their own account "user-123".
    Then no error is returned.
    And the user is soft-deleted in the database.

  # Non-admin deletes another user returns forbidden
  Scenario: Non-admin deletes another user returns forbidden
    Given an employer user is authenticated with ID "user-123".
    And another user exists with ID "user-456".
    When the employer deletes user "user-456".
    Then a forbidden error is returned.
```

**Step 5: Append role change scenarios**

Append to `features/user.feature`:

```gherkin

  # Admin changes user role to employer
  Scenario: Admin changes user role to employer
    Given an admin user is authenticated.
    And a user exists with ID "user-123" and role "employee".
    When the admin changes user "user-123" role to "employer".
    Then no error is returned.
    And the user has role "employer" in the database.

  # Admin changes user role to employee
  Scenario: Admin changes user role to employee
    Given an admin user is authenticated.
    And a user exists with ID "user-123" and role "employer".
    When the admin changes user "user-123" role to "employee".
    Then no error is returned.
    And the user has role "employee" in the database.

  # Admin changes user role to admin
  Scenario: Admin changes user role to admin
    Given an admin user is authenticated.
    And a user exists with ID "user-123" and role "employer".
    When the admin changes user "user-123" role to "admin".
    Then no error is returned.
    And the user has role "admin" in the database.

  # Admin tries to change own role returns forbidden
  Scenario: Admin tries to change own role returns forbidden
    Given an admin user is authenticated with ID "admin-123".
    When the admin changes their own role to "employer".
    Then a forbidden error is returned.

  # Non-admin changes role returns forbidden
  Scenario: Non-admin changes role returns forbidden
    Given an employer user is authenticated.
    And a user exists with ID "user-123".
    When the employer changes user "user-123" role to "admin".
    Then a forbidden error is returned.

  # Change role of non-existent user returns 404 error
  Scenario: Change role of non-existent user returns 404 error
    Given an admin user is authenticated.
    And no user exists with ID "nonexistent-123".
    When the admin changes user "nonexistent-123" role to "employer".
    Then a not found error is returned.

  # Change role to invalid value returns 400 error
  Scenario: Change role to invalid value returns 400 error
    Given an admin user is authenticated.
    And a user exists with ID "user-123".
    When the admin changes user "user-123" role to "invalid_role".
    Then a bad request error is returned.

  # Change role to same value succeeds (idempotent)
  Scenario: Change role to same value succeeds (idempotent)
    Given an admin user is authenticated.
    And a user exists with ID "user-123" and role "employer".
    When the admin changes user "user-123" role to "employer".
    Then no error is returned.
    And the user has role "employer" in the database.
```

**Step 6: Verify**

Run: `wc -l features/user.feature`
Expected: ~185 lines

**Step 7: Commit**

```bash
git add features/user.feature
git commit -m "feat: add Gherkin feature file for user scenarios"
```

---

## Task 5: Convert `features/task.txt` → `features/task.feature`

**Files:**
- Create: `features/task.feature`
- Delete (after verification): `features/task.txt`

**Note:** This is the largest file (~55 scenarios). Created in multiple parts.

**Step 1: Create `features/task.feature` (Part 1: Create and List scenarios)**

```gherkin
Feature: Task lifecycle and permissions

  # Admin creates task
  Scenario: Admin creates task
    Given an admin user is authenticated.
    When the admin creates a task with title "New Task".
    Then no error is returned.
    And the task exists in the database with status "todo".

  # Employer creates task
  Scenario: Employer creates task
    Given an employer user is authenticated.
    When the employer creates a task with title "New Task".
    Then no error is returned.
    And the task exists in the database with status "todo".

  # Employee creates task returns forbidden
  Scenario: Employee creates task returns forbidden
    Given an employee user is authenticated.
    When the employee creates a task with title "New Task".
    Then a forbidden error is returned.

  # Create task with empty title returns 400 error
  Scenario: Create task with empty title returns 400 error
    Given an admin user is authenticated.
    When the admin creates a task with title "".
    Then a bad request error is returned.

  # Create task with missing title returns 400 error
  Scenario: Create task with missing title returns 400 error
    Given an admin user is authenticated.
    When the admin creates a task without a title.
    Then a bad request error is returned.

  # Create task with very long title succeeds
  Scenario: Create task with very long title succeeds
    Given an admin user is authenticated.
    When the admin creates a task with a 500-character title.
    Then no error is returned.

  # Admin lists all tasks
  Scenario: Admin lists all tasks
    Given an admin user is authenticated.
    And multiple tasks exist in the database.
    When the admin requests the task list.
    Then no error is returned.
    And all tasks are available.

  # Employer lists tasks
  Scenario: Employer lists tasks
    Given an employer user is authenticated.
    And multiple tasks exist in the database.
    When the employer requests the task list.
    Then no error is returned.
    And all tasks are available.

  # Employee lists only assigned tasks
  Scenario: Employee lists only assigned tasks
    Given an employee user is authenticated with ID "emp-123".
    And a task exists with status todo.
    And the task is assigned to the employee.
    And another task exists without assignment.
    When the employee requests the task list.
    Then no error is returned.
    And only the assigned task is available.

  # Employee with no assignments gets empty list
  Scenario: Employee with no assignments gets empty list
    Given an employee user is authenticated.
    And tasks exist but none are assigned to the employee.
    When the employee requests the task list.
    Then no error is returned.
    And an empty list is available.

  # List tasks with pagination
  Scenario: List tasks with pagination
    Given an admin user is authenticated.
    And 50 tasks exist in the database.
    When the admin requests the task list with limit 10 and offset 20.
    Then no error is returned.
    And 10 tasks are available.
    And the tasks are from offset 20.

  # List tasks includes archived
  Scenario: List tasks includes archived
    Given an admin user is authenticated.
    And a task exists with status todo.
    And the task is archived.
    When the admin requests the task list including archived.
    Then no error is returned.
    And the archived task is available.

  # List tasks filtered by status
  Scenario: List tasks filtered by status
    Given an admin user is authenticated.
    And a task exists with status "todo".
    And a task exists with status "in_progress".
    When the admin requests the task list filtered by status "todo".
    Then no error is returned.
    And only tasks with status "todo" are available.

  # List tasks sorted by due date
  Scenario: List tasks sorted by due date
    Given an admin user is authenticated.
    And multiple tasks exist with different due dates.
    When the admin requests the task list sorted by due date ascending.
    Then no error is returned.
    And the tasks are ordered by due date ascending.
```

**Step 2: Append status transition scenarios**

Append to `features/task.feature`:

```gherkin

  # Valid transition: todo to in_progress
  Scenario: Valid transition: todo to in_progress
    Given a task exists with status todo.
    When the admin changes the task status to "in_progress".
    Then no error is returned.
    And the task has status "in_progress" in the database.

  # Valid transition: todo to pending
  Scenario: Valid transition: todo to pending
    Given a task exists with status todo.
    When the admin changes the task status to "pending".
    Then no error is returned.
    And the task has status "pending" in the database.

  # Valid transition: pending to todo
  Scenario: Valid transition: pending to todo
    Given a task exists with status pending.
    When the admin changes the task status to "todo".
    Then no error is returned.
    And the task has status "todo" in the database.

  # Valid transition: pending to in_progress
  Scenario: Valid transition: pending to in_progress
    Given a task exists with status pending.
    When the admin changes the task status to "in_progress".
    Then no error is returned.
    And the task has status "in_progress" in the database.

  # Valid transition: in_progress to completed
  Scenario: Valid transition: in_progress to completed
    Given a task exists with status in_progress.
    When the admin changes the task status to "completed".
    Then no error is returned.
    And the task has status "completed" in the database.

  # Valid transition: in_progress to pending
  Scenario: Valid transition: in_progress to pending
    Given a task exists with status in_progress.
    When the admin changes the task status to "pending".
    Then no error is returned.
    And the task has status "pending" in the database.

  # Invalid transition: todo to completed
  Scenario: Invalid transition: todo to completed
    Given a task exists with status todo.
    When the admin changes the task status to "completed".
    Then an error is returned.
    And the task still has status "todo" in the database.

  # Invalid transition: completed to any
  Scenario: Invalid transition: completed to any
    Given a task exists with status completed.
    When the admin changes the task status to "todo".
    Then an error is returned.
    And the task still has status "completed" in the database.

  # Assigned user changes status
  Scenario: Assigned user changes status
    Given a task exists with status todo.
    And a user exists with role employer.
    And the task is assigned to the user.
    When the assigned user changes the task status to "in_progress".
    Then no error is returned.
    And the task has status "in_progress" in the database.

  # Unassigned employee changes status returns forbidden
  Scenario: Unassigned employee changes status returns forbidden
    Given a task exists with status todo.
    And an employee user is authenticated.
    And the task is not assigned to the employee.
    When the employee changes the task status to "in_progress".
    Then a forbidden error is returned.
    And the task still has status "todo" in the database.

  # Non-admin/non-employer changes unassigned task returns forbidden
  Scenario: Non-admin/non-employer changes unassigned task returns forbidden
    Given a task exists with status todo.
    And an employee user is authenticated.
    And the task is not assigned to the employee.
    When the employee changes the task status to "in_progress".
    Then a forbidden error is returned.

  # Change status of non-existent task returns 404 error
  Scenario: Change status of non-existent task returns 404 error
    Given an admin user is authenticated.
    And no task exists with ID "nonexistent-123".
    When the admin changes task "nonexistent-123" status to "in_progress".
    Then a not found error is returned.
```

**Step 3: Append title update scenarios**

Append to `features/task.feature`:

```gherkin

  # Admin updates title
  Scenario: Admin updates title
    Given a task exists with title "Old Title".
    When the admin updates the task title to "New Title".
    Then no error is returned.
    And the task has title "New Title" in the database.

  # Employer updates title
  Scenario: Employer updates title
    Given an employer user is authenticated.
    And a task exists with title "Old Title".
    When the employer updates the task title to "New Title".
    Then no error is returned.
    And the task has title "New Title" in the database.

  # Assigned user updates title
  Scenario: Assigned user updates title
    Given a task exists with title "Old Title".
    And a user exists with role employer.
    And the task is assigned to the user.
    When the assigned user updates the task title to "New Title".
    Then no error is returned.
    And the task has title "New Title" in the database.

  # Unassigned employee updates title returns forbidden
  Scenario: Unassigned employee updates title returns forbidden
    Given a task exists with title "Old Title".
    And an employee user is authenticated.
    And the task is not assigned to the employee.
    When the employee updates the task title to "New Title".
    Then a forbidden error is returned.
    And the task still has title "Old Title" in the database.

  # Update title to empty returns 400 error
  Scenario: Update title to empty returns 400 error
    Given a task exists with title "Old Title".
    When the admin updates the task title to "".
    Then a bad request error is returned.

  # Update non-existent task returns 404 error
  Scenario: Update non-existent task returns 404 error
    Given an admin user is authenticated.
    And no task exists with ID "nonexistent-123".
    When the admin updates task "nonexistent-123" title to "New Title".
    Then a not found error is returned.
```

**Step 4: Append priority and description scenarios**

Append to `features/task.feature`:

```gherkin

  # Admin sets priority
  Scenario: Admin sets priority
    Given a task exists with priority "medium".
    When the admin sets the task priority to "high".
    Then no error is returned.
    And the task has priority "high" in the database.

  # Employer sets priority
  Scenario: Employer sets priority
    Given an employer user is authenticated.
    And a task exists with priority "medium".
    When the employer sets the task priority to "high".
    Then no error is returned.
    And the task has priority "high" in the database.

  # Employee sets priority returns forbidden
  Scenario: Employee sets priority returns forbidden
    Given an employee user is authenticated.
    And a task exists with priority "medium".
    When the employee sets the task priority to "high".
    Then a forbidden error is returned.

  # Set invalid priority value returns 400 error
  Scenario: Set invalid priority value returns 400 error
    Given a task exists.
    When the admin sets the task priority to "invalid".
    Then a bad request error is returned.

  # Set priority on non-existent task returns 404 error
  Scenario: Set priority on non-existent task returns 404 error
    Given an admin user is authenticated.
    And no task exists with ID "nonexistent-123".
    When the admin sets task "nonexistent-123" priority to "high".
    Then a not found error is returned.

  # Admin updates description
  Scenario: Admin updates description
    Given a task exists with description "Old description".
    When the admin updates the task description to "New description".
    Then no error is returned.
    And the task has description "New description" in the database.

  # Employer updates description
  Scenario: Employer updates description
    Given an employer user is authenticated.
    And a task exists with description "Old description".
    When the employer updates the task description to "New description".
    Then no error is returned.
    And the task has description "New description" in the database.

  # Assigned user updates description
  Scenario: Assigned user updates description
    Given a task exists with description "Old description".
    And a user exists with role employer.
    And the task is assigned to the user.
    When the assigned user updates the task description to "New description".
    Then no error is returned.
    And the task has description "New description" in the database.

  # Unassigned employee updates description returns forbidden
  Scenario: Unassigned employee updates description returns forbidden
    Given a task exists with description "Old description".
    And an employee user is authenticated.
    And the task is not assigned to the employee.
    When the employee updates the task description to "New description".
    Then a forbidden error is returned.

  # Update description on non-existent task returns 404 error
  Scenario: Update description on non-existent task returns 404 error
    Given an admin user is authenticated.
    And no task exists with ID "nonexistent-123".
    When the admin updates task "nonexistent-123" description to "New description".
    Then a not found error is returned.

  # Admin sets due date
  Scenario: Admin sets due date
    Given a task exists without a due date.
    When the admin sets the task due date to "2027-03-01T10:00:00Z".
    Then no error is returned.
    And the task has due date "2027-03-01T10:00:00Z" in the database.

  # Employer sets due date
  Scenario: Employer sets due date
    Given an employer user is authenticated.
    And a task exists without a due date.
    When the employer sets the task due date to "2027-03-01T10:00:00Z".
    Then no error is returned.
    And the task has due date "2027-03-01T10:00:00Z" in the database.

  # Employee sets due date returns forbidden
  Scenario: Employee sets due date returns forbidden
    Given an employee user is authenticated.
    And a task exists without a due date.
    When the employee sets the task due date to "2027-03-01T10:00:00Z".
    Then a forbidden error is returned.
```

**Step 5: Append assignment scenarios**

Append to `features/task.feature`:

```gherkin

  # Admin assigns task to employer
  Scenario: Admin assigns task to employer
    Given a task exists with status todo.
    And a user exists with role employer.
    When the admin assigns the task to the employer.
    Then no error is returned.
    And the task is assigned to the employer in the database.

  # Employer assigns task
  Scenario: Employer assigns task
    Given an employer user is authenticated.
    And a task exists with status todo.
    And another user exists with role employer.
    When the employer assigns the task to the other employer.
    Then no error is returned.
    And the task is assigned to the other employer in the database.

  # Admin assigns task to employee fails
  Scenario: Admin assigns task to employee fails
    Given a task exists with status todo.
    And a user exists with role employee.
    When the admin assigns the task to the employee.
    Then an error is returned.
    And the task is not assigned in the database.

  # Employee assigns task returns forbidden
  Scenario: Employee assigns task returns forbidden
    Given an employee user is authenticated.
    And a task exists with status todo.
    And a user exists with role employer.
    When the employee assigns the task to the employer.
    Then a forbidden error is returned.

  # Assign to non-existent user returns 404 error
  Scenario: Assign to non-existent user returns 404 error
    Given an admin user is authenticated.
    And a task exists with status todo.
    And no user exists with ID "nonexistent-123".
    When the admin assigns the task to user "nonexistent-123".
    Then a not found error is returned.

  # Admin unassigns task
  Scenario: Admin unassigns task
    Given a task exists with status todo.
    And the task is assigned to a user.
    When the admin unassigns the task.
    Then no error is returned.
    And the task is not assigned in the database.

  # Employer unassigns task
  Scenario: Employer unassigns task
    Given an employer user is authenticated.
    And a task exists with status todo.
    And the task is assigned to a user.
    When the employer unassigns the task.
    Then no error is returned.
    And the task is not assigned in the database.

  # Employee unassigns task returns forbidden
  Scenario: Employee unassigns task returns forbidden
    Given an employee user is authenticated.
    And a task exists with status todo.
    And the task is assigned to a user.
    When the employee unassigns the task.
    Then a forbidden error is returned.
```

**Step 6: Append archive/reopen/delete scenarios**

Append to `features/task.feature`:

```gherkin

  # Admin archives task
  Scenario: Admin archives task
    Given a task exists with status todo.
    When the admin archives the task.
    Then no error is returned.
    And the task is archived in the database.

  # Employer archives task
  Scenario: Employer archives task
    Given an employer user is authenticated.
    And a task exists with status todo.
    When the employer archives the task.
    Then no error is returned.
    And the task is archived in the database.

  # Employee archives task returns forbidden
  Scenario: Employee archives task returns forbidden
    Given an employee user is authenticated.
    And a task exists with status todo.
    When the employee archives the task.
    Then a forbidden error is returned.

  # Archive already archived task is idempotent
  Scenario: Archive already archived task is idempotent
    Given a task exists with status todo.
    And the task is archived.
    When the admin archives the task again.
    Then no error is returned.

  # Admin reopens archived task
  Scenario: Admin reopens archived task
    Given a task exists with status todo.
    And the task is archived.
    When the admin reopens the task.
    Then no error is returned.
    And the task is not archived in the database.

  # Employer reopens archived task
  Scenario: Employer reopens archived task
    Given an employer user is authenticated.
    And a task exists with status todo.
    And the task is archived.
    When the employer reopens the task.
    Then no error is returned.
    And the task is not archived in the database.

  # Employee reopens task returns forbidden
  Scenario: Employee reopens task returns forbidden
    Given an employee user is authenticated.
    And a task exists with status todo.
    And the task is archived.
    When the employee reopens the task.
    Then a forbidden error is returned.

  # Admin deletes task
  Scenario: Admin deletes task
    Given a task exists with status todo.
    When the admin deletes the task.
    Then no error is returned.
    And the task is soft-deleted in the database.

  # Employer (creator) deletes own task
  Scenario: Employer (creator) deletes own task
    Given an employer user is authenticated.
    And a task exists created by the employer.
    When the employer deletes the task.
    Then no error is returned.
    And the task is soft-deleted in the database.

  # Employer deletes another's task returns forbidden
  Scenario: Employer deletes another's task returns forbidden
    Given an employer user is authenticated.
    And a task exists created by another user.
    When the employer deletes the task.
    Then a forbidden error is returned.

  # Employee deletes task returns forbidden
  Scenario: Employee deletes task returns forbidden
    Given an employee user is authenticated.
    And a task exists with status todo.
    When the employee deletes the task.
    Then a forbidden error is returned.
```

**Step 7: Verify**

Run: `wc -l features/task.feature`
Expected: ~350 lines

**Step 8: Commit**

```bash
git add features/task.feature
git commit -m "feat: add Gherkin feature file for task scenarios"
```

---

## Task 6: Update `plans/permanent/acceptance-test-framework.md`

**Files:**
- Modify: `plans/permanent/acceptance-test-framework.md`

**Step 1: Update the Overview section**

Replace lines 1-9:

```markdown
# Acceptance Test Framework for clean-arch-go

## Overview

**When to read this file:** When writing or translating acceptance test scenarios from `.feature` files into Go component tests in `tests/`. Also read when debugging a test that doesn't match the intent of its `.feature` source.

---

## Principle

Two-artifact approach — `.feature` files are the source of truth, component tests are the runnable form:

```
.feature (Gherkin spec) → component_test.go (tests/) → go test
```

Gherkin Feature files describe behaviour in domain language using standard `Given/When/Then` syntax. Component tests at the HTTP layer are written to mirror those scenarios exactly. The source filename and first `Given` line number are embedded in each test name for traceability.
```

**Step 2: Update File Format section (lines 86-112)**

Replace with:

```markdown
## File Format (Gherkin)

Gherkin files use standard `.feature` extension with the following structure:

```gherkin
Feature: Task lifecycle and permissions

  # Description of this test case.
  Scenario: Admin creates task
    Given an admin user is authenticated.
    Given a task exists with status todo.
    Given the task is assigned to the user.

    When the admin creates a task with title "New Task".

    Then no error is returned.
    And the task exists in the database with status "todo".
```

### Structure Rules

- `Feature:` is the top-level keyword — one per file, describes the domain
- `Scenario:` marks each independent test case
- `Given`, `When`, `Then` are case-sensitive keywords (Title Case)
- `And` continues the preceding step type (Given/When/Then)
- Lines starting with `#` are comments
- Blank lines separate scenarios for readability
- Each `Scenario:` block becomes one `t.Run` subtest
```

**Step 3: Update test name convention (lines 78-83)**

Replace with:

```markdown
**Test name convention:**
```
"<source-file>:<first-Given-line> - <description>"
```
Example: `"task.feature:8 - Admin creates task"`
```

**Step 4: Update Directory Structure (lines 39-55)**

Replace with:

```markdown
## Directory Structure

```
features/                # Gherkin .feature spec files (never auto-modified)
  task.feature
  user.feature
  comment.feature
  activity.feature
  tag.feature
tests/
  helpers_test.go         # HTTP helpers and DB assertion functions
  setup_test.go           # SetupComponentTest and TestFixtures
  component_test.go       # Main test runner with t.Run blocks
  task_test.go            # Task-related test functions
  user_test.go            # User-related test functions
  comment_test.go         # Comment-related test functions
```
```

**Step 5: Update Rules section (lines 749-758)**

Replace with:

```markdown
## Rules

- Never modify a `.feature` acceptance test without explicit permission.
- The test name **must** embed the `.feature` filename and first `Given` line number — this is the traceability link.
- Every `Given`, `When`, and `Then` step must map to a real helper call — no no-op stubs.
- `require` for setup (Given/When failures stop the test), `assert` for Then assertions.
- Each test function is isolated with its own data via API calls — never share state between tests.
- Every test must be seen to fail before it passes (per `AGENTS.md`).
- If a `.feature` scenario cannot be translated, write the test as a failing stub with a `t.Skip("not yet implemented: ...")` and report it.
- `.feature` files are committed; tests are committed; no intermediate generated files exist.
```

**Step 6: Update Full Example section (lines 762-824)**

Replace with:

```markdown
## Full Example

`.feature` source (`features/task.feature`):

```gherkin
Feature: Task lifecycle and permissions

  # Admin creates task
  Scenario: Admin creates task
    Given an admin user is authenticated.
    When the admin creates a task with title "New Task".
    Then no error is returned.
    And the task exists in the database with status "todo".

  # Employee creates task returns forbidden
  Scenario: Employee creates task returns forbidden
    Given an employee user is authenticated.
    When the employee creates a task with title "New Task".
    Then a forbidden error is returned.
```

Component test (`tests/task_test.go`):

```go
// Scenarios from task.feature — keep in sync with that file.
func testCreateTask(t *testing.T, f *TestFixtures) {
    t.Helper()

    // task.feature:8 - Admin creates task
    t.Run("task.feature:8 - Admin creates task", func(t *testing.T) {
        title := fmt.Sprintf("New-Task-%d", time.Now().UnixNano())
        resp, body := createTask(t, f.AuthToken, map[string]any{"title": title})
        require.Equal(t, http.StatusOK, resp.StatusCode, "unexpected status: %s", string(body))

        taskID := getTaskIDByTitle(t, f.DB, title)
        assert.NotEmpty(t, taskID)
        assertTaskFieldEquals[string](t, f.DB, taskID, "status", "todo")
    })

    // task.feature:17 - Employee creates task returns forbidden
    t.Run("task.feature:17 - Employee creates task returns forbidden", func(t *testing.T) {
        email := fmt.Sprintf("emp-%d@test.com", time.Now().UnixNano())
        resp, body := postUser(t, map[string]any{"role": "employee", "name": "Emp", "email": email, "password": "pass"})
        require.Equal(t, http.StatusOK, resp.StatusCode, "setup failed: %s", string(body))

        loginResp, loginBody := login(t, map[string]any{"email": email, "password": "pass"})
        require.Equal(t, http.StatusOK, loginResp.StatusCode, "login failed: %s", string(loginBody))

        var loginResult struct {
            Token string `json:"token"`
        }
        require.NoError(t, json.Unmarshal(loginBody, &loginResult))
        empToken := loginResult.Token

        resp, body = createTask(t, empToken, map[string]any{"title": "New Task"})
        assert.Equal(t, http.StatusForbidden, resp.StatusCode, "unexpected status: %s", string(body))
    })
}
```
```

**Step 7: Commit**

```bash
git add plans/permanent/acceptance-test-framework.md
git commit -m "docs: update acceptance-test-framework for Gherkin syntax"
```

---

## Task 7: Remove old `.txt` files

**Files:**
- Delete: `features/tag.txt`
- Delete: `features/activity.txt`
- Delete: `features/comment.txt`
- Delete: `features/user.txt`
- Delete: `features/task.txt`

**Step 1: Remove the old files**

```bash
rm features/tag.txt features/activity.txt features/comment.txt features/user.txt features/task.txt
```

**Step 2: Commit**

```bash
git add -A
git commit -m "refactor: remove old .txt feature files, replaced by .feature Gherkin files"
```

---

## Task 8: Verify migration completeness

**Step 1: Verify all .feature files exist**

Run: `ls features/*.feature`
Expected:
```
features/activity.feature
features/comment.feature
features/tag.feature
features/task.feature
features/user.feature
```

**Step 2: Verify no .txt files remain**

Run: `ls features/*.txt`
Expected: No such file or empty output

**Step 3: Verify tests still pass**

Run: `make test`
Expected: All tests pass

**Step 4: Final commit (if needed)**

If any changes from verification fixes:

```bash
git add -A
git commit -m "fix: resolve any issues from Gherkin migration verification"
```

---

## Summary

| Task | Description | Files Changed |
|------|-------------|---------------|
| 1 | Convert tag.txt → tag.feature | Create `features/tag.feature` |
| 2 | Convert activity.txt → activity.feature | Create `features/activity.feature` |
| 3 | Convert comment.txt → comment.feature | Create `features/comment.feature` |
| 4 | Convert user.txt → user.feature | Create `features/user.feature` |
| 5 | Convert task.txt → task.feature | Create `features/task.feature` |
| 6 | Update documentation | Modify `plans/permanent/acceptance-test-framework.md` |
| 7 | Remove old files | Delete 5 `.txt` files |
| 8 | Verify completeness | Run tests |

**Total scenarios migrated:** ~119 (task: ~55, user: ~28, comment: ~14, tag: ~10, activity: ~12)

**No code changes required** — the component tests in `tests/` continue to work as-is since they only need the line number references to be valid.
