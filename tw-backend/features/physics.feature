Feature: Orbital Physics
  As a developer
  I want accurate orbital mechanics
  So that the simulation reflects realistic celestial motion

  Scenario: Planet Orbit
    Given a star with mass 1.989e30 kg
    And a planet at distance 149.6e6 km
    When I simulate one full orbit
    Then the orbital period should be approximately 365 days
    And the planet should return to its initial position

  Scenario: Gravity Calculation
    Given an object of mass 100 kg on the planet surface
    When I calculate the gravitational force
    Then the force should be approximately 980 Newtons
