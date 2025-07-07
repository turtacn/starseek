package middleware

import (
	"fmt"
	"net/http"
	"runtime/debug"

	"github.com/gin-gonic/gin"
	"github.com/turtacn/starseek/internal/common/errorx"
	"go.uber.org/zap"
)

// RecoveryMiddleware returns a Gin middleware that recovers from panics.
func RecoveryMiddleware(logger *zap.Logger) gin.HandlerFunc {
	// Name the logger for this middleware
	logger = logger.Named("RecoveryMiddleware")

	return func(c *gin.Context) {
		defer func() {
			if r := recover(); r != nil {
				// Log the panic with stack trace
				logger.Error("Recovered from panic",
					zap.Any("panic", r),
					zap.String("stack", string(debug.Stack())),
				)

				// Respond with a 500 Internal Server Error
				err := errorx.ErrInternalServer.Wrap(fmt.Errorf("unhandled panic: %v", r))
				c.JSON(http.StatusInternalServerError, gin.H{
					"code":    err.Code,
					"message": err.Message,
				})
				c.Abort() // Abort the request chain
			}
		}()
		c.Next()
	}
}

//Personal.AI order the ending
