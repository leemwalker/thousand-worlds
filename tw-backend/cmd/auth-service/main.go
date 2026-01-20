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

	// Configuration
	natsURL := os.Getenv("NATS_URL")
	if natsURL == "" {
		natsURL = nats.DefaultURL
	}

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Warn().Msg("DATABASE_URL not set, using default")
		dbURL = "postgresql://admin:password@localhost:5432/tw_core?sslmode=disable"
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

	// Connect to Database
	db, err := auth.ConnectDB(dbURL)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to connect to database")
	}
	defer db.Close()
	log.Info().Msg("Connected to Database")

	// Initialize Components
	repo := auth.NewPostgresRepository(db)

	signingKey := []byte(os.Getenv("JWT_SIGNING_KEY"))
	if len(signingKey) == 0 {
		signingKey = []byte("default-signing-key-do-not-use-in-prod")
	}

	authConfig := &auth.Config{
		SecretKey:       signingKey,
		TokenExpiration: 24 * time.Hour, // Could be env var
	}

	authService := auth.NewService(authConfig, repo)
	rateLimiter := auth.NewRateLimiter(redisClient)

	// Initialize Handler
	handler := NewAuthHandler(nc, authService, rateLimiter)

	// Subscribe to Login
	_, err = nc.Subscribe("auth.login", func(msg *nats.Msg) {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := handler.HandleLogin(ctx, msg); err != nil {
			log.Error().Err(err).Msg("Failed to handle login")
		}
	})
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to subscribe to auth.login")
	}

	// Subscribe to Register
	_, err = nc.Subscribe("auth.register", func(msg *nats.Msg) {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := handler.HandleRegister(ctx, msg); err != nil {
			log.Error().Err(err).Msg("Failed to handle register")
		}
	})
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to subscribe to auth.register")
	}

	log.Info().Msg("Auth Service Started")

	// Wait for shutdown signal
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	<-sig

	log.Info().Msg("Shutting down...")
}
