Feature: Comment CRUD and author restrictions

  Scenario: Admin adds comment
    Given a task exists with status todo.
    And an admin user is authenticated.
    When the admin adds a comment "This is a comment" to the task.
    Then no error is returned.
    And the comment exists on the task.

  Scenario: Employer adds comment
    Given a task exists with status todo.
    And an employer user is authenticated.
    When the employer adds a comment "This is a comment" to the task.
    Then no error is returned.
    And the comment exists on the task.

  Scenario: Assigned user adds comment
    Given a task exists with status todo.
    And a user exists with role employer.
    And the task is assigned to the user.
    When the assigned user adds a comment "This is a comment" to the task.
    Then no error is returned.
    And the comment exists on the task.

  Scenario: Unassigned user adds comment returns forbidden
    Given a task exists with status todo.
    And an employee user is authenticated.
    And the task is not assigned to the employee.
    When the employee adds a comment "This is a comment" to the task.
    Then a forbidden error is returned.

  Scenario: Add comment with empty content returns 400 error
    Given a task exists with status todo.
    And an admin user is authenticated.
    When the admin adds a comment "" to the task.
    Then a bad request error is returned.

  Scenario: Add comment to non-existent task returns 404 error
    Given an admin user is authenticated.
    And no task exists with ID "nonexistent-123".
    When the admin adds a comment "This is a comment" to task "nonexistent-123".
    Then a not found error is returned.

  Scenario: Author updates own comment
    Given a task exists with status todo.
    And a comment exists on the task by user "user-123".
    And user "user-123" is authenticated.
    When the author updates the comment to "Updated content".
    Then no error is returned.
    And the comment has content "Updated content" in the database.

  Scenario: Non-author updates comment returns forbidden
    Given a task exists with status todo.
    And a comment exists on the task by user "user-123".
    And another user is authenticated with ID "user-456".
    When the non-author updates the comment to "Updated content".
    Then a forbidden error is returned.

  Scenario: Admin updates any comment
    Given a task exists with status todo.
    And a comment exists on the task by user "user-123".
    And an admin user is authenticated.
    When the admin updates the comment to "Updated by admin".
    Then no error is returned.
    And the comment has content "Updated by admin" in the database.

  Scenario: Update comment with empty content returns 400 error
    Given a task exists with status todo.
    And a comment exists on the task.
    And the comment author is authenticated.
    When the author updates the comment to "".
    Then a bad request error is returned.

  Scenario: Update comment on non-existent task returns 404 error
    Given an admin user is authenticated.
    And no task exists with ID "nonexistent-123".
    And a comment ID "comment-123" is provided.
    When the admin updates comment "comment-123" on task "nonexistent-123".
    Then a not found error is returned.

  Scenario: Author deletes own comment
    Given a task exists with status todo.
    And a comment exists on the task by user "user-123".
    And user "user-123" is authenticated.
    When the author deletes the comment.
    Then no error is returned.
    And the comment is deleted in the database.

  Scenario: Non-author deletes comment returns forbidden
    Given a task exists with status todo.
    And a comment exists on the task by user "user-123".
    And another user is authenticated with ID "user-456".
    When the non-author deletes the comment.
    Then a forbidden error is returned.

  Scenario: Admin deletes any comment
    Given a task exists with status todo.
    And a comment exists on the task by user "user-123".
    And an admin user is authenticated.
    When the admin deletes the comment.
    Then no error is returned.
    And the comment is deleted in the database.

  Scenario: Delete non-existent comment returns 404 error
    Given a task exists with status todo.
    And an admin user is authenticated.
    And no comment exists with ID "nonexistent-123".
    When the admin deletes comment "nonexistent-123".
    Then a not found error is returned.
