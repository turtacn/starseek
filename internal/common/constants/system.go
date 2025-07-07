package constants

import (
	"fmt"
	"runtime"
	"time"
)

// System metadata constants
const (
	ServiceName        = "StarSeek"
	ServiceDescription = "Full-text search middleware for columnar databases"
	ServiceVersion     = "1.0.0"
	ServiceBuildTime   = "2025-01-01T00:00:00Z"
	ServiceCommitHash  = "dev"
	ServiceGoVersion   = runtime.Version()
)

// API version constants
const (
	APIVersion   = "v1"
	APIPrefix    = "/api/" + APIVersion
	APIUserAgent = ServiceName + "/" + ServiceVersion
	APIMediaType = "application/json"
)

// API path constants
const (
	// Search endpoints
	SearchPath        = APIPrefix + "/search"
	SearchStatsPath   = APIPrefix + "/search/stats"
	SearchSuggestPath = APIPrefix + "/search/suggest"

	// Index management endpoints
	IndexesPath       = APIPrefix + "/indexes"
	IndexPath         = APIPrefix + "/indexes/{id}"
	IndexStatsPath    = APIPrefix + "/indexes/{id}/stats"
	IndexValidatePath = APIPrefix + "/indexes/validate"

	// Admin endpoints
	AdminPath   = APIPrefix + "/admin"
	HealthPath  = "/health"
	ReadyPath   = "/ready"
	MetricsPath = "/metrics"
	ConfigPath  = AdminPath + "/config"
	CachePath   = AdminPath + "/cache"
	LogsPath    = AdminPath + "/logs"

	// Authentication endpoints
	AuthPath    = APIPrefix + "/auth"
	LoginPath   = AuthPath + "/login"
	RefreshPath = AuthPath + "/refresh"
	LogoutPath  = AuthPath + "/logout"
)

// System resource limits
const (
	MaxMemoryUsage        = 8 * 1024 * 1024 * 1024 // 8GB
	MaxCPUUsage           = 80                     // 80%
	MaxDiskUsage          = 90                     // 90%
	MaxFileDescriptors    = 65536
	MaxGoroutines         = 100000
	MaxNetworkConnections = 10000
)

// Monitoring metric names
const (
	// Request metrics
	MetricRequestsTotal   = "starseek_requests_total"
	MetricRequestDuration = "starseek_request_duration_seconds"
	MetricRequestSize     = "starseek_request_size_bytes"
	MetricResponseSize    = "starseek_response_size_bytes"

	// Search metrics
	MetricSearchTotal    = "starseek_search_total"
	MetricSearchDuration = "starseek_search_duration_seconds"
	MetricSearchResults  = "starseek_search_results_total"
	MetricSearchErrors   = "starseek_search_errors_total"

	// Database metrics
	MetricDatabaseConnections   = "starseek_database_connections"
	MetricDatabaseQueryDuration = "starseek_database_query_duration_seconds"
	MetricDatabaseQueryTotal    = "starseek_database_query_total"
	MetricDatabaseErrors        = "starseek_database_errors_total"

	// Cache metrics
	MetricCacheHits       = "starseek_cache_hits_total"
	MetricCacheMisses     = "starseek_cache_misses_total"
	MetricCacheSize       = "starseek_cache_size_bytes"
	MetricCacheOperations = "starseek_cache_operations_total"

	// System metrics
	MetricMemoryUsage     = "starseek_memory_usage_bytes"
	MetricCPUUsage        = "starseek_cpu_usage_percent"
	MetricGoroutines      = "starseek_goroutines_total"
	MetricFileDescriptors = "starseek_file_descriptors_total"
)

// Service status constants
const (
	StatusHealthy     = "healthy"
	StatusUnhealthy   = "unhealthy"
	StatusUnknown     = "unknown"
	StatusStarting    = "starting"
	StatusStopping    = "stopping"
	StatusMaintenance = "maintenance"
)

// Time format constants
const (
	TimeFormatRFC3339     = time.RFC3339
	TimeFormatRFC3339Nano = time.RFC3339Nano
	TimeFormatISO8601     = "2006-01-02T15:04:05Z07:00"
	TimeFormatDate        = "2006-01-02"
	TimeFormatTime        = "15:04:05"
	TimeFormatDateTime    = "2006-01-02 15:04:05"
	TimeFormatLog         = "2006/01/02 15:04:05"
)

// HTTP header constants
const (
	HeaderContentType     = "Content-Type"
	HeaderAccept          = "Accept"
	HeaderAuthorization   = "Authorization"
	HeaderUserAgent       = "User-Agent"
	HeaderXTraceID        = "X-Trace-ID"
	HeaderXRequestID      = "X-Request-ID"
	HeaderXForwardedFor   = "X-Forwarded-For"
	HeaderXRealIP         = "X-Real-IP"
	HeaderXRateLimit      = "X-RateLimit-Remaining"
	HeaderXRateLimitReset = "X-RateLimit-Reset"
)

// System events
const (
	EventSystemStartup      = "system.startup"
	EventSystemShutdown     = "system.shutdown"
	EventConfigReload       = "config.reload"
	EventHealthCheck        = "health.check"
	EventCacheClear         = "cache.clear"
	EventIndexCreate        = "index.create"
	EventIndexUpdate        = "index.update"
	EventIndexDelete        = "index.delete"
	EventSearchQuery        = "search.query"
	EventDatabaseConnect    = "database.connect"
	EventDatabaseDisconnect = "database.disconnect"
)

// Build and version information
type BuildInfo struct {
	Version    string `json:"version"`
	BuildTime  string `json:"build_time"`
	CommitHash string `json:"commit_hash"`
	GoVersion  string `json:"go_version"`
	Platform   string `json:"platform"`
	Compiler   string `json:"compiler"`
}

// GetBuildInfo returns the build information
func GetBuildInfo() BuildInfo {
	return BuildInfo{
		Version:    ServiceVersion,
		BuildTime:  ServiceBuildTime,
		CommitHash: ServiceCommitHash,
		GoVersion:  runtime.Version(),
		Platform:   fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH),
		Compiler:   runtime.Compiler,
	}
}

// System limits validation
type SystemLimits struct {
	MaxMemoryUsage        int64   `json:"max_memory_usage"`
	MaxCPUUsage           float64 `json:"max_cpu_usage"`
	MaxDiskUsage          float64 `json:"max_disk_usage"`
	MaxFileDescriptors    int     `json:"max_file_descriptors"`
	MaxGoroutines         int     `json:"max_goroutines"`
	MaxNetworkConnections int     `json:"max_network_connections"`
}

// GetSystemLimits returns the system limits
func GetSystemLimits() SystemLimits {
	return SystemLimits{
		MaxMemoryUsage:        MaxMemoryUsage,
		MaxCPUUsage:           MaxCPUUsage,
		MaxDiskUsage:          MaxDiskUsage,
		MaxFileDescriptors:    MaxFileDescriptors,
		MaxGoroutines:         MaxGoroutines,
		MaxNetworkConnections: MaxNetworkConnections,
	}
}

// Service capabilities
const (
	CapabilityFullTextSearch = "full_text_search"
	CapabilityMultiDatabase  = "multi_database"
	CapabilityRealTimeIndex  = "real_time_index"
	CapabilityHighlighting   = "highlighting"
	CapabilityRanking        = "ranking"
	CapabilityCaching        = "caching"
	CapabilityMonitoring     = "monitoring"
	CapabilityAuthentication = "authentication"
	CapabilityRateLimit      = "rate_limit"
	CapabilityBatchOperation = "batch_operation"
)

// GetServiceCapabilities returns the service capabilities
func GetServiceCapabilities() []string {
	return []string{
		CapabilityFullTextSearch,
		CapabilityMultiDatabase,
		CapabilityRealTimeIndex,
		CapabilityHighlighting,
		CapabilityRanking,
		CapabilityCaching,
		CapabilityMonitoring,
		CapabilityAuthentication,
		CapabilityRateLimit,
		CapabilityBatchOperation,
	}
}

// Default system configurations
var (
	DefaultSystemConfig = map[string]interface{}{
		"service_name":         ServiceName,
		"service_version":      ServiceVersion,
		"api_version":          APIVersion,
		"max_memory_usage":     MaxMemoryUsage,
		"max_cpu_usage":        MaxCPUUsage,
		"max_file_descriptors": MaxFileDescriptors,
		"max_goroutines":       MaxGoroutines,
	}

	DefaultMonitoringConfig = map[string]interface{}{
		"metrics_path":     MetricsPath,
		"health_path":      HealthPath,
		"ready_path":       ReadyPath,
		"metrics_interval": 15 * time.Second,
		"health_interval":  30 * time.Second,
	}
)

// System status information
type SystemStatus struct {
	Service      string        `json:"service"`
	Version      string        `json:"version"`
	Status       string        `json:"status"`
	Uptime       time.Duration `json:"uptime"`
	StartTime    time.Time     `json:"start_time"`
	LastCheck    time.Time     `json:"last_check"`
	Capabilities []string      `json:"capabilities"`
}

// Exit codes
const (
	ExitCodeSuccess      = 0
	ExitCodeError        = 1
	ExitCodeConfigError  = 2
	ExitCodeNetworkError = 3
	ExitCodeSignal       = 130
)

//Personal.AI order the ending
