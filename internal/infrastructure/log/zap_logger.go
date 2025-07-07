package log

import (
	"fmt"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Logger defines the interface for logging operations.
type Logger interface {
	Info(msg string, fields ...zap.Field)
	Warn(msg string, fields ...zap.Field)
	Error(msg string, fields ...zap.Field)
	Debug(msg string, fields ...zap.Field)
	Fatal(msg string, fields ...zap.Field)
	WithOptions(opts ...zap.Option) *zap.Logger // Allow custom options for context
	Named(name string) *zap.Logger              // Allow naming logger instances
}

// ZapLogger implements the Logger interface using Zap.
type ZapLogger struct {
	*zap.Logger
}

// Config holds the configuration for the ZapLogger.
type Config struct {
	Level      string `mapstructure:"level"`       // Log level (debug, info, warn, error, fatal)
	Format     string `mapstructure:"format"`      // Output format (json, console)
	OutputPath string `mapstructure:"output_path"` // Output file path (e.g., "stdout", "stderr", or a file path)
}

// NewLogger initializes and returns a new ZapLogger instance based on the provided configuration.
func NewLogger(cfg Config) (Logger, error) {
	var zapConfig zap.Config

	if cfg.Format == "json" {
		zapConfig = zap.NewProductionConfig()
	} else {
		zapConfig = zap.NewDevelopmentConfig()
	}

	// Set output path
	if cfg.OutputPath == "" {
		cfg.OutputPath = "stdout" // Default to stdout if not specified
	}
	zapConfig.OutputPaths = []string{cfg.OutputPath}
	zapConfig.ErrorOutputPaths = []string{cfg.OutputPath} // Errors also go to the main output

	// Set log level
	var level zapcore.Level
	if err := level.UnmarshalText([]byte(cfg.Level)); err != nil {
		return nil, fmt.Errorf("invalid log level: %s", cfg.Level)
	}
	zapConfig.Level = zap.NewAtomicLevelAt(level)

	// Build the logger
	logger, err := zapConfig.Build(zap.AddCallerSkip(1)) // Skip one caller to show the actual call site
	if err != nil {
		return nil, fmt.Errorf("failed to build zap logger: %w", err)
	}

	return &ZapLogger{logger}, nil
}

// Implement the Logger interface methods by calling the underlying zap.Logger methods.
func (l *ZapLogger) Info(msg string, fields ...zap.Field) {
	l.Logger.Info(msg, fields...)
}

func (l *ZapLogger) Warn(msg string, fields ...zap.Field) {
	l.Logger.Warn(msg, fields...)
}

func (l *ZapLogger) Error(msg string, fields ...zap.Field) {
	l.Logger.Error(msg, fields...)
}

func (l *ZapLogger) Debug(msg string, fields ...zap.Field) {
	l.Logger.Debug(msg, fields...)
}

func (l *ZapLogger) Fatal(msg string, fields ...zap.Field) {
	l.Logger.Fatal(msg, fields...)
}

func (l *ZapLogger) WithOptions(opts ...zap.Option) *zap.Logger {
	return l.Logger.WithOptions(opts...)
}

func (l *ZapLogger) Named(name string) *zap.Logger {
	return l.Logger.Named(name)
}

//Personal.AI order the ending
