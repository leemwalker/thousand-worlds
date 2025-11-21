package logging

import (
	"context"
	"net/http"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

type contextKey string

const (
	correlationIDKey contextKey = "correlation_id"
	loggerKey        contextKey = "logger"
)

// InitLogger initializes the global logger.
func InitLogger() {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.RFC3339})
}

// Middleware adds a correlation ID to the request context and logs the request.
func Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		correlationID := r.Header.Get("X-Correlation-ID")
		if correlationID == "" {
			correlationID = uuid.New().String()
		}

		// Create a logger with the correlation ID
		logger := log.With().Str("correlation_id", correlationID).Logger()

		// Add logger and correlation ID to context
		ctx := context.WithValue(r.Context(), correlationIDKey, correlationID)
		ctx = context.WithValue(ctx, loggerKey, logger)

		start := time.Now()

		// Log request start
		logger.Info().
			Str("method", r.Method).
			Str("path", r.URL.Path).
			Msg("Request started")

		next.ServeHTTP(w, r.WithContext(ctx))

		// Log request completion
		logger.Info().
			Str("method", r.Method).
			Str("path", r.URL.Path).
			Dur("duration", time.Since(start)).
			Msg("Request completed")
	})
}

// FromContext returns the logger from the context, or the global logger if not found.
func FromContext(ctx context.Context) *zerolog.Logger {
	if logger, ok := ctx.Value(loggerKey).(zerolog.Logger); ok {
		return &logger
	}
	return &log.Logger
}

// GetCorrelationID returns the correlation ID from the context.
func GetCorrelationID(ctx context.Context) string {
	if id, ok := ctx.Value(correlationIDKey).(string); ok {
		return id
	}
	return ""
}
