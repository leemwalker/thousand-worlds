package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"

	"mud-platform-backend/cmd/game-server/api"
	"mud-platform-backend/cmd/game-server/websocket"
	"mud-platform-backend/internal/ai/ollama"
	"mud-platform-backend/internal/auth"
	"mud-platform-backend/internal/game/entry"
	"mud-platform-backend/internal/game/processor"
	"mud-platform-backend/internal/lobby"
	"mud-platform-backend/internal/metrics"
	"mud-platform-backend/internal/repository"
	"mud-platform-backend/internal/world/interview"
)

func main() {
	log.Println("Starting Thousand Worlds Game Server...")

	// Create context for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Load JWT secret from environment
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		// Generate a secret key if not provided (dev only)
		log.Println("WARNING: JWT_SECRET not set, generating random key (dev mode)")
		secretKey, err := auth.GenerateSecretKey()
		if err != nil {
			log.Fatal("Failed to generate secret key:", err)
		}
		jwtSecret = string(secretKey)
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
	interviewRepo := interview.NewRepository(db)

	// Initialize pgxpool for WorldRepository
	poolConfig, err := pgxpool.ParseConfig(dbDSN)
	if err != nil {
		log.Fatal("Failed to parse database URL for pgxpool:", err)
	}
	dbPool, err := pgxpool.NewWithConfig(context.Background(), poolConfig)
	if err != nil {
		log.Fatal("Failed to connect to database with pgxpool:", err)
	}
	defer dbPool.Close()
	worldRepo := repository.NewPostgresWorldRepository(dbPool)

	// Initialize services
	authConfig := &auth.Config{
		SecretKey:       []byte(jwtSecret),
		TokenExpiration: 24 * time.Hour,
	}
	authService := auth.NewService(authConfig, authRepo)
	lobbyService := lobby.NewService(authRepo)
	entryService := entry.NewService(interviewRepo)

	ollamaClient := ollama.NewClient(os.Getenv("OLLAMA_HOST"), "llama3.1:8b") // 8B model with increased container memory
	interviewService := interview.NewServiceWithRepository(ollamaClient, interviewRepo)

	// Initialize session manager and rate limiter
	var sessionManager *auth.SessionManager
	var rateLimiter *auth.RateLimiter
	if redisClient != nil {
		sessionManager = auth.NewSessionManager(redisClient)
		rateLimiter = auth.NewRateLimiter(redisClient)
	}

	// Initialize look service for lobby commands
	lookService := lobby.NewLookService(authRepo, worldRepo, interviewRepo)

	// Initialize game processor
	gameProcessor := processor.NewGameProcessor(authRepo, worldRepo, lookService)

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
	sessionHandler := api.NewSessionHandler(authRepo)
	entryHandler := api.NewEntryHandler(entryService)
	worldHandler := api.NewWorldHandler(worldRepo)
	wsHandler := websocket.NewHandler(hub, lobbyService, authRepo, descGen)

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

	// CORS configuration
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
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
