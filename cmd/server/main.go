package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/niaga-labs/niaga-labs-pet-lib-common/auth"
	"github.com/niaga-labs/niaga-labs-pet-lib-common/database"
	"github.com/niaga-labs/niaga-labs-pet-lib-common/health"
	"github.com/niaga-labs/niaga-labs-pet-lib-common/kafka"
	"github.com/niaga-labs/niaga-labs-pet-lib-common/logger"
	"github.com/niaga-labs/niaga-labs-pet-lib-common/middleware"
	"github.com/niaga-labs/niaga-labs-pet-service-review/internal/application"
	"github.com/niaga-labs/niaga-labs-pet-service-review/internal/config"
	"github.com/niaga-labs/niaga-labs-pet-service-review/internal/handler"
	"github.com/niaga-labs/niaga-labs-pet-service-review/internal/repository"
	"go.uber.org/zap"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to load config: %v\n", err)
		os.Exit(1)
	}

	// Initialize logger
	log, err := logger.NewNamed(cfg.AppEnv, "service-review")
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to create logger: %v\n", err)
		os.Exit(1)
	}
	defer func() { _ = log.Sync() }()

	log.Info("starting service-review", zap.String("port", cfg.Port))

	// Connect to database
	dbConfig := database.PostgresConfig{
		Host:     cfg.DBConfig.Host,
		Port:     cfg.DBConfig.Port,
		User:     cfg.DBConfig.User,
		Password: cfg.DBConfig.Password,
		DBName:   cfg.DBConfig.DBName,
		SSLMode:  cfg.DBConfig.SSLMode,
	}
	db, err := database.Connect(dbConfig, log)
	if err != nil {
		log.Fatal("failed to connect to database", zap.Error(err))
	}

	// Run database migrations.
	//
	// KPD-58: this used to AutoMigrate ReviewModel in development and call
	// RunMigrations everywhere else -- but the repository had no migrations/
	// directory, so every non-development boot died here. 001_create_reviews now
	// owns the schema in every environment, including development, so dev and
	// production can no longer drift apart.
	dbURL := dbConfig.DatabaseURL()
	if err := database.RunMigrations(dbURL, "migrations", log); err != nil {
		log.Fatal("failed to run migrations", zap.Error(err))
	}

	// Initialize JWT manager
	jwtManager := auth.NewJWTManager(
		cfg.JWTConfig.Secret,
		15*time.Minute,
		7*24*time.Hour,
	)

	// Initialize Kafka producer
	kafkaProducer := kafka.NewProducer(cfg.KafkaConfig.Brokers, log)
	defer func() { _ = kafkaProducer.Close() }()

	// Initialize repository and service
	reviewRepo := repository.NewGormReviewRepository(db)
	reviewService := application.NewReviewService(reviewRepo, kafkaProducer, log)

	// Initialize handler
	reviewHandler := handler.NewReviewHandler(reviewService)

	// Setup Gin router
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()

	router.Use(middleware.RecoveryMiddleware(log))
	router.Use(middleware.LoggerMiddleware(log))
	router.Use(middleware.RequestIDMiddleware())
	router.Use(middleware.CORSMiddleware())
	router.Use(middleware.SecurityHeadersMiddleware())

	// Register health check
	healthHandler := health.NewHandler(db, "service-review")
	healthHandler.RegisterRoutes(router)

	// Register routes
	reviewHandler.RegisterRoutes(&router.RouterGroup, jwtManager)

	// Create HTTP server
	srv := &http.Server{
		Addr:         cfg.Port,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server
	go func() {
		log.Info("HTTP server starting", zap.String("addr", cfg.Port))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal("HTTP server error", zap.Error(err))
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info("shutting down service-review...")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Error("HTTP server forced shutdown", zap.Error(err))
	}

	log.Info("service-review stopped")
}
