package api

import (
	"encoding/json"
	"net/http"
	"runtime"
	"sync/atomic"
	"time"
)

// HealthHandler provides health and readiness check endpoints
// Used by Kubernetes/load balancers for service discovery
type HealthHandler struct {
	startTime      time.Time
	isReady        atomic.Bool
	connectedUsers atomic.Int64
}

// NewHealthHandler creates a new health check handler
func NewHealthHandler() *HealthHandler {
	h := &HealthHandler{
		startTime: time.Now(),
	}
	h.isReady.Store(true) // Start as ready
	return h
}

// SetReady sets the readiness status
// Call with false during graceful shutdown to drain connections
func (h *HealthHandler) SetReady(ready bool) {
	h.isReady.Store(ready)
}

// SetConnectedUsers updates the connected user count
func (h *HealthHandler) SetConnectedUsers(count int64) {
	h.connectedUsers.Store(count)
}

// HealthResponse represents the health check response
type HealthResponse struct {
	Status         string  `json:"status"`
	Uptime         string  `json:"uptime"`
	ConnectedUsers int64   `json:"connected_users"`
	Goroutines     int     `json:"goroutines"`
	MemoryMB       float64 `json:"memory_mb"`
}

// LivenessProbe checks if the service is alive
// Returns 200 if the service is running
// Never returns unhealthy unless the service is completely broken
func (h *HealthHandler) LivenessProbe(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	json.NewEncoder(w).Encode(map[string]string{
		"status": "alive",
	})
}

// ReadinessProbe checks if the service is ready to accept traffic
// Returns 200 if ready, 503 if not ready (during startup or graceful shutdown)
func (h *HealthHandler) ReadinessProbe(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if !h.isReady.Load() {
		w.WriteHeader(http.StatusServiceUnavailable)
		json.NewEncoder(w).Encode(map[string]string{
			"status": "not_ready",
			"reason": "graceful_shutdown_in_progress",
		})
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"status": "ready",
	})
}

// HealthCheck provides detailed health information
// Includes metrics useful for monitoring and load balancing decisions
func (h *HealthHandler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Get memory stats
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	memoryMB := float64(m.Alloc) / 1024 / 1024

	uptime := time.Since(h.startTime)

	response := HealthResponse{
		Status:         "healthy",
		Uptime:         uptime.String(),
		ConnectedUsers: h.connectedUsers.Load(),
		Goroutines:     runtime.NumGoroutine(),
		MemoryMB:       memoryMB,
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// RegisterRoutes registers health check routes on the provided mux
func (h *HealthHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/health", h.HealthCheck)
	mux.HandleFunc("/health/live", h.LivenessProbe)
	mux.HandleFunc("/health/ready", h.ReadinessProbe)
}
