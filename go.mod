module github.com/turtacn/starseek

go 1.20.2

require (
	// Web框架 - Gin HTTP web framework
	github.com/gin-gonic/gin v1.9.1

	// Redis客户端 - Redis client for Go
	github.com/go-redis/redis/v8 v8.11.5

	// 日志系统 - Structured logger for Go
	github.com/sirupsen/logrus v1.9.3

	// 配置管理 - Configuration management
	github.com/spf13/viper v1.16.0

	// 监控指标 - Prometheus metrics
	github.com/prometheus/client_golang v1.16.0

	// 数据库驱动 - Database drivers
	github.com/go-sql-driver/mysql v1.7.1        // MySQL/StarRocks driver
	github.com/ClickHouse/clickhouse-go/v2 v2.12.1 // ClickHouse driver
	github.com/lib/pq v1.10.9                     // PostgreSQL driver for Doris

	// JWT认证 - JWT authentication
	github.com/golang-jwt/jwt/v5 v5.0.0

	// UUID生成 - UUID generation
	github.com/google/uuid v1.3.0

	// 数据验证 - Data validation
	github.com/go-playground/validator/v10 v10.14.1

	// JSON处理 - JSON utilities
	github.com/json-iterator/go v1.1.12

	// 分词器 - Chinese tokenizer
	github.com/yanyiwu/gojieba v1.3.0

	// 时间处理 - Time utilities
	github.com/golang-module/carbon/v2 v2.2.3

	// HTTP客户端 - HTTP client
	github.com/go-resty/resty/v2 v2.7.0

	// 缓存 - In-memory cache
	github.com/patrickmn/go-cache v2.1.0+incompatible

	// 并发控制 - Concurrency control
	golang.org/x/sync v0.3.0

	// 上下文 - Context utilities
	golang.org/x/net v0.12.0

	// 加密 - Cryptography
	golang.org/x/crypto v0.11.0

	// 文本处理 - Text processing
	golang.org/x/text v0.11.0

	// 限流器 - Rate limiter
	go.uber.org/ratelimit v0.3.0

	// 链路追踪 - Distributed tracing
	go.opentelemetry.io/otel v1.16.0
	go.opentelemetry.io/otel/trace v1.16.0
	go.opentelemetry.io/otel/exporters/jaeger v1.16.0

	// 协程池 - Goroutine pool
	github.com/panjf2000/ants/v2 v2.8.1

	// 位图操作 - Bitmap operations
	github.com/RoaringBitmap/roaring v1.3.0

	// 配置文件格式支持 - Configuration formats
	gopkg.in/yaml.v3 v3.0.1
)

require (
	// 测试相关依赖 - Testing dependencies
	github.com/stretchr/testify v1.8.4
	github.com/golang/mock v1.6.0
	github.com/onsi/ginkgo/v2 v2.11.0
	github.com/onsi/gomega v1.27.8

	// 基准测试 - Benchmarking
	github.com/pkg/profile v1.7.0
)

require (
	// Gin框架间接依赖 - Gin framework indirect dependencies
	github.com/bytedance/sonic v1.9.1 // indirect
	github.com/chenzhuoyu/base64x v0.0.0-20221115062448-fe3a3abad311 // indirect
	github.com/gabriel-vasile/mimetype v1.4.2 // indirect
	github.com/gin-contrib/sse v0.1.0 // indirect
	github.com/go-playground/locales v0.14.1 // indirect
	github.com/go-playground/universal-translator v0.18.1 // indirect
	github.com/goccy/go-json v0.10.2 // indirect
	github.com/klauspost/cpuid/v2 v2.2.4 // indirect
	github.com/leodido/go-urn v1.2.4 // indirect
	github.com/mattn/go-isatty v0.0.19 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.2 // indirect
	github.com/pelletier/go-toml/v2 v2.0.8 // indirect
	github.com/twitchyliquid64/golang-asm v0.15.1 // indirect
	github.com/ugorji/go/codec v1.2.11 // indirect
	golang.org/x/arch v0.3.0 // indirect

	// Redis客户端间接依赖 - Redis client indirect dependencies
	github.com/cespare/xxhash/v2 v2.2.0 // indirect
	github.com/dgryski/go-rendezvous v0.0.0-20200823014737-9f7001d12a5f // indirect

	// Logrus间接依赖 - Logrus indirect dependencies
	golang.org/x/sys v0.10.0 // indirect

	// Viper间接依赖 - Viper indirect dependencies
	github.com/fsnotify/fsnotify v1.6.0 // indirect
	github.com/hashicorp/hcl v1.0.0 // indirect
	github.com/magiconair/properties v1.8.7 // indirect
	github.com/mitchellh/mapstructure v1.5.0 // indirect
	github.com/spf13/afero v1.9.5 // indirect
	github.com/spf13/cast v1.5.1 // indirect
	github.com/spf13/jwalterweatherman v1.1.0 // indirect
	github.com/spf13/pflag v1.0.5 // indirect
	github.com/subosito/gotenv v1.4.2 // indirect
	gopkg.in/ini.v1 v1.67.0 // indirect

	// Prometheus间接依赖 - Prometheus indirect dependencies
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/matttproud/golang_protobuf_extensions v1.0.4 // indirect
	github.com/prometheus/client_model v0.4.0 // indirect
	github.com/prometheus/common v0.44.0 // indirect
	github.com/prometheus/procfs v0.11.0 // indirect
	google.golang.org/protobuf v1.31.0 // indirect

	// ClickHouse驱动间接依赖 - ClickHouse driver indirect dependencies
	github.com/ClickHouse/ch-go v0.58.2 // indirect
	github.com/andybalholm/brotli v1.0.5 // indirect
	github.com/go-faster/city v1.0.1 // indirect
	github.com/go-faster/errors v0.6.1 // indirect
	github.com/klauspost/compress v1.16.7 // indirect
	github.com/paulmach/orb v0.10.0 // indirect
	github.com/pierrec/lz4/v4 v4.1.18 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/segmentio/asm v1.2.0 // indirect
	github.com/shopspring/decimal v1.3.1 // indirect
	go.opentelemetry.io/otel/metric v1.16.0 // indirect

	// OpenTelemetry间接依赖 - OpenTelemetry indirect dependencies
	go.opentelemetry.io/otel/sdk v1.16.0 // indirect
	go.uber.org/atomic v1.11.0 // indirect
	go.uber.org/multierr v1.11.0 // indirect

	// 测试框架间接依赖 - Testing framework indirect dependencies
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	gopkg.in/check.v1 v1.0.0-20201130134442-10cb98267c6c // indirect
)

// 替换指令 - Replace directives for local development
// replace github.com/turtacn/starseek => ./

//Personal.AI order the ending
