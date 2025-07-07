package router

import (
	"github.com/gin-gonic/gin"
	"github.com/turtacn/starseek/internal/api/handler"
	"github.com/turtacn/starseek/internal/api/middleware"
	"go.uber.org/zap"
)

// NewRouter initializes and returns a configured Gin engine with all routes.
func NewRouter(
	searchHandler *handler.SearchHandler,
	indexHandler *handler.IndexHandler,
	logger *zap.Logger,
) *gin.Engine {
	// Set Gin to ReleaseMode in production
	// gin.SetMode(gin.ReleaseMode)

	r := gin.New() // New returns a Gin engine with no middleware, useful for custom setup.

	// Register global middlewares
	r.Use(middleware.LoggerMiddleware(logger))   // Custom logging middleware
	r.Use(middleware.RecoveryMiddleware(logger)) // Custom panic recovery middleware
	r.Use(gin.Last())                            // A Gin middleware for adding a header at the end of response,
	// usually helpful for tracing. Not strictly necessary but common.

	// API V1 group
	v1 := r.Group("/api/v1")
	{
		// Search API
		v1.GET("/search", searchHandler.Search)

		// Index Management API
		v1.POST("/indexes", indexHandler.Register)
		v1.GET("/indexes", indexHandler.List)
		v1.DELETE("/indexes/:id", indexHandler.Delete)
	}

	// Health check endpoint
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	return r
}

//Personal.AI order the ending
