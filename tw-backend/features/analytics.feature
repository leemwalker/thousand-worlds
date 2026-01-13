Feature: Analytics Service
  In order to track the health and status of the simulation
  As a system administrator
  I need to collect and query metrics efficiently and reliably

  Scenario: Recording simulation metrics
    Given the analytics service is running
    When I record a global stats snapshot
    Then the snapshot should be persisted successfully
    And the circuit breaker state should be "closed"

  Scenario: Handling database failures with Circuit Breaker
    Given the analytics service is running
    But the database connection is lost
    When I record a global stats snapshot
    Then the operation should fail gracefully
    When I record 5 more global stats snapshots
    Then the circuit breaker state should be "open"
    And subsequent requests should fail fast
