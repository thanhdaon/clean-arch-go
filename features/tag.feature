Feature: Tag add/remove operations

  Scenario: Admin adds tag
    Given a task exists with status todo.
    And an admin user is authenticated.
    When the admin adds a tag "backend" to the task.
    Then no error is returned.
    And the tag "backend" exists on the task.

  Scenario: Employer adds tag
    Given a task exists with status todo.
    And an employer user is authenticated.
    When the employer adds a tag "backend" to the task.
    Then no error is returned.
    And the tag "backend" exists on the task.

  Scenario: Employee adds tag returns forbidden
    Given a task exists with status todo.
    And an employee user is authenticated.
    When the employee adds a tag "backend" to the task.
    Then a forbidden error is returned.

  Scenario: Add tag with empty name returns 400 error
    Given a task exists with status todo.
    And an admin user is authenticated.
    When the admin adds a tag "" to the task.
    Then a bad request error is returned.

  Scenario: Add tag to non-existent task returns 404 error
    Given an admin user is authenticated.
    And no task exists with ID "nonexistent-123".
    When the admin adds a tag "backend" to task "nonexistent-123".
    Then a not found error is returned.

  Scenario: Admin removes tag
    Given a task exists with status todo.
    And the task has tag "backend".
    And an admin user is authenticated.
    When the admin removes the tag from the task.
    Then no error is returned.
    And the tag no longer exists on the task.

  Scenario: Employer removes tag
    Given a task exists with status todo.
    And the task has tag "backend".
    And an employer user is authenticated.
    When the employer removes the tag from the task.
    Then no error is returned.
    And the tag no longer exists on the task.

  Scenario: Employee removes tag returns forbidden
    Given a task exists with status todo.
    And the task has tag "backend".
    And an employee user is authenticated.
    When the employee removes the tag from the task.
    Then a forbidden error is returned.

  Scenario: Remove non-existent tag returns 404 error
    Given a task exists with status todo.
    And an admin user is authenticated.
    And no tag exists with ID "nonexistent-123".
    When the admin removes tag "nonexistent-123" from the task.
    Then a not found error is returned.

  Scenario: Remove tag from non-existent task returns 404 error
    Given an admin user is authenticated.
    And no task exists with ID "nonexistent-123".
    When the admin removes a tag from task "nonexistent-123".
    Then a not found error is returned.
