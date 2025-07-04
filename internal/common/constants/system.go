package constants

import (
	"runtime"
	"time"
)

// =============================================================================
// System Metadata Constants
// =============================================================================

const (
	// Application information
	ApplicationName        = "StarSeek"
	ApplicationDescription = "Universal Search Engine for Column-Store Databases"
	ApplicationVendor      = "Personal.AI"

	// Version information (these would typically be set via build flags)
	Version               = "1.0.0"
	VersionMajor          = 1
	VersionMinor          = 0
	VersionPatch          = 0
	VersionPreRelease     = ""
	VersionBuildMetadata  = ""

	// Build information placeholders (set via ldflags during build)
	BuildVersion          = "dev"
	BuildCommit           = "unknown"
	BuildDate             = "unknown"
	BuildGoVersion        = runtime.Version()
	BuildOS               = runtime.GOOS
	BuildArch             = runtime.GOARCH

	// Release information
	ReleaseChannel        = "stable"  // stable, beta, alpha
	ReleaseCodename       = "Phoenix"
	MinimumGoVersion      = "1.21"
)

// =============================================================================
// API Version and Path Constants
// =============================================================================

const (
	// API versioning
	APIVersionV1          = "v1"
	APIVersionV2          = "v2"
	CurrentAPIVersion     = APIVersionV1

	// API path prefixes
	APIPathPrefix         = "/api"
	APIPathV1             = "/api/v1"
	APIPathV2             = "/api/v2"

	// API endpoint paths
	SearchEndpointPath    = "/search"
	IndexEndpointPath     = "/indexes"
	AdminEndpointPath     = "/admin"
	MetricsEndpointPath   = "/metrics"
	HealthEndpointPath    = "/health"
	DocsEndpointPath      = "/docs"

	// Full API paths
	SearchAPIPath         = APIPathV1 + SearchEndpointPath
	IndexAPIPath          = APIPathV1 + IndexEndpointPath
	AdminAPIPath          = APIPathV1 + AdminEndpointPath

	// Static resource paths
	StaticFilesPath       = "/static"
	AssetsPath            = "/assets"
	TemplatesPath         = "/templates"

	// OpenAPI/Swagger paths
	SwaggerUIPath         = "/swagger"
	OpenAPISpecPath       = "/openapi.json"
	RedocPath             = "/redoc"
)

// =============================================================================
// System Resource Limits
// =============================================================================

const (
	// Memory limits (in bytes)
	MaxMemoryUsage        = 8 * 1024 * 1024 * 1024  // 8GB
	MaxHeapSize           = 6 * 1024 * 1024 * 1024  // 6GB
	MaxStackSize          = 8 * 1024 * 1024         // 8MB

	// File system limits
	MaxOpenFiles          = 65536
	MaxFileSize           = 1024 * 1024 * 1024      // 1GB
	MaxLogFileSize        = 100 * 1024 * 1024       // 100MB
	MaxTempFileSize       = 500 * 1024 * 1024       // 500MB

	// Network limits
	MaxConnections        = 10000
	MaxConcurrentRequests = 5000
	MaxRequestBodySize    = 32 * 1024 * 1024        // 32MB
	MaxResponseSize       = 100 * 1024 * 1024       // 100MB

	// Processing limits
	MaxGoroutines         = 50000
	MaxChannelBuffer      = 10000
	MaxWorkerThreads      = 1000
	MaxTaskQueue          = 100000

	// Cache limits
	MaxCacheEntries       = 1000000
	MaxCacheKeySize       = 1024
	MaxCacheValueSize     = 10 * 1024 * 1024        // 10MB

	// Database limits
	MaxQueryComplexity    = 1000
	MaxJoinTables         = 20
	MaxWhereConditions    = 100
	MaxOrderByFields      = 20
	MaxGroupByFields      = 20
)

// =============================================================================
// Monitoring Metrics Names
// =============================================================================

const (
	// HTTP metrics
	MetricHTTPRequestsTotal        = "starseek_http_requests_total"
	MetricHTTPRequestDuration      = "starseek_http_request_duration_seconds"
	MetricHTTPRequestSize          = "starseek_http_request_size_bytes"
	MetricHTTPResponseSize         = "starseek_http_response_size_bytes"
	MetricHTTPActiveConnections    = "starseek_http_active_connections"

	// Search metrics
	MetricSearchRequestsTotal      = "starseek_search_requests_total"
	MetricSearchDuration           = "starseek_search_duration_seconds"
	MetricSearchResultsCount       = "starseek_search_results_count"
	MetricSearchErrorsTotal        = "starseek_search_errors_total"
	MetricSearchCacheHits          = "starseek_search_cache_hits_total"
	MetricSearchCacheMisses        = "starseek_search_cache_misses_total"

	// Database metrics
	MetricDBConnectionsActive      = "starseek_db_connections_active"
	MetricDBConnectionsIdle        = "starseek_db_connections_idle"
	MetricDBQueryDuration          = "starseek_db_query_duration_seconds"
	MetricDBQueriesTotal           = "starseek_db_queries_total"
	MetricDBErrorsTotal            = "starseek_db_errors_total"
	MetricDBSlowQueries            = "starseek_db_slow_queries_total"

	// Cache metrics
	MetricCacheHitsTotal           = "starseek_cache_hits_total"
	MetricCacheMissesTotal         = "starseek_cache_misses_total"
	MetricCacheEvictionsTotal      = "starseek_cache_evictions_total"
	MetricCacheMemoryUsage         = "starseek_cache_memory_usage_bytes"
	MetricCacheEntriesCount        = "starseek_cache_entries_count"

	// System metrics
	MetricSystemMemoryUsage        = "starseek_system_memory_usage_bytes"
	MetricSystemCPUUsage           = "starseek_system_cpu_usage_percent"
	MetricSystemGoroutinesCount    = "starseek_system_goroutines_count"
	MetricSystemGCDuration         = "starseek_system_gc_duration_seconds"
	MetricSystemOpenFiles          = "starseek_system_open_files_count"

	// Index metrics
	MetricIndexesCount             = "starseek_indexes_count"
	MetricIndexRefreshDuration     = "starseek_index_refresh_duration_seconds"
	MetricIndexSizeBytes           = "starseek_index_size_bytes"
	MetricIndexOperationsTotal     = "starseek_index_operations_total"

	// Rate limiting metrics
	MetricRateLimitExceeded        = "starseek_rate_limit_exceeded_total"
	MetricRateLimitAllowed         = "starseek_rate_limit_allowed_total"
	MetricRateLimitCurrent         = "starseek_rate_limit_current_requests"
)

// =============================================================================
// System Status Constants
// =============================================================================

const (
	// Health status values
	HealthStatusHealthy            = "healthy"
	HealthStatusUnhealthy          = "unhealthy"
	HealthStatusDegraded           = "degraded"
	HealthStatusUnknown            = "unknown"

	// Service status values
	ServiceStatusStarting          = "starting"
	ServiceStatusRunning           = "running"
	ServiceStatusStopping          = "stopping"
	ServiceStatusStopped           = "stopped"
	ServiceStatusFailed            = "failed"
	ServiceStatusMaintenance       = "maintenance"

	// Readiness status values
	ReadinessStatusReady           = "ready"
	ReadinessStatusNotReady        = "not_ready"
	ReadinessStatusWaiting         = "waiting"

	// Liveness status values
	LivenessStatusAlive            = "alive"
	LivenessStatusDead             = "dead"
	LivenessStatusChecking         = "checking"

	// Database status values
	DatabaseStatusConnected        = "connected"
	DatabaseStatusDisconnected     = "disconnected"
	DatabaseStatusConnecting       = "connecting"
	DatabaseStatusError            = "error"

	// Cache status values
	CacheStatusAvailable           = "available"
	CacheStatusUnavailable         = "unavailable"
	CacheStatusPartiallyAvailable  = "partially_available"

	// Index status values
	IndexStatusActive              = "active"
	IndexStatusInactive            = "inactive"
	IndexStatusBuilding            = "building"
	IndexStatusError               = "error"
	IndexStatusDeleted             = "deleted"
)

// =============================================================================
// Time Format Constants
// =============================================================================

const (
	// Standard time formats
	TimeFormatRFC3339              = time.RFC3339
	TimeFormatRFC3339Nano          = time.RFC3339Nano
	TimeFormatISO8601              = "2006-01-02T15:04:05Z07:00"
	TimeFormatISO8601Milli         = "2006-01-02T15:04:05.000Z07:00"
	TimeFormatISO8601Micro         = "2006-01-02T15:04:05.000000Z07:00"

	// Custom time formats
	TimeFormatDate                 = "2006-01-02"
	TimeFormatTime                 = "15:04:05"
	TimeFormatDateTime             = "2006-01-02 15:04:05"
	TimeFormatDateTimeWithTZ       = "2006-01-02 15:04:05 MST"
	TimeFormatUnix                 = "unix"
	TimeFormatUnixMilli            = "unix_milli"
	TimeFormatUnixMicro            = "unix_micro"
	TimeFormatUnixNano             = "unix_nano"

	// Log time formats
	TimeFormatLog                  = "2006-01-02 15:04:05.000"
	TimeFormatLogWithTZ            = "2006-01-02 15:04:05.000 MST"
	TimeFormatLogCompact           = "060102-150405"

	// File name time formats
	TimeFormatFilename             = "20060102_150405"
	TimeFormatFilenameMilli        = "20060102_150405_000"

	// HTTP time formats
	TimeFormatHTTP                 = time.RFC1123
	TimeFormatHTTPCookie           = "Mon, 02 Jan 2006 15:04:05 GMT"

	// Database time formats
	TimeFormatMySQL                = "2006-01-02 15:04:05"
	TimeFormatPostgres             = "2006-01-02 15:04:05.000000"
	TimeFormatSQLite               = "2006-01-02 15:04:05"
)

// =============================================================================
// System Environment Constants
// =============================================================================

const (
	// Environment names
	EnvironmentDevelopment         = "development"
	EnvironmentStaging             = "staging"
	EnvironmentProduction          = "production"
	EnvironmentTest                = "test"
	EnvironmentLocal               = "local"

	// Environment short names
	EnvDev                         = "dev"
	EnvStage                       = "stage"
	EnvProd                        = "prod"

	// Default environment
	DefaultEnvironment             = EnvironmentDevelopment

	// Profile names
	ProfileDefault                 = "default"
	ProfileDebug                   = "debug"
	ProfileRelease                 = "release"
	ProfileBenchmark               = "benchmark"
)

// =============================================================================
// System Defaults
// =============================================================================

const (
	// Default timeouts
	DefaultSystemTimeout           = 30 * time.Second
	DefaultShutdownGracePeriod     = 15 * time.Second
	DefaultHealthCheckInterval     = 30 * time.Second
	DefaultMetricsCollectionInterval = 15 * time.Second

	// Default retry settings
	DefaultMaxRetryAttempts        = 3
	DefaultRetryBaseDelay          = 1 * time.Second
	DefaultRetryMaxDelay           = 30 * time.Second
	DefaultRetryMultiplier         = 2.0

	// Default buffer sizes
	DefaultChannelBufferSize       = 1000
	DefaultIOBufferSize            = 32 * 1024
	DefaultLogBufferSize           = 64 * 1024

	// Default worker pool settings
	DefaultWorkerPoolSize          = 100
	DefaultWorkerQueueSize         = 10000
	DefaultWorkerTimeout           = 5 * time.Minute

	// Default rate limiting
	DefaultRateLimit               = 1000  // requests per minute
	DefaultBurstLimit              = 100   // burst requests
	DefaultRateLimitWindow         = 1 * time.Minute
)

// =============================================================================
// System Headers and Content Types
// =============================================================================

const (
	// HTTP headers
	HeaderContentType              = "Content-Type"
	HeaderContentLength            = "Content-Length"
	HeaderContentEncoding          = "Content-Encoding"
	HeaderAccept                   = "Accept"
	HeaderAcceptEncoding           = "Accept-Encoding"
	HeaderAuthorization            = "Authorization"
	HeaderCacheControl             = "Cache-Control"
	HeaderUserAgent                = "User-Agent"
	HeaderXRealIP                  = "X-Real-IP"
	HeaderXForwardedFor            = "X-Forwarded-For"
	HeaderXRequestID               = "X-Request-ID"
	HeaderXCorrelationID           = "X-Correlation-ID"

	// Custom headers
	HeaderStarSeekVersion          = "X-StarSeek-Version"
	HeaderStarSeekRequestID        = "X-StarSeek-Request-ID"
	HeaderStarSeekTimestamp        = "X-StarSeek-Timestamp"
	HeaderStarSeekRateLimit        = "X-StarSeek-RateLimit-Limit"
	HeaderStarSeekRateRemaining    = "X-StarSeek-RateLimit-Remaining"
	HeaderStarSeekRateReset        = "X-StarSeek-RateLimit-Reset"

	// Content types
	ContentTypeJSON                = "application/json"
	ContentTypeXML                 = "application/xml"
	ContentTypeHTML                = "text/html"
	ContentTypePlain               = "text/plain"
	ContentTypeForm                = "application/x-www-form-urlencoded"
	ContentTypeMultipart           = "multipart/form-data"
	ContentTypeOctetStream         = "application/octet-stream"

	// Encoding types
	EncodingGzip                   = "gzip"
	EncodingDeflate                = "deflate"
	EncodingBrotli                 = "br"
	EncodingIdentity               = "identity"
)

// =============================================================================
// System Information Functions
// =============================================================================

// GetSystemInfo returns basic system information
func GetSystemInfo() map[string]interface{} {
	return map[string]interface{}{
		"name":         ApplicationName,
		"version":      Version,
		"build":        BuildVersion,
		"commit":       BuildCommit,
		"build_date":   BuildDate,
		"go_version":   BuildGoVersion,
		"os":           BuildOS,
		"arch":         BuildArch,
		"environment":  DefaultEnvironment,
	}
}

// GetVersionInfo returns detailed version information
func GetVersionInfo() map[string]interface{} {
	return map[string]interface{}{
		"version":        Version,
		"major":          VersionMajor,
		"minor":          VersionMinor,
		"patch":          VersionPatch,
		"pre_release":    VersionPreRelease,
		"build_metadata": VersionBuildMetadata,
		"build_version":  BuildVersion,
		"build_commit":   BuildCommit,
		"build_date":     BuildDate,
	}
}

// GetBuildInfo returns build-related information
func GetBuildInfo() map[string]interface{} {
	return map[string]interface{}{
		"version":    BuildVersion,
		"commit":     BuildCommit,
		"date":       BuildDate,
		"go_version": BuildGoVersion,
		"os":         BuildOS,
		"arch":       BuildArch,
		"channel":    ReleaseChannel,
		"codename":   ReleaseCodename,
	}
}

// IsValidEnvironment checks if the given environment name is valid
func IsValidEnvironment(env string) bool {
	validEnvs := []string{
		EnvironmentDevelopment, EnvDev,
		EnvironmentStaging, EnvStage,
		EnvironmentProduction, EnvProd,
		EnvironmentTest,
		EnvironmentLocal,
	}

	for _, validEnv := range validEnvs {
		if env == validEnv {
			return true
		}
	}
	return false
}

// IsProductionEnvironment checks if the environment is production
func IsProductionEnvironment(env string) bool {
	return env == EnvironmentProduction || env == EnvProd
}

// IsDevelopmentEnvironment checks if the environment is development
func IsDevelopmentEnvironment(env string) bool {
	return env == EnvironmentDevelopment || env == EnvDev || env == EnvironmentLocal
}

// GetDefaultTimeFormat returns the default time format for the given context
func GetDefaultTimeFormat(context string) string {
	switch context {
	case "log":
		return TimeFormatLog
	case "api":
		return TimeFormatRFC3339
	case "filename":
		return TimeFormatFilename
	case "database":
		return TimeFormatMySQL
	case "http":
		return TimeFormatHTTP
	default:
		return TimeFormatRFC3339
	}
}

//Personal.AI order the ending
