package mobile

import (
	"time"
)

// User represents a user account
type User struct {
	UserID    string    `json:"user_id"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
}

// LoginResponse represents the response from login
type LoginResponse struct {
	Token string `json:"token"`
	User  *User  `json:"user"`
}

// ErrorResponse represents an API error response
type ErrorResponse struct {
	ErrorData struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
}

// Error implements the error interface for ErrorResponse
func (e *ErrorResponse) Error() string {
	if e.ErrorData.Message != "" {
		return e.ErrorData.Message
	}
	return e.ErrorData.Code
}
