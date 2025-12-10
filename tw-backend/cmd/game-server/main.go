package main

import (
	"context"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"

	"tw-backend/cmd/game-server/api"
	"tw-backend/cmd/game-server/websocket"
	"tw-backend/internal/ai/ollama"
	"tw-backend/internal/auth"
	"tw-backend/internal/game/entry"
	"tw-backend/internal/game/processor"
	"tw-backend/internal/game/services/entity"
	"tw-backend/internal/game/services/look"
	"tw-backend/internal/lobby"
	"tw-backend/internal/metrics"
	"tw-backend/internal/player"
	"tw-backend/internal/repository"
	"tw-backend/internal/skills"
	"tw-backend/internal/world/interview"
	"tw-backend/internal/worldgen/weather"
)

func main() {
	// Setup logging
	logFile, err := os.OpenFile("server.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatal("Failed to open log file:", err)
	}
	defer logFile.Close()

	// Create multi-writer for both stdout and file
	mw := io.MultiWriter(os.Stdout, logFile)
	log.SetOutput(mw)

	log.Println("Starting Thousand Worlds Game Server...")

	// Create context for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Load JWT secret from environment - REQUIRED
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		log.Fatal("FATAL: JWT_SECRET environment variable must be set. Generate with: openssl rand -hex 32")
	}
	if len(jwtSecret) < 32 {
		log.Fatal("FATAL: JWT_SECRET must be at least 32 characters long for security")
	}

	// Database connection
	dbDSN := os.Getenv("DATABASE_URL")
	if dbDSN == "" {
		dbDSN = "postgres://postgres:postgres@127.0.0.1:5432/thousand_worlds?sslmode=disable"
	}

	log.Printf("Connecting to database...")
	db, err := auth.ConnectDB(dbDSN)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer db.Close()

	// Redis connection for sessions and rate limiting
	redisAddr := os.Getenv("REDIS_ADDR")
	if redisAddr == "" {
		redisAddr = "localhost:6379"
	}

	log.Printf("Connecting to Redis...")
	redisClient := redis.NewClient(&redis.Options{
		Addr: redisAddr,
		DB:   0,
	})

	// Verify Redis connection
	if err := redisClient.Ping(ctx).Err(); err != nil {
		log.Printf("WARNING: Failed to connect to Redis: %v", err)
		log.Printf("Session management and rate limiting will be disabled")
		redisClient = nil
	}

	// Initialize repositories
	authRepo := auth.NewPostgresRepository(db)

	// Initialize pgxpool for WorldRepository and InterviewRepository
	poolConfig, err := pgxpool.ParseConfig(dbDSN)
	if err != nil {
		log.Fatal("Failed to parse database URL for pgxpool:", err)
	}
	dbPool, err := pgxpool.NewWithConfig(context.Background(), poolConfig)
	if err != nil {
		log.Fatal("Failed to connect to database with pgxpool:", err)
	}
	// defer dbPool.Close() // Defer close in main function scope

	worldRepo := repository.NewPostgresWorldRepository(dbPool)
	interviewRepo := interview.NewRepository(dbPool)

	// Initialize services
	authConfig := &auth.Config{
		SecretKey:       []byte(jwtSecret),
		TokenExpiration: 24 * time.Hour,
	}
	authService := auth.NewService(authConfig, authRepo)
	lobbyService := lobby.NewService(authRepo)
	entryService := entry.NewService(interviewRepo)

	ollamaClient := ollama.NewClient(os.Getenv("OLLAMA_HOST"), "llama3.2:3b") // 3B model for faster response times
	interviewService := interview.NewServiceWithRepository(ollamaClient, interviewRepo, worldRepo)

	// Initialize session manager and rate limiter
	var sessionManager *auth.SessionManager
	var rateLimiter *auth.RateLimiter
	if redisClient != nil {
		sessionManager = auth.NewSessionManager(redisClient)
		rateLimiter = auth.NewRateLimiter(redisClient)
	}

	// Initialize weather service
	weatherRepo := weather.NewPostgresRepository(db)
	weatherService := weather.NewService(weatherRepo)

	// Initialize Entity Service
	entityService := entity.NewService()

	// Initialize look service for lobby commands
	lookService := look.NewLookService(worldRepo, weatherService, entityService, interviewRepo)

	// Initialize spatial service
	spatialService := player.NewSpatialService(authRepo, worldRepo)

	// Initialize skills repository (needed for map service)
	skillsRepo := skills.NewRepository(dbPool)

	// Initialize game processor
	gameProcessor := processor.NewGameProcessor(authRepo, worldRepo, lookService, entityService, interviewService, spatialService, weatherService, skillsRepo)

	// Create and start the Hub
	hub := websocket.NewHub(gameProcessor)
	gameProcessor.SetHub(hub)
	go hub.Run(ctx)

	// Create health check handler
	healthHandler := api.NewHealthHandler()

	// Update connected users count periodically
	go func() {
		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				count := int64(hub.GetClientCount())
				healthHandler.SetConnectedUsers(count)
			}
		}
	}()

	// Initialize DescriptionGenerator
	descGen := lobby.NewDescriptionGenerator(worldRepo, authRepo)

	// Initialize handlers
	authHandler := api.NewAuthHandler(authService, sessionManager, rateLimiter)
	interviewHandler := api.NewInterviewHandler(interviewService)
	sessionHandler := api.NewSessionHandler(authRepo, lookService)
	entryHandler := api.NewEntryHandler(entryService)
	worldHandler := api.NewWorldHandler(worldRepo)
	wsHandler := websocket.NewHandler(hub, lobbyService, authRepo, descGen)

	// Skills service and handler
	skillsService := skills.NewService(skillsRepo)
	skillsHandler := api.NewSkillsHandler(skillsService)

	// Router setup
	r := chi.NewRouter()

	// Request logging
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	// Custom metrics middleware - but NOT for WebSocket (it breaks hijacking)
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Skip metrics wrapping for WebSocket to preserve hijacker
			if r.URL.Path == "/api/game/ws" {
				next.ServeHTTP(w, r)
				return
			}
			metrics.Middleware(next).ServeHTTP(w, r)
		})
	})

	// CORS configuration - Load allowed origins from environment
	// TODO_SECURITY: Update CORS_ALLOWED_ORIGINS when you have a production domain
	// Current default is for development only
	corsOrigins := os.Getenv("CORS_ALLOWED_ORIGINS")
	if corsOrigins == "" {
		// Default for local development
		corsOrigins = "http://localhost:5173"
		log.Println("INFO: Using default CORS origins for development:", corsOrigins)
	}

	// Split comma-separated origins
	allowedOrigins := strings.Split(corsOrigins, ",")
	for i := range allowedOrigins {
		allowedOrigins[i] = strings.TrimSpace(allowedOrigins[i])
	}

	// Validate CORS configuration security
	for _, origin := range allowedOrigins {
		if origin == "*" {
			log.Fatal("FATAL: Wildcard (*) CORS origin is not allowed for security. Specify exact origins.")
		}
	}

	log.Printf("INFO: CORS allowed origins: %v", allowedOrigins)

	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   allowedOrigins,
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	// Metrics endpoint
	r.Handle("/metrics", metrics.Handler())

	// API Routes
	r.Route("/api", func(r chi.Router) {
		// Public routes (no auth required)
		r.Post("/auth/register", authHandler.Register)
		r.Post("/auth/login", authHandler.Login)

		// Protected routes (auth required)
		r.Group(func(r chi.Router) {
			r.Use(api.AuthMiddleware(authService))

			r.Get("/auth/me", authHandler.GetMe)
			r.Post("/auth/logout", authHandler.Logout)

			// World Interview routes
			r.Post("/world/interview/start", interviewHandler.StartInterview)
			r.Post("/world/interview/message", interviewHandler.ProcessMessage)
			r.Get("/world/interview/active", interviewHandler.GetActiveInterview)
			r.Post("/world/interview/finalize", interviewHandler.FinalizeInterview)

			// Game Session routes
			r.Post("/game/characters", sessionHandler.CreateCharacter)
			r.Get("/game/characters", sessionHandler.GetCharacters)
			r.Post("/game/join", sessionHandler.JoinGame)
			r.Get("/game/entry-options", entryHandler.GetEntryOptions)
			r.Get("/game/worlds", worldHandler.ListWorlds)

			// Skills
			r.Get("/game/skills", skillsHandler.HandleGetSkills)

			// WebSocket endpoint
			r.Get("/game/ws", wsHandler.ServeHTTP)
		})
	})

	// Health check
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// Start HTTP server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	server := &http.Server{
		Addr:         ":" + port,
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Graceful shutdown
	go func() {
		sigint := make(chan os.Signal, 1)
		signal.Notify(sigint, os.Interrupt, syscall.SIGTERM)
		<-sigint

		log.Println("Shutting down server...")
		cancel()

		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer shutdownCancel()

		if err := server.Shutdown(shutdownCtx); err != nil {
			log.Printf("Server shutdown error: %v", err)
		}
	}()

	log.Printf("Server listening on port %s", port)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatal("Server error:", err)
	}

	log.Println("Server stopped")
}
