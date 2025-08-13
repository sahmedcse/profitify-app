package middleware

import (
	"time"

	"profitify-backend/pkg/logger"

	"github.com/gin-gonic/gin"
)

func Log() gin.HandlerFunc {
	return func(c *gin.Context) {
        log := logger.Get()
		c.Set("logger", log)
		c.Next()
        start := time.Now()
        c.Next()
        latency := time.Since(start)
        status := c.Writer.Status()

        requestID := c.GetHeader("X-Request-Id")
        if requestID == "" {
            requestID = c.Writer.Header().Get("X-Request-Id")
        }
        log.Infof("%s %s %d %s %s", c.Request.Method, c.Request.URL.Path, status, latency, c.ClientIP())

	}
}
