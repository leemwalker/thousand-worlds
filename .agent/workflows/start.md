---
description: Start all services with Docker Compose
---
Start the full development environment.

1. Navigate to backend: `cd tw-backend`
// turbo
2. Run `make up` to start PostgreSQL, Redis, NATS, and Ollama
// turbo
3. Wait for services to be healthy
4. Run `make run` to start the game server
5. Frontend available at http://localhost:8080
