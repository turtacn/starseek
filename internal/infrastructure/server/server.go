// internal/infrastructure/server/server.go
package server

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/turtacn/starseek/internal/config"                // Adjust module path
	"github.com/turtacn/starseek/internal/infrastructure/logger" // Adjust module path
)

// Server represents the HTTP server instance.
type Server struct {
	httpServer *http.Server
	logger     logger.Logger
	router     *gin.Engine // Keep a reference to the Gin engine if needed for future extensions
	config     *config.ServerConfig
}

// NewServer creates a new HTTP server instance.
func NewServer(
	router *gin.Engine,
	cfg *config.ServerConfig,
	log logger.Logger,
) *Server {
	httpServer := &http.Server{
		Addr:    cfg.Addr(),
		Handler: router,
		// Optional: configure read/write timeouts to prevent slowloris attacks
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	return &Server{
		httpServer: httpServer,
		logger:     log,
		router:     router,
		config:     cfg,
	}
}

// Run starts the HTTP server and listens for incoming requests.
// It blocks until the server is shut down or an error occurs.
func (s *Server) Run() error {
	s.logger.Infof("HTTP server starting on %s", s.httpServer.Addr)

	// In a goroutine, start listening. This allows the main goroutine
	// to proceed for graceful shutdown logic.
	err := s.httpServer.ListenAndServe()
	if err != nil && err != http.ErrServerClosed {
		s.logger.Errorf("HTTP server failed to start: %v", err)
		return err
	}

	s.logger.Info("HTTP server stopped gracefully.")
	return nil
}

// Shutdown gracefully shuts down the HTTP server.
// It waits for at most `ctx`'s deadline for current connections to finish.
func (s *Server) Shutdown(ctx context.Context) error {
	s.logger.Info("HTTP server shutting down...")

	// The context here usually comes from a signal listener (e.g., os.Interrupt, syscall.SIGTERM)
	// and often has a short timeout.
	if err := s.httpServer.Shutdown(ctx); err != nil {
		s.logger.Errorf("HTTP server shutdown failed: %v", err)
		return err
	}

	s.logger.Info("HTTP server gracefully shut down.")
	return nil
}

// GetAddr returns the address the server is listening on.
func (s *Server) GetAddr() string {
	return s.httpServer.Addr
}
