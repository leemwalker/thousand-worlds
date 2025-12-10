package api

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"

	"tw-backend/internal/world/interview"
)

type InterviewHandler struct {
	service *interview.InterviewService
}

func NewInterviewHandler(service *interview.InterviewService) *InterviewHandler {
	return &InterviewHandler{service: service}
}

type StartInterviewResponse struct {
	SessionID uuid.UUID `json:"session_id"`
	Question  string    `json:"question"`
}

func (h *InterviewHandler) StartInterview(w http.ResponseWriter, r *http.Request) {
	userID := getUserIDFromContext(r.Context())
	if userID == uuid.Nil {
		respondError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	session, question, err := h.service.StartInterview(r.Context(), userID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to start interview: "+err.Error())
		return
	}

	respondJSON(w, http.StatusCreated, StartInterviewResponse{
		SessionID: session.ID,
		Question:  question,
	})
}

type MessageRequest struct {
	SessionID uuid.UUID `json:"session_id"`
	Message   string    `json:"message"`
}

type MessageResponse struct {
	Question  string `json:"question"`
	Completed bool   `json:"completed"`
}

func (h *InterviewHandler) ProcessMessage(w http.ResponseWriter, r *http.Request) {
	userID := getUserIDFromContext(r.Context())
	if userID == uuid.Nil {
		respondError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	var req MessageRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Verify session belongs to user (optional, but good practice)
	// The service doesn't expose a way to check ownership easily without loading the session first.
	// For now, we rely on the service to handle session lookup.
	// Ideally, we should check if the session belongs to the user.

	question, completed, err := h.service.ProcessResponse(r.Context(), userID, req.Message)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to process message: "+err.Error())
		return
	}

	respondJSON(w, http.StatusOK, MessageResponse{
		Question:  question,
		Completed: completed,
	})
}

func (h *InterviewHandler) GetActiveInterview(w http.ResponseWriter, r *http.Request) {
	userID := getUserIDFromContext(r.Context())
	if userID == uuid.Nil {
		respondError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	session, err := h.service.GetActiveInterview(r.Context(), userID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to get active interview: "+err.Error())
		return
	}

	if session == nil {
		respondError(w, http.StatusNotFound, "No active interview found")
		return
	}

	// Resume the interview to get the last question
	_, question, err := h.service.ResumeInterview(r.Context(), userID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to resume interview: "+err.Error())
		return
	}

	respondJSON(w, http.StatusOK, StartInterviewResponse{
		SessionID: session.ID,
		Question:  question,
	})
}

func (h *InterviewHandler) FinalizeInterview(w http.ResponseWriter, r *http.Request) {
	userID := getUserIDFromContext(r.Context())
	if userID == uuid.Nil {
		respondError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	var req struct {
		SessionID uuid.UUID `json:"session_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	config, err := h.service.CompleteInterview(r.Context(), userID, req.SessionID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to finalize interview: "+err.Error())
		return
	}

	respondJSON(w, http.StatusOK, config)
}
