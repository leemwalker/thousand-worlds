---
description: Run database migrations
---
Apply database migrations to PostgreSQL.

1. Ensure PostgreSQL is running: `docker ps | grep postgres`
// turbo
2. Navigate to backend: `cd tw-backend`
// turbo
3. Run `make migrate` to apply all pending migrations
4. Verify with: `psql -h localhost -U admin -d mud_core -c '\dt'`
