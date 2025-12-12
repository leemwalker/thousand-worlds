package errors

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestAppError_Error(t *testing.T) {
	tests := []struct {
		name     string
		appErr   *AppError
		expected string
	}{
		{
			name: "error without underlying error",
			appErr: &AppError{
				Code:    "TEST_ERROR",
				Message: "Test message",
			},
			expected: "Test message",
		},
		{
			name: "error with underlying error",
			appErr: &AppError{
				Code:    "TEST_ERROR",
				Message: "Test message",
				Err:     errors.New("underlying error"),
			},
			expected: "Test message: underlying error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.appErr.Error(); got != tt.expected {
				t.Errorf("AppError.Error() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestAppError_Unwrap(t *testing.T) {
	underlying := errors.New("underlying error")
	appErr := &AppError{
		Code:    "TEST",
		Message: "Test",
		Err:     underlying,
	}

	if got := appErr.Unwrap(); got != underlying {
		t.Errorf("AppError.Unwrap() = %v, want %v", got, underlying)
	}

	// Test with nil underlying error
	appErrNoUnderlying := &AppError{
		Code:    "TEST",
		Message: "Test",
	}
	if got := appErrNoUnderlying.Unwrap(); got != nil {
		t.Errorf("AppError.Unwrap() = %v, want nil", got)
	}
}

func TestWrap(t *testing.T) {
	underlying := errors.New("underlying error")
	wrapped := Wrap(ErrNotFound, "Custom message", underlying)

	if wrapped.Code != ErrNotFound.Code {
		t.Errorf("Wrap() Code = %v, want %v", wrapped.Code, ErrNotFound.Code)
	}
	if wrapped.Message != "Custom message" {
		t.Errorf("Wrap() Message = %v, want %v", wrapped.Message, "Custom message")
	}
	if wrapped.HTTPStatus != ErrNotFound.HTTPStatus {
		t.Errorf("Wrap() HTTPStatus = %v, want %v", wrapped.HTTPStatus, ErrNotFound.HTTPStatus)
	}
	if wrapped.Err != underlying {
		t.Errorf("Wrap() Err = %v, want %v", wrapped.Err, underlying)
	}
}

func TestNew(t *testing.T) {
	appErr := New("CUSTOM_CODE", "Custom message", http.StatusTeapot)

	if appErr.Code != "CUSTOM_CODE" {
		t.Errorf("New() Code = %v, want %v", appErr.Code, "CUSTOM_CODE")
	}
	if appErr.Message != "Custom message" {
		t.Errorf("New() Message = %v, want %v", appErr.Message, "Custom message")
	}
	if appErr.HTTPStatus != http.StatusTeapot {
		t.Errorf("New() HTTPStatus = %v, want %v", appErr.HTTPStatus, http.StatusTeapot)
	}
}

func TestRespondWithError_AppError(t *testing.T) {
	recorder := httptest.NewRecorder()

	appErr := &AppError{
		Code:       "TEST_ERROR",
		Message:    "Test error message",
		HTTPStatus: http.StatusBadRequest,
	}

	RespondWithError(recorder, appErr)

	// Check status code
	if recorder.Code != http.StatusBadRequest {
		t.Errorf("RespondWithError() status = %v, want %v", recorder.Code, http.StatusBadRequest)
	}

	// Check content type
	contentType := recorder.Header().Get("Content-Type")
	if contentType != "application/json" {
		t.Errorf("RespondWithError() content-type = %v, want %v", contentType, "application/json")
	}

	// Check response body
	var response ErrorResponse
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if response.Error.Code != "TEST_ERROR" {
		t.Errorf("RespondWithError() response code = %v, want %v", response.Error.Code, "TEST_ERROR")
	}
	if response.Error.Message != "Test error message" {
		t.Errorf("RespondWithError() response message = %v, want %v", response.Error.Message, "Test error message")
	}
}

func TestRespondWithError_NonAppError(t *testing.T) {
	recorder := httptest.NewRecorder()

	regularErr := errors.New("regular error")
	RespondWithError(recorder, regularErr)

	// Non-AppError should be treated as internal server error
	if recorder.Code != http.StatusInternalServerError {
		t.Errorf("RespondWithError() status = %v, want %v", recorder.Code, http.StatusInternalServerError)
	}

	var response ErrorResponse
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if response.Error.Code != "UNKNOWN_ERROR" {
		t.Errorf("RespondWithError() response code = %v, want %v", response.Error.Code, "UNKNOWN_ERROR")
	}
}

func TestPredefinedErrors(t *testing.T) {
	tests := []struct {
		err        *AppError
		code       string
		httpStatus int
	}{
		{ErrInvalidInput, "INVALID_INPUT", http.StatusBadRequest},
		{ErrUnauthorized, "UNAUTHORIZED", http.StatusUnauthorized},
		{ErrForbidden, "FORBIDDEN", http.StatusForbidden},
		{ErrNotFound, "NOT_FOUND", http.StatusNotFound},
		{ErrConflict, "CONFLICT", http.StatusConflict},
		{ErrInternalServer, "INTERNAL_ERROR", http.StatusInternalServerError},
	}

	for _, tt := range tests {
		t.Run(tt.code, func(t *testing.T) {
			if tt.err.Code != tt.code {
				t.Errorf("Error code = %v, want %v", tt.err.Code, tt.code)
			}
			if tt.err.HTTPStatus != tt.httpStatus {
				t.Errorf("HTTP status = %v, want %v", tt.err.HTTPStatus, tt.httpStatus)
			}
		})
	}
}
