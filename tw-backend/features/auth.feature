Feature: Authentication Service
  As a user
  I want to register and login
  So that I can access the game

  Scenario: User Registration Success
    Given the auth service is initialized
    When I register with email "test@example.com", username "PlayerOne", and password "securePass123"
    Then the user should be created successfully
    And I should be able to login with "test@example.com" and "securePass123"

  Scenario: Duplicate Registration
    Given the auth service is initialized
    And a user exists with email "existing@example.com"
    When I register with email "existing@example.com", username "DuplicateUser", and password "anyPass"
    Then the registration should fail with "user already exists" error

  Scenario: Login with Invalid Credentials
    Given the auth service is initialized
    And a user exists with email "user@example.com" and password "correctPass"
    When I attempt to login with "user@example.com" and "wrongPass"
    Then the login should fail
    And no token should be returned

  Scenario: Token Validation
    Given the auth service is initialized
    And I have a valid token for user "tokenUser"
    When I validate the token
    Then the token should be valid
    And the claims should contain the correct user ID
