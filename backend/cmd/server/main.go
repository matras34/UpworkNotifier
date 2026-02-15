package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/joho/godotenv"
	"github.com/matras34/UpworkNotifier/internal/api"
	"github.com/matras34/UpworkNotifier/internal/api/handlers"
	"github.com/matras34/UpworkNotifier/internal/api/middleware"
	"github.com/matras34/UpworkNotifier/internal/auth"
	"github.com/matras34/UpworkNotifier/internal/cache"
	"github.com/matras34/UpworkNotifier/internal/config"
	"github.com/matras34/UpworkNotifier/internal/crypto"
	"github.com/matras34/UpworkNotifier/internal/db"
	sshpkg "github.com/matras34/UpworkNotifier/internal/ssh"
	wspkg "github.com/matras34/UpworkNotifier/internal/websocket"
	"github.com/matras34/UpworkNotifier/pkg/logger"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		fmt.Println("No .env file found, using environment variables")
	}

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		fmt.Printf("Failed to load configuration: %v\n", err)
		os.Exit(1)
	}

	// Initialize logger
	log := logger.New(logger.INFO)
	log.Info("Starting Web SSH Service")

	// Initialize database
	database, err := db.NewDatabase(
		cfg.GetDatabaseDSN(),
		cfg.Database.MaxConns,
		cfg.Database.MinConns,
	)
	if err != nil {
		log.Fatal("Failed to connect to database", map[string]interface{}{"error": err.Error()})
	}
	defer database.Close()
	log.Info("Database connected")

	// Initialize Redis cache
	redisCache, err := cache.NewCache(
		cfg.GetRedisAddr(),
		cfg.Redis.Password,
		cfg.Redis.DB,
	)
	if err != nil {
		log.Fatal("Failed to connect to Redis", map[string]interface{}{"error": err.Error()})
	}
	defer redisCache.Close()
	log.Info("Redis connected")

	// Initialize KMS
	kms, err := crypto.NewKMS(cfg.Security.EncryptionKey)
	if err != nil {
		log.Fatal("Failed to initialize KMS", map[string]interface{}{"error": err.Error()})
	}
	log.Info("KMS initialized")

	// Initialize JWT manager
	jwtManager := auth.NewJWTManager(cfg.Auth.JWTSecret, cfg.Auth.JWTExpiration)
	log.Info("JWT manager initialized")

	// Initialize SSH session manager
	sessionManager := sshpkg.NewSessionManager(
		cfg.SSH.MaxConcurrentSessions,
		cfg.SSH.SessionIdleTimeout,
		cfg.SSH.SessionMaxDuration,
	)
	log.Info("SSH session manager initialized")

	// Initialize WebSocket hub
	wsHub := wspkg.NewHub()
	go wsHub.Run()
	log.Info("WebSocket hub started")

	// Initialize handlers
	authHandler := handlers.NewAuthHandler(database, redisCache, jwtManager, log, cfg.Auth.BcryptCost)
	sshHandler := handlers.NewSSHHandler(database, redisCache, kms, log, sessionManager, wsHub)

	// Initialize middleware
	authMiddleware := middleware.NewAuthMiddleware(jwtManager)
	rateLimiter := middleware.NewRateLimiter(redisCache, cfg.Security.RateLimitPerMinute)
	loggerMiddleware := middleware.Logger(log)

	// Initialize router
	router := api.NewRouter(
		authHandler,
		sshHandler,
		authMiddleware,
		rateLimiter,
		loggerMiddleware,
		cfg.Security.AllowedOrigins,
	)

	// Create HTTP server
	srv := &http.Server{
		Addr:         ":" + cfg.Server.Port,
		Handler:      router,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
		IdleTimeout:  120 * time.Second,
	}

	// Start server in goroutine
	go func() {
		log.Info("Server starting", map[string]interface{}{"port": cfg.Server.Port})
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal("Server failed to start", map[string]interface{}{"error": err.Error()})
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info("Server shutting down...")

	// Graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), cfg.Server.ShutdownTimeout)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Error("Server forced to shutdown", map[string]interface{}{"error": err.Error()})
	}

	log.Info("Server stopped")
}
