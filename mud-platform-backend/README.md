# Thousand Worlds - Backend

## Development Setup

### Prerequisites
- Go 1.21+
- PostgreSQL 14+ with PostGIS extension
- Make (optional)

### Database Setup

1. Create database:
```bash
createdb thousand_worlds
psql thousand_worlds -c "CREATE EXTENSION postgis;"
```

2. Run migrations:
```bash
# Install migrate tool
go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

# Run migrations
migrate -path migrations/postgres -database "postgresql://postgres:postgres@localhost:5432/thousand_worlds?sslmode=disable" up
```

### Running the Server

1. Install dependencies:
```bash
go mod download
```

2. Set environment variables:
```bash
export DATABASE_URL="postgresql://postgres:postgres@localhost:5432/thousand_worlds?sslmode=disable"
export JWT_SECRET="your-secret-key"
export PORT="8080"
```

3. Run server:
```bash
go run cmd/game-server/main.go
```

Or use the dev script:
```bash
chmod +x scripts/dev.sh
./scripts/dev.sh
```

### API Endpoints

#### Authentication
- `POST /api/auth/register` - Create account
- `POST /api/auth/login` - Login
- `GET /api/auth/me` - Get current user (requires auth)

#### WebSocket
- `WS /api/game/ws` - Game WebSocket connection (requires auth)

### Testing

Run all tests:
```bash
go test ./...
```

Run with coverage:
```bash
go test -cover ./...
```

### Environment Variables

- `DATABASE_URL` - PostgreSQL connection string (required)
- `JWT_SECRET` - Secret key for JWT tokens (required)
- `PORT` - HTTP server port (default: 8080)

## Project Structure

```
cmd/
  game-server/          # Main game server
    api/               # HTTP API handlers
    websocket/         # WebSocket handlers
internal/
  auth/               # Authentication logic
  game/               # Game logic
    processor/        # Command processor
migrations/
  postgres/           # Database migrations
```
