package logger

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

// Logger 定义统一的日志记录器接口
type Logger interface {
	// Debug 记录调试级别日志
	Debug(msg string, fields ...Field)
	// Info 记录信息级别日志
	Info(msg string, fields ...Field)
	// Warn 记录警告级别日志
	Warn(msg string, fields ...Field)
	// Error 记录错误级别日志
	Error(msg string, fields ...Field)
	// Fatal 记录致命错误日志并退出程序
	Fatal(msg string, fields ...Field)
	// Panic 记录panic级别日志并触发panic
	Panic(msg string, fields ...Field)

	// With 添加字段到日志上下文
	With(fields ...Field) Logger
	// WithContext 添加上下文信息
	WithContext(ctx context.Context) Logger

	// SetLevel 动态设置日志级别
	SetLevel(level Level)
	// GetLevel 获取当前日志级别
	GetLevel() Level

	// Sync 同步日志缓冲区
	Sync() error
}

// Field 日志字段结构
type Field struct {
	Key   string
	Value interface{}
}

// Level 日志级别
type Level int

const (
	DebugLevel Level = iota
	InfoLevel
	WarnLevel
	ErrorLevel
	FatalLevel
	PanicLevel
)

// String 返回日志级别的字符串表示
func (l Level) String() string {
	switch l {
	case DebugLevel:
		return "debug"
	case InfoLevel:
		return "info"
	case WarnLevel:
		return "warn"
	case ErrorLevel:
		return "error"
	case FatalLevel:
		return "fatal"
	case PanicLevel:
		return "panic"
	default:
		return "unknown"
	}
}

// zapLogger 基于zap的日志实现
type zapLogger struct {
	logger *zap.Logger
	config *Config
	level  zap.AtomicLevel
}

// Config 日志配置
type Config struct {
	Level      string `yaml:"level" json:"level"`             // 日志级别
	Format     string `yaml:"format" json:"format"`           // 日志格式: json, console
	OutputPath string `yaml:"output_path" json:"output_path"` // 输出路径

	// 文件轮转配置
	Rotation RotationConfig `yaml:"rotation" json:"rotation"`

	// 开发模式配置
	Development bool `yaml:"development" json:"development"`

	// 采样配置
	Sampling SamplingConfig `yaml:"sampling" json:"sampling"`
}

// RotationConfig 日志轮转配置
type RotationConfig struct {
	MaxSize    int  `yaml:"max_size" json:"max_size"`       // 单个文件最大大小(MB)
	MaxAge     int  `yaml:"max_age" json:"max_age"`         // 文件保留天数
	MaxBackups int  `yaml:"max_backups" json:"max_backups"` // 最大备份文件数
	Compress   bool `yaml:"compress" json:"compress"`       // 是否压缩
}

// SamplingConfig 采样配置
type SamplingConfig struct {
	Initial    int `yaml:"initial" json:"initial"`       // 初始采样率
	Thereafter int `yaml:"thereafter" json:"thereafter"` // 后续采样率
}

// 全局日志实例
var globalLogger Logger

// NewLogger 创建新的日志记录器
func NewLogger(config *Config) (Logger, error) {
	// 解析日志级别
	level := parseLevel(config.Level)
	atomicLevel := zap.NewAtomicLevelAt(level)

	// 构建编码器配置
	encoderConfig := buildEncoderConfig(config)

	// 构建编码器
	var encoder zapcore.Encoder
	switch config.Format {
	case "json":
		encoder = zapcore.NewJSONEncoder(encoderConfig)
	case "console":
		encoder = zapcore.NewConsoleEncoder(encoderConfig)
	default:
		encoder = zapcore.NewJSONEncoder(encoderConfig)
	}

	// 构建输出写入器
	writers, err := buildWriters(config)
	if err != nil {
		return nil, fmt.Errorf("failed to build writers: %w", err)
	}

	writeSyncer := zapcore.NewMultiWriteSyncer(writers...)

	// 构建核心
	core := zapcore.NewCore(encoder, writeSyncer, atomicLevel)

	// 添加采样
	if config.Sampling.Initial > 0 && config.Sampling.Thereafter > 0 {
		core = zapcore.NewSamplerWithOptions(
			core,
			time.Second,
			config.Sampling.Initial,
			config.Sampling.Thereafter,
		)
	}

	// 构建选项
	options := buildOptions(config)

	// 创建zap logger
	zapLog := zap.New(core, options...)

	return &zapLogger{
		logger: zapLog,
		config: config,
		level:  atomicLevel,
	}, nil
}

// parseLevel 解析日志级别字符串
func parseLevel(levelStr string) zapcore.Level {
	switch levelStr {
	case "debug":
		return zapcore.DebugLevel
	case "info":
		return zapcore.InfoLevel
	case "warn":
		return zapcore.WarnLevel
	case "error":
		return zapcore.ErrorLevel
	case "fatal":
		return zapcore.FatalLevel
	case "panic":
		return zapcore.PanicLevel
	default:
		return zapcore.InfoLevel
	}
}

// buildEncoderConfig 构建编码器配置
func buildEncoderConfig(config *Config) zapcore.EncoderConfig {
	encoderConfig := zap.NewProductionEncoderConfig()

	if config.Development {
		encoderConfig = zap.NewDevelopmentEncoderConfig()
	}

	// 时间格式
	encoderConfig.TimeKey = "timestamp"
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

	// 级别格式
	encoderConfig.LevelKey = "level"
	encoderConfig.EncodeLevel = zapcore.LowercaseLevelEncoder

	// 调用者格式
	encoderConfig.CallerKey = "caller"
	encoderConfig.EncodeCaller = zapcore.ShortCallerEncoder

	// 堆栈跟踪
	encoderConfig.StacktraceKey = "stacktrace"

	return encoderConfig
}

// buildWriters 构建输出写入器
func buildWriters(config *Config) ([]zapcore.WriteSyncer, error) {
	var writers []zapcore.WriteSyncer

	// 控制台输出
	writers = append(writers, zapcore.AddSync(os.Stdout))

	// 文件输出
	if config.OutputPath != "" {
		// 确保目录存在
		dir := filepath.Dir(config.OutputPath)
		if err := os.MkdirAll(dir, 0755); err != nil {
			return nil, fmt.Errorf("failed to create log directory: %w", err)
		}

		// 创建轮转写入器
		rotateWriter := &lumberjack.Logger{
			Filename:   config.OutputPath,
			MaxSize:    config.Rotation.MaxSize,
			MaxAge:     config.Rotation.MaxAge,
			MaxBackups: config.Rotation.MaxBackups,
			Compress:   config.Rotation.Compress,
		}

		writers = append(writers, zapcore.AddSync(rotateWriter))
	}

	return writers, nil
}

// buildOptions 构建zap选项
func buildOptions(config *Config) []zap.Option {
	var options []zap.Option

	// 添加调用者信息
	options = append(options, zap.AddCaller())
	options = append(options, zap.AddCallerSkip(1))

	// 开发模式选项
	if config.Development {
		options = append(options, zap.Development())
		options = append(options, zap.AddStacktrace(zapcore.ErrorLevel))
	} else {
		options = append(options, zap.AddStacktrace(zapcore.FatalLevel))
	}

	return options
}

// Debug 记录调试级别日志
func (l *zapLogger) Debug(msg string, fields ...Field) {
	l.logger.Debug(msg, convertFields(fields...)...)
}

// Info 记录信息级别日志
func (l *zapLogger) Info(msg string, fields ...Field) {
	l.logger.Info(msg, convertFields(fields...)...)
}

// Warn 记录警告级别日志
func (l *zapLogger) Warn(msg string, fields ...Field) {
	l.logger.Warn(msg, convertFields(fields...)...)
}

// Error 记录错误级别日志
func (l *zapLogger) Error(msg string, fields ...Field) {
	l.logger.Error(msg, convertFields(fields...)...)
}

// Fatal 记录致命错误日志并退出程序
func (l *zapLogger) Fatal(msg string, fields ...Field) {
	l.logger.Fatal(msg, convertFields(fields...)...)
}

// Panic 记录panic级别日志并触发panic
func (l *zapLogger) Panic(msg string, fields ...Field) {
	l.logger.Panic(msg, convertFields(fields...)...)
}

// With 添加字段到日志上下文
func (l *zapLogger) With(fields ...Field) Logger {
	return &zapLogger{
		logger: l.logger.With(convertFields(fields...)...),
		config: l.config,
		level:  l.level,
	}
}

// WithContext 添加上下文信息
func (l *zapLogger) WithContext(ctx context.Context) Logger {
	var fields []Field

	// 提取链路追踪ID
	if traceID := extractTraceID(ctx); traceID != "" {
		fields = append(fields, Field{Key: "trace_id", Value: traceID})
	}

	// 提取请求ID
	if requestID := extractRequestID(ctx); requestID != "" {
		fields = append(fields, Field{Key: "request_id", Value: requestID})
	}

	// 提取用户ID
	if userID := extractUserID(ctx); userID != "" {
		fields = append(fields, Field{Key: "user_id", Value: userID})
	}

	return l.With(fields...)
}

// SetLevel 动态设置日志级别
func (l *zapLogger) SetLevel(level Level) {
	zapLevel := convertLevel(level)
	l.level.SetLevel(zapLevel)
}

// GetLevel 获取当前日志级别
func (l *zapLogger) GetLevel() Level {
	zapLevel := l.level.Level()
	return convertZapLevel(zapLevel)
}

// Sync 同步日志缓冲区
func (l *zapLogger) Sync() error {
	return l.logger.Sync()
}

// convertFields 转换字段格式
func convertFields(fields ...Field) []zap.Field {
	zapFields := make([]zap.Field, len(fields))
	for i, field := range fields {
		zapFields[i] = zap.Any(field.Key, field.Value)
	}
	return zapFields
}

// convertLevel 转换日志级别
func convertLevel(level Level) zapcore.Level {
	switch level {
	case DebugLevel:
		return zapcore.DebugLevel
	case InfoLevel:
		return zapcore.InfoLevel
	case WarnLevel:
		return zapcore.WarnLevel
	case ErrorLevel:
		return zapcore.ErrorLevel
	case FatalLevel:
		return zapcore.FatalLevel
	case PanicLevel:
		return zapcore.PanicLevel
	default:
		return zapcore.InfoLevel
	}
}

// convertZapLevel 转换zap日志级别
func convertZapLevel(level zapcore.Level) Level {
	switch level {
	case zapcore.DebugLevel:
		return DebugLevel
	case zapcore.InfoLevel:
		return InfoLevel
	case zapcore.WarnLevel:
		return WarnLevel
	case zapcore.ErrorLevel:
		return ErrorLevel
	case zapcore.FatalLevel:
		return FatalLevel
	case zapcore.PanicLevel:
		return PanicLevel
	default:
		return InfoLevel
	}
}

// 上下文提取函数
func extractTraceID(ctx context.Context) string {
	if traceID := ctx.Value("trace_id"); traceID != nil {
		if id, ok := traceID.(string); ok {
			return id
		}
	}
	return ""
}

func extractRequestID(ctx context.Context) string {
	if requestID := ctx.Value("request_id"); requestID != nil {
		if id, ok := requestID.(string); ok {
			return id
		}
	}
	return ""
}

func extractUserID(ctx context.Context) string {
	if userID := ctx.Value("user_id"); userID != nil {
		if id, ok := userID.(string); ok {
			return id
		}
	}
	return ""
}

// 便利函数
func String(key, value string) Field {
	return Field{Key: key, Value: value}
}

func Int(key string, value int) Field {
	return Field{Key: key, Value: value}
}

func Int64(key string, value int64) Field {
	return Field{Key: key, Value: value}
}

func Float64(key string, value float64) Field {
	return Field{Key: key, Value: value}
}

func Bool(key string, value bool) Field {
	return Field{Key: key, Value: value}
}

func Time(key string, value time.Time) Field {
	return Field{Key: key, Value: value}
}

func Duration(key string, value time.Duration) Field {
	return Field{Key: key, Value: value}
}

func Error(err error) Field {
	return Field{Key: "error", Value: err}
}

func Any(key string, value interface{}) Field {
	return Field{Key: key, Value: value}
}

// 全局函数
func Init(config *Config) error {
	logger, err := NewLogger(config)
	if err != nil {
		return err
	}
	globalLogger = logger
	return nil
}

func GetLogger() Logger {
	if globalLogger == nil {
		// 使用默认配置初始化
		config := &Config{
			Level:       "info",
			Format:      "console",
			Development: true,
		}
		logger, _ := NewLogger(config)
		globalLogger = logger
	}
	return globalLogger
}

func Debug(msg string, fields ...Field) {
	GetLogger().Debug(msg, fields...)
}

func Info(msg string, fields ...Field) {
	GetLogger().Info(msg, fields...)
}

func Warn(msg string, fields ...Field) {
	GetLogger().Warn(msg, fields...)
}

func Error(msg string, fields ...Field) {
	GetLogger().Error(msg, fields...)
}

func Fatal(msg string, fields ...Field) {
	GetLogger().Fatal(msg, fields...)
}

func Panic(msg string, fields ...Field) {
	GetLogger().Panic(msg, fields...)
}

func With(fields ...Field) Logger {
	return GetLogger().With(fields...)
}

func WithContext(ctx context.Context) Logger {
	return GetLogger().WithContext(ctx)
}

func Sync() error {
	return GetLogger().Sync()
}

// DefaultConfig 返回默认配置
func DefaultConfig() *Config {
	return &Config{
		Level:       "info",
		Format:      "json",
		OutputPath:  "./logs/starseek.log",
		Development: false,
		Rotation: RotationConfig{
			MaxSize:    100, // 100MB
			MaxAge:     30,  // 30天
			MaxBackups: 10,  // 10个备份文件
			Compress:   true,
		},
		Sampling: SamplingConfig{
			Initial:    100,
			Thereafter: 100,
		},
	}
}

// DevelopmentConfig 返回开发环境配置
func DevelopmentConfig() *Config {
	return &Config{
		Level:       "debug",
		Format:      "console",
		OutputPath:  "",
		Development: true,
		Rotation: RotationConfig{
			MaxSize:    10,
			MaxAge:     7,
			MaxBackups: 3,
			Compress:   false,
		},
	}
}

// ProductionConfig 返回生产环境配置
func ProductionConfig() *Config {
	return &Config{
		Level:       "info",
		Format:      "json",
		OutputPath:  "/var/log/starseek/starseek.log",
		Development: false,
		Rotation: RotationConfig{
			MaxSize:    500, // 500MB
			MaxAge:     90,  // 90天
			MaxBackups: 30,  // 30个备份文件
			Compress:   true,
		},
		Sampling: SamplingConfig{
			Initial:    1000,
			Thereafter: 1000,
		},
	}
}

// CreateLoggerWithWriters 使用自定义写入器创建日志记录器
func CreateLoggerWithWriters(config *Config, writers ...io.Writer) (Logger, error) {
	// 解析日志级别
	level := parseLevel(config.Level)
	atomicLevel := zap.NewAtomicLevelAt(level)

	// 构建编码器配置
	encoderConfig := buildEncoderConfig(config)

	// 构建编码器
	var encoder zapcore.Encoder
	switch config.Format {
	case "json":
		encoder = zapcore.NewJSONEncoder(encoderConfig)
	case "console":
		encoder = zapcore.NewConsoleEncoder(encoderConfig)
	default:
		encoder = zapcore.NewJSONEncoder(encoderConfig)
	}

	// 构建写入器
	var writeSyncers []zapcore.WriteSyncer
	for _, writer := range writers {
		writeSyncers = append(writeSyncers, zapcore.AddSync(writer))
	}

	writeSyncer := zapcore.NewMultiWriteSyncer(writeSyncers...)

	// 构建核心
	core := zapcore.NewCore(encoder, writeSyncer, atomicLevel)

	// 构建选项
	options := buildOptions(config)

	// 创建zap logger
	zapLog := zap.New(core, options...)

	return &zapLogger{
		logger: zapLog,
		config: config,
		level:  atomicLevel,
	}, nil
}

//Personal.AI order the ending
