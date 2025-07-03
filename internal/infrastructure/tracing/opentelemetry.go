package tracing

import (
	"context"
	"fmt"
	"time"

	// OpenTelemetry SDK 核心
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc" // OTLP gRPC 导出器
	"go.opentelemetry.io/otel/propagation"                            // 传播器
	"go.opentelemetry.io/otel/sdk/resource"                           // 资源
	sdktrace "go.opentelemetry.io/otel/sdk/trace"                     // 追踪 SDK
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"                // 语义约定
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	// OpenTelemetry 贡献库（用于常见框架和库的自动化插桩）
	otelgin "go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"       // Gin 中间件
	otelredis "go.opentelemetry.io/contrib/instrumentation/github.com/go-redis/redis/otelredis"  // Redis 客户端
	otelsql "go.opentelemetry.io/contrib/instrumentation/go.opentelemetry.io/otel/trace/otelsql" // database/sql 驱动

	// 项目内部依赖
	ferrors "github.com/turtacn/starseek/internal/common/errors"
	"github.com/turtacn/starseek/internal/infrastructure/logger" // 引入日志包
)

// 全局的 TracerProvider，用于创建 Tracer 实例。
// 它应该在应用启动时被初始化一次。
var globalTracerProvider *sdktrace.TracerProvider

// InitOpenTelemetry 初始化 OpenTelemetry SDK，包括 TracerProvider 和 SpanExporter。
//
// serviceName: 您的应用程序或服务的名称，将作为资源属性附加到所有 Span。
// collectorEndpoint: OpenTelemetry Collector 的 gRPC 端点 (例如 "localhost:4317")。
// log: 用于记录初始化过程中错误的日志接口。
func InitOpenTelemetry(serviceName string, collectorEndpoint string, log logger.Logger) (func(context.Context) error, error) {
	// 1. 创建 OTLP gRPC Span 导出器
	// 使用 WithInsecure() 禁用 TLS，在生产环境中应使用 WithTransportCredentials 或 WithTLSCredentials。
	// WithEndpoint() 指定 Collector 的地址。
	// WithTimeout() 设置连接和发送的超时时间。
	conn, err := grpc.DialContext(context.Background(), collectorEndpoint,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(), // 阻塞直到连接成功或超时
		grpc.WithTimeout(5*time.Second),
	)
	if err != nil {
		log.Error(fmt.Sprintf("Failed to create gRPC connection to OpenTelemetry Collector: %v", err),
			zap.String("collector_endpoint", collectorEndpoint))
		return nil, ferrors.NewInternalError(fmt.Sprintf("failed to connect to OTLP collector: %v", err), err)
	}

	exporter, err := otlptracegrpc.New(context.Background(), otlptracegrpc.WithGRPCConn(conn))
	if err != nil {
		log.Error(fmt.Sprintf("Failed to create OTLP gRPC exporter: %v", err))
		return nil, ferrors.NewInternalError(fmt.Sprintf("failed to create OTLP exporter: %v", err), err)
	}

	// 2. 创建资源（Resource），用于标识您的服务。
	// 服务名称 (ServiceName) 是最重要的属性。
	r, err := resource.New(context.Background(),
		resource.WithAttributes(
			semconv.ServiceNameKey.String(serviceName),
			semconv.ServiceVersionKey.String("1.0.0"), // 可以从应用配置中获取版本
			// 其他有用的属性，例如部署环境，主机名等
			// attribute.String("environment", "production"),
			// attribute.String("host.name", "my-server-01"),
		),
	)
	if err != nil {
		log.Error(fmt.Sprintf("Failed to create resource: %v", err))
		return nil, ferrors.NewInternalError(fmt.Sprintf("failed to create OpenTelemetry resource: %v", err), err)
	}

	// 3. 创建 TracerProvider
	// BatchSpanProcessor 会异步地批处理和发送 Span，提高性能。
	globalTracerProvider = sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),                // 使用批处理导出器
		sdktrace.WithResource(r),                      // 附加资源信息
		sdktrace.WithSampler(sdktrace.AlwaysSample()), // 采样策略：始终采样所有 Span
		// sdktrace.WithSampler(sdktrace.ParentBased(sdktrace.TraceIDRatioBased(0.5))), // 基于父Span和50%概率采样
	)

	// 4. 设置全局 TracerProvider 和 Propagator
	// 这使得所有使用 otel.Tracer() 的代码都能获取到配置好的 Tracer。
	// TextMapPropagator 用于在不同服务间传播追踪上下文 (例如 HTTP Headers)。
	otel.SetTracerProvider(globalTracerProvider)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{}, // W3C Trace Context (推荐)
		propagation.Baggage{},      // W3C Baggage
	))

	log.Info(fmt.Sprintf("OpenTelemetry initialized for service: %s, collector: %s", serviceName, collectorEndpoint))

	// 返回一个 shutdown 函数，用于在应用关闭时优雅地关闭 TracerProvider 和 exporter。
	return func(ctx context.Context) error {
		shutdownCtx, cancel := context.WithTimeout(ctx, 5*time.Second) // 留5秒钟让 Span 发送完成
		defer cancel()

		if err := globalTracerProvider.Shutdown(shutdownCtx); err != nil {
			log.Error(fmt.Sprintf("Failed to shutdown OpenTelemetry TracerProvider: %v", err))
			return ferrors.NewInternalError(fmt.Sprintf("failed to shutdown OpenTelemetry: %v", err), err)
		}
		if err := conn.Close(); err != nil {
			log.Error(fmt.Sprintf("Failed to close gRPC connection to OpenTelemetry Collector: %v", err))
			return ferrors.NewInternalError(fmt.Sprintf("failed to close OTLP gRPC connection: %v", err), err)
		}
		log.Info("OpenTelemetry gracefully shut down.")
		return nil
	}, nil
}

// NewTracer 返回一个 OpenTelemetry Tracer 实例。
//
// instrumentationName: 通常是调用 Tracer 的模块或库的名称，例如 "github.com/turtacn/starseek/internal/api/http"
// 或 "github.com/turtacn/starseek/internal/application/service"。
func NewTracer(instrumentationName string) otel.Tracer {
	// 从全局 TracerProvider 获取 Tracer 实例。
	return otel.Tracer(instrumentationName)
}

// GinMiddleware 返回一个 Gin 中间件，用于自动追踪 HTTP 请求。
func GinMiddleware(serviceName string) gin.HandlerFunc {
	// otelgin.Middleware 会自动创建并结束 Span，并将 Span 上下文注入 Gin Context。
	return otelgin.Middleware(serviceName)
}

// WrapSQLDriverName 用于包装 SQL 驱动名称，使其能够被 OpenTelemetry 追踪。
// 示例: otelsql.Open("postgres", dbDSN, otelsql.WithTracerProvider(otel.GetTracerProvider()))
// 注意：这只包装了驱动，实际的数据库连接还需要使用 sql.Open 函数
func WrapSQLDriverName(driverName string) string {
	return otelsql.OpenDB(driverName).Driver().String() // otelsql.OpenDB(driverName) 返回一个 *sql.DB，其 Driver() 返回包装后的 DriverName
}

// NewRedisHook 返回一个 Redis Hook，用于自动追踪 Redis 操作。
//
// tracerProvider: 用于创建 Tracer 的 TracerProvider。
// hookOptions: 可选的配置选项，例如自定义 Span 名称。
//
// 使用方法:
// rdb := redis.NewClient(...)
// rdb.AddHook(NewRedisHook(otel.GetTracerProvider()))
func NewRedisHook(tracerProvider *sdktrace.TracerProvider, hookOptions ...otelredis.Option) otelredis.TracingHook {
	return otelredis.NewTracingHook(hookOptions...)
}
