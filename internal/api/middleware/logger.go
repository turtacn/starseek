package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// LoggerMiddleware returns a Gin middleware that logs HTTP requests.
func LoggerMiddleware(logger *zap.Logger) gin.HandlerFunc {
	// Name the logger for this middleware
	logger = logger.Named("RequestLogger")

	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		rawQuery := c.Request.URL.RawQuery

		// Process the request
		c.Next()

		// Log after the request has been processed
		latency := time.Since(start)
		clientIP := c.ClientIP()
		method := c.Request.Method
		statusCode := c.Writer.Status()
		errorMessage := c.Errors.ByType(gin.ErrorTypePrivate).String() // Get any errors set by Gin

		fields := []zap.Field{
			zap.Int("status", statusCode),
			zap.String("method", method),
			zap.String("path", path),
			zap.String("ip", clientIP),
			zap.Duration("latency", latency),
			zap.Int("body_size", c.Writer.Size()),
		}

		if rawQuery != "" {
			fields = append(fields, zap.String("query", rawQuery))
		}
		if errorMessage != "" {
			fields = append(fields, zap.String("error", errorMessage))
		}

		if statusCode >= 400 && statusCode < 500 {
			logger.Warn("HTTP request failed (client error)", fields...)
		} else if statusCode >= 500 {
			logger.Error("HTTP request failed (server error)", fields...)
		} else {
			logger.Info("HTTP request", fields...)
		}
	}
}

//Personal.AI order the ending
