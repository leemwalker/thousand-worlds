package errors

import (
	"fmt"
	"net/http"
)

// Domain-specific error codes for consistent API responses

// Authentication errors
var (
	ErrAuthInvalidCredentials = &AppError{Code: "AUTH_INVALID_CREDENTIALS", Message: "Invalid email or password", HTTPStatus: http.StatusUnauthorized}
	ErrAuthTokenExpired       = &AppError{Code: "AUTH_TOKEN_EXPIRED", Message: "Authentication token has expired", HTTPStatus: http.StatusUnauthorized}
	ErrAuthTokenInvalid       = &AppError{Code: "AUTH_TOKEN_INVALID", Message: "Authentication token is invalid", HTTPStatus: http.StatusUnauthorized}
	ErrAuthRateLimited        = &AppError{Code: "AUTH_RATE_LIMITED", Message: "Too many attempts, please try again later", HTTPStatus: http.StatusTooManyRequests}
)

// User errors
var (
	ErrUserNotFound = &AppError{Code: "USER_NOT_FOUND", Message: "User not found", HTTPStatus: http.StatusNotFound}
	ErrUserExists   = &AppError{Code: "USER_EXISTS", Message: "User already exists", HTTPStatus: http.StatusConflict}
	ErrEmailInvalid = &AppError{Code: "EMAIL_INVALID", Message: "Invalid email format", HTTPStatus: http.StatusBadRequest}
	ErrPasswordWeak = &AppError{Code: "PASSWORD_WEAK", Message: "Password does not meet requirements", HTTPStatus: http.StatusBadRequest}
)

// Character errors
var (
	ErrCharacterNotFound    = &AppError{Code: "CHARACTER_NOT_FOUND", Message: "Character not found", HTTPStatus: http.StatusNotFound}
	ErrCharacterExists      = &AppError{Code: "CHARACTER_EXISTS", Message: "Character already exists in this world", HTTPStatus: http.StatusConflict}
	ErrCharacterNotOwned    = &AppError{Code: "CHARACTER_NOT_OWNED", Message: "This character does not belong to you", HTTPStatus: http.StatusForbidden}
	ErrCharacterNameInvalid = &AppError{Code: "CHARACTER_NAME_INVALID", Message: "Character name is invalid", HTTPStatus: http.StatusBadRequest}
)

// World errors
var (
	ErrWorldNotFound       = &AppError{Code: "WORLD_NOT_FOUND", Message: "World not found", HTTPStatus: http.StatusNotFound}
	ErrWorldPrivate        = &AppError{Code: "WORLD_PRIVATE", Message: "This world is private", HTTPStatus: http.StatusForbidden}
	ErrWorldFull           = &AppError{Code: "WORLD_FULL", Message: "World has reached maximum players", HTTPStatus: http.StatusConflict}
	ErrInterviewInProgress = &AppError{Code: "INTERVIEW_IN_PROGRESS", Message: "World interview already in progress", HTTPStatus: http.StatusConflict}
	ErrInterviewNotFound   = &AppError{Code: "INTERVIEW_NOT_FOUND", Message: "No active interview found", HTTPStatus: http.StatusNotFound}
)

// Game session errors
var (
	ErrSessionNotFound   = &AppError{Code: "SESSION_NOT_FOUND", Message: "Session not found", HTTPStatus: http.StatusNotFound}
	ErrSessionExpired    = &AppError{Code: "SESSION_EXPIRED", Message: "Session has expired", HTTPStatus: http.StatusUnauthorized}
	ErrAlreadyInGame     = &AppError{Code: "ALREADY_IN_GAME", Message: "Already in a game session", HTTPStatus: http.StatusConflict}
	ErrNotInGame         = &AppError{Code: "NOT_IN_GAME", Message: "Not currently in a game session", HTTPStatus: http.StatusBadRequest}
	ErrWebSocketRequired = &AppError{Code: "WEBSOCKET_REQUIRED", Message: "WebSocket connection required", HTTPStatus: http.StatusUpgradeRequired}
)

// Game action errors
var (
	ErrInvalidCommand      = &AppError{Code: "INVALID_COMMAND", Message: "Invalid command", HTTPStatus: http.StatusBadRequest}
	ErrInsufficientStamina = &AppError{Code: "INSUFFICIENT_STAMINA", Message: "Not enough stamina for this action", HTTPStatus: http.StatusBadRequest}
	ErrMovementBlocked     = &AppError{Code: "MOVEMENT_BLOCKED", Message: "Cannot move in that direction", HTTPStatus: http.StatusBadRequest}
	ErrTargetNotFound      = &AppError{Code: "TARGET_NOT_FOUND", Message: "Target not found", HTTPStatus: http.StatusNotFound}
	ErrTargetOutOfRange    = &AppError{Code: "TARGET_OUT_OF_RANGE", Message: "Target is out of range", HTTPStatus: http.StatusBadRequest}
	ErrCooldown            = &AppError{Code: "ACTION_COOLDOWN", Message: "Action is on cooldown", HTTPStatus: http.StatusTooManyRequests}
)

// Inventory errors
var (
	ErrItemNotFound      = &AppError{Code: "ITEM_NOT_FOUND", Message: "Item not found", HTTPStatus: http.StatusNotFound}
	ErrInventoryFull     = &AppError{Code: "INVENTORY_FULL", Message: "Inventory is full", HTTPStatus: http.StatusBadRequest}
	ErrCannotEquip       = &AppError{Code: "CANNOT_EQUIP", Message: "Cannot equip this item", HTTPStatus: http.StatusBadRequest}
	ErrInsufficientSkill = &AppError{Code: "INSUFFICIENT_SKILL", Message: "Skill level too low for this item", HTTPStatus: http.StatusBadRequest}
)

// Crafting errors
var (
	ErrRecipeNotFound        = &AppError{Code: "RECIPE_NOT_FOUND", Message: "Recipe not found", HTTPStatus: http.StatusNotFound}
	ErrMissingIngredients    = &AppError{Code: "MISSING_INGREDIENTS", Message: "Missing required ingredients", HTTPStatus: http.StatusBadRequest}
	ErrCraftingStationNeeded = &AppError{Code: "CRAFTING_STATION_NEEDED", Message: "Crafting station required", HTTPStatus: http.StatusBadRequest}
)

// Database errors
var (
	ErrDatabaseConnection = &AppError{Code: "DATABASE_ERROR", Message: "Database connection error", HTTPStatus: http.StatusServiceUnavailable}
	ErrDatabaseTimeout    = &AppError{Code: "DATABASE_TIMEOUT", Message: "Database operation timed out", HTTPStatus: http.StatusGatewayTimeout}
)

// Helper functions for dynamic errors

// NewNotFound returns a NotFound error with a custom message
func NewNotFound(format string, args ...any) error {
	return &AppError{
		Code:       ErrNotFound.Code,
		Message:    fmt.Sprintf(format, args...),
		HTTPStatus: ErrNotFound.HTTPStatus,
	}
}

// NewInvalidInput returns an InvalidInput error with a custom message
func NewInvalidInput(format string, args ...any) error {
	return &AppError{
		Code:       ErrInvalidInput.Code,
		Message:    fmt.Sprintf(format, args...),
		HTTPStatus: ErrInvalidInput.HTTPStatus,
	}
}

// NewInternalError returns an AppError for internal errors
func NewInternalError(format string, args ...any) error {
	return &AppError{
		Code:       ErrInternalServer.Code,
		Message:    fmt.Sprintf(format, args...),
		HTTPStatus: ErrInternalServer.HTTPStatus,
	}
}
