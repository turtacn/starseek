package models

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"time"
)

// ServerConfig 服务器配置结构体
type ServerConfig struct {
	// 基础配置
	Host           string        `json:"host" yaml:"host" validate:"required"`
	Port           int           `json:"port" yaml:"port" validate:"required,min=1,max=65535"`
	Environment    string        `json:"environment" yaml:"environment" validate:"oneof=development staging production"`
	Debug          bool          `json:"debug" yaml:"debug"`

	// 超时配置
	ReadTimeout    time.Duration `json:"read_timeout" yaml:"read_timeout"`
	WriteTimeout   time.Duration `json:"write_timeout" yaml:"write_timeout"`
	IdleTimeout    time.Duration `json:"idle_timeout" yaml:"idle_timeout"`
	ShutdownTimeout time.Duration `json:"shutdown_timeout" yaml:"shutdown_timeout"`

	// 连接配置
	MaxHeaderBytes int           `json:"max_header_bytes" yaml:"max_header_bytes"`
	MaxConnections int           `json:"max_connections" yaml:"max_connections"`
	KeepAlive      bool          `json:"keep_alive" yaml:"keep_alive"`

	// TLS配置
	TLS            *TLSConfig    `json:"tls,omitempty" yaml:"tls,omitempty"`

	// HTTP配置
	HTTP           *HTTPConfig   `json:"http,omitempty" yaml:"http,omitempty"`

	// gRPC配置
	GRPC           *GRPCConfig   `json:"grpc,omitempty" yaml:"grpc,omitempty"`

	// 中间件配置
	Middleware     *MiddlewareConfig `json:"middleware,omitempty" yaml:"middleware,omitempty"`

	// 限流配置
	RateLimit      *RateLimitConfig `json:"rate_limit,omitempty" yaml:"rate_limit,omitempty"`

	// 健康检查配置
	HealthCheck    *HealthCheckConfig `json:"health_check,omitempty" yaml:"health_check,omitempty"`

	// 性能配置
	Performance    *PerformanceConfig `json:"performance,omitempty" yaml:"performance,omitempty"`
}

// DatabaseConfig 数据库配置结构体
type DatabaseConfig struct {
	// 基础连接信息
	Driver         string        `json:"driver" yaml:"driver" validate:"required,oneof=starrocks clickhouse doris mysql postgresql"`
	Host           string        `json:"host" yaml:"host" validate:"required"`
	Port           int           `json:"port" yaml:"port" validate:"required,min=1,max=65535"`
	Database       string        `json:"database" yaml:"database" validate:"required"`
	Username       string        `json:"username" yaml:"username" validate:"required"`
	Password       string        `json:"password" yaml:"password" validate:"required"`

	// SSL配置
	SSLMode        string        `json:"ssl_mode" yaml:"ssl_mode"`
	SSLCert        string        `json:"ssl_cert" yaml:"ssl_cert"`
	SSLKey         string        `json:"ssl_key" yaml:"ssl_key"`
	SSLRootCert    string        `json:"ssl_root_cert" yaml:"ssl_root_cert"`

	// 连接池配置
	Pool           *PoolConfig   `json:"pool,omitempty" yaml:"pool,omitempty"`

	// 超时配置
	ConnectTimeout time.Duration `json:"connect_timeout" yaml:"connect_timeout"`
	ReadTimeout    time.Duration `json:"read_timeout" yaml:"read_timeout"`
	WriteTimeout   time.Duration `json:"write_timeout" yaml:"write_timeout"`
	QueryTimeout   time.Duration `json:"query_timeout" yaml:"query_timeout"`

	// 重试策略
	Retry          *RetryConfig  `json:"retry,omitempty" yaml:"retry,omitempty"`

	// 事务配置
	Transaction    *TransactionConfig `json:"transaction,omitempty" yaml:"transaction,omitempty"`

	// 监控配置
	Monitoring     *DBMonitoringConfig `json:"monitoring,omitempty" yaml:"monitoring,omitempty"`

	// 字符集和时区
	Charset        string        `json:"charset" yaml:"charset"`
	Timezone       string        `json:"timezone" yaml:"timezone"`

	// 额外参数
	Parameters     map[string]string `json:"parameters,omitempty" yaml:"parameters,omitempty"`

	// 数据库特定配置
	StarRocksConfig *StarRocksConfig `json:"starrocks_config,omitempty" yaml:"starrocks_config,omitempty"`
	ClickHouseConfig *ClickHouseConfig `json:"clickhouse_config,omitempty" yaml:"clickhouse_config,omitempty"`
	DorisConfig     *DorisConfig      `json:"doris_config,omitempty" yaml:"doris_config,omitempty"`
}

// CacheConfig 缓存配置结构体
type CacheConfig struct {
	// 缓存类型
	Type           CacheType     `json:"type" yaml:"type" validate:"required,oneof=memory redis memcached"`

	// 连接信息
	Servers        []string      `json:"servers" yaml:"servers"`
	Username       string        `json:"username" yaml:"username"`
	Password       string        `json:"password" yaml:"password"`
	Database       int           `json:"database" yaml:"database"`

	// 连接池配置
	Pool           *CachePoolConfig `json:"pool,omitempty" yaml:"pool,omitempty"`

	// TTL配置
	DefaultTTL     time.Duration `json:"default_ttl" yaml:"default_ttl"`
	MaxTTL         time.Duration `json:"max_ttl" yaml:"max_ttl"`
	CleanupInterval time.Duration `json:"cleanup_interval" yaml:"cleanup_interval"`

	// 内存限制
	MaxMemory      int64         `json:"max_memory" yaml:"max_memory"`
	EvictionPolicy EvictionPolicy `json:"eviction_policy" yaml:"eviction_policy"`

	// 序列化配置
	Serialization  *SerializationConfig `json:"serialization,omitempty" yaml:"serialization,omitempty"`

	// 压缩配置
	Compression    *CompressionConfig `json:"compression,omitempty" yaml:"compression,omitempty"`

	// 分区配置
	Sharding       *ShardingConfig `json:"sharding,omitempty" yaml:"sharding,omitempty"`

	// 监控配置
	Monitoring     *CacheMonitoringConfig `json:"monitoring,omitempty" yaml:"monitoring,omitempty"`

	// 预热配置
	Warmup         *WarmupConfig `json:"warmup,omitempty" yaml:"warmup,omitempty"`

	// 故障转移配置
	Failover       *FailoverConfig `json:"failover,omitempty" yaml:"failover,omitempty"`
}

// LoggingConfig 日志配置结构体
type LoggingConfig struct {
	// 基础配置
	Level          LogLevel      `json:"level" yaml:"level" validate:"required,oneof=debug info warn error fatal panic"`
	Format         LogFormat     `json:"format" yaml:"format" validate:"oneof=json text"`
	Encoding       string        `json:"encoding" yaml:"encoding" validate:"oneof=json console"`

	// 输出配置
	Outputs        []*OutputConfig `json:"outputs" yaml:"outputs"`

	// 文件配置
	File           *FileLogConfig `json:"file,omitempty" yaml:"file,omitempty"`

	// 控制台配置
	Console        *ConsoleLogConfig `json:"console,omitempty" yaml:"console,omitempty"`

	// 系统日志配置
	Syslog         *SyslogConfig `json:"syslog,omitempty" yaml:"syslog,omitempty"`

	// 结构化日志配置
	Structured     *StructuredLogConfig `json:"structured,omitempty" yaml:"structured,omitempty"`

	// 采样配置
	Sampling       *SamplingConfig `json:"sampling,omitempty" yaml:"sampling,omitempty"`

	// 钩子配置
	Hooks          []*HookConfig `json:"hooks,omitempty" yaml:"hooks,omitempty"`

	// 上下文配置
	Context        *ContextConfig `json:"context,omitempty" yaml:"context,omitempty"`

	// 性能配置
	Performance    *LogPerformanceConfig `json:"performance,omitempty" yaml:"performance,omitempty"`
}

// MonitoringConfig 监控配置结构体
type MonitoringConfig struct {
	// 基础配置
	Enabled        bool          `json:"enabled" yaml:"enabled"`
	Interval       time.Duration `json:"interval" yaml:"interval"`

	// 指标配置
	Metrics        *MetricsConfig `json:"metrics,omitempty" yaml:"metrics,omitempty"`

	// 追踪配置
	Tracing        *TracingConfig `json:"tracing,omitempty" yaml:"tracing,omitempty"`

	// 性能监控配置
	Profiling      *ProfilingConfig `json:"profiling,omitempty" yaml:"profiling,omitempty"`

	// 健康检查配置
	Health         *HealthMonitoringConfig `json:"health,omitempty" yaml:"health,omitempty"`

	// 告警配置
	Alerting       *AlertingConfig `json:"alerting,omitempty" yaml:"alerting,omitempty"`

	// 数据源配置
	DataSources    []*DataSourceConfig `json:"data_sources,omitempty" yaml:"data_sources,omitempty"`

	// 仪表板配置
	Dashboard      *DashboardConfig `json:"dashboard,omitempty" yaml:"dashboard,omitempty"`

	// 日志监控配置
	LogMonitoring  *LogMonitoringConfig `json:"log_monitoring,omitempty" yaml:"log_monitoring,omitempty"`
}

// SecurityConfig 安全配置结构体
type SecurityConfig struct {
	// 认证配置
	Authentication *AuthenticationConfig `json:"authentication,omitempty" yaml:"authentication,omitempty"`

	// 授权配置
	Authorization  *AuthorizationConfig `json:"authorization,omitempty" yaml:"authorization,omitempty"`

	// 加密配置
	Encryption     *EncryptionConfig `json:"encryption,omitempty" yaml:"encryption,omitempty"`

	// JWT配置
	JWT            *JWTConfig    `json:"jwt,omitempty" yaml:"jwt,omitempty"`

	// OAuth配置
	OAuth          *OAuthConfig  `json:"oauth,omitempty" yaml:"oauth,omitempty"`

	// CORS配置
	CORS           *CORSConfig   `json:"cors,omitempty" yaml:"cors,omitempty"`

	// CSRF配置
	CSRF           *CSRFConfig   `json:"csrf,omitempty" yaml:"csrf,omitempty"`

	// 会话配置
	Session        *SessionConfig `json:"session,omitempty" yaml:"session,omitempty"`

	// API密钥配置
	APIKey         *APIKeyConfig `json:"api_key,omitempty" yaml:"api_key,omitempty"`

	// 防护配置
	Protection     *ProtectionConfig `json:"protection,omitempty" yaml:"protection,omitempty"`

	// 审计配置
	Audit          *AuditConfig  `json:"audit,omitempty" yaml:"audit,omitempty"`
}

// 子配置结构体定义

// TLSConfig TLS配置
type TLSConfig struct {
	Enabled        bool          `json:"enabled" yaml:"enabled"`
	CertFile       string        `json:"cert_file" yaml:"cert_file"`
	KeyFile        string        `json:"key_file" yaml:"key_file"`
	CAFile         string        `json:"ca_file" yaml:"ca_file"`
	InsecureSkipVerify bool      `json:"insecure_skip_verify" yaml:"insecure_skip_verify"`
	MinVersion     string        `json:"min_version" yaml:"min_version"`
	MaxVersion     string        `json:"max_version" yaml:"max_version"`
	CipherSuites   []string      `json:"cipher_suites" yaml:"cipher_suites"`
	ClientAuth     ClientAuthType `json:"client_auth" yaml:"client_auth"`
}

// HTTPConfig HTTP配置
type HTTPConfig struct {
	Enabled        bool          `json:"enabled" yaml:"enabled"`
	CORS           *CORSConfig   `json:"cors,omitempty" yaml:"cors,omitempty"`
	Compression    bool          `json:"compression" yaml:"compression"`
	RequestTimeout time.Duration `json:"request_timeout" yaml:"request_timeout"`
	BodyLimit      int64         `json:"body_limit" yaml:"body_limit"`
}

// GRPCConfig gRPC配置
type GRPCConfig struct {
	Enabled        bool          `json:"enabled" yaml:"enabled"`
	Port           int           `json:"port" yaml:"port"`
	MaxRecvMsgSize int           `json:"max_recv_msg_size" yaml:"max_recv_msg_size"`
	MaxSendMsgSize int           `json:"max_send_msg_size" yaml:"max_send_msg_size"`
	Reflection     bool          `json:"reflection" yaml:"reflection"`
}

// MiddlewareConfig 中间件配置
type MiddlewareConfig struct {
	Recovery       bool          `json:"recovery" yaml:"recovery"`
	Logger         bool          `json:"logger" yaml:"logger"`
	CORS           bool          `json:"cors" yaml:"cors"`
	RateLimit      bool          `json:"rate_limit" yaml:"rate_limit"`
	Timeout        bool          `json:"timeout" yaml:"timeout"`
	Compress       bool          `json:"compress" yaml:"compress"`
	RequestID      bool          `json:"request_id" yaml:"request_id"`
	Authentication bool          `json:"authentication" yaml:"authentication"`
	Authorization  bool          `json:"authorization" yaml:"authorization"`
}

// RateLimitConfig 限流配置
type RateLimitConfig struct {
	Enabled        bool          `json:"enabled" yaml:"enabled"`
	Rate           int           `json:"rate" yaml:"rate"`
	Burst          int           `json:"burst" yaml:"burst"`
	Window         time.Duration `json:"window" yaml:"window"`
	KeyFunc        string        `json:"key_func" yaml:"key_func"`
	Store          string        `json:"store" yaml:"store"`
}

// HealthCheckConfig 健康检查配置
type HealthCheckConfig struct {
	Enabled        bool          `json:"enabled" yaml:"enabled"`
	Path           string        `json:"path" yaml:"path"`
	Interval       time.Duration `json:"interval" yaml:"interval"`
	Timeout        time.Duration `json:"timeout" yaml:"timeout"`
	Checks         []string      `json:"checks" yaml:"checks"`
}

// PerformanceConfig 性能配置
type PerformanceConfig struct {
	GOMAXPROCS     int           `json:"gomaxprocs" yaml:"gomaxprocs"`
	GCPercent      int           `json:"gc_percent" yaml:"gc_percent"`
	ReadBufferSize int           `json:"read_buffer_size" yaml:"read_buffer_size"`
	WriteBufferSize int          `json:"write_buffer_size" yaml:"write_buffer_size"`
}

// PoolConfig 连接池配置
type PoolConfig struct {
	MaxOpen        int           `json:"max_open" yaml:"max_open"`
	MaxIdle        int           `json:"max_idle" yaml:"max_idle"`
	MaxLifetime    time.Duration `json:"max_lifetime" yaml:"max_lifetime"`
	MaxIdleTime    time.Duration `json:"max_idle_time" yaml:"max_idle_time"`
	HealthCheck    bool          `json:"health_check" yaml:"health_check"`
}

// RetryConfig 重试配置
type RetryConfig struct {
	MaxAttempts    int           `json:"max_attempts" yaml:"max_attempts"`
	InitialDelay   time.Duration `json:"initial_delay" yaml:"initial_delay"`
	MaxDelay       time.Duration `json:"max_delay" yaml:"max_delay"`
	Multiplier     float64       `json:"multiplier" yaml:"multiplier"`
	Jitter         bool          `json:"jitter" yaml:"jitter"`
}

// TransactionConfig 事务配置
type TransactionConfig struct {
	IsolationLevel string        `json:"isolation_level" yaml:"isolation_level"`
	Timeout        time.Duration `json:"timeout" yaml:"timeout"`
	ReadOnly       bool          `json:"read_only" yaml:"read_only"`
}

// MetricsConfig 指标配置
type MetricsConfig struct {
	Enabled        bool          `json:"enabled" yaml:"enabled"`
	Path           string        `json:"path" yaml:"path"`
	Namespace      string        `json:"namespace" yaml:"namespace"`
	Subsystem      string        `json:"subsystem" yaml:"subsystem"`
	CollectRuntime bool          `json:"collect_runtime" yaml:"collect_runtime"`
	CollectDB      bool          `json:"collect_db" yaml:"collect_db"`
	CollectHTTP    bool          `json:"collect_http" yaml:"collect_http"`
}

// TracingConfig 追踪配置
type TracingConfig struct {
	Enabled        bool          `json:"enabled" yaml:"enabled"`
	Endpoint       string        `json:"endpoint" yaml:"endpoint"`
	ServiceName    string        `json:"service_name" yaml:"service_name"`
	SampleRate     float64       `json:"sample_rate" yaml:"sample_rate"`
	Headers        map[string]string `json:"headers" yaml:"headers"`
}

// AlertingConfig 告警配置
type AlertingConfig struct {
	Enabled        bool          `json:"enabled" yaml:"enabled"`
	Rules          []*AlertRule  `json:"rules" yaml:"rules"`
	Channels       []*AlertChannel `json:"channels" yaml:"channels"`
}

// AlertRule 告警规则
type AlertRule struct {
	Name           string        `json:"name" yaml:"name"`
	Expression     string        `json:"expression" yaml:"expression"`
	Threshold      float64       `json:"threshold" yaml:"threshold"`
	Duration       time.Duration `json:"duration" yaml:"duration"`
	Severity       AlertSeverity `json:"severity" yaml:"severity"`
	Message        string        `json:"message" yaml:"message"`
	Channels       []string      `json:"channels" yaml:"channels"`
}

// AlertChannel 告警通道
type AlertChannel struct {
	Name           string            `json:"name" yaml:"name"`
	Type           AlertChannelType  `json:"type" yaml:"type"`
	Enabled        bool              `json:"enabled" yaml:"enabled"`
	Config         map[string]interface{} `json:"config" yaml:"config"`
}

// AuthenticationConfig 认证配置
type AuthenticationConfig struct {
	Type           AuthenticationType `json:"type" yaml:"type"`
	Providers      []*AuthProvider   `json:"providers" yaml:"providers"`
	DefaultRealm   string            `json:"default_realm" yaml:"default_realm"`
	CacheEnabled   bool              `json:"cache_enabled" yaml:"cache_enabled"`
	CacheTTL       time.Duration     `json:"cache_ttl" yaml:"cache_ttl"`
}

// AuthProvider 认证提供者
type AuthProvider struct {
	Name           string            `json:"name" yaml:"name"`
	Type           string            `json:"type" yaml:"type"`
	Enabled        bool              `json:"enabled" yaml:"enabled"`
	Config         map[string]interface{} `json:"config" yaml:"config"`
}

// JWTConfig JWT配置
type JWTConfig struct {
	Secret         string        `json:"secret" yaml:"secret"`
	Algorithm      string        `json:"algorithm" yaml:"algorithm"`
	Expiration     time.Duration `json:"expiration" yaml:"expiration"`
	RefreshTTL     time.Duration `json:"refresh_ttl" yaml:"refresh_ttl"`
	Issuer         string        `json:"issuer" yaml:"issuer"`
	Audience       string        `json:"audience" yaml:"audience"`
}

// CORSConfig CORS配置
type CORSConfig struct {
	Enabled        bool          `json:"enabled" yaml:"enabled"`
	AllowOrigins   []string      `json:"allow_origins" yaml:"allow_origins"`
	AllowMethods   []string      `json:"allow_methods" yaml:"allow_methods"`
	AllowHeaders   []string      `json:"allow_headers" yaml:"allow_headers"`
	ExposeHeaders  []string      `json:"expose_headers" yaml:"expose_headers"`
	AllowCredentials bool        `json:"allow_credentials" yaml:"allow_credentials"`
	MaxAge         time.Duration `json:"max_age" yaml:"max_age"`
}

// FileLogConfig 文件日志配置
type FileLogConfig struct {
	Enabled        bool          `json:"enabled" yaml:"enabled"`
	Filename       string        `json:"filename" yaml:"filename"`
	MaxSize        int           `json:"max_size" yaml:"max_size"`
	MaxAge         int           `json:"max_age" yaml:"max_age"`
	MaxBackups     int           `json:"max_backups" yaml:"max_backups"`
	Compress       bool          `json:"compress" yaml:"compress"`
	LocalTime      bool          `json:"local_time" yaml:"local_time"`
}

// OutputConfig 输出配置
type OutputConfig struct {
	Type           string            `json:"type" yaml:"type"`
	Level          LogLevel          `json:"level" yaml:"level"`
	Config         map[string]interface{} `json:"config" yaml:"config"`
}

// 枚举类型定义
type (
	CacheType           string
	EvictionPolicy      string
	LogLevel            string
	LogFormat           string
	ClientAuthType      string
	AuthenticationType  string
	AlertSeverity       string
	AlertChannelType    string
)

// CacheType 常量
const (
	CacheTypeMemory     CacheType = "memory"
	CacheTypeRedis      CacheType = "redis"
	CacheTypeMemcached  CacheType = "memcached"
)

// EvictionPolicy 常量
const (
	EvictionPolicyLRU   EvictionPolicy = "lru"
	EvictionPolicyLFU   EvictionPolicy = "lfu"
	EvictionPolicyTTL   EvictionPolicy = "ttl"
	EvictionPolicyRandom EvictionPolicy = "random"
)

// LogLevel 常量
const (
	LogLevelDebug LogLevel = "debug"
	LogLevelInfo  LogLevel = "info"
	LogLevelWarn  LogLevel = "warn"
	LogLevelError LogLevel = "error"
	LogLevelFatal LogLevel = "fatal"
	LogLevelPanic LogLevel = "panic"
)

// LogFormat 常量
const (
	LogFormatJSON LogFormat = "json"
	LogFormatText LogFormat = "text"
)

// ClientAuthType 常量
const (
	ClientAuthTypeNoClientCert   ClientAuthType = "NoClientCert"
	ClientAuthTypeRequestClientCert ClientAuthType = "RequestClientCert"
	ClientAuthTypeRequireAnyClientCert ClientAuthType = "RequireAnyClientCert"
	ClientAuthTypeVerifyClientCertIfGiven ClientAuthType = "VerifyClientCertIfGiven"
	ClientAuthTypeRequireAndVerifyClientCert ClientAuthType = "RequireAndVerifyClientCert"
)

// AuthenticationType 常量
const (
	AuthenticationTypeBasic   AuthenticationType = "basic"
	AuthenticationTypeBearer  AuthenticationType = "bearer"
	AuthenticationTypeJWT     AuthenticationType = "jwt"
	AuthenticationTypeOAuth   AuthenticationType = "oauth"
	AuthenticationTypeLDAP    AuthenticationType = "ldap"
	AuthenticationTypeCustom  AuthenticationType = "custom"
)

// AlertSeverity 常量
const (
	AlertSeverityInfo     AlertSeverity = "info"
	AlertSeverityWarning  AlertSeverity = "warning"
	AlertSeverityError    AlertSeverity = "error"
	AlertSeverityCritical AlertSeverity = "critical"
)

// AlertChannelType 常量
const (
	AlertChannelTypeEmail    AlertChannelType = "email"
	AlertChannelTypeSlack    AlertChannelType = "slack"
	AlertChannelTypeWebhook  AlertChannelType = "webhook"
	AlertChannelTypeSMS      AlertChannelType = "sms"
)

// 验证方法

// Validate 验证ServerConfig
func (c *ServerConfig) Validate() error {
	if c.Host == "" {
		return errors.New("server host is required")
	}

	if c.Port <= 0 || c.Port > 65535 {
		return errors.New("server port must be between 1 and 65535")
	}

	if c.Environment != "" {
		validEnvs := []string{"development", "staging", "production"}
		if !contains(validEnvs, c.Environment) {
			return fmt.Errorf("invalid environment: %s", c.Environment)
		}
	}

	// 验证TLS配置
	if c.TLS != nil && c.TLS.Enabled {
		if err := c.TLS.Validate(); err != nil {
			return fmt.Errorf("invalid TLS config: %w", err)
		}
	}

	return nil
}

// Validate 验证DatabaseConfig
func (c *DatabaseConfig) Validate() error {
	if c.Driver == "" {
		return errors.New("database driver is required")
	}

	if c.Host == "" {
		return errors.New("database host is required")
	}

	if c.Port <= 0 || c.Port > 65535 {
		return errors.New("database port must be between 1 and 65535")
	}

	if c.Database == "" {
		return errors.New("database name is required")
	}

	if c.Username == "" {
		return errors.New("database username is required")
	}

	// 验证驱动类型
	validDrivers := []string{"starrocks", "clickhouse", "doris", "mysql", "postgresql"}
	if !contains(validDrivers, c.Driver) {
		return fmt.Errorf("invalid database driver: %s", c.Driver)
	}

	return nil
}

// Validate 验证CacheConfig
func (c *CacheConfig) Validate() error {
	if c.Type == "" {
		return errors.New("cache type is required")
	}

	validTypes := []string{string(CacheTypeMemory), string(CacheTypeRedis), string(CacheTypeMemcached)}
	if !contains(validTypes, string(c.Type)) {
		return fmt.Errorf("invalid cache type: %s", c.Type)
	}

	if c.Type != CacheTypeMemory && len(c.Servers) == 0 {
		return errors.New("cache servers are required for non-memory cache types")
	}

	return nil
}

// Validate 验证LoggingConfig
func (c *LoggingConfig) Validate() error {
	if c.Level == "" {
		return errors.New("log level is required")
	}

	validLevels := []string{"debug", "info", "warn", "error", "fatal", "panic"}
	if !contains(validLevels, string(c.Level)) {
		return fmt.Errorf("invalid log level: %s", c.Level)
	}

	if c.Format != "" {
		validFormats := []string{"json", "text"}
		if !contains(validFormats, string(c.Format)) {
			return fmt.Errorf("invalid log format: %s", c.Format)
		}
	}

	return nil
}

// Validate 验证TLSConfig
func (c *TLSConfig) Validate() error {
	if !c.Enabled {
		return nil
	}

	if c.CertFile == "" {
		return errors.New("TLS cert file is required when TLS is enabled")
	}

	if c.KeyFile == "" {
		return errors.New("TLS key file is required when TLS is enabled")
	}

	return nil
}

// 配置合并方法

// Merge 合并ServerConfig
func (c *ServerConfig) Merge(other *ServerConfig) *ServerConfig {
	if other == nil {
		return c
	}

	merged := *c

	if other.Host != "" {
		merged.Host = other.Host
	}
	if other.Port != 0 {
		merged.Port = other.Port
	}
	if other.Environment != "" {
		merged.Environment = other.Environment
	}
	if other.ReadTimeout != 0 {
		merged.ReadTimeout = other.ReadTimeout
	}
	if other.WriteTimeout != 0 {
		merged.WriteTimeout = other.WriteTimeout
	}

	// 合并嵌套配置
	if other.TLS != nil {
		if merged.TLS == nil {
			merged.TLS = other.TLS
		} else {
			merged.TLS = merged.TLS.Merge(other.TLS)
		}
	}

	return &merged
}

// Merge 合并TLSConfig
func (c *TLSConfig) Merge(other *TLSConfig) *TLSConfig {
	if other == nil {
		return c
	}

	merged := *c

	if other.CertFile != "" {
		merged.CertFile = other.CertFile
	}
	if other.KeyFile != "" {
		merged.KeyFile = other.KeyFile
	}
	if other.CAFile != "" {
		merged.CAFile = other.CAFile
	}

	return &merged
}

// 默认配置方法

// DefaultServerConfig 返回默认服务器配置
func DefaultServerConfig() *ServerConfig {
	return &ServerConfig{
		Host:           "0.0.0.0",
		Port:           8080,
		Environment:    "development",
		Debug:          false,
		ReadTimeout:    30 * time.Second,
		WriteTimeout:   30 * time.Second,
		IdleTimeout:    60 * time.Second,
		ShutdownTimeout: 30 * time.Second,
		MaxHeaderBytes: 1 << 20, // 1MB
		MaxConnections: 1000,
		KeepAlive:      true,
	}
}

// DefaultDatabaseConfig 返回默认数据库配置
func DefaultDatabaseConfig() *DatabaseConfig {
	return &DatabaseConfig{
		ConnectTimeout: 10 * time.Second,
		ReadTimeout:    30 * time.Second,
		WriteTimeout:   30 * time.Second,
		QueryTimeout:   60 * time.Second,
		Charset:        "utf8mb4",
		Timezone:       "UTC",
		Pool: &PoolConfig{
			MaxOpen:     10,
			MaxIdle:     5,
			MaxLifetime: time.Hour,
			MaxIdleTime: 30 * time.Minute,
			HealthCheck: true,
		},
		Retry: &RetryConfig{
			MaxAttempts:  3,
			InitialDelay: time.Second,
			MaxDelay:     30 * time.Second,
			Multiplier:   2.0,
			Jitter:       true,
		},
	}
}

// DefaultCacheConfig 返回默认缓存配置
func DefaultCacheConfig() *CacheConfig {
	return &CacheConfig{
		Type:           CacheTypeMemory,
		DefaultTTL:     time.Hour,
		MaxTTL:         24 * time.Hour,
		CleanupInterval: 10 * time.Minute,
		MaxMemory:      100 * 1024 * 1024, // 100MB
		EvictionPolicy: EvictionPolicyLRU,
	}
}

// DefaultLoggingConfig 返回默认日志配置
func DefaultLoggingConfig() *LoggingConfig {
	return &LoggingConfig{
		Level:    LogLevelInfo,
		Format:   LogFormatJSON,
		Encoding: "json",
		Console: &ConsoleLogConfig{
			Enabled: true,
		},
	}
}

// 工具函数

// contains 检查字符串数组是否包含指定值
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// ToJSON 转换为JSON字符串
func (c *ServerConfig) ToJSON() (string, error) {
	data, err := json.Marshal(c)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// GetAddress 获取服务器地址
func (c *ServerConfig) GetAddress() string {
	return fmt.Sprintf("%s:%d", c.Host, c.Port)
}

// GetConnectionString 获取数据库连接字符串
func (c *DatabaseConfig) GetConnectionString() string {
	switch c.Driver {
	case "mysql":
		return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=%s&parseTime=true&loc=%s",
			c.Username, c.Password, c.Host, c.Port, c.Database, c.Charset, url.QueryEscape(c.Timezone))
	case "postgresql":
		return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s timezone=%s",
			c.Host, c.Port, c.Username, c.Password, c.Database, c.SSLMode, c.Timezone)
	default:
		return fmt.Sprintf("%s://%s:%s@%s:%d/%s",
			c.Driver, c.Username, c.Password, c.Host, c.Port, c.Database)
	}
}

// IsProduction 检查是否为生产环境
func (c *ServerConfig) IsProduction() bool {
	return c.Environment == "production"
}

// IsDevelopment 检查是否为开发环境
func (c *ServerConfig) IsDevelopment() bool {
	return c.Environment == "development"
}

// String 返回字符串表示
func (c *ServerConfig) String() string {
	return fmt.Sprintf("ServerConfig{Host:%s, Port:%d, Environment:%s}", c.Host, c.Port, c.Environment)
}

// String 返回字符串表示
func (c *DatabaseConfig) String() string {
	return fmt.Sprintf("DatabaseConfig{Driver:%s, Host:%s, Port:%d, Database:%s}",
		c.Driver, c.Host, c.Port, c.Database)
}

//Personal.AI order the ending
