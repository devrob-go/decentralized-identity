package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"did-manager/internal/handler"
	"did-manager/internal/repository"
	"did-manager/internal/services"
	"did-manager/pkg/blockchain"
	"did-manager/pkg/did"
	"did-manager/pkg/queue"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"github.com/rs/zerolog"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Printf("Warning: .env file not found")
	}

	// Initialize logger
	logger := zerolog.New(os.Stdout).With().Timestamp().Logger()
	zerolog.SetGlobalLevel(zerolog.InfoLevel)

	// Database connection
	db, err := connectDB()
	if err != nil {
		logger.Fatal().Err(err).Msg("Failed to connect to database")
	}
	defer db.Close()

	// Initialize repositories
	didRepo := repository.NewDIDRepository(db)
	queueRepo := repository.NewBlockchainJobRepository(db)

	// Initialize blockchain client
	blockchainClient, err := blockchain.NewEthereumClient(
		os.Getenv("ETHEREUM_RPC_URL"),
		os.Getenv("ETHEREUM_PRIVATE_KEY"),
		os.Getenv("ETHEREUM_CONTRACT_ADDRESS"),
	)
	if err != nil {
		logger.Warn().Err(err).Msg("Failed to initialize blockchain client, running in offline mode")
		blockchainClient = nil
	}
	defer func() {
		if blockchainClient != nil {
			blockchainClient.Close()
		}
	}()

	// Initialize NATS queue
	queueClient, err := queue.NewNATSQueue(os.Getenv("NATS_URL"))
	if err != nil {
		logger.Warn().Err(err).Msg("Failed to initialize NATS queue, running in local mode")
		queueClient = nil
	}
	defer func() {
		if queueClient != nil {
			queueClient.Close()
		}
	}()

	// Initialize DID generator
	didGen := did.NewGenerator()

	// Initialize services
	didService := services.NewDIDService(didRepo, queueRepo, didGen, blockchainClient, queueClient)

	// Initialize handlers
	didHandler := handler.NewDIDHandler(didService)

	// Setup Gin router
	router := gin.Default()

	// Add middleware
	router.Use(gin.Recovery())
	router.Use(gin.Logger())

	// Register routes
	didHandler.RegisterRoutes(router)

	// Start background worker for blockchain queue processing
	if blockchainClient != nil && queueClient != nil {
		go startBackgroundWorker(didService, logger)
	}

	// Start HTTP server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8082"
	}

	srv := &http.Server{
		Addr:    ":" + port,
		Handler: router,
	}

	// Start server in a goroutine
	go func() {
		logger.Info().Msgf("Starting DID Manager server on port %s", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal().Err(err).Msg("Failed to start server")
		}
	}()

	// Wait for interrupt signal to gracefully shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info().Msg("Shutting down server...")

	// Graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Fatal().Err(err).Msg("Server forced to shutdown")
	}

	logger.Info().Msg("Server exited")
}

// connectDB establishes a connection to the PostgreSQL database
func connectDB() (*sql.DB, error) {
	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_NAME"),
		os.Getenv("DB_SSLMODE"),
	)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Test connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	// Set connection pool settings
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(25)
	db.SetConnMaxLifetime(5 * time.Minute)

	return db, nil
}

// startBackgroundWorker starts a background worker to process blockchain jobs
func startBackgroundWorker(didService *services.DIDService, logger zerolog.Logger) {
	ticker := time.NewTicker(30 * time.Second) // Process every 30 seconds
	defer ticker.Stop()

	logger.Info().Msg("Starting background blockchain job processor")

	for {
		select {
		case <-ticker.C:
			if err := didService.ProcessBlockchainQueue(); err != nil {
				logger.Error().Err(err).Msg("Failed to process blockchain queue")
			}
		}
	}
}
