package dto

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// BaseResponse 基础响应结构体
type BaseResponse struct {
	Code      int         `json:"code"`
	Message   string      `json:"message"`
	Success   bool        `json:"success"`
	Data      interface{} `json:"data,omitempty"`
	Error     *ErrorInfo  `json:"error,omitempty"`
	RequestID string      `json:"request_id,omitempty"`
	TraceID   string      `json:"trace_id,omitempty"`
	Timestamp time.Time   `json:"timestamp"`
	Duration  int64       `json:"duration,omitempty"` // milliseconds
	Version   string      `json:"version,omitempty"`
	Path      string      `json:"path,omitempty"`
	Method    string      `json:"method,omitempty"`
}

// ErrorInfo 错误信息结构体
type ErrorInfo struct {
	Code        string                 `json:"code"`
	Message     string                 `json:"message"`
	Details     []string               `json:"details,omitempty"`
	Cause       string                 `json:"cause,omitempty"`
	StackTrace  string                 `json:"stack_trace,omitempty"`
	Context     map[string]interface{} `json:"context,omitempty"`
	Timestamp   time.Time              `json:"timestamp"`
	Severity    ErrorSeverity          `json:"severity"`
	Retryable   bool                   `json:"retryable"`
	HelpURL     string                 `json:"help_url,omitempty"`
	ErrorID     string                 `json:"error_id,omitempty"`
	UserMessage string                 `json:"user_message,omitempty"`
}

// PaginationInfo 分页信息结构体
type PaginationInfo struct {
	Page         int    `json:"page"`
	Size         int    `json:"size"`
	Total        int64  `json:"total"`
	TotalPages   int    `json:"total_pages"`
	HasNext      bool   `json:"has_next"`
	HasPrevious  bool   `json:"has_previous"`
	NextCursor   string `json:"next_cursor,omitempty"`
	PrevCursor   string `json:"prev_cursor,omitempty"`
	Offset       int    `json:"offset,omitempty"`
	Limit        int    `json:"limit,omitempty"`
}

// PerformanceMetrics 性能指标结构体
type PerformanceMetrics struct {
	QueryTime       int64 `json:"query_time"`        // microseconds
	ProcessTime     int64 `json:"process_time"`      // microseconds
	SerializeTime   int64 `json:"serialize_time"`    // microseconds
	TotalTime       int64 `json:"total_time"`        // microseconds
	DatabaseTime    int64 `json:"database_time"`     // microseconds
	CacheTime       int64 `json:"cache_time"`        // microseconds
	NetworkTime     int64 `json:"network_time"`      // microseconds
	MemoryUsage     int64 `json:"memory_usage"`      // bytes
	CPUUsage        float64 `json:"cpu_usage"`       // percentage
	CacheHitRate    float64 `json:"cache_hit_rate"`  // percentage
	QPS             float64 `json:"qps"`             // queries per second
	Concurrency     int   `json:"concurrency"`       // concurrent requests
}

// SearchResponse 搜索响应结构体
type SearchResponse struct {
	BaseResponse

	// 搜索结果
	Results      []*SearchHit       `json:"results"`
	Total        int64              `json:"total"`
	MaxScore     float64            `json:"max_score,omitempty"`

	// 分页信息
	Pagination   *PaginationInfo    `json:"pagination,omitempty"`

	// 性能指标
	Performance  *PerformanceMetrics `json:"performance,omitempty"`

	// 聚合结果
	Aggregations map[string]*AggregationResult `json:"aggregations,omitempty"`

	// 建议结果
	Suggestions  map[string]*SuggestionResult `json:"suggestions,omitempty"`

	// 搜索元数据
	Metadata     *SearchMetadata    `json:"metadata,omitempty"`

	// 警告信息
	Warnings     []string           `json:"warnings,omitempty"`

	// 调试信息
	Debug        *SearchDebugInfo   `json:"debug,omitempty"`
}

// SearchHit 搜索命中结果
type SearchHit struct {
	ID          string                 `json:"id"`
	Index       string                 `json:"index,omitempty"`
	Type        string                 `json:"type,omitempty"`
	Score       float64                `json:"score,omitempty"`
	Source      map[string]interface{} `json:"source"`
	Highlight   map[string][]string    `json:"highlight,omitempty"`
	Sort        []interface{}          `json:"sort,omitempty"`
	Explanation *ScoreExplanation      `json:"explanation,omitempty"`
	Version     int64                  `json:"version,omitempty"`
	SeqNo       int64                  `json:"seq_no,omitempty"`
	PrimaryTerm int64                  `json:"primary_term,omitempty"`
}

// ScoreExplanation 得分解释
type ScoreExplanation struct {
	Value       float64              `json:"value"`
	Description string               `json:"description"`
	Details     []*ScoreExplanation  `json:"details,omitempty"`
}

// AggregationResult 聚合结果
type AggregationResult struct {
	Type         string                    `json:"type"`
	Value        interface{}               `json:"value,omitempty"`
	Buckets      []*AggregationBucket      `json:"buckets,omitempty"`
	DocCount     int64                     `json:"doc_count,omitempty"`
	DocCountErrorUpperBound int64          `json:"doc_count_error_upper_bound,omitempty"`
	SumOtherDocCount int64                 `json:"sum_other_doc_count,omitempty"`
	SubAggregations map[string]*AggregationResult `json:"sub_aggregations,omitempty"`
}

// AggregationBucket 聚合桶
type AggregationBucket struct {
	Key             interface{}               `json:"key"`
	KeyAsString     string                    `json:"key_as_string,omitempty"`
	DocCount        int64                     `json:"doc_count"`
	From            interface{}               `json:"from,omitempty"`
	To              interface{}               `json:"to,omitempty"`
	SubAggregations map[string]*AggregationResult `json:"sub_aggregations,omitempty"`
}

// SuggestionResult 建议结果
type SuggestionResult struct {
	Text    string             `json:"text"`
	Offset  int                `json:"offset"`
	Length  int                `json:"length"`
	Options []*SuggestionOption `json:"options"`
}

// SuggestionOption 建议选项
type SuggestionOption struct {
	Text        string                 `json:"text"`
	Score       float64                `json:"score"`
	Freq        int                    `json:"freq,omitempty"`
	Highlighted string                 `json:"highlighted,omitempty"`
	CollateMatch bool                  `json:"collate_match,omitempty"`
	Contexts    map[string][]string    `json:"contexts,omitempty"`
	Payload     map[string]interface{} `json:"payload,omitempty"`
}

// SearchMetadata 搜索元数据
type SearchMetadata struct {
	Database        string            `json:"database"`
	Table           string            `json:"table"`
	Index           string            `json:"index,omitempty"`
	QueryType       string            `json:"query_type"`
	ParsedQuery     string            `json:"parsed_query,omitempty"`
	FilteredCount   int64             `json:"filtered_count,omitempty"`
	IndexedCount    int64             `json:"indexed_count,omitempty"`
	CacheUsed       bool              `json:"cache_used"`
	CacheKey        string            `json:"cache_key,omitempty"`
	ShardInfo       []*ShardInfo      `json:"shard_info,omitempty"`
	NodeInfo        []*NodeInfo       `json:"node_info,omitempty"`
	TimedOut        bool              `json:"timed_out"`
	PartialResults  bool              `json:"partial_results"`
}

// ShardInfo 分片信息
type ShardInfo struct {
	Index       string `json:"index"`
	Shard       int    `json:"shard"`
	Node        string `json:"node"`
	Total       int64  `json:"total"`
	Successful  int64  `json:"successful"`
	Failed      int64  `json:"failed"`
	Skipped     int64  `json:"skipped"`
}

// NodeInfo 节点信息
type NodeInfo struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Host    string `json:"host"`
	Version string `json:"version"`
	Roles   []string `json:"roles"`
}

// SearchDebugInfo 搜索调试信息
type SearchDebugInfo struct {
	Query           string                 `json:"query"`
	ParsedQuery     map[string]interface{} `json:"parsed_query"`
	RewrittenQuery  string                 `json:"rewritten_query,omitempty"`
	ExecutionPlan   string                 `json:"execution_plan,omitempty"`
	ProfileData     map[string]interface{} `json:"profile_data,omitempty"`
	SlowOperations  []string               `json:"slow_operations,omitempty"`
	CacheStatistics map[string]interface{} `json:"cache_statistics,omitempty"`
}

// IndexResponse 索引响应结构体
type IndexResponse struct {
	BaseResponse

	// 索引信息
	Index        *IndexInfo         `json:"index,omitempty"`
	Indexes      []*IndexInfo       `json:"indexes,omitempty"`

	// 操作结果
	Operation    *OperationResult   `json:"operation,omitempty"`

	// 性能指标
	Performance  *PerformanceMetrics `json:"performance,omitempty"`

	// 分页信息
	Pagination   *PaginationInfo    `json:"pagination,omitempty"`
}

// IndexInfo 索引信息
type IndexInfo struct {
	ID           string                 `json:"id"`
	Name         string                 `json:"name"`
	Database     string                 `json:"database"`
	Table        string                 `json:"table"`
	Type         string                 `json:"type"`
	Engine       string                 `json:"engine"`
	Status       IndexStatus            `json:"status"`
	Health       IndexHealth            `json:"health"`
	CreatedAt    time.Time              `json:"created_at"`
	UpdatedAt    time.Time              `json:"updated_at"`
	Version      int64                  `json:"version"`
	Settings     map[string]interface{} `json:"settings,omitempty"`
	Mappings     map[string]interface{} `json:"mappings,omitempty"`
	Aliases      []string               `json:"aliases,omitempty"`
	Stats        *IndexStats            `json:"stats,omitempty"`
	Shards       *ShardStats            `json:"shards,omitempty"`
	Description  string                 `json:"description,omitempty"`
	Tags         map[string]string      `json:"tags,omitempty"`
}

// IndexStats 索引统计信息
type IndexStats struct {
	DocumentCount int64                  `json:"document_count"`
	SizeInBytes   int64                  `json:"size_in_bytes"`
	IndexSize     int64                  `json:"index_size"`
	QueryCount    int64                  `json:"query_count"`
	QueryLatency  *LatencyStats          `json:"query_latency,omitempty"`
	IndexLatency  *LatencyStats          `json:"index_latency,omitempty"`
	FieldStats    map[string]*FieldStats `json:"field_stats,omitempty"`
	LastUpdated   time.Time              `json:"last_updated"`
}

// LatencyStats 延迟统计
type LatencyStats struct {
	Min     int64   `json:"min"`     // microseconds
	Max     int64   `json:"max"`     // microseconds
	Avg     int64   `json:"avg"`     // microseconds
	P50     int64   `json:"p50"`     // microseconds
	P95     int64   `json:"p95"`     // microseconds
	P99     int64   `json:"p99"`     // microseconds
	Count   int64   `json:"count"`
	Sum     int64   `json:"sum"`     // microseconds
}

// FieldStats 字段统计
type FieldStats struct {
	Type         string      `json:"type"`
	Count        int64       `json:"count"`
	NullCount    int64       `json:"null_count"`
	UniqueCount  int64       `json:"unique_count"`
	MinValue     interface{} `json:"min_value,omitempty"`
	MaxValue     interface{} `json:"max_value,omitempty"`
	AvgLength    float64     `json:"avg_length,omitempty"`
	Cardinality  int64       `json:"cardinality,omitempty"`
}

// ShardStats 分片统计
type ShardStats struct {
	Total      int                    `json:"total"`
	Successful int                    `json:"successful"`
	Failed     int                    `json:"failed"`
	Skipped    int                    `json:"skipped"`
	Details    []*ShardDetail         `json:"details,omitempty"`
}

// ShardDetail 分片详情
type ShardDetail struct {
	Index       string    `json:"index"`
	Shard       int       `json:"shard"`
	Primary     bool      `json:"primary"`
	Node        string    `json:"node"`
	State       string    `json:"state"`
	Docs        int64     `json:"docs"`
	Store       int64     `json:"store"`      // bytes
	Segments    int       `json:"segments"`
	LastActive  time.Time `json:"last_active"`
}

// OperationResult 操作结果
type OperationResult struct {
	Type          OperationType          `json:"type"`
	Status        OperationStatus        `json:"status"`
	StartTime     time.Time              `json:"start_time"`
	EndTime       *time.Time             `json:"end_time,omitempty"`
	Duration      int64                  `json:"duration,omitempty"` // milliseconds
	Progress      *OperationProgress     `json:"progress,omitempty"`
	Result        map[string]interface{} `json:"result,omitempty"`
	Errors        []string               `json:"errors,omitempty"`
	Warnings      []string               `json:"warnings,omitempty"`
	TaskID        string                 `json:"task_id,omitempty"`
	RetryCount    int                    `json:"retry_count,omitempty"`
	NextRetryAt   *time.Time             `json:"next_retry_at,omitempty"`
}

// OperationProgress 操作进度
type OperationProgress struct {
	Phase           string    `json:"phase"`
	Percentage      float64   `json:"percentage"`
	Current         int64     `json:"current"`
	Total           int64     `json:"total"`
	EstimatedEnd    *time.Time `json:"estimated_end,omitempty"`
	RemainingTime   int64     `json:"remaining_time,omitempty"` // seconds
	ThroughputPerSec float64  `json:"throughput_per_sec,omitempty"`
	Message         string    `json:"message,omitempty"`
}

// HealthResponse 健康检查响应结构体
type HealthResponse struct {
	BaseResponse

	// 服务状态
	Status       HealthStatus           `json:"status"`
	Version      string                 `json:"version"`
	Uptime       int64                  `json:"uptime"` // seconds

	// 检查结果
	Checks       map[string]*HealthCheck `json:"checks"`

	// 依赖服务状态
	Dependencies map[string]*DependencyStatus `json:"dependencies,omitempty"`

	// 系统信息
	System       *SystemInfo            `json:"system,omitempty"`

	// 性能指标
	Performance  *PerformanceMetrics    `json:"performance,omitempty"`
}

// HealthCheck 健康检查项
type HealthCheck struct {
	Name        string                 `json:"name"`
	Status      HealthStatus           `json:"status"`
	Message     string                 `json:"message,omitempty"`
	Duration    int64                  `json:"duration,omitempty"` // milliseconds
	LastCheck   time.Time              `json:"last_check"`
	Details     map[string]interface{} `json:"details,omitempty"`
	Critical    bool                   `json:"critical"`
	Timeout     int64                  `json:"timeout,omitempty"` // milliseconds
}

// DependencyStatus 依赖服务状态
type DependencyStatus struct {
	Name        string                 `json:"name"`
	Type        string                 `json:"type"`
	Status      HealthStatus           `json:"status"`
	Version     string                 `json:"version,omitempty"`
	Endpoint    string                 `json:"endpoint,omitempty"`
	Latency     int64                  `json:"latency,omitempty"` // milliseconds
	LastCheck   time.Time              `json:"last_check"`
	ErrorCount  int64                  `json:"error_count"`
	SuccessRate float64                `json:"success_rate"`
	Config      map[string]interface{} `json:"config,omitempty"`
}

// SystemInfo 系统信息
type SystemInfo struct {
	Hostname     string                 `json:"hostname"`
	Platform     string                 `json:"platform"`
	Architecture string                 `json:"architecture"`
	CPUCount     int                    `json:"cpu_count"`
	Memory       *MemoryInfo            `json:"memory"`
	Disk         *DiskInfo              `json:"disk"`
	Network      *NetworkInfo           `json:"network,omitempty"`
	GoVersion    string                 `json:"go_version"`
	BuildInfo    *BuildInfo             `json:"build_info,omitempty"`
	Environment  map[string]string      `json:"environment,omitempty"`
}

// MemoryInfo 内存信息
type MemoryInfo struct {
	Total       int64   `json:"total"`        // bytes
	Available   int64   `json:"available"`    // bytes
	Used        int64   `json:"used"`         // bytes
	Free        int64   `json:"free"`         // bytes
	UsagePercent float64 `json:"usage_percent"`
	SwapTotal   int64   `json:"swap_total"`   // bytes
	SwapUsed    int64   `json:"swap_used"`    // bytes
}

// DiskInfo 磁盘信息
type DiskInfo struct {
	Total       int64   `json:"total"`        // bytes
	Used        int64   `json:"used"`         // bytes
	Free        int64   `json:"free"`         // bytes
	UsagePercent float64 `json:"usage_percent"`
	IOStats     *IOStats `json:"io_stats,omitempty"`
}

// IOStats IO统计
type IOStats struct {
	ReadBytes    int64 `json:"read_bytes"`
	WriteBytes   int64 `json:"write_bytes"`
	ReadCount    int64 `json:"read_count"`
	WriteCount   int64 `json:"write_count"`
	ReadTime     int64 `json:"read_time"`     // milliseconds
	WriteTime    int64 `json:"write_time"`    // milliseconds
}

// NetworkInfo 网络信息
type NetworkInfo struct {
	Interfaces []*NetworkInterface    `json:"interfaces,omitempty"`
	Stats      *NetworkStats          `json:"stats,omitempty"`
}

// NetworkInterface 网络接口
type NetworkInterface struct {
	Name        string   `json:"name"`
	HardwareAddr string  `json:"hardware_addr,omitempty"`
	Flags       []string `json:"flags,omitempty"`
	Addrs       []string `json:"addrs,omitempty"`
	MTU         int      `json:"mtu,omitempty"`
}

// NetworkStats 网络统计
type NetworkStats struct {
	BytesSent   int64 `json:"bytes_sent"`
	BytesRecv   int64 `json:"bytes_recv"`
	PacketsSent int64 `json:"packets_sent"`
	PacketsRecv int64 `json:"packets_recv"`
	ErrorsIn    int64 `json:"errors_in"`
	ErrorsOut   int64 `json:"errors_out"`
	DropsIn     int64 `json:"drops_in"`
	DropsOut    int64 `json:"drops_out"`
}

// BuildInfo 构建信息
type BuildInfo struct {
	Version     string    `json:"version"`
	GitCommit   string    `json:"git_commit,omitempty"`
	GitBranch   string    `json:"git_branch,omitempty"`
	BuildTime   time.Time `json:"build_time"`
	GoVersion   string    `json:"go_version"`
	Compiler    string    `json:"compiler"`
	Platform    string    `json:"platform"`
}

// MetricsResponse 监控指标响应结构体
type MetricsResponse struct {
	BaseResponse

	// 指标数据
	Metrics      []*MetricData          `json:"metrics"`

	// 时间范围
	TimeRange    *TimeRange             `json:"time_range"`

	// 聚合信息
	Aggregations map[string]*MetricAggregation `json:"aggregations,omitempty"`

	// 性能指标
	Performance  *PerformanceMetrics    `json:"performance,omitempty"`

	// 分页信息
	Pagination   *PaginationInfo        `json:"pagination,omitempty"`
}

// MetricData 指标数据
type MetricData struct {
	Name        string                 `json:"name"`
	Type        MetricType             `json:"type"`
	Unit        string                 `json:"unit,omitempty"`
	Description string                 `json:"description,omitempty"`
	Labels      map[string]string      `json:"labels,omitempty"`
	Values      []*MetricValue         `json:"values"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// MetricValue 指标值
type MetricValue struct {
	Timestamp time.Time `json:"timestamp"`
	Value     float64   `json:"value"`
	Labels    map[string]string `json:"labels,omitempty"`
}

// MetricAggregation 指标聚合
type MetricAggregation struct {
	Type        string    `json:"type"`
	Value       float64   `json:"value"`
	Count       int64     `json:"count"`
	Min         float64   `json:"min,omitempty"`
	Max         float64   `json:"max,omitempty"`
	Avg         float64   `json:"avg,omitempty"`
	Sum         float64   `json:"sum,omitempty"`
	StdDev      float64   `json:"std_dev,omitempty"`
	Percentiles map[string]float64 `json:"percentiles,omitempty"`
}

// TimeRange 时间范围
type TimeRange struct {
	Start    time.Time `json:"start"`
	End      time.Time `json:"end"`
	Duration int64     `json:"duration"` // seconds
	Step     int64     `json:"step,omitempty"` // seconds
}

// ConfigResponse 配置响应结构体
type ConfigResponse struct {
	BaseResponse

	// 配置数据
	Config       *ConfigData            `json:"config,omitempty"`
	Configs      []*ConfigData          `json:"configs,omitempty"`

	// 分页信息
	Pagination   *PaginationInfo        `json:"pagination,omitempty"`
}

// ConfigData 配置数据
type ConfigData struct {
	Key         string                 `json:"key"`
	Value       interface{}            `json:"value"`
	Type        string                 `json:"type"`
	Namespace   string                 `json:"namespace,omitempty"`
	Environment string                 `json:"environment,omitempty"`
	Version     string                 `json:"version"`
	Description string                 `json:"description,omitempty"`
	Tags        map[string]string      `json:"tags,omitempty"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
	CreatedBy   string                 `json:"created_by,omitempty"`
	UpdatedBy   string                 `json:"updated_by,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// 枚举类型定义
type (
	ErrorSeverity    string
	IndexStatus      string
	IndexHealth      string
	OperationType    string
	OperationStatus  string
	HealthStatus     string
	MetricType       string
)

// 枚举常量定义
const (
	// ErrorSeverity 常量
	ErrorSeverityLow      ErrorSeverity = "low"
	ErrorSeverityMedium   ErrorSeverity = "medium"
	ErrorSeverityHigh     ErrorSeverity = "high"
	ErrorSeverityCritical ErrorSeverity = "critical"

	// IndexStatus 常量
	IndexStatusCreating   IndexStatus = "creating"
	IndexStatusActive     IndexStatus = "active"
	IndexStatusInactive   IndexStatus = "inactive"
	IndexStatusBuilding   IndexStatus = "building"
	IndexStatusRebuilding IndexStatus = "rebuilding"
	IndexStatusDeleting   IndexStatus = "deleting"
	IndexStatusError      IndexStatus = "error"

	// IndexHealth 常量
	IndexHealthGreen  IndexHealth = "green"
	IndexHealthYellow IndexHealth = "yellow"
	IndexHealthRed    IndexHealth = "red"
	IndexHealthGray   IndexHealth = "gray"

	// OperationType 常量
	OperationTypeCreate   OperationType = "create"
	OperationTypeUpdate   OperationType = "update"
	OperationTypeDelete   OperationType = "delete"
	OperationTypeRebuild  OperationType = "rebuild"
	OperationTypeOptimize OperationType = "optimize"
	OperationTypeBackup   OperationType = "backup"
	OperationTypeRestore  OperationType = "restore"

	// OperationStatus 常量
	OperationStatusPending    OperationStatus = "pending"
	OperationStatusRunning    OperationStatus = "running"
	OperationStatusCompleted  OperationStatus = "completed"
	OperationStatusFailed     OperationStatus = "failed"
	OperationStatusCancelled  OperationStatus = "cancelled"
	OperationStatusTimeout    OperationStatus = "timeout"

	// HealthStatus 常量
	HealthStatusHealthy   HealthStatus = "healthy"
	HealthStatusDegraded  HealthStatus = "degraded"
	HealthStatusUnhealthy HealthStatus = "unhealthy"
	HealthStatusUnknown   HealthStatus = "unknown"

	// MetricType 常量
	MetricTypeCounter   MetricType = "counter"
	MetricTypeGauge     MetricType = "gauge"
	MetricTypeHistogram MetricType = "histogram"
	MetricTypeSummary   MetricType = "summary"
)

// 响应构造函数

// NewSuccessResponse 创建成功响应
func NewSuccessResponse(data interface{}) *BaseResponse {
	return &BaseResponse{
		Code:      http.StatusOK,
		Message:   "Success",
		Success:   true,
		Data:      data,
		Timestamp: time.Now(),
	}
}

// NewErrorResponse 创建错误响应
func NewErrorResponse(code int, message string, err error) *BaseResponse {
	var errorInfo *ErrorInfo
	if err != nil {
		errorInfo = &ErrorInfo{
			Code:      "INTERNAL_ERROR",
			Message:   err.Error(),
			Timestamp: time.Now(),
			Severity:  ErrorSeverityMedium,
			Retryable: false,
		}
	}

	return &BaseResponse{
		Code:      code,
		Message:   message,
		Success:   false,
		Error:     errorInfo,
		Timestamp: time.Now(),
	}
}

// NewSearchResponse 创建搜索响应
func NewSearchResponse(hits []*SearchHit, total int64, pagination *PaginationInfo) *SearchResponse {
	return &SearchResponse{
		BaseResponse: *NewSuccessResponse(nil),
		Results:      hits,
		Total:        total,
		Pagination:   pagination,
	}
}

// NewIndexResponse 创建索引响应
func NewIndexResponse(index *IndexInfo) *IndexResponse {
	return &IndexResponse{
		BaseResponse: *NewSuccessResponse(nil),
		Index:        index,
	}
}

// NewIndexListResponse 创建索引列表响应
func NewIndexListResponse(indexes []*IndexInfo, pagination *PaginationInfo) *IndexResponse {
	return &IndexResponse{
		BaseResponse: *NewSuccessResponse(nil),
		Indexes:      indexes,
		Pagination:   pagination,
	}
}

// NewHealthResponse 创建健康检查响应
func NewHealthResponse(status HealthStatus, checks map[string]*HealthCheck) *HealthResponse {
	code := http.StatusOK
	if status != HealthStatusHealthy {
		code = http.StatusServiceUnavailable
	}

	return &HealthResponse{
		BaseResponse: BaseResponse{
			Code:      code,
			Message:   "Health check completed",
			Success:   status == HealthStatusHealthy,
			Timestamp: time.Now(),
		},
		Status: status,
		Checks: checks,
	}
}

// NewMetricsResponse 创建指标响应
func NewMetricsResponse(metrics []*MetricData, timeRange *TimeRange) *MetricsResponse {
	return &MetricsResponse{
		BaseResponse: *NewSuccessResponse(nil),
		Metrics:      metrics,
		TimeRange:    timeRange,
	}
}

// NewConfigResponse 创建配置响应
func NewConfigResponse(config *ConfigData) *ConfigResponse {
	return &ConfigResponse{
		BaseResponse: *NewSuccessResponse(nil),
		Config:       config,
	}
}

// NewConfigListResponse 创建配置列表响应
func NewConfigListResponse(configs []*ConfigData, pagination *PaginationInfo) *ConfigResponse {
	return &ConfigResponse{
		BaseResponse: *NewSuccessResponse(nil),
		Configs:      configs,
		Pagination:   pagination,
	}
}

// 响应方法

// SetRequestID 设置请求ID
func (r *BaseResponse) SetRequestID(requestID string) *BaseResponse {
	r.RequestID = requestID
	return r
}

// SetTraceID 设置追踪ID
func (r *BaseResponse) SetTraceID(traceID string) *BaseResponse {
	r.TraceID = traceID
	return r
}

// SetDuration 设置处理时长
func (r *BaseResponse) SetDuration(duration time.Duration) *BaseResponse {
	r.Duration = duration.Milliseconds()
	return r
}

// SetVersion 设置版本
func (r *BaseResponse) SetVersion(version string) *BaseResponse {
	r.Version = version
	return r
}

// SetPath 设置请求路径
func (r *BaseResponse) SetPath(path string) *BaseResponse {
	r.Path = path
	return r
}

// SetMethod 设置请求方法
func (r *BaseResponse) SetMethod(method string) *BaseResponse {
	r.Method = method
	return r
}

// AddWarning 添加警告信息
func (r *SearchResponse) AddWarning(warning string) *SearchResponse {
	r.Warnings = append(r.Warnings, warning)
	return r
}

// SetAggregations 设置聚合结果
func (r *SearchResponse) SetAggregations(aggregations map[string]*AggregationResult) *SearchResponse {
	r.Aggregations = aggregations
	return r
}

// SetSuggestions 设置建议结果
func (r *SearchResponse) SetSuggestions(suggestions map[string]*SuggestionResult) *SearchResponse {
	r.Suggestions = suggestions
	return r
}

// SetMetadata 设置搜索元数据
func (r *SearchResponse) SetMetadata(metadata *SearchMetadata) *SearchResponse {
	r.Metadata = metadata
	return r
}

// SetPerformance 设置性能指标
func (r *SearchResponse) SetPerformance(performance *PerformanceMetrics) *SearchResponse {
	r.Performance = performance
	return r
}

// SetDebug 设置调试信息
func (r *SearchResponse) SetDebug(debug *SearchDebugInfo) *SearchResponse {
	r.Debug = debug
	return r
}

// SetOperation 设置操作结果
func (r *IndexResponse) SetOperation(operation *OperationResult) *IndexResponse {
	r.Operation = operation
	return r
}

// AddDependency 添加依赖服务状态
func (r *HealthResponse) AddDependency(name string, dependency *DependencyStatus) *HealthResponse {
	if r.Dependencies == nil {
		r.Dependencies = make(map[string]*DependencyStatus)
	}
	r.Dependencies[name] = dependency
	return r
}

// SetSystem 设置系统信息
func (r *HealthResponse) SetSystem(system *SystemInfo) *HealthResponse {
	r.System = system
	return r
}

// SetUptime 设置运行时间
func (r *HealthResponse) SetUptime(uptime time.Duration) *HealthResponse {
	r.Uptime = int64(uptime.Seconds())
	return r
}

// AddMetric 添加指标数据
func (r *MetricsResponse) AddMetric(metric *MetricData) *MetricsResponse {
	r.Metrics = append(r.Metrics, metric)
	return r
}

// SetTimeRange 设置时间范围
func (r *MetricsResponse) SetTimeRange(start, end time.Time) *MetricsResponse {
	r.TimeRange = &TimeRange{
		Start:    start,
		End:      end,
		Duration: int64(end.Sub(start).Seconds()),
	}
	return r
}

// JSON序列化方法

// ToJSON 转换为JSON字符串
func (r *BaseResponse) ToJSON() (string, error) {
	data, err := json.Marshal(r)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// ToJSONIndent 转换为格式化的JSON字符串
func (r *BaseResponse) ToJSONIndent() (string, error) {
	data, err := json.MarshalIndent(r, "", "  ")
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// String 返回字符串表示
func (r *BaseResponse) String() string {
	return fmt.Sprintf("Response{Code:%d, Success:%t, Message:%s}", r.Code, r.Success, r.Message)
}

// IsSuccess 检查是否成功
func (r *BaseResponse) IsSuccess() bool {
	return r.Success && r.Code >= 200 && r.Code < 300
}

// IsError 检查是否错误
func (r *BaseResponse) IsError() bool {
	return !r.Success || r.Code >= 400
}

// GetStatusText 获取状态文本
func (r *BaseResponse) GetStatusText() string {
	return http.StatusText(r.Code)
}

// 分页计算方法

// NewPaginationInfo 创建分页信息
func NewPaginationInfo(page, size int, total int64) *PaginationInfo {
	totalPages := int((total + int64(size) - 1) / int64(size))
	offset := (page - 1) * size

	return &PaginationInfo{
		Page:        page,
		Size:        size,
		Total:       total,
		TotalPages:  totalPages,
		HasNext:     page < totalPages,
		HasPrevious: page > 1,
		Offset:      offset,
		Limit:       size,
	}
}

// SetCursors 设置游标
func (p *PaginationInfo) SetCursors(nextCursor, prevCursor string) *PaginationInfo {
	p.NextCursor = nextCursor
	p.PrevCursor = prevCursor
	return p
}

//Personal.AI order the ending
