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
	userIDKey        contextKey = "user_id"
)

// InitLogger initializes the global logger.
func InitLogger() {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.RFC3339})
}

// responseWriter wraps http.ResponseWriter to capture status code
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
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

		// Wrap response writer to capture status code
		ww := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

		// Log request start
		logger.Info().
			Str("method", r.Method).
			Str("path", r.URL.Path).
			Str("remote_addr", r.RemoteAddr).
			Msg("Request started")

		next.ServeHTTP(ww, r.WithContext(ctx))

		// Log request completion
		logger.Info().
			Str("method", r.Method).
			Str("path", r.URL.Path).
			Int("status", ww.statusCode).
			Dur("duration_ms", time.Since(start)).
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

// LogError logs an error with context
func LogError(ctx context.Context, err error, message string, fields map[string]interface{}) {
	logger := FromContext(ctx)
	event := logger.Error().Err(err)
	
	for k, v := range fields {
		event = event.Interface(k, v)
	}
	
	event.Msg(message)
}

// LogInfo logs an info message with context
func LogInfo(ctx context.Context, message string, fields map[string]interface{}) {
	logger := FromContext(ctx)
	event := logger.Info()
	
	for k, v := range fields {
		event = event.Interface(k, v)
	}
	
	event.Msg(message)
}

// LogWarning logs a warning message with context
func LogWarning(ctx context.Context, message string, fields map[string]interface{}) {
	logger := FromContext(ctx)
	event := logger.Warn()
	
	for k, v := range fields {
		event = event.Interface(k, v)
	}
	
	event.Msg(message)
}
