---
description: TDD workflow - write tests first, then implement
---
Test-Driven Development workflow for new features.

## Phase 1: Define Acceptance Criteria
1. Create a test file: `<package>/<feature>_test.go`
2. Write test function signatures with `t.Skip("TODO: implement")`
3. Define the expected behavior in test names:
   - `TestFeature_WhenCondition_ShouldBehavior`

## Phase 2: Write Failing Tests
1. Implement test bodies with assertions
2. Run tests - they should FAIL (red)
// turbo
3. `cd tw-backend && go test -v ./internal/<package>/...`

## Phase 3: Implement
1. Write minimal code to make tests pass
2. Run tests - they should PASS (green)
// turbo
3. `cd tw-backend && go test -v ./internal/<package>/...`

## Phase 4: Refactor
1. Clean up implementation
2. Ensure tests still pass
3. Check coverage: `go test -cover ./internal/<package>/...`

## Tips
- One behavior per test
- Use table-driven tests for similar cases
- Mock external dependencies (DB, HTTP, etc.)
