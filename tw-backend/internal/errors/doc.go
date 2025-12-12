// Package errors provides standardized error handling for the Thousand Worlds API.
//
// # Core Types
//
//   - AppError: Application-level error with HTTP context, error code, and message
//   - ErrorResponse: JSON structure for API error responses
//
// # Usage
//
// Using predefined errors:
//
//	if user == nil {
//	    return errors.ErrNotFound
//	}
//
// Wrapping errors with context:
//
//	if err := db.Query(...); err != nil {
//	    return errors.Wrap(errors.ErrInternalServer, "failed to query users", err)
//	}
//
// Creating custom errors:
//
//	return errors.New("CUSTOM_ERROR", "Something went wrong", http.StatusBadRequest)
//
// Responding to HTTP requests:
//
//	func handler(w http.ResponseWriter, r *http.Request) {
//	    if err := doSomething(); err != nil {
//	        errors.RespondWithError(w, err)
//	        return
//	    }
//	}
//
// # Error Categories
//
// Domain-specific errors are defined in domain.go:
//   - Authentication: ErrAuthInvalidCredentials, ErrAuthTokenExpired, etc.
//   - User: ErrUserNotFound, ErrUserExists, etc.
//   - Character: ErrCharacterNotFound, ErrCharacterNotOwned, etc.
//   - World: ErrWorldNotFound, ErrWorldPrivate, etc.
//   - Session: ErrSessionExpired, ErrAlreadyInGame, etc.
//   - Game Actions: ErrInvalidCommand, ErrInsufficientStamina, etc.
//   - Inventory: ErrItemNotFound, ErrInventoryFull, etc.
//   - Crafting: ErrRecipeNotFound, ErrMissingIngredients, etc.
package errors
