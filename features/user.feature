Feature: Auth, user CRUD, role management

  Scenario: Login with valid credentials returns JWT token
    Given a user exists with role employer.
    And the user has email "test@example.com" and password "password123".
    When the user logs in with email "test@example.com" and password "password123".
    Then no error is returned.
    And a JWT token is available.

  Scenario: Login with invalid password returns 401 error
    Given a user exists with role employer.
    And the user has email "test@example.com" and password "password123".
    When the user logs in with email "test@example.com" and password "wrongpassword".
    Then an unauthorized error is returned.

  Scenario: Login with non-existent email returns 401 error
    Given no user exists with email "nonexistent@example.com".
    When the user logs in with email "nonexistent@example.com" and password "password123".
    Then an unauthorized error is returned.

  Scenario: Login with missing email returns 400 error
    Given no preconditions.
    When the user logs in with empty email and password "password123".
    Then a bad request error is returned.

  Scenario: Login with missing password returns 400 error
    Given no preconditions.
    When the user logs in with email "test@example.com" and empty password.
    Then a bad request error is returned.

  Scenario: Login with empty credentials returns 400 error
    Given no preconditions.
    When the user logs in with empty email and empty password.
    Then a bad request error is returned.

  Scenario: Admin creates user with role admin
    Given an admin user is authenticated.
    When the admin creates a user with role "admin", name "Admin User", email "admin@example.com", and password "password123".
    Then no error is returned.
    And the user exists in the database with role "admin".

  Scenario: Admin creates user with role employer
    Given an admin user is authenticated.
    When the admin creates a user with role "employer", name "Employer User", email "employer@example.com", and password "password123".
    Then no error is returned.
    And the user exists in the database with role "employer".

  Scenario: Admin creates user with role employee
    Given an admin user is authenticated.
    When the admin creates a user with role "employee", name "Employee User", email "employee@example.com", and password "password123".
    Then no error is returned.
    And the user exists in the database with role "employee".

  Scenario: Create user with duplicate email returns 409 conflict
    Given a user exists with email "existing@example.com".
    When the admin creates a user with email "existing@example.com".
    Then a conflict error is returned.

  Scenario: Create user with invalid email format returns 400 error
    Given an admin user is authenticated.
    When the admin creates a user with email "invalid-email".
    Then a bad request error is returned.

  Scenario: Create user with missing required fields returns 400 error
    Given an admin user is authenticated.
    When the admin creates a user with empty name.
    Then a bad request error is returned.

  Scenario: Create user with empty name returns 400 error
    Given an admin user is authenticated.
    When the admin creates a user with name "".
    Then a bad request error is returned.

  Scenario: Create user with weak password returns 400 error
    Given an admin user is authenticated.
    When the admin creates a user with password "123".
    Then a bad request error is returned.

  Scenario: Admin lists all users
    Given an admin user is authenticated.
    And multiple users exist in the database.
    When the admin requests the user list.
    Then no error is returned.
    And all users are available.

  Scenario: Non-admin lists users returns forbidden
    Given an employer user is authenticated.
    When the employer requests the user list.
    Then a forbidden error is returned.

  Scenario: List users with empty database returns empty list
    Given an admin user is authenticated.
    And no users exist except the admin.
    When the admin requests the user list.
    Then no error is returned.
    And an empty list is available.

  Scenario: Admin gets any user by ID
    Given an admin user is authenticated.
    And a user exists with ID "user-123".
    When the admin requests user "user-123".
    Then no error is returned.
    And the user details are available.

  Scenario: User gets own profile
    Given a user is authenticated with ID "user-123".
    When the user requests their own profile "user-123".
    Then no error is returned.
    And the user details are available.

  Scenario: User gets another user's profile returns forbidden
    Given a user is authenticated with ID "user-123".
    And another user exists with ID "user-456".
    When the user requests profile "user-456".
    Then a forbidden error is returned.

  Scenario: User updates own profile
    Given a user is authenticated with ID "user-123".
    When the user updates their profile with name "New Name".
    Then no error is returned.
    And the user has name "New Name" in the database.

  Scenario: Admin updates another user's profile
    Given an admin user is authenticated.
    And a user exists with ID "user-123".
    When the admin updates user "user-123" with name "New Name".
    Then no error is returned.
    And the user has name "New Name" in the database.

  Scenario: Non-admin updates another user's profile returns forbidden
    Given an employer user is authenticated with ID "user-123".
    And another user exists with ID "user-456".
    When the employer updates user "user-456" with name "New Name".
    Then a forbidden error is returned.

  Scenario: Update with empty name returns 400 error
    Given a user is authenticated with ID "user-123".
    When the user updates their profile with name "".
    Then a bad request error is returned.

  Scenario: Admin deletes another user
    Given an admin user is authenticated.
    And a user exists with ID "user-123".
    When the admin deletes user "user-123".
    Then no error is returned.
    And the user is soft-deleted in the database.

  Scenario: User deletes own account
    Given a user is authenticated with ID "user-123".
    When the user deletes their own account "user-123".
    Then no error is returned.
    And the user is soft-deleted in the database.

  Scenario: Non-admin deletes another user returns forbidden
    Given an employer user is authenticated with ID "user-123".
    And another user exists with ID "user-456".
    When the employer deletes user "user-456".
    Then a forbidden error is returned.

  Scenario: Admin changes user role to employer
    Given an admin user is authenticated.
    And a user exists with ID "user-123" and role "employee".
    When the admin changes user "user-123" role to "employer".
    Then no error is returned.
    And the user has role "employer" in the database.

  Scenario: Admin changes user role to employee
    Given an admin user is authenticated.
    And a user exists with ID "user-123" and role "employer".
    When the admin changes user "user-123" role to "employee".
    Then no error is returned.
    And the user has role "employee" in the database.

  Scenario: Admin changes user role to admin
    Given an admin user is authenticated.
    And a user exists with ID "user-123" and role "employer".
    When the admin changes user "user-123" role to "admin".
    Then no error is returned.
    And the user has role "admin" in the database.

  Scenario: Admin tries to change own role returns forbidden
    Given an admin user is authenticated with ID "admin-123".
    When the admin changes their own role to "employer".
    Then a forbidden error is returned.

  Scenario: Non-admin changes role returns forbidden
    Given an employer user is authenticated.
    And a user exists with ID "user-123".
    When the employer changes user "user-123" role to "admin".
    Then a forbidden error is returned.

  Scenario: Change role of non-existent user returns 404 error
    Given an admin user is authenticated.
    And no user exists with ID "nonexistent-123".
    When the admin changes user "nonexistent-123" role to "employer".
    Then a not found error is returned.

  Scenario: Change role to invalid value returns 400 error
    Given an admin user is authenticated.
    And a user exists with ID "user-123".
    When the admin changes user "user-123" role to "invalid_role".
    Then a bad request error is returned.

  Scenario: Change role to same value succeeds (idempotent)
    Given an admin user is authenticated.
    And a user exists with ID "user-123" and role "employer".
    When the admin changes user "user-123" role to "employer".
    Then no error is returned.
    And the user has role "employer" in the database.
