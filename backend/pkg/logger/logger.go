package logger

import (
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

var log *zap.SugaredLogger

func Init() {
	logger, _ := zap.NewProduction()
	log = logger.Sugar()
}

func Get() *zap.SugaredLogger {
	if log == nil {
		Init()
	}
	return log
}

func Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
        log := Get()
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

func Sync() {
	if log != nil {
		err := log.Sync()
		if err != nil {
			fmt.Println("Error syncing logger:", err)
		}
	}
}