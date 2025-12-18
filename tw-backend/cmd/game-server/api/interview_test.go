package api

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"tw-backend/internal/world/interview"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockInterviewer
type MockInterviewer struct {
	mock.Mock
}

func (m *MockInterviewer) StartInterview(ctx context.Context, userID uuid.UUID) (*interview.InterviewSession, string, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.String(1), args.Error(2)
	}
	return args.Get(0).(*interview.InterviewSession), args.String(1), args.Error(2)
}

func (m *MockInterviewer) ProcessResponse(ctx context.Context, userID uuid.UUID, message string) (string, bool, error) {
	args := m.Called(ctx, userID, message)
	return args.String(0), args.Bool(1), args.Error(2)
}

func (m *MockInterviewer) GetActiveInterview(ctx context.Context, userID uuid.UUID) (*interview.InterviewSession, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*interview.InterviewSession), args.Error(1)
}

func (m *MockInterviewer) ResumeInterview(ctx context.Context, userID uuid.UUID) (*interview.InterviewSession, string, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.String(1), args.Error(2)
	}
	return args.Get(0).(*interview.InterviewSession), args.String(1), args.Error(2)
}

func (m *MockInterviewer) CompleteInterview(ctx context.Context, userID, sessionID uuid.UUID) (*interview.WorldConfiguration, error) {
	args := m.Called(ctx, userID, sessionID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*interview.WorldConfiguration), args.Error(1)
}

func TestStartInterview(t *testing.T) {
	mockService := new(MockInterviewer)
	handler := NewInterviewHandler(mockService)

	userID := uuid.New()
	sessionID := uuid.New()
	question := "What is your world name?"

	// Mock expectations
	mockService.On("StartInterview", mock.Anything, userID).
		Return(&interview.InterviewSession{ID: sessionID}, question, nil)

	// Create request with context
	req, _ := http.NewRequest("POST", "/interview/start", nil)
	ctx := context.WithValue(req.Context(), "userID", userID.String())
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()
	handler.StartInterview(rr, req)

	assert.Equal(t, http.StatusCreated, rr.Code)

	var resp StartInterviewResponse
	json.NewDecoder(rr.Body).Decode(&resp)
	assert.Equal(t, sessionID, resp.SessionID)
	assert.Equal(t, question, resp.Question)
}

func TestStartInterview_Unauthorized(t *testing.T) {
	mockService := new(MockInterviewer)
	handler := NewInterviewHandler(mockService)

	req, _ := http.NewRequest("POST", "/interview/start", nil)
	rr := httptest.NewRecorder()
	handler.StartInterview(rr, req)

	assert.Equal(t, http.StatusUnauthorized, rr.Code)
}

func TestProcessMessage(t *testing.T) {
	mockService := new(MockInterviewer)
	handler := NewInterviewHandler(mockService)

	userID := uuid.New()
	sessionID := uuid.New()
	msgReq := MessageRequest{SessionID: sessionID, Message: "My World"}

	mockService.On("ProcessResponse", mock.Anything, userID, "My World").
		Return("Next question?", false, nil)

	body, _ := json.Marshal(msgReq)
	req, _ := http.NewRequest("POST", "/interview/message", bytes.NewBuffer(body))
	ctx := context.WithValue(req.Context(), "userID", userID.String())
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()
	handler.ProcessMessage(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var resp MessageResponse
	json.NewDecoder(rr.Body).Decode(&resp)
	assert.Equal(t, "Next question?", resp.Question)
	assert.False(t, resp.Completed)
}

func TestProcessMessage_Error(t *testing.T) {
	mockService := new(MockInterviewer)
	handler := NewInterviewHandler(mockService)

	userID := uuid.New()
	msgReq := MessageRequest{Message: "Bad input"}

	mockService.On("ProcessResponse", mock.Anything, userID, "Bad input").
		Return("", false, errors.New("processing failed"))

	body, _ := json.Marshal(msgReq)
	req, _ := http.NewRequest("POST", "/interview/message", bytes.NewBuffer(body))
	ctx := context.WithValue(req.Context(), "userID", userID.String())
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()
	handler.ProcessMessage(rr, req)

	assert.Equal(t, http.StatusInternalServerError, rr.Code)
}

func TestGetActiveInterview(t *testing.T) {
	mockService := new(MockInterviewer)
	handler := NewInterviewHandler(mockService)

	userID := uuid.New()
	sessionID := uuid.New()

	mockService.On("GetActiveInterview", mock.Anything, userID).
		Return(&interview.InterviewSession{ID: sessionID}, nil)

	mockService.On("ResumeInterview", mock.Anything, userID).
		Return(&interview.InterviewSession{ID: sessionID}, "Resume Q", nil)

	req, _ := http.NewRequest("GET", "/interview/active", nil)
	ctx := context.WithValue(req.Context(), "userID", userID.String())
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()
	handler.GetActiveInterview(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
}
