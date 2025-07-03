package logger

import (
	"context"
	"fmt"
	"os"
	"strings"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/turtacn/starseek/internal/common/config"         // 引入配置包
	ferrors "github.com/turtacn/starseek/internal/common/errors" // 引入自定义错误包
	"github.com/turtacn/starseek/internal/common/types"          // 引入上下文类型
)

// Logger 接口定义了项目通用的日志方法。
// 所有的日志记录都应该通过此接口进行，便于未来切换底层日志库。
type Logger interface {
	Debug(msg string, fields ...zap.Field)
	Info(msg string, fields ...zap.Field)
	Warn(msg string, fields ...zap.Field)
	Error(msg string, fields ...zap.Field)
	Fatal(msg string, fields ...zap.Field) // Fatal 会在记录日志后调用 os.Exit(1)

	// WithContext 从 context 中提取请求ID等信息，返回一个带有这些字段的子 Logger。
	// 这对于将请求特定信息附加到日志中非常有用。
	WithContext(ctx context.Context) Logger

	// With 返回一个带有额外字段的子 Logger。
	With(fields ...zap.Field) Logger
}

// logger 是 Logger 接口的 Zap 实现。
// 内部嵌入了 *zap.Logger 以便直接使用其方法，同时提供自定义的封装和扩展。
type logger struct {
	*zap.Logger
}

// NewZapLogger 根据 LoggerConfig 初始化并返回一个 Logger 接口的实例。
func NewZapLogger(cfg *config.LoggerConfig) (Logger, error) {
	// 1. 解析日志级别
	logLevel, err := zapcore.ParseLevel(cfg.Level)
	if err != nil {
		return nil, ferrors.NewConfigError(fmt.Sprintf("invalid logger level '%s': %v", cfg.Level, err), err)
	}

	// 2. 配置编码器 (Encoder) - JSON 或 Console (文本)
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder        // 时间格式：2006-01-02T15:04:05.000Z0700
	encoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder // 控制台带颜色，JSON不影响
	encoderConfig.TimeKey = "timestamp"                          // 时间字段名
	encoderConfig.LevelKey = "level"                             // 级别字段名
	encoderConfig.CallerKey = "caller"                           // 调用者字段名
	encoderConfig.MessageKey = "message"                         // 消息字段名
	encoderConfig.StacktraceKey = "stacktrace"                   // 堆栈字段名

	var encoder zapcore.Encoder
	switch strings.ToLower(cfg.Format) {
	case "json":
		encoder = zapcore.NewJSONEncoder(encoderConfig)
	case "text":
		encoder = zapcore.NewConsoleEncoder(encoderConfig)
	default:
		return nil, ferrors.NewConfigError(fmt.Sprintf("invalid logger format '%s': must be 'json' or 'text'", cfg.Format), nil)
	}

	// 3. 配置输出目标 (WriteSyncer)
	var writeSyncer zapcore.WriteSyncer
	if strings.ToLower(cfg.OutputPath) == "stdout" {
		writeSyncer = zapcore.AddSync(os.Stdout)
	} else if strings.ToLower(cfg.OutputPath) == "stderr" {
		writeSyncer = zapcore.AddSync(os.Stderr)
	} else {
		// 写入文件
		// 这里可以使用 lumberjack 库来实现日志切割、压缩等功能
		// 例如：import "gopkg.in/natefinch/lumberjack.v2"
		// w := zapcore.AddSync(&lumberjack.Logger{
		// 	Filename:   cfg.OutputPath,
		// 	MaxSize:    100, // megabytes
		// 	MaxBackups: 3,
		// 	MaxAge:     28, // days
		// 	Compress:   true, // compress rotated files
		// })
		// writeSyncer = w
		//
		// 为简单起见，这里直接使用 os.OpenFile，生产环境强烈建议使用 lumberjack
		file, err := os.OpenFile(cfg.OutputPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return nil, ferrors.NewConfigError(fmt.Sprintf("failed to open log file '%s': %v", cfg.OutputPath, err), err)
		}
		writeSyncer = zapcore.AddSync(file)
	}

	// 4. 创建 Core
	core := zapcore.NewCore(
		encoder,
		writeSyncer,
		logLevel, // 实际的日志级别
	)

	// 5. 构建 Zap Logger
	// AddCaller() 报告调用者 (文件名:行号)
	// AddStacktrace() 在 Error 或更高等级记录堆栈信息
	// Development() 开启开发模式，更易读的输出，但性能稍差
	// Build() 创建 Logger
	zapLogger := zap.New(core, zap.AddCaller(), zap.AddStacktrace(zap.ErrorLevel))

	// zap.ReplaceGlobals(zapLogger) // 可以选择替换全局的zap logger，但不推荐

	return &logger{zapLogger}, nil
}

// Ensure logger implements Logger interface
var _ Logger = (*logger)(nil)

// Debug 实现了 Logger 接口的 Debug 方法。
func (l *logger) Debug(msg string, fields ...zap.Field) {
	l.Logger.Debug(msg, fields...)
}

// Info 实现了 Logger 接口的 Info 方法。
func (l *logger) Info(msg string, fields ...zap.Field) {
	l.Logger.Info(msg, fields...)
}

// Warn 实现了 Logger 接口的 Warn 方法。
func (l *logger) Warn(msg string, fields ...zap.Field) {
	l.Logger.Warn(msg, fields...)
}

// Error 实现了 Logger 接口的 Error 方法。
func (l *logger) Error(msg string, fields ...zap.Field) {
	l.Logger.Error(msg, fields...)
}

// Fatal 实现了 Logger 接口的 Fatal 方法。
func (l *logger) Fatal(msg string, fields ...zap.Field) {
	l.Logger.Fatal(msg, fields...)
}

// WithContext 从 context 中提取请求ID等信息，返回一个带有这些字段的子 Logger。
func (l *logger) WithContext(ctx context.Context) Logger {
	fields := []zap.Field{}

	// 尝试从 Context 中获取 RequestID
	if requestID := types.GetRequestIDFromContext(ctx); requestID != "" {
		fields = append(fields, zap.String("request_id", requestID))
	}

	// 尝试从 Context 中获取 AuthUserID
	if userID := types.GetAuthUserIDFromContext(ctx); userID != "" {
		fields = append(fields, zap.String("user_id", userID))
	}

	// 返回一个新的 logger 实例，带有这些字段
	return l.With(fields...)
}

// With 返回一个带有额外字段的子 Logger。
func (l *logger) With(fields ...zap.Field) Logger {
	return &logger{l.Logger.With(fields...)}
}
