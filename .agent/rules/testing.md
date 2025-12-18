---
trigger: always_on
---

## Testing Standards (Strict Enforcement)

### Unit Testing (Minimum 80% Coverage)
* **Go:** Use the standard `testing` package or `testify`. Table-driven tests are preferred.
* **Frontend:** Use Vitest or Jest.
* **Rule:** Every utility function, logic block, and component method must have unit tests covering success, failure, and edge cases.

### Integration Testing (100% Coverage)
* **Scope:** All integration points (DB connections, API endpoints, external services) must be tested.
* **Method:** Use Docker containers (e.g., `testcontainers-go`) to spin up dependencies for true integration testing.

### User Flow / E2E Testing (100% Coverage)
* **Tool:** Playwright.
* **Scope:** Every critical user journey (Happy Path + Common Error Paths) must have an automated E2E test.
* **Stability:** Use strict locators (e.g., `getByRole`, `getByText`) to prevent flaky tests. Do not use brittle CSS/XPath selectors.