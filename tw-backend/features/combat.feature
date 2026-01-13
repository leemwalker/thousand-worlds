Feature: Combat System
  As a player
  I want to engage in combat
  So that I can defeat enemies

  Scenario: Joining Combat
    Given a character with agility 10
    When the character joins combat
    Then the character should be listed as a combatant
    And the combatant's HP should generally match the character's MaxHP

  Scenario: Executing an Attack
    Given I have two combatants "Alice" and "Bob"
    And "Alice" queues an attack against "Bob"
    When the combat simulation ticks for 2 seconds
    Then "Alice" should execute the attack
    And a combat event should be generated

  Scenario: Attack Timing based on Agility
    Given a fast character "Fast" with agility 100
    And a slow character "Slow" with agility 0
    When "Fast" queues an attack against "Slow"
    And "Slow" queues an attack against "Fast"
    And the combat simulation ticks for 2 seconds
    Then "Fast" should attack before "Slow"
