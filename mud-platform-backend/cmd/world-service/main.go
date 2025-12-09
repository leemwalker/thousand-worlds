package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/nats-io/nats.go"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"mud-platform-backend/internal/auth"
	"mud-platform-backend/internal/eventstore"
	"mud-platform-backend/internal/spatial"
	"mud-platform-backend/internal/world"
	"mud-platform-backend/internal/worldgen/weather"
)

func main() {
	// Configure logging
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.RFC3339})

	log.Info().Msg("Starting World Service...")

	// Load configuration from environment
	config := loadConfig()

	// Connect to PostgreSQL
	log.Info().Str("db_url", maskPassword(config.DatabaseURL)).Msg("Connecting to PostgreSQL")
	dbPool, err := pgxpool.New(context.Background(), config.DatabaseURL)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to connect to database")
	}
	defer dbPool.Close()

	// Verify database connection
	if err := dbPool.Ping(context.Background()); err != nil {
		log.Fatal().Err(err).Msg("Failed to ping database")
	}
	log.Info().Msg("Database connection established")

	// Initialize event store
	eventStore := eventstore.NewPostgresEventStore(dbPool)

	// Connect to NATS
	log.Info().Str("nats_url", config.NATSURL).Msg("Connecting to NATS")
	nc, err := nats.Connect(config.NATSURL)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to connect to NATS")
	}
	defer nc.Close()
	log.Info().Msg("NATS connection established")

	// Create NATS publisher wrapper for ticker manager
	natsPublisher := &NATSPublisherWrapper{nc: nc}

	// Initialize weather service
	// We need sql.DB for weather repo
	sqlDB, err := auth.ConnectDB(config.DatabaseURL)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to connect to database (sql.DB)")
	}
	defer sqlDB.Close()

	weatherRepo := weather.NewPostgresRepository(sqlDB)
	weatherService := weather.NewService(weatherRepo)

	// Create NATS area broadcaster
	areaBroadcaster := &NATSAreaBroadcaster{nc: nc}

	// Initialize world registry and ticker manager
	registry := world.NewRegistry()
	tickerManager := world.NewTickerManager(registry, eventStore, natsPublisher, weatherService, areaBroadcaster)

	log.Info().Msg("World Service initialized successfully")

	// TODO: Set up NATS subscriptions for world commands
	// - world.create -> spawn new world
	// - world.pause -> pause ticker
	// - world.resume -> resume ticker
	// - world.query -> get world status

	// Set up graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	log.Info().Msg("World Service is running. Press Ctrl+C to stop.")

	// Wait for shutdown signal
	sig := <-sigChan
	log.Info().Str("signal", sig.String()).Msg("Received shutdown signal")

	// Graceful shutdown
	log.Info().Msg("Stopping all world tickers...")
	tickerManager.StopAll()

	log.Info().Msg("Closing NATS connection...")
	nc.Close()

	log.Info().Msg("Closing database connection...")
	dbPool.Close()

	log.Info().Msg("World Service stopped gracefully")
}

// NATSPublisherWrapper wraps NATS connection for world.NATSPublisher interface
type NATSPublisherWrapper struct {
	nc *nats.Conn
}

func (n *NATSPublisherWrapper) Publish(subject string, data []byte) error {
	return n.nc.Publish(subject, data)
}

type Config struct {
	DatabaseURL string
	NATSURL     string
}

func loadConfig() Config {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://admin:password123@localhost:5432/mud_core?sslmode=disable"
	}

	natsURL := os.Getenv("NATS_URL")
	if natsURL == "" {
		natsURL = "nats://localhost:4222"
	}

	return Config{
		DatabaseURL: dbURL,
		NATSURL:     natsURL,
	}
}

// maskPassword masks the password in database URL for logging
// NATSAreaBroadcaster implements world.AreaBroadcaster via NATS
type NATSAreaBroadcaster struct {
	nc *nats.Conn
}

func (b *NATSAreaBroadcaster) BroadcastToArea(center spatial.Position, radius float64, msgType string, data interface{}) {
	// Publish to NATS for game-server to pick up
	// Subject: world.broadcast.area
	// Payload: { Center, Radius, Type, Data }
	// For now, just logging or simple publish
	// TODO: Define shared struct for this payload
}

func maskPassword(dbURL string) string {
	// Simple masking - in production use proper URL parsing
	return "postgres://admin:****@localhost:5432/mud_core"
}
