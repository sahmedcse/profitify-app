package logger

import (
	"fmt"
	"os"
	"sync"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	instance *zap.SugaredLogger
	once     sync.Once
	initErr  error
)

// Config holds logger configuration
type Config struct {
	Level       string
	Environment string
	OutputPaths []string
}

// Init initializes the logger with the given configuration
func Init(cfg *Config) error {
	once.Do(func() {
		instance, initErr = buildLogger(cfg)
	})
	return initErr
}

// InitWithDefaults initializes logger with default configuration
func InitWithDefaults() error {
	return Init(&Config{
		Level:       getEnv("LOG_LEVEL", "info"),
		Environment: getEnv("ENVIRONMENT", "development"),
		OutputPaths: []string{"stdout"},
	})
}

// Get returns the logger instance, initializing with defaults if needed
func Get() *zap.SugaredLogger {
	if instance == nil {
		if err := InitWithDefaults(); err != nil {
			// Fallback to a basic logger if initialization fails
			fallback, _ := zap.NewProduction()
			return fallback.Sugar()
		}
	}
	return instance
}

// Sync flushes any buffered log entries
func Sync() error {
	if instance != nil {
		return instance.Sync()
	}
	return nil
}

// WithFields returns a logger with additional fields
func WithFields(fields map[string]interface{}) *zap.SugaredLogger {
	logger := Get()
	for k, v := range fields {
		logger = logger.With(k, v)
	}
	return logger
}

// buildLogger creates a new logger based on configuration
func buildLogger(cfg *Config) (*zap.SugaredLogger, error) {
	var zapCfg zap.Config
	
	if cfg.Environment == "production" {
		zapCfg = zap.NewProductionConfig()
		zapCfg.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	} else {
		zapCfg = zap.NewDevelopmentConfig()
		zapCfg.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	}
	
	// Set log level
	level, err := zapcore.ParseLevel(cfg.Level)
	if err != nil {
		level = zapcore.InfoLevel
	}
	zapCfg.Level = zap.NewAtomicLevelAt(level)
	
	// Set output paths
	if len(cfg.OutputPaths) > 0 {
		zapCfg.OutputPaths = cfg.OutputPaths
	}
	
	// Add caller information
	zapCfg.Development = cfg.Environment != "production"
	
	// Build the logger
	logger, err := zapCfg.Build(
		zap.AddCallerSkip(1), // Skip one level to show actual caller
		zap.AddStacktrace(zapcore.ErrorLevel),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to build logger: %w", err)
	}
	
	return logger.Sugar(), nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}