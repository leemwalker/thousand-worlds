Feature: Economy System
  As a player
  I want to craft items and trade resources
  So that I can improve my equipment and acquire wealth

  Scenario: Crafting an item
    Given the crafting service is available
    And I have the required ingredients for "Iron Sword"
    When I attempt to craft "Iron Sword"
    Then the item "Iron Sword" should be added to my inventory
    And the ingredients should be removed from my inventory

  Scenario: Crafting with insufficient ingredients
    Given the crafting service is available
    And I do not have the required ingredients for "Golden Shield"
    When I attempt to craft "Golden Shield"
    Then the crafting attempt should fail
    And my inventory should remain unchanged

  Scenario: Executing a trade route
    Given a merchant at "Capital City"
    And a planned trade route to "Mining Outpost"
    When the merchant starts the route
    And the travel time elapses
    Then the merchant should arrive at "Mining Outpost"
    And the trade should be completed successfully
