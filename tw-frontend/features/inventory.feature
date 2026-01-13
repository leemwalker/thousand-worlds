Feature: Inventory Management
  As a player
  I want to view and manage my items
  So that I can use resources and equipment

  Scenario: Viewing Inventory
    Given the player has "5" items in their inventory
    When the player opens the inventory panel
    Then the inventory list should show "5" items
    And the first item should be "Iron Ore"

  Scenario: Filtering Inventory
    Given the player has the following items:
      | Name      | Type     |
      | Iron Ore  | Resource |
      | Health Kit| Consumable|
      | Blaster   | Weapon   |
    When the player filters by "Resource"
    Then the inventory list should show only "Iron Ore"

  Scenario: Using an Item
    Given the player has a "Health Kit"
    When the player clicks "Use" on "Health Kit"
    Then the "Health Kit" count should decrease by 1
    And the player health should increase
