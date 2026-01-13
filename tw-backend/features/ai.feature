Feature: AI & NPC Controls
  As a game system
  I want NPCs to make intelligent decisions and distribute load
  So that the world is alive and performant

  Scenario: NPC Goal Selection
    Given the desire engine is running
    And I have an NPC "Guard" in "castle_hall" with context "Suspicious of strangers"
    When the system requests a decision for "Guard"
    Then the engine should publish an AI request with prompt containing "Suspicious of strangers"
    When the AI service responds with "SAY Who goes there?"
    Then the engine should publish a spatial action "SAY Who goes there?"

  Scenario: AI Load Scheduling
    Given an AI scheduler with 4 buckets
    When I register 8 entities
    Then they should be distributed evenly across buckets
    And tick 0 should process 2 entities
    And tick 1 should process 2 entities
    And tick 5 should process the same entities as tick 1
