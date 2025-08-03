package logger

import (
	"os"
	"strings"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Environment variable names for logger configuration
const (
	EnvLogLevel    = "LOG_LEVEL"
	EnvEnvironment = "ENVIRONMENT"
)

// LogLevel represents different log levels
type LogLevel string

// Available log levels
const (
	DebugLevel LogLevel = "debug"
	InfoLevel  LogLevel = "info"
	WarnLevel  LogLevel = "warn"
	ErrorLevel LogLevel = "error"
)

// Environment represents different deployment environments
type Environment string

// Available deployment environments
const (
	Development Environment = "development"
	Production  Environment = "production"
	Testing     Environment = "testing"
)

// Config holds logger configuration
type Config struct {
	Level       LogLevel
	Environment Environment
}

// NewFromEnv creates a logger configuration from environment variables
func NewFromEnv() Config {
	return Config{
		Level:       getLogLevelFromEnv(),
		Environment: getEnvironmentFromEnv(),
	}
}

// getLogLevelFromEnv reads log level from LOG_LEVEL environment variable
func getLogLevelFromEnv() LogLevel {
	levelStr := strings.TrimSpace(os.Getenv(EnvLogLevel))
	level := strings.ToLower(levelStr)
	switch level {
	case "debug":
		return DebugLevel
	case "info":
		return InfoLevel
	case "warn":
		return WarnLevel
	case "error":
		return ErrorLevel
	default:
		return InfoLevel
	}
}

// getEnvironmentFromEnv reads environment from ENVIRONMENT environment variable
func getEnvironmentFromEnv() Environment {
	envStr := strings.TrimSpace(os.Getenv(EnvEnvironment))
	env := strings.ToLower(envStr)
	switch env {
	case "development", "dev":
		return Development
	case "production", "prod":
		return Production
	case "testing", "test":
		return Testing
	default:
		return Production
	}
}

// New creates a new zap logger based on configuration
func New(config Config) (*zap.Logger, error) {
	var zapConfig zap.Config

	switch config.Environment {
	case Production:
		zapConfig = zap.NewProductionConfig()
		// JSON format for production (easier for log aggregators)
		zapConfig.Encoding = "json"

	case Testing:
		zapConfig = zap.NewDevelopmentConfig()
		zapConfig.Encoding = "console"
		zapConfig.DisableCaller = true

	case Development:
		fallthrough
	default:
		zapConfig = zap.NewDevelopmentConfig()
		zapConfig.Encoding = "console"
		zapConfig.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	}

	// Always log to stdout/stderr (Kubernetes will handle collection)
	zapConfig.OutputPaths = []string{"stdout"}
	zapConfig.ErrorOutputPaths = []string{"stderr"}

	// Set log level
	zapConfig.Level = zap.NewAtomicLevelAt(mapLogLevel(config.Level))

	return zapConfig.Build()
}

// NewForProduction creates a production logger
func NewForProduction() (*zap.Logger, error) {
	config := Config{
		Level:       InfoLevel,
		Environment: Production,
	}
	return New(config)
}

// NewForDevelopment creates a development logger
func NewForDevelopment() (*zap.Logger, error) {
	config := Config{
		Level:       DebugLevel,
		Environment: Development,
	}
	return New(config)
}

// NewForTesting creates a testing logger
func NewForTesting() (*zap.Logger, error) {
	config := Config{
		Level:       ErrorLevel,
		Environment: Testing,
	}
	return New(config)
}

// NewNop returns a no-op logger
func NewNop() *zap.Logger {
	return zap.NewNop()
}

// mapLogLevel converts our LogLevel to zap level
func mapLogLevel(level LogLevel) zapcore.Level {
	switch level {
	case DebugLevel:
		return zapcore.DebugLevel
	case InfoLevel:
		return zapcore.InfoLevel
	case WarnLevel:
		return zapcore.WarnLevel
	case ErrorLevel:
		return zapcore.ErrorLevel
	default:
		return zapcore.InfoLevel
	}
}