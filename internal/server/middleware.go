package server

import (
	"log/slog"
	"time"

	"github.com/gin-gonic/gin"
)

// SlogMiddleware replaces gin.Logger() with structured JSON logging.
func SlogMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		slog.Info("request",
			"method", c.Request.Method,
			"path", c.Request.URL.Path,
			"status", c.Writer.Status(),
			"ms", time.Since(start).Milliseconds(),
			"ip", c.ClientIP(),
		)
	}
}
