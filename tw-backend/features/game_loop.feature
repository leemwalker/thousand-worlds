Feature: Game Loop
  As a system
  I want to process game ticks consistently
  So that the simulation advances smoothly

  Scenario: Tick Processing
    Given the game engine is running
    When 60 ticks pass
    Then the simulation time should advance by 1 second
    And all registered subsystems should have updated

  Scenario: Event Handling
    Given a "PLAYER_JOINED" event is queued
    When the event loop processes the queue
    Then the "PlayerManager" should receive the event
    And the active player count should increase by 1
