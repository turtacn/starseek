// internal/api/http/middleware.go
package http

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/propagation"
	semconv "go.opentelemetry.io/otel/semconv/v1.24.0"
	"go.opentelemetry.io/otel/trace"

	"github.com/turtacn/starseek/internal/infrastructure/logger"  // Adjust module path
	"github.com/turtacn/starseek/internal/infrastructure/metrics" // Adjust module path
)

const (
	RequestIDKey = "requestID"
	TraceIDKey   = "traceID" // Key to store Trace ID in Gin context
)

// RequestIDMiddleware generates a unique request ID and attaches it to the context and response header.
func RequestIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := c.Request.Header.Get("X-Request-ID")
		if requestID == "" {
			requestID = uuid.New().String()
		}
		c.Set(RequestIDKey, requestID)
		c.Writer.Header().Set("X-Request-ID", requestID)

		// Also attach to context.Context for downstream logging/tracing
		ctx := context.WithValue(c.Request.Context(), RequestIDKey, requestID)
		c.Request = c.Request.WithContext(ctx)

		c.Next()
	}
}

// LoggingMiddleware logs details about each HTTP request.
func LoggingMiddleware(log logger.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		rawQuery := c.Request.URL.RawQuery

		// Process request
		c.Next()

		duration := time.Since(start)
		clientIP := c.ClientIP()
		method := c.Request.Method
		statusCode := c.Writer.Status()
		errorMessage := c.Errors.ByType(gin.ErrorTypePrivate).String() // Errors internal to Gin (e.g. from Recovery)

		// Get Request ID and Trace ID from context if set by previous middlewares
		requestID, _ := c.Get(RequestIDKey)
		traceID, _ := c.Get(TraceIDKey) // Set by TracingMiddleware

		logEntry := log.WithFields(map[string]interface{}{
			"status_code": statusCode,
			"method":      method,
			"path":        path,
			"query":       rawQuery,
			"ip":          clientIP,
			"latency":     duration,
			"user_agent":  c.Request.UserAgent(),
			"request_id":  requestID,
			"trace_id":    traceID, // Add trace ID to log
		})

		if errorMessage != "" {
			logEntry = logEntry.WithField("error", errorMessage)
		}

		if statusCode >= http.StatusInternalServerError {
			logEntry.Error("HTTP Request Error")
		} else if statusCode >= http.StatusBadRequest {
			logEntry.Warn("HTTP Request Client Error")
		} else {
			logEntry.Info("HTTP Request")
		}
	}
}

// MetricsMiddleware records HTTP request metrics.
func MetricsMiddleware(m *metrics.Metrics) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		c.Next() // Process request

		duration := time.Since(start).Seconds()
		statusCode := fmt.Sprintf("%d", c.Writer.Status())
		method := c.Request.Method
		path := c.FullPath() // Use FullPath for consistent metric labels across parameterized routes

		m.HTTPRequestsTotal.WithLabelValues(method, path, statusCode).Inc()
		m.HTTPRequestsDuration.WithLabelValues(method, path, statusCode).Observe(duration)
	}
}

// TracingMiddleware initializes OpenTelemetry spans for each request.
func TracingMiddleware(tp *trace.TracerProvider) gin.HandlerFunc {
	propagator := propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{})
	tracer := tp.Tracer("http-server") // Get a tracer from the provider

	return func(c *gin.Context) {
		// Extract span context from incoming request headers
		ctx := propagator.Extract(c.Request.Context(), propagation.HeaderCarrier(c.Request.Header))

		// Start a new span
		opts := []trace.SpanStartOption{
			trace.WithAttributes(
				semconv.HTTPMethodKey.String(c.Request.Method),
				semconv.HTTPTargetKey.String(c.Request.URL.Path),
				semconv.HTTPURLKey.String(c.Request.URL.String()),
				semconv.NetHostNameKey.String(c.Request.Host),
				semconv.HTTPRequestContentLength(int64(c.Request.ContentLength)),
				attribute.String("http.client_ip", c.ClientIP()),
				attribute.String("http.user_agent", c.Request.UserAgent()),
			),
			trace.WithSpanKind(trace.SpanKindServer), // Indicate this is a server-side span
		}
		ctx, span := tracer.Start(ctx, fmt.Sprintf("%s %s", c.Request.Method, c.FullPath()), opts...)
		defer span.End()

		// Inject the new context back into the Gin context
		c.Request = c.Request.WithContext(ctx)

		// Store trace ID in Gin context for logging
		spanContext := span.SpanContext()
		if spanContext.IsValid() {
			c.Set(TraceIDKey, spanContext.TraceID().String())
		}

		c.Next()

		// Record response details
		if c.Writer.Status() >= http.StatusBadRequest {
			span.SetStatus(codes.Error, fmt.Sprintf("HTTP %d", c.Writer.Status()))
			if errorMessage := c.Errors.String(); errorMessage != "" {
				span.RecordError(fmt.Errorf(errorMessage))
				span.SetAttributes(attribute.String("gin.error", errorMessage))
			}
		} else {
			span.SetStatus(codes.Ok, "OK")
		}
		span.SetAttributes(semconv.HTTPStatusCodeKey.Int(c.Writer.Status()))
		span.SetAttributes(semconv.HTTPResponseContentLength(int64(c.Writer.Size())))
	}
}

// RecoveryMiddleware recovers from panics and returns a 500 Internal Server Error.
func RecoveryMiddleware(log logger.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				// Get Request ID for context
				requestID, _ := c.Get(RequestIDKey)

				// Log the panic
				log.WithFields(map[string]interface{}{
					"request_id": requestID,
					"error":      err,
					"stack":      c.Errors.ByType(gin.ErrorTypePrivate).String(), // Gin captures stack trace if Default() is used
					"path":       c.Request.URL.Path,
					"method":     c.Request.Method,
				}).Errorf("Panic recovered: %v", err)

				c.AbortWithStatus(http.StatusInternalServerError)
				c.JSON(http.StatusInternalServerError, gin.H{
					"code":    "INTERNAL_SERVER_ERROR",
					"message": "An unexpected error occurred.",
				})
			}
		}()
		c.Next()
	}
}
