package main

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/nats-io/nats.go"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"tw-backend/internal/ai/gateway"
)

func main() {
	// Setup logging
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})

	natsURL := os.Getenv("NATS_URL")
	if natsURL == "" {
		natsURL = "nats://localhost:4222"
	}

	log.Info().Str("url", natsURL).Msg("Connecting to NATS")
	nc, err := nats.Connect(natsURL)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to connect to NATS")
	}
	defer nc.Close()

	ollamaClient := gateway.NewOllamaClient()

	// Start Worker
	gateway.StartWorker(nc, ollamaClient)

	// Start Listener
	if err := gateway.StartListener(nc); err != nil {
		log.Fatal().Err(err).Msg("Failed to start listener")
	}

	log.Info().Msg("AI Gateway Service Started")

	// Wait for shutdown signal
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	<-sig

	log.Info().Msg("Shutting down AI Gateway Service")
}
