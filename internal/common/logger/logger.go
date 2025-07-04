package logger

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"

	"internal/common/constants"
)

// Logger defines the interface for structured logging
type Logger interface {
	// Standard logging methods
	Debug(msg string, fields ...Field)
	Info(msg string, fields ...Field)
	Warn(msg string, fields ...Field)
	Error(msg string, fields ...Field)
	Fatal(msg string, fields ...Field)
	Panic(msg string, fields ...Field)

	// Formatted logging methods
	Debugf(template string, args ...interface{})
	Infof(template string, args ...interface{})
	Warnf(template string, args ...interface{})
	Errorf(template string, args ...interface{})
	Fatalf(template string, args ...interface{})
	Panicf(template string, args ...interface{})

	// Context-aware logging
	DebugContext(ctx context.Context, msg string, fields ...Field)
	InfoContext(ctx context.Context, msg string, fields ...Field)
	WarnContext(ctx context.Context, msg string, fields ...Field)
	ErrorContext(ctx context.Context, msg string, fields ...Field)

	// Logger management
	With(fields ...Field) Logger
	WithError(err error) Logger
	WithContext(ctx context.Context) Logger
	Named(name string) Logger

	// Level control
	SetLevel(level Level)
	GetLevel() Level
	IsLevelEnabled(level Level) bool

	// Sync ensures all buffered log entries are written
	Sync() error
}

// Field represents a key-value pair for structured logging
type Field struct {
	Key   string
	Value interface{}
	Type  FieldType
}

// FieldType represents the type of a log field
type FieldType int

const (
	StringType FieldType = iota
	IntType
	Int64Type
	Float64Type
	BoolType
	TimeType
	DurationType
	ErrorType
	ObjectType
	ArrayType
)

// Level represents log levels
type Level int8

const (
	DebugLevel Level = iota - 1
	InfoLevel
	WarnLevel
	ErrorLevel
	FatalLevel
	PanicLevel
)

// String returns the string representation of the log level
func (l Level) String() string {
	switch l {
	case DebugLevel:
		return "DEBUG"
	case InfoLevel:
		return "INFO"
	case WarnLevel:
		return "WARN"
	case ErrorLevel:
		return "ERROR"
	case FatalLevel:
		return "FATAL"
	case PanicLevel:
		return "PANIC"
	default:
		return "UNKNOWN"
	}
}

// ParseLevel parses a string to Level
func ParseLevel(s string) (Level, error) {
	switch s {
	case "debug", "DEBUG":
		return DebugLevel, nil
	case "info", "INFO":
		return InfoLevel, nil
	case "warn", "WARN", "warning", "WARNING":
		return WarnLevel, nil
	case "error", "ERROR":
		return ErrorLevel, nil
	case "fatal", "FATAL":
		return FatalLevel, nil
	case "panic", "PANIC":
		return PanicLevel, nil
	default:
		return InfoLevel, fmt.Errorf("invalid log level: %s", s)
	}
}

// Config holds logger configuration
type Config struct {
	Level            string        `yaml:"level" json:"level"`
	Format           string        `yaml:"format" json:"format"`                         // json, console
	Output           string        `yaml:"output" json:"output"`                         // stdout, stderr, file path
	EnableCaller     bool          `yaml:"enable_caller" json:"enable_caller"`
	EnableStackTrace bool          `yaml:"enable_stacktrace" json:"enable_stacktrace"`
	TimeFormat       string        `yaml:"time_format" json:"time_format"`

	// File rotation settings
	Rotation RotationConfig `yaml:"rotation" json:"rotation"`

	// Development settings
	Development bool `yaml:"development" json:"development"`

	// Sampling settings
	Sampling SamplingConfig `yaml:"sampling" json:"sampling"`
}

// RotationConfig holds log rotation configuration
type RotationConfig struct {
	Enabled    bool `yaml:"enabled" json:"enabled"`
	MaxSize    int  `yaml:"max_size" json:"max_size"`       // megabytes
	MaxBackups int  `yaml:"max_backups" json:"max_backups"` // number of backups
	MaxAge     int  `yaml:"max_age" json:"max_age"`         // days
	Compress   bool `yaml:"compress" json:"compress"`
}

// SamplingConfig holds log sampling configuration
type SamplingConfig struct {
	Enabled    bool          `yaml:"enabled" json:"enabled"`
	Tick       time.Duration `yaml:"tick" json:"tick"`
	Initial    int           `yaml:"initial" json:"initial"`
	Thereafter int           `yaml:"thereafter" json:"thereafter"`
}

// zapLogger implements Logger interface using zap
type zapLogger struct {
	zap    *zap.Logger
	config *Config
	level  zap.AtomicLevel
	mu     sync.RWMutex
}

// Global logger instance
var (
	globalLogger Logger
	once         sync.Once
)

// Field creation functions
func String(key, value string) Field {
	return Field{Key: key, Value: value, Type: StringType}
}

func Int(key string, value int) Field {
	return Field{Key: key, Value: value, Type: IntType}
}

func Int64(key string, value int64) Field {
	return Field{Key: key, Value: value, Type: Int64Type}
}

func Float64(key string, value float64) Field {
	return Field{Key: key, Value: value, Type: Float64Type}
}

func Bool(key string, value bool) Field {
	return Field{Key: key, Value: value, Type: BoolType}
}

func Time(key string, value time.Time) Field {
	return Field{Key: key, Value: value, Type: TimeType}
}

func Duration(key string, value time.Duration) Field {
	return Field{Key: key, Value: value, Type: DurationType}
}

func Error(err error) Field {
	return Field{Key: "error", Value: err, Type: ErrorType}
}

func Object(key string, value interface{}) Field {
	return Field{Key: key, Value: value, Type: ObjectType}
}

func Array(key string, value interface{}) Field {
	return Field{Key: key, Value: value, Type: ArrayType}
}

// NewLogger creates a new logger with the given configuration
func NewLogger(config *Config) (Logger, error) {
	if config == nil {
		config = getDefaultConfig()
	}

	// Parse log level
	level, err := ParseLevel(config.Level)
	if err != nil {
		return nil, fmt.Errorf("invalid log level: %w", err)
	}

	// Create atomic level for dynamic level changes
	atomicLevel := zap.NewAtomicLevelAt(zapcore.Level(level))

	// Create encoder config
	encoderConfig := getEncoderConfig(config)

	// Create encoder
	var encoder zapcore.Encoder
	switch config.Format {
	case "json":
		encoder = zapcore.NewJSONEncoder(encoderConfig)
	case "console":
		encoder = zapcore.NewConsoleEncoder(encoderConfig)
	default:
		encoder = zapcore.NewJSONEncoder(encoderConfig)
	}

	// Create writer syncer
	writeSyncer, err := getWriteSyncer(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create write syncer: %w", err)
	}

	// Create core
	core := zapcore.NewCore(encoder, writeSyncer, atomicLevel)

	// Add sampling if enabled
	if config.Sampling.Enabled {
		core = zapcore.NewSamplerWithOptions(
			core,
			config.Sampling.Tick,
			config.Sampling.Initial,
			config.Sampling.Thereafter,
		)
	}

	// Create logger options
	options := []zap.Option{
		zap.AddCaller(),
		zap.AddCallerSkip(1),
	}

	if config.EnableStackTrace {
		options = append(options, zap.AddStacktrace(zapcore.ErrorLevel))
	}

	if config.Development {
		options = append(options, zap.Development())
	}

	// Create zap logger
	zapLog := zap.New(core, options...)

	return &zapLogger{
		zap:    zapLog,
		config: config,
		level:  atomicLevel,
	}, nil
}

// GetGlobalLogger returns the global logger instance
func GetGlobalLogger() Logger {
	once.Do(func() {
		logger, err := NewLogger(getDefaultConfig())
		if err != nil {
			panic(fmt.Sprintf("failed to create global logger: %v", err))
		}
		globalLogger = logger
	})
	return globalLogger
}

// SetGlobalLogger sets the global logger instance
func SetGlobalLogger(logger Logger) {
	globalLogger = logger
}

// InitGlobalLogger initializes the global logger with the given configuration
func InitGlobalLogger(config *Config) error {
	logger, err := NewLogger(config)
	if err != nil {
		return err
	}
	SetGlobalLogger(logger)
	return nil
}

// Implementation of Logger interface

func (l *zapLogger) Debug(msg string, fields ...Field) {
	l.zap.Debug(msg, l.convertFields(fields...)...)
}

func (l *zapLogger) Info(msg string, fields ...Field) {
	l.zap.Info(msg, l.convertFields(fields...)...)
}

func (l *zapLogger) Warn(msg string, fields ...Field) {
	l.zap.Warn(msg, l.convertFields(fields...)...)
}

func (l *zapLogger) Error(msg string, fields ...Field) {
	l.zap.Error(msg, l.convertFields(fields...)...)
}

func (l *zapLogger) Fatal(msg string, fields ...Field) {
	l.zap.Fatal(msg, l.convertFields(fields...)...)
}

func (l *zapLogger) Panic(msg string, fields ...Field) {
	l.zap.Panic(msg, l.convertFields(fields...)...)
}

func (l *zapLogger) Debugf(template string, args ...interface{}) {
	l.zap.Debug(fmt.Sprintf(template, args...))
}

func (l *zapLogger) Infof(template string, args ...interface{}) {
	l.zap.Info(fmt.Sprintf(template, args...))
}

func (l *zapLogger) Warnf(template string, args ...interface{}) {
	l.zap.Warn(fmt.Sprintf(template, args...))
}

func (l *zapLogger) Errorf(template string, args ...interface{}) {
	l.zap.Error(fmt.Sprintf(template, args...))
}

func (l *zapLogger) Fatalf(template string, args ...interface{}) {
	l.zap.Fatal(fmt.Sprintf(template, args...))
}

func (l *zapLogger) Panicf(template string, args ...interface{}) {
	l.zap.Panic(fmt.Sprintf(template, args...))
}

func (l *zapLogger) DebugContext(ctx context.Context, msg string, fields ...Field) {
	l.zap.Debug(msg, l.convertFieldsWithContext(ctx, fields...)...)
}

func (l *zapLogger) InfoContext(ctx context.Context, msg string, fields ...Field) {
	l.zap.Info(msg, l.convertFieldsWithContext(ctx, fields...)...)
}

func (l *zapLogger) WarnContext(ctx context.Context, msg string, fields ...Field) {
	l.zap.Warn(msg, l.convertFieldsWithContext(ctx, fields...)...)
}

func (l *zapLogger) ErrorContext(ctx context.Context, msg string, fields ...Field) {
	l.zap.Error(msg, l.convertFieldsWithContext(ctx, fields...)...)
}

func (l *zapLogger) With(fields ...Field) Logger {
	return &zapLogger{
		zap:    l.zap.With(l.convertFields(fields...)...),
		config: l.config,
		level:  l.level,
	}
}

func (l *zapLogger) WithError(err error) Logger {
	return l.With(Error(err))
}

func (l *zapLogger) WithContext(ctx context.Context) Logger {
	return &zapLogger{
		zap:    l.zap.With(l.extractContextFields(ctx)...),
		config: l.config,
		level:  l.level,
	}
}

func (l *zapLogger) Named(name string) Logger {
	return &zapLogger{
		zap:    l.zap.Named(name),
		config: l.config,
		level:  l.level,
	}
}

func (l *zapLogger) SetLevel(level Level) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.level.SetLevel(zapcore.Level(level))
}

func (l *zapLogger) GetLevel() Level {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return Level(l.level.Level())
}

func (l *zapLogger) IsLevelEnabled(level Level) bool {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return l.level.Enabled(zapcore.Level(level))
}

func (l *zapLogger) Sync() error {
	return l.zap.Sync()
}

// Helper methods

func (l *zapLogger) convertFields(fields ...Field) []zap.Field {
	zapFields := make([]zap.Field, 0, len(fields))
	for _, field := range fields {
		zapFields = append(zapFields, l.convertField(field))
	}
	return zapFields
}

func (l *zapLogger) convertField(field Field) zap.Field {
	switch field.Type {
	case StringType:
		return zap.String(field.Key, field.Value.(string))
	case IntType:
		return zap.Int(field.Key, field.Value.(int))
	case Int64Type:
		return zap.Int64(field.Key, field.Value.(int64))
	case Float64Type:
		return zap.Float64(field.Key, field.Value.(float64))
	case BoolType:
		return zap.Bool(field.Key, field.Value.(bool))
	case TimeType:
		return zap.Time(field.Key, field.Value.(time.Time))
	case DurationType:
		return zap.Duration(field.Key, field.Value.(time.Duration))
	case ErrorType:
		return zap.Error(field.Value.(error))
	case ObjectType:
		return zap.Any(field.Key, field.Value)
	case ArrayType:
		return zap.Any(field.Key, field.Value)
	default:
		return zap.Any(field.Key, field.Value)
	}
}

func (l *zapLogger) convertFieldsWithContext(ctx context.Context, fields ...Field) []zap.Field {
	contextFields := l.extractContextFields(ctx)
	userFields := l.convertFields(fields...)
	return append(contextFields, userFields...)
}

func (l *zapLogger) extractContextFields(ctx context.Context) []zap.Field {
	var fields []zap.Field

	// Extract request ID from context
	if requestID := ctx.Value("request_id"); requestID != nil {
		if id, ok := requestID.(string); ok {
			fields = append(fields, zap.String("request_id", id))
		}
	}

	// Extract correlation ID from context
	if correlationID := ctx.Value("correlation_id"); correlationID != nil {
		if id, ok := correlationID.(string); ok {
			fields = append(fields, zap.String("correlation_id", id))
		}
	}

	// Extract user ID from context
	if userID := ctx.Value("user_id"); userID != nil {
		if id, ok := userID.(string); ok {
			fields = append(fields, zap.String("user_id", id))
		}
	}

	return fields
}

// Configuration helpers

func getDefaultConfig() *Config {
	return &Config{
		Level:            constants.DefaultLogLevel,
		Format:           "json",
		Output:           "stdout",
		EnableCaller:     true,
		EnableStackTrace: false,
		TimeFormat:       constants.TimeFormatRFC3339,
		Rotation: RotationConfig{
			Enabled:    false,
			MaxSize:    constants.DefaultLogMaxSize,
			MaxBackups: constants.DefaultLogMaxBackups,
			MaxAge:     constants.DefaultLogMaxAge,
			Compress:   constants.DefaultLogCompress,
		},
		Development: false,
		Sampling: SamplingConfig{
			Enabled:    false,
			Tick:       time.Second,
			Initial:    100,
			Thereafter: 100,
		},
	}
}

func getEncoderConfig(config *Config) zapcore.EncoderConfig {
	encoderConfig := zap.NewProductionEncoderConfig()

	if config.Development {
		encoderConfig = zap.NewDevelopmentEncoderConfig()
	}

	// Customize time format
	if config.TimeFormat != "" {
		switch config.TimeFormat {
		case "unix":
			encoderConfig.EncodeTime = zapcore.EpochTimeEncoder
		case "unix_milli":
			encoderConfig.EncodeTime = zapcore.EpochMillisTimeEncoder
		case "unix_nano":
			encoderConfig.EncodeTime = zapcore.EpochNanosTimeEncoder
		default:
			encoderConfig.EncodeTime = zapcore.TimeEncoderOfLayout(config.TimeFormat)
		}
	}

	// Customize level encoding
	if config.Format == "console" {
		encoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	} else {
		encoderConfig.EncodeLevel = zapcore.LowercaseLevelEncoder
	}

	// Customize caller encoding
	if config.EnableCaller {
		encoderConfig.EncodeCaller = zapcore.ShortCallerEncoder
	}

	return encoderConfig
}

func getWriteSyncer(config *Config) (zapcore.WriteSyncer, error) {
	switch config.Output {
	case "stdout":
		return zapcore.AddSync(os.Stdout), nil
	case "stderr":
		return zapcore.AddSync(os.Stderr), nil
	default:
		// File output
		if config.Rotation.Enabled {
			// Use lumberjack for log rotation
			return zapcore.AddSync(&lumberjack.Logger{
				Filename:   config.Output,
				MaxSize:    config.Rotation.MaxSize,
				MaxBackups: config.Rotation.MaxBackups,
				MaxAge:     config.Rotation.MaxAge,
				Compress:   config.Rotation.Compress,
			}), nil
		} else {
			// Create directory if it doesn't exist
			dir := filepath.Dir(config.Output)
			if err := os.MkdirAll(dir, 0755); err != nil {
				return nil, fmt.Errorf("failed to create log directory: %w", err)
			}

			file, err := os.OpenFile(config.Output, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
			if err != nil {
				return nil, fmt.Errorf("failed to open log file: %w", err)
			}
			return zapcore.AddSync(file), nil
		}
	}
}

// Convenience functions using global logger

func Debug(msg string, fields ...Field) {
	GetGlobalLogger().Debug(msg, fields...)
}

func Info(msg string, fields ...Field) {
	GetGlobalLogger().Info(msg, fields...)
}

func Warn(msg string, fields ...Field) {
	GetGlobalLogger().Warn(msg, fields...)
}

func Error(msg string, fields ...Field) {
	GetGlobalLogger().Error(msg, fields...)
}

func Fatal(msg string, fields ...Field) {
	GetGlobalLogger().Fatal(msg, fields...)
}

func Panic(msg string, fields ...Field) {
	GetGlobalLogger().Panic(msg, fields...)
}

func Debugf(template string, args ...interface{}) {
	GetGlobalLogger().Debugf(template, args...)
}

func Infof(template string, args ...interface{}) {
	GetGlobalLogger().Infof(template, args...)
}

func Warnf(template string, args ...interface{}) {
	GetGlobalLogger().Warnf(template, args...)
}

func Errorf(template string, args ...interface{}) {
	GetGlobalLogger().Errorf(template, args...)
}

func Fatalf(template string, args ...interface{}) {
	GetGlobalLogger().Fatalf(template, args...)
}

func Panicf(template string, args ...interface{}) {
	GetGlobalLogger().Panicf(template, args...)
}

func DebugContext(ctx context.Context, msg string, fields ...Field) {
	GetGlobalLogger().DebugContext(ctx, msg, fields...)
}

func InfoContext(ctx context.Context, msg string, fields ...Field) {
	GetGlobalLogger().InfoContext(ctx, msg, fields...)
}

func WarnContext(ctx context.Context, msg string, fields ...Field) {
	GetGlobalLogger().WarnContext(ctx, msg, fields...)
}

func ErrorContext(ctx context.Context, msg string, fields ...Field) {
	GetGlobalLogger().ErrorContext(ctx, msg, fields...)
}

func With(fields ...Field) Logger {
	return GetGlobalLogger().With(fields...)
}

func WithError(err error) Logger {
	return GetGlobalLogger().WithError(err)
}

func WithContext(ctx context.Context) Logger {
	return GetGlobalLogger().WithContext(ctx)
}

func Named(name string) Logger {
	return GetGlobalLogger().Named(name)
}

func SetLevel(level Level) {
	GetGlobalLogger().SetLevel(level)
}

func GetLevel() Level {
	return GetGlobalLogger().GetLevel()
}

func IsLevelEnabled(level Level) bool {
	return GetGlobalLogger().IsLevelEnabled(level)
}

func Sync() error {
	return GetGlobalLogger().Sync()
}

// MultiWriter allows writing to multiple outputs
type MultiWriter struct {
	writers []io.Writer
}

func NewMultiWriter(writers ...io.Writer) *MultiWriter {
	return &MultiWriter{writers: writers}
}

func (mw *MultiWriter) Write(p []byte) (n int, err error) {
	for _, w := range mw.writers {
		n, err = w.Write(p)
		if err != nil {
			return
		}
	}
	return len(p), nil
}

//Personal.AI order the ending
