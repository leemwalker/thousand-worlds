Feature: Game Controls
  As a player
  I want to control the flow of time
  So that I can observe long-term geological processes

  Scenario: Pause Simulation
    Given the simulation is running at "1x" speed
    When I click the "Pause" button
    Then the simulation time should stop increasing
    And the "Pause" button should indicate active state

  Scenario: Fast Forward
    Given the simulation is running
    When I click the "Speed Up" button
    Then the simulation speed should increase to "2x"
    And the ticks per second should double

  Scenario: Reset Simulation
    Given the simulation has been running for "100" years
    When I click the "Reset World" button
    And I confirm the action
    Then the world should regenerate
    And the simulation time should return to "Year 0"
