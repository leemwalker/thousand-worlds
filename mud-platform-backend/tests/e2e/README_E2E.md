# End-to-End Tests

This directory contains end-to-end (E2E) tests for the Thousand Worlds platform. These tests run against a live running environment (backend + database) to verify critical user journeys.

## Prerequisites

- Running Backend Server (`localhost:8080`)
- Running Database (`localhost:5432`)
- Go 1.21+

## Running Tests

To run the mobile user journey test:

```bash
# From mud-platform-backend directory
go test -v ./tests/e2e/mobile_user_journey_test.go
```

## Configuration

You can configure the test environment using environment variables:

| Variable | Default | Description |
|----------|---------|-------------|
| `TEST_BASE_URL` | `http://localhost:8080` | Backend API URL |
| `TEST_WS_URL` | `ws://localhost:8080` | Backend WebSocket URL |
| `TEST_DB_DSN` | `postgres://admin:password123@localhost:5432/mud_core?sslmode=disable` | Database connection string |

## Test Scenarios

### Mobile User Journey (`TestMobileUserJourney`)

Simulates a new mobile user:
1.  **Account Creation**: Registers a new account.
2.  **Login**: Authenticates and retrieves a token.
3.  **Lobby Interaction**: Connects to WebSocket, sends `look`, `say`, and `tell` commands.
4.  **World Creation**: Completes the interview with the statue.
5.  **Verification**: Checks if the world was created in the database.
6.  **Entry**: Enters the newly created world.

## Troubleshooting

-   **Connection Refused**: Ensure the backend server is running (`./launch.sh`).
-   **Database Error**: Check the `TEST_DB_DSN` matches your local postgres setup.
-   **Timeout**: If the server is slow (e.g., during world generation), you might need to increase timeouts in the test file.
