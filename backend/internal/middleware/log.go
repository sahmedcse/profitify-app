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
		
		start := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery
		
		c.Next()
		
		latency := time.Since(start)
		status := c.Writer.Status()
		clientIP := c.ClientIP()
		method := c.Request.Method
		errorMessage := c.Errors.ByType(gin.ErrorTypePrivate).String()
		
		if raw != "" {
			path = path + "?" + raw
		}
		
		fields := map[string]any{
			"method":     method,
			"path":       path,
			"status":     status,
			"latency_ms": latency.Milliseconds(),
			"client_ip":  clientIP,
			"user_agent": c.Request.UserAgent(),
		}
		
		logWithFields := logger.WithFields(fields)
		
		if len(c.Errors) > 0 {
			logWithFields.Errorf("Request failed: %s", errorMessage)
		} else if status >= 500 {
			logWithFields.Error("Internal server error")
		} else if status >= 400 {
			logWithFields.Warn("Client error")
		} else {
			logWithFields.Info("Request completed")
		}
	}
}
