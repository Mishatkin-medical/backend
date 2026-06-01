package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"mishatkin/auth-service/internal/config"
	"mishatkin/auth-service/internal/delivery/http"
	"mishatkin/auth-service/internal/domain/usecase"
	"mishatkin/auth-service/internal/repository/postgres"
	"mishatkin/auth-service/internal/repository/redis"
	"mishatkin/shared/pkg/logger"
	"mishatkin/shared/pkg/postgres"
)

func main() {
	// Initialize logger
	log := logger.New("auth-service")
	log.Info().Msg("Starting auth-service...")

	// Load configuration
	cfg := config.Load()

	// Set log level
	log.SetLevel(cfg.LogLevel)

	// Initialize context
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Initialize PostgreSQL
	log.Info().Msg("Connecting to PostgreSQL...")
	db, err := postgres.NewPool(ctx, &postgres.Config{
		URL: cfg.DatabaseURL,
		MaxConns: 10,
		MinConns: 2,
	})
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to connect to PostgreSQL")
	}
	defer db.Close()

	// Run migrations
	log.Info().Msg("Running migrations...")
	if err := postgres.RunMigrations(ctx, db, "migrations"); err != nil {
		log.Fatal().Err(err).Msg("Failed to run migrations")
	}

	// Initialize Redis
	log.Info().Msg("Connecting to Redis...")
	redisClient, err := redis.NewClient(&redis.Config{
		URL: cfg.RedisURL,
	})
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to connect to Redis")
	}
	defer redisClient.Close()

	// Initialize repositories
	userRepo := postgres.NewUserRepository(db)
	tokenRepo := postgres.NewTokenRepository(db)
	tokenCache := redis.NewTokenCache(redisClient)

	// Initialize use cases
	authUseCase := usecase.NewAuthUseCase(
		userRepo,
		tokenRepo,
		tokenCache,
		cfg.JWTSecret,
		time.Duration(cfg.AccessTTL),
		time.Duration(cfg.RefreshTTL),
		cfg.BcryptCost,
	)

	// Initialize HTTP handler
	handler := http.NewHandler(authUseCase, log)

	// Setup router
	router := http.SetupRouter(handler, log)

	// Create HTTP server
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Port),
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in goroutine
	go func() {
		log.Info().Int("port", cfg.Port).Msg("HTTP server starting...")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal().Err(err).Msg("HTTP server error")
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info().Msg("Shutting down server...")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Error().Err(err).Msg("Server forced to shutdown")
	}

	log.Info().Msg("Server exited")
}

