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

	"tw-backend/internal/analytics"
	"tw-backend/internal/auth"
	"tw-backend/internal/ecosystem"
	"tw-backend/internal/events"
	"tw-backend/internal/repository"
	"tw-backend/internal/storage"
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
			OwnerID:       uuid.MustParse("00000000-0000-0000-0000-000000000001"), // System owned
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

	// 5.5. MinIO Connection (L3: Snapshot Storage)
	minioEndpoint := os.Getenv("MINIO_ENDPOINT")
	if minioEndpoint == "" {
		minioEndpoint = "localhost:9000"
	}
	minioAccessKey := os.Getenv("MINIO_ACCESS_KEY")
	if minioAccessKey == "" {
		minioAccessKey = "admin"
	}
	minioSecretKey := os.Getenv("MINIO_SECRET_KEY")
	if minioSecretKey == "" {
		minioSecretKey = "password123"
	}

	snapshotStore, err := storage.NewSnapshotStore(
		minioEndpoint,
		minioAccessKey,
		minioSecretKey,
		"world-snapshots",
		false, // useSSL
	)
	if err != nil {
		log.Warn().Err(err).Msg("MinIO unavailable - heightmap snapshots disabled")
	} else {
		// Ensure bucket exists
		if err := snapshotStore.EnsureBucket(ctx); err != nil {
			log.Warn().Err(err).Msg("Failed to create MinIO bucket")
		} else {
			runner.SetSnapshotStore(snapshotStore)
			log.Info().Str("endpoint", minioEndpoint).Msg("Connected to MinIO snapshot storage")
		}
	}

	// 5.6. Crash Recovery Intergration (L4)
	saveRepo := repository.NewPostgresSaveRepository(db)

	eventConsumer, err := events.NewEventConsumer(natsURL)
	var recoveryService *ecosystem.RecoveryService

	if err != nil {
		log.Warn().Err(err).Msg("Event consumer unavailable - crash recovery disabled")
	} else {
		defer eventConsumer.Close()
		if snapshotStore != nil {
			recoveryService = ecosystem.NewRecoveryService(snapshotStore, eventConsumer, saveRepo)
			runner.SetRecoveryService(recoveryService)
			log.Info().Msg("Crash recovery service initialized")
		}
	}

	// 5.7. Analytics/Metrics Integration (Phase 5)
	analyticsURL := os.Getenv("ANALYTICS_URL")
	if analyticsURL != "" {
		analyticsService, err := analytics.NewService(analyticsURL)
		if err != nil {
			log.Warn().Err(err).Msg("Analytics service unavailable - metrics collection disabled")
		} else {
			runner.SetMetricsCollector(analyticsService)
			log.Info().Msg("Connected to TimescaleDB analytics service")
			defer analyticsService.Close()
		}
	} else {
		log.Info().Msg("ANALYTICS_URL not set - metrics collection disabled")
	}

	// 6. Initialize State & Recovery
	log.Info().Msg("Initializing simulation state...")
	runner.InitializePopulationSimulator(seed)

	// Determine start year (prefer recovery > DB > 0)
	startYear := runner.GetCurrentYear()

	// Attempt Recovery
	if recoveryService != nil {
		log.Info().Msg("Checking for recovery save...")
		if res, err := recoveryService.RecoverWorld(ctx, world.ID); err == nil {
			if res.CurrentYear > startYear {
				startYear = res.CurrentYear
				geology.SphereHeightmap = res.Heightmap
				log.Info().
					Int64("year", startYear).
					Int("events", res.EventsReplayed).
					Dur("duration", res.RecoveryDuration).
					Msg("State recovered from snapshot + event replay")
			} else {
				log.Info().Int64("recovery_year", res.CurrentYear).Int64("db_year", startYear).Msg("Recovery state is older than DB, ignoring")
			}
		} else if err != repository.ErrSaveNotFound {
			log.Warn().Err(err).Msg("Failed to recover world")
		} else {
			log.Info().Msg("No recovery save found")
		}
	}

	if startYear == 0 {
		log.Info().Msg("Initializing fresh geology...")
		geology.InitializeGeology(0) // Default resolution calculation
	} else if geology.SphereHeightmap == nil {
		// We have a year but no heightmap (DB only, no snapshot/recovery)
		// We must regenerate or load from somewhere else.
		// For now, regenerating is the fallback (though it loses terrain changes not in DB)
		log.Info().Int64("year", startYear).Msg("Resuming from DB state (Geology re-generation)")
		geology.InitializeGeology(0)
	}

	// 7. Start Simulation
	log.Info().Int64("start_year", startYear).Msg("Starting simulation loop")

	// Handle shutdown gracefully
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	go func() {
		if err := runner.Start(startYear); err != nil {
			log.Fatal().Err(err).Msg("Simulation runner failed")
		}
	}()
	log.Info().Msg("Simulation running. Press Ctrl+C to stop.")

	// 7. Wait for shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	log.Info().Msg("Shutting down...")
	runner.Stop()
	log.Info().Msg("Shutdown complete")
}
