// internal/api/http/router.go
package http

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/turtacn/starseek/internal/infrastructure/logger"  // Adjust module path
	"github.com/turtacn/starseek/internal/infrastructure/metrics" // Adjust module path
	"go.opentelemetry.io/otel/sdk/trace"
)

// NewRouter initializes and configures the Gin HTTP router.
func NewRouter(
	handler *Handler,
	log logger.Logger,
	metrics *metrics.Metrics,
	tracerProvider *trace.TracerProvider,
) *gin.Engine {
	// Use gin.New() for more control over middleware.
	// gin.Default() includes Logger, Recovery, and Static middlewares by default.
	router := gin.New()

	// Global Middlewares (order matters)
	router.Use(RecoveryMiddleware(log))           // Recovery should be first to catch panics from other middlewares/handlers
	router.Use(RequestIDMiddleware())             // Generate/extract request ID
	router.Use(TracingMiddleware(tracerProvider)) // Initialize OpenTelemetry tracing
	router.Use(LoggingMiddleware(log))            // Log requests with IDs and trace context
	router.Use(MetricsMiddleware(metrics))        // Record request metrics

	// Health check endpoint
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// API V1 Group
	v1 := router.Group("/api/v1")
	{
		// Index Endpoints
		v1.POST("/indices", handler.RegisterIndexHandler)

		// Search Endpoints
		v1.POST("/indices/:index_name/_search", handler.SearchHandler)

		// Add more API routes here as needed...
	}

	return router
}
