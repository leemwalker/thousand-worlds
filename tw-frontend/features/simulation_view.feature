Feature: Simulation View
  As a player
  I want to inspect the generated world
  So that I can understand its geography and climate

  Scenario: View Elevation Map
    Given the simulation is running
    When I select the "Elevation" view layer
    Then the map should display a heightmap gradient
    And high elevations should be white/grey
    And oceans should be blue

  Scenario: Inspect Province
    Given I am viewing the world map
    When I click on a land province
    Then I should see a tooltip with province details
    And the details should include "Biome", "Elevation", and "Temperature"

  Scenario: Zoom and Pan
    Given I am viewing the world map
    When I use the mouse wheel to scroll up
    Then the map camera should zoom in
    And details should become clearer
