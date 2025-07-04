package constants

import (
	"fmt"
	"time"
)

// =============================================================================
// Configuration File Paths
// =============================================================================

const (
	// Default configuration file paths
	DefaultConfigFile        = "configs/config.yaml"
	DefaultDevConfigFile     = "configs/config.dev.yaml"
	DefaultProdConfigFile    = "configs/config.prod.yaml"
	DefaultTestConfigFile    = "configs/config.test.yaml"

	// Configuration directory
	DefaultConfigDir         = "configs"

	// Log configuration file
	DefaultLogConfigFile     = "configs/log.yaml"
)

// =============================================================================
// Environment Variable Names
// =============================================================================

const (
	// Application environment variables
	EnvVarConfigFile         = "STARSEEK_CONFIG_FILE"
	EnvVarEnvironment        = "STARSEEK_ENVIRONMENT"
	EnvVarLogLevel           = "STARSEEK_LOG_LEVEL"
	EnvVarServerPort         = "STARSEEK_SERVER_PORT"

	// Database environment variables
	EnvVarDatabaseType       = "STARSEEK_DATABASE_TYPE"
	EnvVarDatabaseHost       = "STARSEEK_DATABASE_HOST"
	EnvVarDatabasePort       = "STARSEEK_DATABASE_PORT"
	EnvVarDatabaseUser       = "STARSEEK_DATABASE_USER"
	EnvVarDatabasePassword   = "STARSEEK_DATABASE_PASSWORD"
	EnvVarDatabaseName       = "STARSEEK_DATABASE_NAME"

	// Cache environment variables
	EnvVarCacheType          = "STARSEEK_CACHE_TYPE"
	EnvVarRedisHost          = "STARSEEK_REDIS_HOST"
	EnvVarRedisPort          = "STARSEEK_REDIS_PORT"
	EnvVarRedisPassword      = "STARSEEK_REDIS_PASSWORD"
	EnvVarRedisDatabase      = "STARSEEK_REDIS_DATABASE"

	// Security environment variables
	EnvVarJWTSecret          = "STARSEEK_JWT_SECRET"
	EnvVarAPIKey             = "STARSEEK_API_KEY"
	EnvVarTLSEnabled         = "STARSEEK_TLS_ENABLED"
	EnvVarTLSCertFile        = "STARSEEK_TLS_CERT_FILE"
	EnvVarTLSKeyFile         = "STARSEEK_TLS_KEY_FILE"
)

// =============================================================================
// Server Configuration Constants
// =============================================================================

const (
	// Default server settings
	DefaultServerPort            = 8080
	DefaultServerHost            = "0.0.0.0"
	DefaultGRPCPort             = 9090
	DefaultAdminPort            = 8081
	DefaultMetricsPort          = 9091

	// Server timeout settings
	DefaultReadTimeout          = 30 * time.Second
	DefaultWriteTimeout         = 30 * time.Second
	DefaultIdleTimeout          = 120 * time.Second
	DefaultShutdownTimeout      = 30 * time.Second
	DefaultRequestTimeout       = 60 * time.Second

	// Server limits
	DefaultMaxHeaderBytes       = 1024 * 1024      // 1MB
	DefaultMaxRequestSize       = 10 * 1024 * 1024 // 10MB
	DefaultMaxConcurrentRequest = 1000

	// Keep-alive settings
	DefaultKeepAliveTimeout     = 60 * time.Second
	DefaultKeepAliveInterval    = 30 * time.Second
)

// =============================================================================
// Database Configuration Constants
// =============================================================================

const (
	// Connection pool settings
	DefaultDBMaxOpenConns       = 100
	DefaultDBMaxIdleConns       = 50
	DefaultDBConnMaxLifetime    = 1 * time.Hour
	DefaultDBConnMaxIdleTime    = 30 * time.Minute

	// Database timeout settings
	DefaultDBConnectTimeout     = 10 * time.Second
	DefaultDBQueryTimeout       = 30 * time.Second
	DefaultDBTransactionTimeout = 60 * time.Second

	// Database retry settings
	DefaultDBMaxRetries         = 3
	DefaultDBRetryInterval      = 1 * time.Second
	DefaultDBRetryBackoff       = 2.0

	// Database batch settings
	DefaultDBBatchSize          = 1000
	DefaultDBMaxBatchSize       = 10000
	DefaultDBBatchTimeout       = 30 * time.Second

	// Database monitoring
	DefaultDBHealthCheckInterval = 30 * time.Second
	DefaultDBSlowQueryThreshold  = 5 * time.Second
)

// =============================================================================
// Cache Configuration Constants
// =============================================================================

const (
	// Redis connection settings
	DefaultRedisPort            = 6379
	DefaultRedisDatabase        = 0
	DefaultRedisPoolSize        = 100
	DefaultRedisMinIdleConns    = 10
	DefaultRedisMaxRetries      = 3
	DefaultRedisRetryDelay      = 100 * time.Millisecond

	// Redis timeout settings
	DefaultRedisDialTimeout     = 5 * time.Second
	DefaultRedisReadTimeout     = 3 * time.Second
	DefaultRedisWriteTimeout    = 3 * time.Second
	DefaultRedisPoolTimeout     = 4 * time.Second
	DefaultRedisIdleTimeout     = 5 * time.Minute

	// Cache TTL settings
	DefaultCacheTTL             = 1 * time.Hour
	DefaultSearchResultTTL      = 30 * time.Minute
	DefaultIndexMetadataTTL     = 24 * time.Hour
	DefaultTokenizationTTL      = 6 * time.Hour
	DefaultRankingCacheTTL      = 2 * time.Hour

	// Memory cache settings
	DefaultMemoryCacheSize      = 100 * 1024 * 1024 // 100MB
	DefaultMemoryCacheItems     = 10000
	DefaultMemoryCacheEvictionInterval = 10 * time.Minute
)

// =============================================================================
// API Configuration Constants
// =============================================================================

const (
	// API versioning
	DefaultAPIVersion           = "v1"
	DefaultAPIPrefix            = "/api"

	// Request limits
	DefaultMaxQueryLength       = 1000
	DefaultMaxFilterConditions  = 50
	DefaultMaxSortFields        = 10
	DefaultMaxResultSize        = 10000
	DefaultMaxSearchFields      = 20

	// Pagination settings
	DefaultPageSize             = 20
	DefaultMaxPageSize          = 1000
	DefaultMaxPageNumber        = 10000

	// Rate limiting
	DefaultRateLimit            = 1000    // requests per minute
	DefaultRateLimitWindow      = 1 * time.Minute
	DefaultRateLimitBurst       = 100

	// API timeout settings
	DefaultAPITimeout           = 30 * time.Second
	DefaultAPIKeepAlive         = 30 * time.Second
)

// =============================================================================
// Search Configuration Constants
// =============================================================================

const (
	// Query processing limits
	DefaultMaxSearchTerms       = 100
	DefaultMaxBooleanClauses    = 50
	DefaultMaxNestedDepth       = 10
	DefaultMaxWildcardTerms     = 10

	// Search performance settings
	DefaultSearchTimeout        = 30 * time.Second
	DefaultIndexScanLimit       = 1000000
	DefaultSearchConcurrency    = 10
	DefaultSearchBatchSize      = 1000

	// Ranking settings
	DefaultRankingAlgorithm     = "BM25"
	DefaultMaxRankingFactors    = 20
	DefaultRankingCacheSize     = 10000

	// Highlighting settings
	DefaultHighlightFragmentSize = 100
	DefaultMaxHighlightFragments = 10
	DefaultHighlightPreTag      = "<em>"
	DefaultHighlightPostTag     = "</em>"

	// Suggestion settings
	DefaultMaxSuggestions       = 10
	DefaultSuggestionMinLength  = 2
	DefaultSuggestionMaxLength  = 50
)

// =============================================================================
// Indexing Configuration Constants
// =============================================================================

const (
	// Index management settings
	DefaultMaxIndexes           = 1000
	DefaultMaxColumnsPerIndex   = 100
	DefaultIndexRefreshInterval = 5 * time.Minute
	DefaultIndexStatisticsInterval = 1 * time.Hour

	// Tokenization settings
	DefaultMaxTokenLength       = 50
	DefaultMaxTokensPerDocument = 10000
	DefaultTokenizationTimeout  = 10 * time.Second

	// Index validation
	DefaultIndexValidationTimeout = 30 * time.Second
	DefaultMaxIndexValidationRetries = 3
)

// =============================================================================
// Logging Configuration Constants
// =============================================================================

const (
	// Log levels
	DefaultLogLevel             = "info"
	DefaultLogFormat            = "json"
	DefaultLogOutput            = "stdout"

	// Log file settings
	DefaultLogMaxSize           = 100    // MB
	DefaultLogMaxBackups        = 10
	DefaultLogMaxAge            = 30     // days
	DefaultLogCompress          = true

	// Structured logging
	DefaultLogWithCaller        = true
	DefaultLogWithStackTrace    = false
	DefaultLogTimestampFormat   = "2006-01-02T15:04:05.000Z07:00"
)

// =============================================================================
// Monitoring Configuration Constants
// =============================================================================

const (
	// Metrics collection
	DefaultMetricsEnabled       = true
	DefaultMetricsInterval      = 15 * time.Second
	DefaultMetricsPath          = "/metrics"

	// Health check settings
	DefaultHealthCheckPath      = "/health"
	DefaultReadinessCheckPath   = "/ready"
	DefaultLivenessCheckPath    = "/live"
	DefaultHealthCheckTimeout   = 5 * time.Second
	DefaultHealthCheckInterval  = 30 * time.Second

	// Performance monitoring
	DefaultSlowRequestThreshold = 5 * time.Second
	DefaultMaxMetricsHistory    = 1000
	DefaultMetricsRetention     = 24 * time.Hour
)

// =============================================================================
// Security Configuration Constants
// =============================================================================

const (
	// Authentication settings
	DefaultJWTExpiration        = 24 * time.Hour
	DefaultJWTRefreshWindow     = 6 * time.Hour
	DefaultAPIKeyLength         = 32
	DefaultPasswordMinLength    = 8

	// TLS settings
	DefaultTLSEnabled           = false
	DefaultTLSMinVersion        = "1.2"
	DefaultTLSCipherSuites      = "ECDHE-RSA-AES128-GCM-SHA256,ECDHE-RSA-AES256-GCM-SHA384"

	// CORS settings
	DefaultCORSEnabled          = true
	DefaultCORSMaxAge           = 24 * time.Hour
	DefaultCORSAllowCredentials = false
)

// =============================================================================
// Performance Tuning Constants
// =============================================================================

const (
	// Concurrency settings
	DefaultWorkerPoolSize       = 100
	DefaultMaxGoroutines        = 10000
	DefaultChannelBufferSize    = 1000

	// Memory settings
	DefaultMaxMemoryUsage       = 8 * 1024 * 1024 * 1024 // 8GB
	DefaultGCTargetPercentage   = 100
	DefaultGCMaxPause           = 10 * time.Millisecond

	// Network settings
	DefaultMaxIdleConnsPerHost  = 100
	DefaultIdleConnTimeout      = 90 * time.Second
	DefaultTLSHandshakeTimeout  = 10 * time.Second
	DefaultExpectContinueTimeout = 1 * time.Second
)

// =============================================================================
// Validation Ranges
// =============================================================================

const (
	// Port ranges
	MinPort                     = 1024
	MaxPort                     = 65535

	// Size ranges
	MinPageSize                 = 1
	MaxPageSizeLimit            = 10000
	MinBatchSize                = 1
	MaxBatchSizeLimit           = 100000

	// Timeout ranges
	MinTimeout                  = 1 * time.Second
	MaxTimeout                  = 10 * time.Minute

	// Connection ranges
	MinConnections              = 1
	MaxConnections              = 10000

	// Memory ranges
	MinMemorySize               = 1024 * 1024    // 1MB
	MaxMemorySize               = 64 * 1024 * 1024 * 1024 // 64GB
)

// =============================================================================
// Validation Functions
// =============================================================================

// ValidatePort checks if a port number is valid
func ValidatePort(port int) error {
	if port < MinPort || port > MaxPort {
		return fmt.Errorf("port %d is out of range [%d, %d]", port, MinPort, MaxPort)
	}
	return nil
}

// ValidatePageSize checks if page size is valid
func ValidatePageSize(size int) error {
	if size < MinPageSize || size > MaxPageSizeLimit {
		return fmt.Errorf("page size %d is out of range [%d, %d]", size, MinPageSize, MaxPageSizeLimit)
	}
	return nil
}

// ValidateBatchSize checks if batch size is valid
func ValidateBatchSize(size int) error {
	if size < MinBatchSize || size > MaxBatchSizeLimit {
		return fmt.Errorf("batch size %d is out of range [%d, %d]", size, MinBatchSize, MaxBatchSizeLimit)
	}
	return nil
}

// ValidateTimeout checks if timeout duration is valid
func ValidateTimeout(timeout time.Duration) error {
	if timeout < MinTimeout || timeout > MaxTimeout {
		return fmt.Errorf("timeout %v is out of range [%v, %v]", timeout, MinTimeout, MaxTimeout)
	}
	return nil
}

// ValidateConnections checks if connection count is valid
func ValidateConnections(count int) error {
	if count < MinConnections || count > MaxConnections {
		return fmt.Errorf("connection count %d is out of range [%d, %d]", count, MinConnections, MaxConnections)
	}
	return nil
}

// ValidateMemorySize checks if memory size is valid
func ValidateMemorySize(size int64) error {
	if size < MinMemorySize || size > MaxMemorySize {
		return fmt.Errorf("memory size %d is out of range [%d, %d]", size, MinMemorySize, MaxMemorySize)
	}
	return nil
}

// ValidateQueryLength checks if query length is valid
func ValidateQueryLength(length int) error {
	if length <= 0 || length > DefaultMaxQueryLength {
		return fmt.Errorf("query length %d is out of range [1, %d]", length, DefaultMaxQueryLength)
	}
	return nil
}

// ValidateResultSize checks if result size is valid
func ValidateResultSize(size int) error {
	if size <= 0 || size > DefaultMaxResultSize {
		return fmt.Errorf("result size %d is out of range [1, %d]", size, DefaultMaxResultSize)
	}
	return nil
}

// =============================================================================
// Configuration Validation Groups
// =============================================================================

// ValidateServerConfig validates server configuration values
func ValidateServerConfig(port int, timeout time.Duration, maxConnections int) error {
	if err := ValidatePort(port); err != nil {
		return fmt.Errorf("server config validation failed: %w", err)
	}
	if err := ValidateTimeout(timeout); err != nil {
		return fmt.Errorf("server config validation failed: %w", err)
	}
	if err := ValidateConnections(maxConnections); err != nil {
		return fmt.Errorf("server config validation failed: %w", err)
	}
	return nil
}

// ValidateDatabaseConfig validates database configuration values
func ValidateDatabaseConfig(maxOpenConns, maxIdleConns int, timeout time.Duration) error {
	if err := ValidateConnections(maxOpenConns); err != nil {
		return fmt.Errorf("database config validation failed: %w", err)
	}
	if err := ValidateConnections(maxIdleConns); err != nil {
		return fmt.Errorf("database config validation failed: %w", err)
	}
	if maxIdleConns > maxOpenConns {
		return fmt.Errorf("database config validation failed: max idle connections (%d) cannot exceed max open connections (%d)", maxIdleConns, maxOpenConns)
	}
	if err := ValidateTimeout(timeout); err != nil {
		return fmt.Errorf("database config validation failed: %w", err)
	}
	return nil
}

// ValidateCacheConfig validates cache configuration values
func ValidateCacheConfig(poolSize int, timeout time.Duration, memorySize int64) error {
	if err := ValidateConnections(poolSize); err != nil {
		return fmt.Errorf("cache config validation failed: %w", err)
	}
	if err := ValidateTimeout(timeout); err != nil {
		return fmt.Errorf("cache config validation failed: %w", err)
	}
	if err := ValidateMemorySize(memorySize); err != nil {
		return fmt.Errorf("cache config validation failed: %w", err)
	}
	return nil
}

// ValidateAPIConfig validates API configuration values
func ValidateAPIConfig(pageSize, maxResults int, timeout time.Duration) error {
	if err := ValidatePageSize(pageSize); err != nil {
		return fmt.Errorf("API config validation failed: %w", err)
	}
	if err := ValidateResultSize(maxResults); err != nil {
		return fmt.Errorf("API config validation failed: %w", err)
	}
	if err := ValidateTimeout(timeout); err != nil {
		return fmt.Errorf("API config validation failed: %w", err)
	}
	return nil
}

// =============================================================================
// Helper Functions
// =============================================================================

// GetDefaultConfigFile returns the default config file path based on environment
func GetDefaultConfigFile(env string) string {
	switch env {
	case "development", "dev":
		return DefaultDevConfigFile
	case "production", "prod":
		return DefaultProdConfigFile
	case "test":
		return DefaultTestConfigFile
	default:
		return DefaultConfigFile
	}
}

// IsValidEnvironment checks if the environment name is valid
func IsValidEnvironment(env string) bool {
	validEnvs := []string{"development", "dev", "production", "prod", "test", "local"}
	for _, validEnv := range validEnvs {
		if env == validEnv {
			return true
		}
	}
	return false
}

// GetEnvironmentDefaults returns environment-specific default values
func GetEnvironmentDefaults(env string) map[string]interface{} {
	defaults := make(map[string]interface{})

	switch env {
	case "development", "dev":
		defaults["log_level"] = "debug"
		defaults["cache_ttl"] = 5 * time.Minute
		defaults["db_pool_size"] = 10
	case "production", "prod":
		defaults["log_level"] = "info"
		defaults["cache_ttl"] = DefaultCacheTTL
		defaults["db_pool_size"] = DefaultDBMaxOpenConns
	case "test":
		defaults["log_level"] = "warn"
		defaults["cache_ttl"] = 1 * time.Minute
		defaults["db_pool_size"] = 5
	default:
		defaults["log_level"] = DefaultLogLevel
		defaults["cache_ttl"] = DefaultCacheTTL
		defaults["db_pool_size"] = DefaultDBMaxOpenConns
	}

	return defaults
}

//Personal.AI order the ending
