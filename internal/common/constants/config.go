package constants

import "time"

// Default configuration values
const (
	// Server defaults
	DefaultHTTPPort     = 8080
	DefaultGRPCPort     = 9090
	DefaultHTTPTimeout  = 30 * time.Second
	DefaultGRPCTimeout  = 30 * time.Second
	DefaultReadTimeout  = 10 * time.Second
	DefaultWriteTimeout = 10 * time.Second
	DefaultIdleTimeout  = 60 * time.Second

	// Database connection defaults
	DefaultDatabaseMaxIdleConns    = 10
	DefaultDatabaseMaxOpenConns    = 100
	DefaultDatabaseConnMaxLifetime = 30 * time.Minute
	DefaultDatabaseConnMaxIdleTime = 15 * time.Minute
	DefaultDatabaseQueryTimeout    = 30 * time.Second
	DefaultDatabasePingTimeout     = 5 * time.Second
	DefaultDatabaseRetryAttempts   = 3
	DefaultDatabaseRetryDelay      = 1 * time.Second

	// Cache defaults
	DefaultRedisTTL        = 1 * time.Hour
	DefaultRedisKeyPrefix  = "starseek:"
	DefaultRedisMaxRetries = 3
	DefaultRedisRetryDelay = 100 * time.Millisecond
	DefaultMemoryCacheSize = 100 * 1024 * 1024 // 100MB
	DefaultMemoryCacheTTL  = 5 * time.Minute

	// Search defaults
	DefaultSearchPageSize              = 20
	DefaultSearchMaxPageSize           = 1000
	DefaultSearchTimeout               = 30 * time.Second
	DefaultSearchMaxResults            = 10000
	DefaultSearchHighlightFragmentSize = 100
	DefaultSearchHighlightMaxFragments = 3

	// Index defaults
	DefaultIndexScanInterval  = 5 * time.Minute
	DefaultIndexMetadataTTL   = 1 * time.Hour
	DefaultIndexStatisticsTTL = 30 * time.Minute
	DefaultIndexBatchSize     = 1000

	// Task execution defaults
	DefaultTaskWorkerCount   = 10
	DefaultTaskQueueSize     = 1000
	DefaultTaskTimeout       = 60 * time.Second
	DefaultTaskRetryAttempts = 3
	DefaultTaskRetryDelay    = 5 * time.Second

	// Rate limiting defaults
	DefaultRateLimitRequests = 1000
	DefaultRateLimitWindow   = 1 * time.Minute
	DefaultRateLimitBurst    = 100

	// Monitoring defaults
	DefaultMetricsPath     = "/metrics"
	DefaultHealthPath      = "/health"
	DefaultReadyPath       = "/ready"
	DefaultMetricsInterval = 15 * time.Second
)

// Configuration file paths
const (
	DefaultConfigFile = "configs/config.yaml"
	DevConfigFile     = "configs/config.dev.yaml"
	ProdConfigFile    = "configs/config.prod.yaml"
	TestConfigFile    = "configs/config.test.yaml"
)

// Environment variable names
const (
	EnvConfigFile   = "STARSEEK_CONFIG_FILE"
	EnvEnvironment  = "STARSEEK_ENVIRONMENT"
	EnvHTTPPort     = "STARSEEK_HTTP_PORT"
	EnvGRPCPort     = "STARSEEK_GRPC_PORT"
	EnvLogLevel     = "STARSEEK_LOG_LEVEL"
	EnvDatabaseURL  = "STARSEEK_DATABASE_URL"
	EnvRedisURL     = "STARSEEK_REDIS_URL"
	EnvJWTSecret    = "STARSEEK_JWT_SECRET"
	EnvAPIKeySecret = "STARSEEK_API_KEY_SECRET"
)

// Database configuration constants
const (
	// StarRocks specific
	StarRocksDefaultPort     = 9030
	StarRocksDefaultUser     = "root"
	StarRocksDefaultDatabase = "starseek"
	StarRocksDefaultCharset  = "utf8mb4"

	// ClickHouse specific
	ClickHouseDefaultPort     = 9000
	ClickHouseDefaultHTTPPort = 8123
	ClickHouseDefaultUser     = "default"
	ClickHouseDefaultDatabase = "starseek"

	// Doris specific
	DorisDefaultPort     = 9030
	DorisDefaultUser     = "root"
	DorisDefaultDatabase = "starseek"
)

// Cache configuration constants
const (
	// Redis specific
	RedisDefaultDB                 = 0
	RedisDefaultPoolSize           = 10
	RedisDefaultMinIdleConns       = 5
	RedisDefaultDialTimeout        = 5 * time.Second
	RedisDefaultReadTimeout        = 3 * time.Second
	RedisDefaultWriteTimeout       = 3 * time.Second
	RedisDefaultPoolTimeout        = 4 * time.Second
	RedisDefaultIdleCheckFrequency = 60 * time.Second
	RedisDefaultIdleTimeout        = 5 * time.Minute
	RedisDefaultMaxConnAge         = 0 * time.Second

	// Memory cache specific
	MemoryCacheDefaultCleanupInterval  = 10 * time.Minute
	MemoryCacheDefaultEvictionInterval = 1 * time.Minute
)

// System limits
const (
	MaxQueryLength        = 1000
	MaxFieldsPerQuery     = 50
	MaxSortFields         = 10
	MaxFilterConditions   = 100
	MaxPageSize           = 1000
	MaxConcurrentQueries  = 1000
	MaxIndexesPerTable    = 20
	MaxTablesPerDatabase  = 1000
	MaxTokensPerQuery     = 100
	MaxHighlightLength    = 500
	MaxCacheKeyLength     = 250
	MaxErrorMessageLength = 1000
)

// Performance tuning constants
const (
	DefaultBatchSize          = 1000
	DefaultBufferSize         = 8192
	DefaultChannelBufferSize  = 1000
	DefaultWorkerPoolSize     = 10
	DefaultConnectionPoolSize = 20
	DefaultQueryConcurrency   = 5
	DefaultParallelism        = 4
)

// Configuration validation functions

// ValidatePort checks if port is in valid range
func ValidatePort(port int) bool {
	return port > 0 && port <= 65535
}

// ValidateTimeout checks if timeout is positive
func ValidateTimeout(timeout time.Duration) bool {
	return timeout > 0
}

// ValidatePageSize checks if page size is within limits
func ValidatePageSize(pageSize int) bool {
	return pageSize > 0 && pageSize <= MaxPageSize
}

// ValidatePoolSize checks if pool size is reasonable
func ValidatePoolSize(poolSize int) bool {
	return poolSize > 0 && poolSize <= 1000
}

// ValidateRetryAttempts checks if retry attempts is reasonable
func ValidateRetryAttempts(attempts int) bool {
	return attempts >= 0 && attempts <= 10
}

// ValidateTTL checks if TTL is positive
func ValidateTTL(ttl time.Duration) bool {
	return ttl > 0
}

// ValidateQueryLength checks if query length is within limits
func ValidateQueryLength(query string) bool {
	return len(query) > 0 && len(query) <= MaxQueryLength
}

// ValidateConcurrency checks if concurrency is reasonable
func ValidateConcurrency(concurrency int) bool {
	return concurrency > 0 && concurrency <= 100
}

// Configuration categories
const (
	ConfigCategoryServer     = "server"
	ConfigCategoryDatabase   = "database"
	ConfigCategoryCache      = "cache"
	ConfigCategoryLogging    = "logging"
	ConfigCategoryMonitoring = "monitoring"
	ConfigCategorySecurity   = "security"
	ConfigCategorySearch     = "search"
	ConfigCategoryIndex      = "index"
	ConfigCategoryTask       = "task"
)

// Default configuration groups
var (
	DefaultServerConfig = map[string]interface{}{
		"http_port":     DefaultHTTPPort,
		"grpc_port":     DefaultGRPCPort,
		"read_timeout":  DefaultReadTimeout,
		"write_timeout": DefaultWriteTimeout,
		"idle_timeout":  DefaultIdleTimeout,
	}

	DefaultDatabaseConfig = map[string]interface{}{
		"max_idle_conns":     DefaultDatabaseMaxIdleConns,
		"max_open_conns":     DefaultDatabaseMaxOpenConns,
		"conn_max_lifetime":  DefaultDatabaseConnMaxLifetime,
		"conn_max_idle_time": DefaultDatabaseConnMaxIdleTime,
		"query_timeout":      DefaultDatabaseQueryTimeout,
		"ping_timeout":       DefaultDatabasePingTimeout,
		"retry_attempts":     DefaultDatabaseRetryAttempts,
		"retry_delay":        DefaultDatabaseRetryDelay,
	}

	DefaultCacheConfig = map[string]interface{}{
		"redis_ttl":         DefaultRedisTTL,
		"redis_key_prefix":  DefaultRedisKeyPrefix,
		"redis_max_retries": DefaultRedisMaxRetries,
		"redis_retry_delay": DefaultRedisRetryDelay,
		"memory_cache_size": DefaultMemoryCacheSize,
		"memory_cache_ttl":  DefaultMemoryCacheTTL,
	}

	DefaultSearchConfig = map[string]interface{}{
		"page_size":               DefaultSearchPageSize,
		"max_page_size":           DefaultSearchMaxPageSize,
		"timeout":                 DefaultSearchTimeout,
		"max_results":             DefaultSearchMaxResults,
		"highlight_fragment_size": DefaultSearchHighlightFragmentSize,
		"highlight_max_fragments": DefaultSearchHighlightMaxFragments,
	}
)

//Personal.AI order the ending
