package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/nats-io/nats.go"
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

	// Initialize Handler
	handler := NewAuthHandler(nc)

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
