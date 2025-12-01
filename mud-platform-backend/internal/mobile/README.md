# Mobile SDK for Thousand Worlds

Go-based Mobile SDK for building mobile applications that interact with the Thousand Worlds MUD game platform.

## Features

- ✅ **Authentication**: Register, login, and session management with JWT tokens
- ✅ **Character Management**: Create characters, list characters, join games
- ✅ **World Management**: List available game worlds
- ✅ **WebSocket Game Client**: Real-time game communication
- ✅ **Push Notifications**: Device registration and notification management
- ✅ **Comprehensive Testing**: 83.6% code coverage with unit and integration tests

## Installation

```bash
go get mud-platform-backend/internal/mobile
```

## Quick Start

```go
package main

import (
    "context"
    "log"
    
    "mud-platform-backend/internal/mobile"
)

func main() {
    // Create client
    client := mobile.NewClient("http://localhost:8080")
    
    // Register new user
    user, err := client.Register(context.Background(), "user@example.com", "SecurePassword123")
    if err != nil {
        log.Fatal(err)
    }
    log.Printf("Registered user: %s", user.UserID)
    
    // Login
    loginResp, err := client.Login(context.Background(), "user@example.com", "SecurePassword123")
    if err != nil {
        log.Fatal(err)
    }
    log.Printf("Login successful, token: %s", loginResp.Token)
    // Token is automatically set in client
    
    // List available worlds
    worlds, err := client.ListWorlds(context.Background())
    if err != nil {
        log.Fatal(err)
    }
    log.Printf("Found %d worlds", len(worlds))
    
    // Create a character
    charReq := &mobile.CreateCharacterRequest{
        WorldID: worlds[0].ID,
        Name:    "MyHero",
        Species: "Human",
    }
    char, err := client.CreateCharacter(context.Background(), charReq)
    if err != nil {
        log.Fatal(err)
    }
    log.Printf("Created character: %s", char.Name)
    
    // Join the game
    joinResp, err := client.JoinGame(context.Background(), char.CharacterID)
    if err != nil {
        log.Fatal(err)
    }
    log.Printf("Joined game: %s", joinResp.Message)
}
```

## WebSocket Usage

```go
// Create WebSocket client
ws := mobile.NewGameWebSocket("ws://localhost:8080", client.GetToken())

// Register message handler
unsubscribe := ws.OnMessage(func(msg *mobile.ServerMessage) {
    log.Printf("Received: %s - %v", msg.Type, msg.Data)
})
defer unsubscribe()

// Connect
err := ws.Connect(context.Background())
if err != nil {
    log.Fatal(err)
}
defer ws.Disconnect()

// Send commands
cmd := &mobile.Command{
    Action:    "move",
    Direction: "north",
}
ws.SendCommand(cmd)
```

## Push Notifications

```go
// Register device for push notifications
err := client.RegisterForPushNotifications(ctx, "device-token-123", "ios")
if err != nil {
    log.Fatal(err)
}

// Get notifications
notifications, err := client.GetNotifications(ctx, true) // unread only
if err != nil {
    log.Fatal(err)
}
for _, notif := range notifications {
    log.Printf("%s: %s", notif.Title, notif.Message)
}

// Update preferences
prefs := &mobile.NotificationPreferences{
    EnableNewWorldAlerts:   true,
    EnableUserSignInAlerts: true,
    PushEnabled:            true,
}
err = client.UpdateNotificationPreferences(ctx, prefs)
```

## API Reference

### Client

#### Authentication
- `NewClient(baseURL string) *Client` - Create a new client
- `Register(ctx, email, password) (*User, error)` - Register new user
- `Login(ctx, email, password) (*LoginResponse, error)` - Login and get token
- `GetMe(ctx) (*User, error)` - Get current user info
- `Logout()` - Clear authentication token

#### Characters
- `CreateCharacter(ctx, req) (*Character, error)` - Create a character
- `GetCharacters(ctx) ([]*Character, error)` - List user's characters
- `JoinGame(ctx, characterID) (*JoinGameResponse, error)` - Join game with character

#### Worlds
- `ListWorlds(ctx) ([]*World, error)` - List available worlds

#### Notifications
- `RegisterForPushNotifications(ctx, deviceToken, platform) error` - Register device
- `GetNotifications(ctx, unreadOnly bool) ([]*Notification, error)` - Get notifications
- `MarkNotificationAsRead(ctx, notificationID) error` - Mark notification as read
- `GetNotificationPreferences(ctx) (*NotificationPreferences, error)` - Get preferences
- `UpdateNotificationPreferences(ctx, prefs) error` - Update preferences

### GameWebSocket

- `NewGameWebSocket(baseURL, token) *GameWebSocket` - Create WebSocket client
- `Connect(ctx) error` - Establish connection
- `Disconnect() error` - Close connection
- `IsConnected() bool` - Check connection status
- `SendCommand(cmd) error` - Send game command
- `OnMessage(handler) func()` - Register message handler (returns unsubscribe function)

## Testing

Run unit tests:
```bash
go test -v ./internal/mobile/...
```

Run with coverage:
```bash
go test -cover ./internal/mobile/...
```

Run integration tests (requires database):
```bash
go test -v ./internal/mobile/mobile_integration_test.go
```

## Mobile API Optimization Opportunities

The current REST API is sufficient for mobile use, but the following endpoints could benefit from mobile-specific versions:

1. **`GET /api/game/worlds`** - Add pagination and filtering
   - Current: Returns all worlds
   - Mobile optimization: Add `?page=1&limit=20&filter=active`

2. **`GET /api/game/characters`** - Add field selection
   - Current: Returns full character objects
   - Mobile optimization: Add `?fields=id,name,world_id` to reduce payload

3. **`GET /api/notifications`** - Add since parameter for incremental updates
   - Current: Returns all notifications
   - Mobile optimization: Add `?since=<timestamp>` to fetch only new ones

4. **WebSocket messages** - Add message compression
   - Current: Plain JSON messages
   - Mobile optimization: Optional GZIP compression for large payloads

5. **Batch operations** - Add batch endpoints
   - New: `POST /api/mobile/batch` to combine multiple requests
   - Reduces round trips for initial app load

## Architecture

The SDK follows a layered architecture:

```
┌─────────────────────────────────────┐
│  Mobile Application (iOS/Android)   │
└─────────────────────────────────────┘
                  ↓
┌─────────────────────────────────────┐
│      Mobile SDK (Go)                 │
│  - Client (HTTP/REST)                │
│  - GameWebSocket (gorilla/websocket) │
│  - Types & Models                    │
└─────────────────────────────────────┘
                  ↓
┌─────────────────────────────────────┐
│    Game Backend API                  │
│  - Chi Router                        │
│  - Auth Middleware                   │
│  - Game Server                       │
└─────────────────────────────────────┘
```

## Thread Safety

The `Client` and `GameWebSocket` types are designed to be thread-safe. You can make concurrent requests from multiple goroutines.

## License

See main project license.
