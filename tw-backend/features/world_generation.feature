Feature: World Generation
  As a player
  I want a realistic world generated with tectonic plates and climate systems
  So that I can explore a diverse and plausible environment

  Scenario: Tectonic Plate Simulation
    Given the world generator is initialized with seed "12345"
    And the planet radius is 6371 km
    When I run the tectonic simulation
    Then I should see at least 5 tectonic plates
    And the elevation map should have values between -11000 and 9000 meters

  Scenario: Erosion Process
    Given a world with valid tectonic elevation
    When I run the hydraulic erosion simulation for 1000 cycles
    Then river channels should be formed
    And sediment should be deposited in lower elevations

  Scenario: Climate and Biome Determination
    Given a world with elevation and temperature maps
    When I calculate biomes based on moisture and temperature
    Then I should see "Desert" biomes in high temperature low moisture areas
    And I should see "Tundra" biomes in low temperature areas
