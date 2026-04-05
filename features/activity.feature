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
