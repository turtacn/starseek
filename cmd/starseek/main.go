// cmd/starseek/main.go
package main

import (
	"context"
	"fmt"
	"github.com/turtacn/starseek/internal/infrastructure/tracing"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/turtacn/starseek/internal/api/http"
	"github.com/turtacn/starseek/internal/application"
	"github.com/turtacn/starseek/internal/common/config"
	domainindex "github.com/turtacn/starseek/internal/domain/index"   // Alias to avoid conflict
	domainsearch "github.com/turtacn/starseek/internal/domain/search" // Alias
	"github.com/turtacn/starseek/internal/domain/tokenizer"
	"github.com/turtacn/starseek/internal/infrastructure/cache"
	"github.com/turtacn/starseek/internal/infrastructure/database"
	"github.com/turtacn/starseek/internal/infrastructure/logger"
	"github.com/turtacn/starseek/internal/infrastructure/metrics"
	"github.com/turtacn/starseek/internal/infrastructure/persistence"
	"github.com/turtacn/starseek/internal/infrastructure/server"
)

const serviceName = "starseek"

func main() {
	// 1. Load Configuration
	cfg, err := config.LoadConfig("./configs") // Assuming config files are in ./configs
	if err != nil {
		fmt.Printf("Failed to load configuration: %v\n", err)
		os.Exit(1)
	}

	// 2. Initialize Logger
	log, err := logger.NewZapLogger(cfg.Logger.Level, cfg.Logger.Format)
	if err != nil {
		fmt.Printf("Failed to initialize logger: %v\n", err)
		os.Exit(1)
	}
	// Use defer to ensure logger is synced on exit
	defer func() {
		if err := log.Sync(); err != nil {
			// Some errors like "sync /dev/stderr: invalid argument" can happen on Linux.
			// Only log if it's a critical sync error.
			if err.Error() != "sync /dev/stderr: invalid argument" {
				fmt.Printf("Failed to sync logger: %v\n", err)
			}
		}
	}()

	log.Info("Application starting up...")

	// 3. Initialize Metrics
	// The metrics server starts in InitMetrics if enabled
	appMetrics := metrics.InitMetrics(&cfg.Metrics, log)

	// 4. Initialize Tracing
	tracerProvider, err := tracing.InitOpenTelemetry(&cfg.Tracing, log)
	if err != nil {
		log.Errorf("Failed to initialize tracing: %v", err)
		// Don't exit, tracing can be non-critical
	}
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		tracing.CloseTracerProvider(ctx, tracerProvider, log)
	}()

	// 5. Initialize Cache Client (Redis)
	redisClient, err := cache.NewRedisClient(&cfg.Redis, log)
	if err != nil {
		log.Fatalf("Failed to initialize Redis client: %v", err)
	}
	defer func() {
		if err := redisClient.Close(); err != nil {
			log.Errorf("Failed to close Redis client: %v", err)
		}
	}()
	log.Info("Redis client initialized.")

	// 6. Initialize Metadata Database Client (MySQL)
	metadataDBClient, err := database.NewMySQLClient(&cfg.MySQL, log) // Assuming metadata DB uses MySQL config
	if err != nil {
		log.Fatalf("Failed to initialize metadata database client: %v", err)
	}
	defer func() {
		if err := metadataDBClient.Close(); err != nil {
			log.Errorf("Failed to close metadata database client: %v", err)
		}
	}()
	log.Info("Metadata database client initialized.")

	// 7. Initialize Main Business Database Client (e.g., another MySQL instance, or a specialized client)
	mainDBClient, err := database.NewDatabaseClient(&cfg.MySQL, log) // Reusing MySQL config for simplicity
	if err != nil {
		log.Fatalf("Failed to initialize main database client: %v", err)
	}
	defer func() {
		if err := mainDBClient.Close(); err != nil {
			log.Errorf("Failed to close main database client: %v", err)
		}
	}()
	log.Info("Main business database client initialized.")

	// Infrastructure/Persistence Layer
	indexMetadataRepo := persistence.NewSQLIndexMetadataRepository(metadataDBClient.DB, log)

	// Domain Layer
	tokenizerService, err := tokenizer.NewTokenizerService(&cfg.Tokenizer, log)
	if err != nil {
		log.Fatalf("Failed to initialize tokenizer service: %v", err)
	}

	domainIndexService := domainindex.NewDomainIndexService(indexMetadataRepo, log)
	dummySearchEngine := domainsearch.NewDummySearchEngine(log) // Placeholder for real search engine
	domainSearchService := domainsearch.NewDomainSearchService(domainIndexService, dummySearchEngine, tokenizerService, log)

	// Application Layer
	appIndexService := application.NewIndexApplicationService(domainIndexService, log)
	appSearchService := application.NewSearchApplicationService(domainSearchService, log)

	// API Layer (HTTP Handlers and Routers)
	httpHandler := http.NewHandler(appIndexService, appSearchService)
	httpRouter := http.NewRouter(httpHandler, log, appMetrics, tracerProvider)

	// HTTP Server
	httpServer := server.NewServer(httpRouter, &cfg.Server, log)

	// Start server in a goroutine
	go func() {
		if err := httpServer.Run(); err != nil {
			log.Errorf("HTTP server stopped with error: %v", err)
			// If server stops unexpectedly, signal main to exit
			os.Exit(1)
		}
	}()

	// Signal Handling for Graceful Shutdown
	quit := make(chan os.Signal, 1)
	// SIGINT (Ctrl+C) and SIGTERM (graceful shutdown by orchestrators like Kubernetes)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit // Block until a signal is received

	log.Info("Shutting down application...")

	// Create a context with a timeout for graceful shutdown
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	// Shut down HTTP server
	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("HTTP server forced to shutdown: %v", err)
	} else {
		log.Info("HTTP server shut down gracefully.")
	}

	log.Info("Application exited successfully.")
}
