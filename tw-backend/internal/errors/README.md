# Errors Package

Standardized error handling for the Thousand Worlds backend API.

## Core Types

### AppError
```go
type AppError struct {
    Code       string // Machine-readable code (e.g., "AUTH_INVALID_CREDENTIALS")
    Message    string // Human-readable message
    HTTPStatus int    // HTTP status code
    Err        error  // Underlying error (for wrapping)
}
```

## Usage

### Creating Errors
```go
// Using predefined errors
return errors.ErrNotFound

// Wrapping with context
return errors.Wrap(errors.ErrInvalidInput, "Invalid email format", nil)

// Creating custom errors
return errors.New("CUSTOM_ERROR", "Custom message", http.StatusBadRequest)
```

### Responding to HTTP Requests
```go
func handler(w http.ResponseWriter, r *http.Request) {
    if err := doSomething(); err != nil {
        errors.RespondWithError(w, err)
        return
    }
}
```

### JSON Response Format
```json
{
  "error": {
    "code": "AUTH_INVALID_CREDENTIALS",
    "message": "Invalid email or password"
  }
}
```

## Error Categories

| File | Domain |
|------|--------|
| `types.go` | Core types: AppError, Wrap, New, RespondWithError |
| `domain.go` | Domain-specific errors by category |

### Available Error Codes

**Authentication:** `AUTH_*`
- `AUTH_INVALID_CREDENTIALS`, `AUTH_TOKEN_EXPIRED`, `AUTH_TOKEN_INVALID`, `AUTH_RATE_LIMITED`

**User:** `USER_*`
- `USER_NOT_FOUND`, `USER_EXISTS`, `EMAIL_INVALID`, `PASSWORD_WEAK`

**Character:** `CHARACTER_*`
- `CHARACTER_NOT_FOUND`, `CHARACTER_EXISTS`, `CHARACTER_NOT_OWNED`, `CHARACTER_NAME_INVALID`

**World:** `WORLD_*`, `INTERVIEW_*`
- `WORLD_NOT_FOUND`, `WORLD_PRIVATE`, `WORLD_FULL`, `INTERVIEW_IN_PROGRESS`, `INTERVIEW_NOT_FOUND`

**Session:** `SESSION_*`
- `SESSION_NOT_FOUND`, `SESSION_EXPIRED`, `ALREADY_IN_GAME`, `NOT_IN_GAME`, `WEBSOCKET_REQUIRED`

**Game Actions:** Various
- `INVALID_COMMAND`, `INSUFFICIENT_STAMINA`, `MOVEMENT_BLOCKED`, `TARGET_NOT_FOUND`, `TARGET_OUT_OF_RANGE`, `ACTION_COOLDOWN`

**Inventory:** `ITEM_*`, `INVENTORY_*`
- `ITEM_NOT_FOUND`, `INVENTORY_FULL`, `CANNOT_EQUIP`, `INSUFFICIENT_SKILL`

**Crafting:** `RECIPE_*`, `CRAFTING_*`
- `RECIPE_NOT_FOUND`, `MISSING_INGREDIENTS`, `CRAFTING_STATION_NEEDED`

**Database:** `DATABASE_*`
- `DATABASE_ERROR`, `DATABASE_TIMEOUT`
