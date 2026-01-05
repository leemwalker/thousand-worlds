package main

import (
	"context"
	"math"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	agones "agones.dev/agones/sdks/go"

	"tw-backend/internal/auth"
	"tw-backend/internal/ecosystem"
	"tw-backend/internal/events"
	"tw-backend/internal/repository"
	"tw-backend/internal/worldgen/astronomy"
)

// Main entrypoint for the Kubernetes/Agones-compatible simulation service
func main() {
	// Setup logging
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: time.RFC3339})
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	log.Info().Msg("Starting Core Physics Service...")

	// 0. Agones SDK Integration
	// Only initialize if we detect Agones environment
	if os.Getenv("AGONES_SDK_HTTP_PORT") != "" || os.Getenv("AGONES_SDK_GRPC_PORT") != "" {
		log.Info().Msg("Detected Agones environment. Initializing SDK...")
		s, err := agones.NewSDK()
		if err != nil {
			log.Fatal().Err(err).Msg("Failed to initialize Agones SDK")
		}

		// Mark as Ready
		if err := s.Ready(); err != nil {
			log.Error().Err(err).Msg("Failed to send Ready signal to Agones")
		} else {
			log.Info().Msg("Agones SDK Ready")
		}

		// Start Health Ping
		go func() {
			ticker := time.NewTicker(2 * time.Second)
			for range ticker.C {
				if err := s.Health(); err != nil {
					log.Warn().Err(err).Msg("Failed to send Health ping")
				}
			}
		}()
	} else {
		log.Info().Msg("No Agones environment detected (Local/Docker mode)")
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 1. Database Connection
	dbDSN := os.Getenv("DATABASE_URL")
	if dbDSN == "" {
		log.Fatal().Msg("DATABASE_URL must be set")
	}

	// Connect using stdlib (for legacy/simple repos)
	db, err := auth.ConnectDB(dbDSN)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to connect to DB (stdlib)")
	}
	defer db.Close()

	// Connect using pgxpool (for WorldRepo)
	poolConfig, err := pgxpool.ParseConfig(dbDSN)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to parse DATABASE_URL")
	}
	poolConfig.MaxConns = 10
	poolConfig.MinConns = 2
	dbPool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to connect to database POOL")
	}
	defer dbPool.Close()
	log.Info().Msg("Connected to Database")

	worldRepo := repository.NewPostgresWorldRepository(dbPool)
	snapshotRepo := ecosystem.NewSimulationSnapshotRepository(db)
	stateRepo := ecosystem.NewRunnerStateRepository(db)

	// 2. NATS Connection
	natsURL := os.Getenv("NATS_URL")
	if natsURL == "" {
		natsURL = "nats://localhost:4222"
	}

	var pub events.Publisher
	if os.Getenv("NATS_DISABLED") == "true" {
		log.Warn().Msg("NATS Disabled - Using NoOp Publisher")
		pub = events.NewNoOpPublisher()
	} else {
		// Stream name is hardcoded in publisher config
		natsPub, err := events.NewNATSPublisher(natsURL)
		if err != nil {
			log.Fatal().Err(err).Msg("Failed to connect to NATS")
		}
		defer natsPub.Close()
		pub = natsPub
		log.Info().Str("url", natsURL).Msg("Connected to NATS JetStream")
	}

	// 3. World Configuration
	worldIDStr := os.Getenv("WORLD_ID")
	if worldIDStr == "" {
		log.Fatal().Msg("WORLD_ID environment variable is required")
	}
	worldID, err := uuid.Parse(worldIDStr)
	if err != nil {
		log.Fatal().Err(err).Msg("Invalid WORLD_ID format")
	}

	world, err := worldRepo.GetWorld(ctx, worldID)
	if err != nil {
		log.Warn().Err(err).Msg("World not found, attempting to create default world")
		// Create default world
		defaultRadius := astronomy.EarthRadiusMeters
		defaultCircumference := 2 * math.Pi * defaultRadius
		val := 100000000.0 // 100M bounds

		newWorld := &repository.World{
			ID:            worldID,
			Name:          "Physics Simulation World",
			OwnerID:       uuid.Nil, // System owned
			Shape:         repository.WorldShapeSphere,
			Radius:        &defaultRadius,
			Circumference: &defaultCircumference,
			BoundsMin:     &repository.Vector3{X: -val, Y: -val, Z: -val},
			BoundsMax:     &repository.Vector3{X: val, Y: val, Z: val},
			Metadata:      map[string]interface{}{"seed": 12345},
			CreatedAt:     time.Now(),
		}
		if createErr := worldRepo.CreateWorld(ctx, newWorld); createErr != nil {
			log.Fatal().Err(createErr).Msg("Failed to create default world")
		}
		world = newWorld
		log.Info().Str("world_id", world.ID.String()).Msg("Created default world")
	}
	log.Info().Str("name", world.Name).Msg("Loaded World")

	// Extract seed from metadata
	seed := int64(0)
	if val, ok := world.Metadata["seed"]; ok {
		switch v := val.(type) {
		case float64:
			seed = int64(v)
		case int64:
			seed = v
		case int:
			seed = int64(v)
		}
	} else {
		log.Warn().Msg("No seed in metadata, using default 0")
	}

	circumference := 2 * math.Pi * astronomy.EarthRadiusMeters
	if world.Circumference != nil {
		circumference = *world.Circumference
	}

	// 4. Initialize Core Systems
	geology := ecosystem.NewWorldGeology(world.ID, seed, circumference)
	geology.EventPublisher = pub // Inject Publisher

	// 5. Initialize Simulation Runner
	simConfig := ecosystem.DefaultConfig(world.ID)
	// Fast Update: 5 ticks/sec, 10 years/tick = 50 years/sec
	simConfig.TickInterval = 200 * time.Millisecond
	simConfig.Speed = ecosystem.SpeedNormal

	// Create Runner with real Repos
	runner := ecosystem.NewSimulationRunner(simConfig, snapshotRepo, stateRepo)

	// Inject Geology
	runner.SetGeology(geology)

	// Initialize subsystems and load/init state
	log.Info().Msg("Initializing simulation state...")
	runner.InitializePopulationSimulator(seed)

	if runner.GetCurrentYear() == 0 {
		log.Info().Msg("Initializing fresh geology...")
		geology.InitializeGeology(0) // Default resolution calculation
	} else {
		log.Info().Int64("year", runner.GetCurrentYear()).Msg("Simulation resumed. (Geology state re-generation TODO)")
		// For now, re-generate to ensure non-nil maps
		geology.InitializeGeology(0)
	}

	// 6. Start Simulation
	startYear := runner.GetCurrentYear()
	if err := runner.Start(startYear); err != nil {
		log.Fatal().Err(err).Msg("Failed to start simulation")
	}

	log.Info().Msg("Simulation running. Press Ctrl+C to stop.")

	// 7. Wait for shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	log.Info().Msg("Shutting down...")
	runner.Stop()
	log.Info().Msg("Shutdown complete")
}
