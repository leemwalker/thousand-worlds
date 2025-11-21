package health

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/nats-io/nats.go"
)

type Pinger interface {
	Ping(ctx context.Context) error
}

type NATSConn interface {
	Status() nats.Status
}

// HealthChecker checks the health of the application and its dependencies.
type HealthChecker struct {
	db    Pinger
	redis Pinger
	nats  NATSConn
}

// NewHealthChecker creates a new HealthChecker.
func NewHealthChecker(db Pinger, redis Pinger, nc NATSConn) *HealthChecker {
	return &HealthChecker{
		db:    db,
		redis: redis,
		nats:  nc,
	}
}

// Check performs the health checks.
func (hc *HealthChecker) Check(ctx context.Context) map[string]string {
	status := make(map[string]string)
	status["status"] = "ok"

	// Check DB
	if hc.db != nil {
		ctx, cancel := context.WithTimeout(ctx, 1*time.Second)
		if err := hc.db.Ping(ctx); err != nil {
			status["database"] = "unhealthy"
			status["status"] = "degraded"
		} else {
			status["database"] = "healthy"
		}
		cancel()
	}

	// Check Redis
	if hc.redis != nil {
		ctx, cancel := context.WithTimeout(ctx, 1*time.Second)
		if err := hc.redis.Ping(ctx); err != nil {
			status["redis"] = "unhealthy"
			status["status"] = "degraded"
		} else {
			status["redis"] = "healthy"
		}
		cancel()
	}

	// Check NATS
	if hc.nats != nil {
		if hc.nats.Status() != nats.CONNECTED {
			status["nats"] = "unhealthy"
			status["status"] = "degraded"
		} else {
			status["nats"] = "healthy"
		}
	}

	return status
}

// Handler returns an HTTP handler for the health check endpoint.
func (hc *HealthChecker) Handler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		status := hc.Check(r.Context())

		w.Header().Set("Content-Type", "application/json")

		// If status is not ok, return 503
		statusCode := http.StatusOK
		if status["status"] != "ok" {
			statusCode = http.StatusServiceUnavailable
		}

		w.WriteHeader(statusCode)
		json.NewEncoder(w).Encode(status)
	}
}
