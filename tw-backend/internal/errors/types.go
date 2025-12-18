package errors

import (
	"encoding/json"
	stdErrors "errors"
	"fmt"
	"net/http"
)

// AppError represents an application-level error with HTTP context
type AppError struct {
	Code       string `json:"code"`    // Machine-readable code (e.g., "AUTH_INVALID_CREDENTIALS")
	Message    string `json:"message"` // Human-readable message
	HTTPStatus int    `json:"-"`       // HTTP status code (not serialized)
	Err        error  `json:"-"`       // Underlying error (not serialized)
}

func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Err)
	}
	return e.Message
}

// Unwrap returns the underlying error for error chain support
func (e *AppError) Unwrap() error {
	return e.Err
}

// Common error templates
var (
	ErrInvalidInput   = &AppError{Code: "INVALID_INPUT", Message: "Invalid input", HTTPStatus: http.StatusBadRequest}
	ErrUnauthorized   = &AppError{Code: "UNAUTHORIZED", Message: "Unauthorized", HTTPStatus: http.StatusUnauthorized}
	ErrForbidden      = &AppError{Code: "FORBIDDEN", Message: "Forbidden", HTTPStatus: http.StatusForbidden}
	ErrNotFound       = &AppError{Code: "NOT_FOUND", Message: "Not found", HTTPStatus: http.StatusNotFound}
	ErrConflict       = &AppError{Code: "CONFLICT", Message: "Conflict", HTTPStatus: http.StatusConflict}
	ErrInternalServer = &AppError{Code: "INTERNAL_ERROR", Message: "Internal server error", HTTPStatus: http.StatusInternalServerError}
)

// Wrap creates a new error wrapping the original with a custom message
func Wrap(base *AppError, message string, err error) *AppError {
	return &AppError{
		Code:       base.Code,
		Message:    message,
		HTTPStatus: base.HTTPStatus,
		Err:        err,
	}
}

// New creates a new AppError with custom values
func New(code string, message string, httpStatus int) *AppError {
	return &AppError{
		Code:       code,
		Message:    message,
		HTTPStatus: httpStatus,
	}
}

// ErrorResponse represents the JSON error response structure
type ErrorResponse struct {
	Error struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
}

// RespondWithError writes an error response to the HTTP writer
func RespondWithError(w http.ResponseWriter, err error) {
	var appErr *AppError
	if !stdErrors.As(err, &appErr) {
		// If not an AppError, treat as internal server error
		appErr = &AppError{
			Code:       "UNKNOWN_ERROR",
			Message:    "An unexpected error occurred",
			HTTPStatus: http.StatusInternalServerError,
			Err:        err,
		}
	}

	response := ErrorResponse{}
	response.Error.Code = appErr.Code
	response.Error.Message = appErr.Message

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(appErr.HTTPStatus)
	_ = json.NewEncoder(w).Encode(response) // Error intentionally ignored - response already committed
}
