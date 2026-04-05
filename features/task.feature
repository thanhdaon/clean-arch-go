Feature: Task lifecycle and permissions

  Scenario: Admin creates task
    Given an admin user is authenticated.
    When the admin creates a task with title "New Task".
    Then no error is returned.
    And the task exists in the database with status "todo".

  Scenario: Employer creates task
    Given an employer user is authenticated.
    When the employer creates a task with title "New Task".
    Then no error is returned.
    And the task exists in the database with status "todo".

  Scenario: Employee creates task returns forbidden
    Given an employee user is authenticated.
    When the employee creates a task with title "New Task".
    Then a forbidden error is returned.

  Scenario: Create task with empty title returns 400 error
    Given an admin user is authenticated.
    When the admin creates a task with title "".
    Then a bad request error is returned.

  Scenario: Create task with missing title returns 400 error
    Given an admin user is authenticated.
    When the admin creates a task without a title.
    Then a bad request error is returned.

  Scenario: Create task with very long title succeeds
    Given an admin user is authenticated.
    When the admin creates a task with a 500-character title.
    Then no error is returned.

  Scenario: Admin lists all tasks
    Given an admin user is authenticated.
    And multiple tasks exist in the database.
    When the admin requests the task list.
    Then no error is returned.
    And all tasks are available.

  Scenario: Employer lists tasks
    Given an employer user is authenticated.
    And multiple tasks exist in the database.
    When the employer requests the task list.
    Then no error is returned.
    And all tasks are available.

  Scenario: Employee lists only assigned tasks
    Given an employee user is authenticated with ID "emp-123".
    And a task exists with status todo.
    And the task is assigned to the employee.
    And another task exists without assignment.
    When the employee requests the task list.
    Then no error is returned.
    And only the assigned task is available.

  Scenario: Employee with no assignments gets empty list
    Given an employee user is authenticated.
    And tasks exist but none are assigned to the employee.
    When the employee requests the task list.
    Then no error is returned.
    And an empty list is available.

  Scenario: List tasks with pagination
    Given an admin user is authenticated.
    And 50 tasks exist in the database.
    When the admin requests the task list with limit 10 and offset 20.
    Then no error is returned.
    And 10 tasks are available.
    And the tasks are from offset 20.

  Scenario: List tasks includes archived
    Given an admin user is authenticated.
    And a task exists with status todo.
    And the task is archived.
    When the admin requests the task list including archived.
    Then no error is returned.
    And the archived task is available.

  Scenario: List tasks filtered by status
    Given an admin user is authenticated.
    And a task exists with status "todo".
    And a task exists with status "in_progress".
    When the admin requests the task list filtered by status "todo".
    Then no error is returned.
    And only tasks with status "todo" are available.

  Scenario: List tasks sorted by due date
    Given an admin user is authenticated.
    And multiple tasks exist with different due dates.
    When the admin requests the task list sorted by due date ascending.
    Then no error is returned.
    And the tasks are ordered by due date ascending.

  Scenario: Valid transition: todo to in_progress
    Given a task exists with status todo.
    When the admin changes the task status to "in_progress".
    Then no error is returned.
    And the task has status "in_progress" in the database.

  Scenario: Valid transition: todo to pending
    Given a task exists with status todo.
    When the admin changes the task status to "pending".
    Then no error is returned.
    And the task has status "pending" in the database.

  Scenario: Valid transition: pending to todo
    Given a task exists with status pending.
    When the admin changes the task status to "todo".
    Then no error is returned.
    And the task has status "todo" in the database.

  Scenario: Valid transition: pending to in_progress
    Given a task exists with status pending.
    When the admin changes the task status to "in_progress".
    Then no error is returned.
    And the task has status "in_progress" in the database.

  Scenario: Valid transition: in_progress to completed
    Given a task exists with status in_progress.
    When the admin changes the task status to "completed".
    Then no error is returned.
    And the task has status "completed" in the database.

  Scenario: Valid transition: in_progress to pending
    Given a task exists with status in_progress.
    When the admin changes the task status to "pending".
    Then no error is returned.
    And the task has status "pending" in the database.

  Scenario: Invalid transition: todo to completed
    Given a task exists with status todo.
    When the admin changes the task status to "completed".
    Then an error is returned.
    And the task still has status "todo" in the database.

  Scenario: Invalid transition: completed to any
    Given a task exists with status completed.
    When the admin changes the task status to "todo".
    Then an error is returned.
    And the task still has status "completed" in the database.

  Scenario: Assigned user changes status
    Given a task exists with status todo.
    And a user exists with role employer.
    And the task is assigned to the user.
    When the assigned user changes the task status to "in_progress".
    Then no error is returned.
    And the task has status "in_progress" in the database.

  Scenario: Unassigned employee changes status returns forbidden
    Given a task exists with status todo.
    And an employee user is authenticated.
    And the task is not assigned to the employee.
    When the employee changes the task status to "in_progress".
    Then a forbidden error is returned.
    And the task still has status "todo" in the database.

  Scenario: Non-admin/non-employer changes unassigned task returns forbidden
    Given a task exists with status todo.
    And an employee user is authenticated.
    And the task is not assigned to the employee.
    When the employee changes the task status to "in_progress".
    Then a forbidden error is returned.

  Scenario: Change status of non-existent task returns 404 error
    Given an admin user is authenticated.
    And no task exists with ID "nonexistent-123".
    When the admin changes task "nonexistent-123" status to "in_progress".
    Then a not found error is returned.

  Scenario: Admin updates title
    Given a task exists with title "Old Title".
    When the admin updates the task title to "New Title".
    Then no error is returned.
    And the task has title "New Title" in the database.

  Scenario: Employer updates title
    Given an employer user is authenticated.
    And a task exists with title "Old Title".
    When the employer updates the task title to "New Title".
    Then no error is returned.
    And the task has title "New Title" in the database.

  Scenario: Assigned user updates title
    Given a task exists with title "Old Title".
    And a user exists with role employer.
    And the task is assigned to the user.
    When the assigned user updates the task title to "New Title".
    Then no error is returned.
    And the task has title "New Title" in the database.

  Scenario: Unassigned employee updates title returns forbidden
    Given a task exists with title "Old Title".
    And an employee user is authenticated.
    And the task is not assigned to the employee.
    When the employee updates the task title to "New Title".
    Then a forbidden error is returned.
    And the task still has title "Old Title" in the database.

  Scenario: Update title to empty returns 400 error
    Given a task exists with title "Old Title".
    When the admin updates the task title to "".
    Then a bad request error is returned.

  Scenario: Update non-existent task returns 404 error
    Given an admin user is authenticated.
    And no task exists with ID "nonexistent-123".
    When the admin updates task "nonexistent-123" title to "New Title".
    Then a not found error is returned.

  Scenario: Admin sets priority
    Given a task exists with priority "medium".
    When the admin sets the task priority to "high".
    Then no error is returned.
    And the task has priority "high" in the database.

  Scenario: Employer sets priority
    Given an employer user is authenticated.
    And a task exists with priority "medium".
    When the employer sets the task priority to "high".
    Then no error is returned.
    And the task has priority "high" in the database.

  Scenario: Employee sets priority returns forbidden
    Given an employee user is authenticated.
    And a task exists with priority "medium".
    When the employee sets the task priority to "high".
    Then a forbidden error is returned.

  Scenario: Set invalid priority value returns 400 error
    Given a task exists.
    When the admin sets the task priority to "invalid".
    Then a bad request error is returned.

  Scenario: Set priority on non-existent task returns 404 error
    Given an admin user is authenticated.
    And no task exists with ID "nonexistent-123".
    When the admin sets task "nonexistent-123" priority to "high".
    Then a not found error is returned.

  Scenario: Admin updates description
    Given a task exists with description "Old description".
    When the admin updates the task description to "New description".
    Then no error is returned.
    And the task has description "New description" in the database.

  Scenario: Employer updates description
    Given an employer user is authenticated.
    And a task exists with description "Old description".
    When the employer updates the task description to "New description".
    Then no error is returned.
    And the task has description "New description" in the database.

  Scenario: Assigned user updates description
    Given a task exists with description "Old description".
    And a user exists with role employer.
    And the task is assigned to the user.
    When the assigned user updates the task description to "New description".
    Then no error is returned.
    And the task has description "New description" in the database.

  Scenario: Unassigned employee updates description returns forbidden
    Given a task exists with description "Old description".
    And an employee user is authenticated.
    And the task is not assigned to the employee.
    When the employee updates the task description to "New description".
    Then a forbidden error is returned.

  Scenario: Update description on non-existent task returns 404 error
    Given an admin user is authenticated.
    And no task exists with ID "nonexistent-123".
    When the admin updates task "nonexistent-123" description to "New description".
    Then a not found error is returned.

  Scenario: Admin sets due date
    Given a task exists without a due date.
    When the admin sets the task due date to "2027-03-01T10:00:00Z".
    Then no error is returned.
    And the task has due date "2027-03-01T10:00:00Z" in the database.

  Scenario: Employer sets due date
    Given an employer user is authenticated.
    And a task exists without a due date.
    When the employer sets the task due date to "2027-03-01T10:00:00Z".
    Then no error is returned.
    And the task has due date "2027-03-01T10:00:00Z" in the database.

  Scenario: Employee sets due date returns forbidden
    Given an employee user is authenticated.
    And a task exists without a due date.
    When the employee sets the task due date to "2027-03-01T10:00:00Z".
    Then a forbidden error is returned.

  Scenario: Admin assigns task to employer
    Given a task exists with status todo.
    And a user exists with role employer.
    When the admin assigns the task to the employer.
    Then no error is returned.
    And the task is assigned to the employer in the database.

  Scenario: Employer assigns task
    Given an employer user is authenticated.
    And a task exists with status todo.
    And another user exists with role employer.
    When the employer assigns the task to the other employer.
    Then no error is returned.
    And the task is assigned to the other employer in the database.

  Scenario: Admin assigns task to employee fails
    Given a task exists with status todo.
    And a user exists with role employee.
    When the admin assigns the task to the employee.
    Then an error is returned.
    And the task is not assigned in the database.

  Scenario: Employee assigns task returns forbidden
    Given an employee user is authenticated.
    And a task exists with status todo.
    And a user exists with role employer.
    When the employee assigns the task to the employer.
    Then a forbidden error is returned.

  Scenario: Assign to non-existent user returns 404 error
    Given an admin user is authenticated.
    And a task exists with status todo.
    And no user exists with ID "nonexistent-123".
    When the admin assigns the task to user "nonexistent-123".
    Then a not found error is returned.

  Scenario: Admin unassigns task
    Given a task exists with status todo.
    And the task is assigned to a user.
    When the admin unassigns the task.
    Then no error is returned.
    And the task is not assigned in the database.

  Scenario: Employer unassigns task
    Given an employer user is authenticated.
    And a task exists with status todo.
    And the task is assigned to a user.
    When the employer unassigns the task.
    Then no error is returned.
    And the task is not assigned in the database.

  Scenario: Employee unassigns task returns forbidden
    Given an employee user is authenticated.
    And a task exists with status todo.
    And the task is assigned to a user.
    When the employee unassigns the task.
    Then a forbidden error is returned.

  Scenario: Admin archives task
    Given a task exists with status todo.
    When the admin archives the task.
    Then no error is returned.
    And the task is archived in the database.

  Scenario: Employer archives task
    Given an employer user is authenticated.
    And a task exists with status todo.
    When the employer archives the task.
    Then no error is returned.
    And the task is archived in the database.

  Scenario: Employee archives task returns forbidden
    Given an employee user is authenticated.
    And a task exists with status todo.
    When the employee archives the task.
    Then a forbidden error is returned.

  Scenario: Archive already archived task is idempotent
    Given a task exists with status todo.
    And the task is archived.
    When the admin archives the task again.
    Then no error is returned.

  Scenario: Admin reopens archived task
    Given a task exists with status todo.
    And the task is archived.
    When the admin reopens the task.
    Then no error is returned.
    And the task is not archived in the database.

  Scenario: Employer reopens archived task
    Given an employer user is authenticated.
    And a task exists with status todo.
    And the task is archived.
    When the employer reopens the task.
    Then no error is returned.
    And the task is not archived in the database.

  Scenario: Employee reopens task returns forbidden
    Given an employee user is authenticated.
    And a task exists with status todo.
    And the task is archived.
    When the employee reopens the task.
    Then a forbidden error is returned.

  Scenario: Admin deletes task
    Given a task exists with status todo.
    When the admin deletes the task.
    Then no error is returned.
    And the task is soft-deleted in the database.

  Scenario: Employer (creator) deletes own task
    Given an employer user is authenticated.
    And a task exists created by the employer.
    When the employer deletes the task.
    Then no error is returned.
    And the task is soft-deleted in the database.

  Scenario: Employer deletes another's task returns forbidden
    Given an employer user is authenticated.
    And a task exists created by another user.
    When the employer deletes the task.
    Then a forbidden error is returned.

  Scenario: Employee deletes task returns forbidden
    Given an employee user is authenticated.
    And a task exists with status todo.
    When the employee deletes the task.
    Then a forbidden error is returned.
