module github.com/turtacn/starseek

go 1.20

require (
	github.com/gin-gonic/gin v1.9.1
	github.com/go-redis/redis/v8 v8.11.5
	github.com/go-sql-driver/mysql v1.7.0
	github.com/google/uuid v1.6.0
	github.com/prometheus/client_golang v1.19.0
	github.com/spf13/viper v1.18.2
	github.com/stretchr/testify v1.9.0 // For testing assertions
	go.opentelemetry.io/otel v1.26.0
	go.opentelemetry.io/otel/exporters/jaeger v1.17.0
	go.opentelemetry.io/otel/exporters/otlp/otlptrace v1.26.0
	go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc v1.26.0
	go.opentelemetry.io/otel/exporters/stdout/stdouttrace v1.26.0
	go.opentelemetry.io/otel/sdk v1.26.0
	go.opentelemetry.io/otel/trace v1.26.0 // Explicitly include if not brought in by sdk
	go.uber.org/zap v1.27.0
)

// Indirect dependencies that will be added by `go mod tidy`
// We'll let `go mod tidy` fill these in. For example:
// require (
// 	go.opentelemetry.io/otel/metric v1.26.0 // Indirect
// 	go.opentelemetry.io/otel/propagation v1.26.0 // Indirect
// 	go.opentelemetry.io/otel/semconv/v1.24.0 v1.26.0 // Indirect
// 	golang.org/x/net v0.24.0 // Indirect
// 	google.golang.org/grpc v1.63.2 // Indirect
// 	google.golang.org/protobuf v1.33.0 // Indirect
//  ... many others
// )
