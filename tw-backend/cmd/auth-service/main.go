package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"
	"tw-backend/internal/auth"

	"github.com/nats-io/nats.go"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func main() {
	// Setup Logging
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})

	// Configuration (Env vars would go here)
	natsURL := os.Getenv("NATS_URL")
	if natsURL == "" {
		natsURL = nats.DefaultURL
	}

	// Connect to NATS
	nc, err := nats.Connect(natsURL, nats.Name("auth-service"))
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to connect to NATS")
	}
	defer nc.Close()

	log.Info().Msg("Connected to NATS")

	// Connect to Redis
	redisAddr := os.Getenv("REDIS_ADDR")
	if redisAddr == "" {
		redisAddr = "localhost:6379"
	}
	redisClient := redis.NewClient(&redis.Options{
		Addr: redisAddr,
	})
	defer redisClient.Close()

	// Initialize Auth Components
	// Keys should come from env vars
	signingKey := []byte(os.Getenv("JWT_SIGNING_KEY"))
	if len(signingKey) == 0 {
		signingKey = []byte("default-signing-key-do-not-use-in-prod")
	}
	encryptionKey := []byte(os.Getenv("JWT_ENCRYPTION_KEY"))
	if len(encryptionKey) != 32 {
		// Fallback for dev/test if not set or invalid length
		// In prod, this should be a fatal error
		encryptionKey = []byte("01234567890123456789012345678901")
	}

	tokenManager, err := auth.NewTokenManager(signingKey, encryptionKey)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create TokenManager")
	}

	passwordHasher := auth.NewPasswordHasher()
	sessionManager := auth.NewSessionManager(redisClient)
	rateLimiter := auth.NewRateLimiter(redisClient)

	// Initialize Handler
	// Explicitly cast to interfaces to ensure compliance, though Go does this implicitly
	handler := NewAuthHandler(nc, tokenManager, passwordHasher, sessionManager, rateLimiter)

	// Subscribe to Login
	_, err = nc.Subscribe("auth.login", func(msg *nats.Msg) {
		// Handle in a goroutine or directly?
		// For high throughput, maybe a worker pool, but for now direct.
		// We need a context, maybe with timeout.
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := handler.HandleLogin(ctx, msg); err != nil {
			log.Error().Err(err).Msg("Failed to handle login")
		}
	})
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to subscribe to auth.login")
	}

	log.Info().Msg("Auth Service Started")

	// Wait for shutdown signal
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	<-sig

	log.Info().Msg("Shutting down...")
}
