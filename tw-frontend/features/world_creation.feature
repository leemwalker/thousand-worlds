Feature: World Creation
  As a user
  I want to create a new unique world
  So that I can start a new simulation

  Scenario: Create New World with Defaults
    Given I am on the main menu
    When I click "New Game"
    Then I should see the "World Creation" modal
    And the "Seed" input should be pre-filled
    And the "create" button should be enabled

  Scenario: Custom Seed Entry
    Given I am on the "World Creation" modal
    When I enter "MyCustomWorld" into the "Seed" field
    And I click "Generate"
    Then the simulation should start with seed "MyCustomWorld"
    And I should see the "Loading" indicator

  Scenario: Invalid Configuration
    Given I am on the "World Creation" modal
    When I set "Planet Size" to an invalid value "-100"
    Then I should see an error message "Size must be positive"
    And the "Generate" button should be disabled
