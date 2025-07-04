package models

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"
)

// IndexConfig 索引配置结构体
type IndexConfig struct {
	// 基础信息
	ID          string            `json:"id,omitempty"`
	Name        string            `json:"name" validate:"required,min=1,max=100"`
	Database    string            `json:"database" validate:"required,min=1,max=100"`
	Table       string            `json:"table" validate:"required,min=1,max=100"`

	// 索引类型和配置
	Type        IndexType         `json:"type" validate:"required,oneof=inverted bitmap bloom_filter ngram"`
	Engine      DatabaseEngine    `json:"engine" validate:"required,oneof=starrocks clickhouse doris"`

	// 列配置
	Columns     []*IndexColumn    `json:"columns" validate:"required,min=1,dive"`

	// 分词器配置
	Tokenizer   *TokenizerConfig  `json:"tokenizer,omitempty"`

	// 索引设置
	Settings    *IndexSettings    `json:"settings,omitempty"`

	// 分区配置
	Partition   *PartitionConfig  `json:"partition,omitempty"`

	// 分桶配置
	Bucket      *BucketConfig     `json:"bucket,omitempty"`

	// 副本配置
	Replication *ReplicationConfig `json:"replication,omitempty"`

	// 压缩配置
	Compression *CompressionConfig `json:"compression,omitempty"`

	// 索引策略
	Strategy    *IndexStrategy    `json:"strategy,omitempty"`

	// 元数据
	Description string            `json:"description,omitempty"`
	Tags        map[string]string `json:"tags,omitempty"`

	// 时间戳
	CreatedAt   time.Time         `json:"created_at,omitempty"`
	UpdatedAt   time.Time         `json:"updated_at,omitempty"`
	CreatedBy   string            `json:"created_by,omitempty"`
	UpdatedBy   string            `json:"updated_by,omitempty"`

	// 版本信息
	Version     int64             `json:"version,omitempty"`
	ConfigHash  string            `json:"config_hash,omitempty"`
}

// IndexMetadata 索引元数据结构体
type IndexMetadata struct {
	// 基础信息
	IndexID     string            `json:"index_id"`
	Name        string            `json:"name"`
	Database    string            `json:"database"`
	Table       string            `json:"table"`

	// 状态信息
	Status      *IndexStatus      `json:"status"`
	Health      IndexHealth       `json:"health"`

	// 统计信息
	Statistics  *IndexStatistics  `json:"statistics"`

	// 分片信息
	Shards      []*ShardMetadata  `json:"shards,omitempty"`

	// 分区信息
	Partitions  []*PartitionMetadata `json:"partitions,omitempty"`

	// 副本信息
	Replicas    []*ReplicaMetadata `json:"replicas,omitempty"`

	// 性能指标
	Performance *PerformanceMetrics `json:"performance,omitempty"`

	// 存储信息
	Storage     *StorageMetadata  `json:"storage,omitempty"`

	// 时间信息
	CreatedAt   time.Time         `json:"created_at"`
	UpdatedAt   time.Time         `json:"updated_at"`
	LastAccessed time.Time        `json:"last_accessed,omitempty"`
	LastModified time.Time        `json:"last_modified,omitempty"`

	// 配置快照
	Config      *IndexConfig      `json:"config,omitempty"`

	// 依赖关系
	Dependencies []string         `json:"dependencies,omitempty"`

	// 标签和注释
	Labels      map[string]string `json:"labels,omitempty"`
	Annotations map[string]string `json:"annotations,omitempty"`
}

// TokenizerConfig 分词器配置结构体
type TokenizerConfig struct {
	// 分词器类型
	Type        TokenizerType     `json:"type" validate:"required,oneof=standard ik jieba custom"`

	// 语言设置
	Language    string            `json:"language,omitempty" validate:"oneof=chinese english japanese korean"`

	// 标准分词器配置
	MaxTokenLength int            `json:"max_token_length,omitempty" validate:"min=1,max=1000"`

	// IK分词器配置
	IKConfig    *IKTokenizerConfig `json:"ik_config,omitempty"`

	// Jieba分词器配置
	JiebaConfig *JiebaTokenizerConfig `json:"jieba_config,omitempty"`

	// 自定义分词器配置
	CustomConfig *CustomTokenizerConfig `json:"custom_config,omitempty"`

	// 过滤器配置
	Filters     []*TokenFilter    `json:"filters,omitempty"`

	// 字符过滤器
	CharFilters []*CharFilter     `json:"char_filters,omitempty"`

	// 停用词配置
	StopWords   *StopWordsConfig  `json:"stop_words,omitempty"`

	// 同义词配置
	Synonyms    *SynonymsConfig   `json:"synonyms,omitempty"`

	// 词典配置
	Dictionary  *DictionaryConfig `json:"dictionary,omitempty"`

	// 输出设置
	LowerCase   bool              `json:"lower_case,omitempty"`
	RemoveDuplicates bool         `json:"remove_duplicates,omitempty"`

	// 调试设置
	Debug       bool              `json:"debug,omitempty"`
	LogLevel    string            `json:"log_level,omitempty"`
}

// IndexStatus 索引状态结构体
type IndexStatus struct {
	// 基础状态
	State       IndexState        `json:"state"`
	Enabled     bool              `json:"enabled"`
	Visible     bool              `json:"visible"`

	// 健康状态
	Health      IndexHealth       `json:"health"`
	HealthMessage string          `json:"health_message,omitempty"`

	// 操作状态
	Building    bool              `json:"building"`
	Rebuilding  bool              `json:"rebuilding"`
	Optimizing  bool              `json:"optimizing"`

	// 进度信息
	Progress    *OperationProgress `json:"progress,omitempty"`

	// 错误信息
	Errors      []string          `json:"errors,omitempty"`
	Warnings    []string          `json:"warnings,omitempty"`

	// 最后操作
	LastOperation *Operation      `json:"last_operation,omitempty"`

	// 时间戳
	StateChangedAt time.Time      `json:"state_changed_at"`
	LastCheckedAt  time.Time      `json:"last_checked_at"`

	// 版本信息
	Version     string            `json:"version,omitempty"`
	BuildVersion string           `json:"build_version,omitempty"`
}

// TableSchema 表结构信息结构体
type TableSchema struct {
	// 基础信息
	Database    string            `json:"database"`
	Table       string            `json:"table"`
	Type        TableType         `json:"type"`
	Engine      string            `json:"engine"`

	// 列定义
	Columns     []*ColumnInfo     `json:"columns" validate:"required,min=1,dive"`

	// 主键
	PrimaryKey  []string          `json:"primary_key,omitempty"`

	// 索引定义
	Indexes     []*IndexDefinition `json:"indexes,omitempty"`

	// 分区定义
	Partitioning *PartitioningDefinition `json:"partitioning,omitempty"`

	// 分桶定义
	Bucketing   *BucketingDefinition `json:"bucketing,omitempty"`

	// 排序键
	SortKey     []string          `json:"sort_key,omitempty"`

	// 表属性
	Properties  map[string]string `json:"properties,omitempty"`

	// 存储格式
	Format      *StorageFormat    `json:"format,omitempty"`

	// 压缩配置
	Compression *CompressionConfig `json:"compression,omitempty"`

	// 统计信息
	Statistics  *TableStatistics  `json:"statistics,omitempty"`

	// 元数据
	Comment     string            `json:"comment,omitempty"`
	CreatedAt   time.Time         `json:"created_at"`
	UpdatedAt   time.Time         `json:"updated_at"`

	// 数据库特定配置
	StarRocksConfig *StarRocksTableConfig `json:"starrocks_config,omitempty"`
	ClickHouseConfig *ClickHouseTableConfig `json:"clickhouse_config,omitempty"`
	DorisConfig     *DorisTableConfig      `json:"doris_config,omitempty"`
}

// ColumnInfo 列信息结构体
type ColumnInfo struct {
	// 基础信息
	Name        string            `json:"name" validate:"required,min=1,max=100"`
	Type        DataType          `json:"type" validate:"required"`
	Size        int               `json:"size,omitempty"`
	Precision   int               `json:"precision,omitempty"`
	Scale       int               `json:"scale,omitempty"`

	// 约束
	Nullable    bool              `json:"nullable"`
	PrimaryKey  bool              `json:"primary_key"`
	Unique      bool              `json:"unique"`
	AutoIncrement bool            `json:"auto_increment"`

	// 默认值
	DefaultValue interface{}      `json:"default_value,omitempty"`
	DefaultExpression string      `json:"default_expression,omitempty"`

	// 编码和压缩
	Encoding    ColumnEncoding    `json:"encoding,omitempty"`
	Compression ColumnCompression `json:"compression,omitempty"`

	// 索引配置
	Indexed     bool              `json:"indexed"`
	IndexType   IndexType         `json:"index_type,omitempty"`

	// 分析配置
	Analyzer    string            `json:"analyzer,omitempty"`
	SearchAnalyzer string         `json:"search_analyzer,omitempty"`

	// 存储配置
	Store       bool              `json:"store"`
	DocValues   bool              `json:"doc_values"`

	// 字段配置
	FieldData   bool              `json:"field_data"`
	IgnoreMalformed bool          `json:"ignore_malformed"`

	// 映射配置
	Properties  map[string]interface{} `json:"properties,omitempty"`
	Fields      map[string]*ColumnInfo `json:"fields,omitempty"`

	// 元数据
	Comment     string            `json:"comment,omitempty"`
	Tags        []string          `json:"tags,omitempty"`

	// 统计信息
	Statistics  *ColumnStatistics `json:"statistics,omitempty"`

	// 数据库特定配置
	DatabaseSpecific map[string]interface{} `json:"database_specific,omitempty"`
}

// 配置相关结构体

// IndexColumn 索引列配置
type IndexColumn struct {
	Name        string            `json:"name" validate:"required"`
	Type        DataType          `json:"type" validate:"required"`
	Analyzer    string            `json:"analyzer,omitempty"`
	Boost       float64           `json:"boost,omitempty"`
	Store       bool              `json:"store,omitempty"`
	Index       bool              `json:"index,omitempty"`
	DocValues   bool              `json:"doc_values,omitempty"`
	Properties  map[string]interface{} `json:"properties,omitempty"`
}

// IndexSettings 索引设置
type IndexSettings struct {
	// 分片设置
	NumberOfShards   int           `json:"number_of_shards,omitempty"`
	NumberOfReplicas int           `json:"number_of_replicas,omitempty"`

	// 刷新设置
	RefreshInterval  time.Duration `json:"refresh_interval,omitempty"`

	// 合并设置
	MergePolicy     string         `json:"merge_policy,omitempty"`
	MaxMergeAtOnce  int           `json:"max_merge_at_once,omitempty"`

	// 缓存设置
	QueryCache      bool          `json:"query_cache,omitempty"`
	RequestCache    bool          `json:"request_cache,omitempty"`

	// 性能设置
	MaxResultWindow int           `json:"max_result_window,omitempty"`
	MaxRescore      int           `json:"max_rescore,omitempty"`

	// 路由设置
	RoutingRequired bool          `json:"routing_required,omitempty"`

	// 其他设置
	Hidden          bool          `json:"hidden,omitempty"`
	ReadOnly        bool          `json:"read_only,omitempty"`

	// 自定义设置
	Custom          map[string]interface{} `json:"custom,omitempty"`
}

// PartitionConfig 分区配置
type PartitionConfig struct {
	Type        PartitionType     `json:"type" validate:"required,oneof=range list hash"`
	Columns     []string          `json:"columns" validate:"required,min=1"`
	Expression  string            `json:"expression,omitempty"`
	Buckets     int               `json:"buckets,omitempty"`
	Ranges      []*PartitionRange `json:"ranges,omitempty"`
	Values      []interface{}     `json:"values,omitempty"`
	AutoCreate  bool              `json:"auto_create,omitempty"`
}

// BucketConfig 分桶配置
type BucketConfig struct {
	Type        BucketType        `json:"type" validate:"required,oneof=hash random"`
	Columns     []string          `json:"columns,omitempty"`
	Buckets     int               `json:"buckets" validate:"required,min=1"`
	Expression  string            `json:"expression,omitempty"`
}

// ReplicationConfig 副本配置
type ReplicationConfig struct {
	Factor      int               `json:"factor" validate:"min=1,max=10"`
	Allocation  *AllocationConfig `json:"allocation,omitempty"`
	Sync        bool              `json:"sync,omitempty"`
}

// CompressionConfig 压缩配置
type CompressionConfig struct {
	Type        CompressionType   `json:"type" validate:"oneof=none gzip lz4 snappy zstd"`
	Level       int               `json:"level,omitempty"`
	BlockSize   int               `json:"block_size,omitempty"`
	Dictionary  string            `json:"dictionary,omitempty"`
}

// IndexStrategy 索引策略
type IndexStrategy struct {
	BuildMode   BuildMode         `json:"build_mode" validate:"oneof=sync async batch"`
	Priority    int               `json:"priority,omitempty"`
	Resources   *ResourceConfig   `json:"resources,omitempty"`
	Schedule    *ScheduleConfig   `json:"schedule,omitempty"`
}

// 分词器特定配置

// IKTokenizerConfig IK分词器配置
type IKTokenizerConfig struct {
	Mode        string            `json:"mode" validate:"oneof=ik_max_word ik_smart"`
	UseSmart    bool              `json:"use_smart,omitempty"`
	EnableRemoteDict bool         `json:"enable_remote_dict,omitempty"`
	RemoteDictUrl string          `json:"remote_dict_url,omitempty"`
}

// JiebaTokenizerConfig Jieba分词器配置
type JiebaTokenizerConfig struct {
	Mode        string            `json:"mode" validate:"oneof=search index"`
	HMM         bool              `json:"hmm,omitempty"`
	UserDict    string            `json:"user_dict,omitempty"`
}

// CustomTokenizerConfig 自定义分词器配置
type CustomTokenizerConfig struct {
	Pattern     string            `json:"pattern,omitempty"`
	Flags       string            `json:"flags,omitempty"`
	Group       int               `json:"group,omitempty"`
	Script      string            `json:"script,omitempty"`
	Language    string            `json:"language,omitempty"`
	Parameters  map[string]interface{} `json:"parameters,omitempty"`
}

// TokenFilter 词元过滤器
type TokenFilter struct {
	Type        string            `json:"type" validate:"required"`
	Parameters  map[string]interface{} `json:"parameters,omitempty"`
}

// CharFilter 字符过滤器
type CharFilter struct {
	Type        string            `json:"type" validate:"required"`
	Parameters  map[string]interface{} `json:"parameters,omitempty"`
}

// StopWordsConfig 停用词配置
type StopWordsConfig struct {
	Enabled     bool              `json:"enabled"`
	Language    string            `json:"language,omitempty"`
	Words       []string          `json:"words,omitempty"`
	Path        string            `json:"path,omitempty"`
	IgnoreCase  bool              `json:"ignore_case,omitempty"`
}

// SynonymsConfig 同义词配置
type SynonymsConfig struct {
	Enabled     bool              `json:"enabled"`
	Path        string            `json:"path,omitempty"`
	Synonyms    []string          `json:"synonyms,omitempty"`
	Format      string            `json:"format,omitempty"`
	IgnoreCase  bool              `json:"ignore_case,omitempty"`
}

// DictionaryConfig 词典配置
type DictionaryConfig struct {
	Path        string            `json:"path,omitempty"`
	Type        string            `json:"type,omitempty"`
	Encoding    string            `json:"encoding,omitempty"`
	UpdateInterval time.Duration  `json:"update_interval,omitempty"`
}

// 统计和元数据相关结构体

// IndexStatistics 索引统计信息
type IndexStatistics struct {
	DocumentCount   int64         `json:"document_count"`
	DeletedCount    int64         `json:"deleted_count"`
	StorageSize     int64         `json:"storage_size"`
	IndexSize       int64         `json:"index_size"`
	SegmentCount    int           `json:"segment_count"`

	// 查询统计
	QueryCount      int64         `json:"query_count"`
	QueryTime       time.Duration `json:"query_time"`
	AvgQueryTime    time.Duration `json:"avg_query_time"`

	// 索引统计
	IndexingCount   int64         `json:"indexing_count"`
	IndexingTime    time.Duration `json:"indexing_time"`
	AvgIndexingTime time.Duration `json:"avg_indexing_time"`

	// 缓存统计
	CacheHitRate    float64       `json:"cache_hit_rate"`
	CacheSize       int64         `json:"cache_size"`

	// 合并统计
	MergeCount      int64         `json:"merge_count"`
	MergeTime       time.Duration `json:"merge_time"`

	// 更新时间
	UpdatedAt       time.Time     `json:"updated_at"`
}

// ShardMetadata 分片元数据
type ShardMetadata struct {
	ID              string        `json:"id"`
	Index           string        `json:"index"`
	Primary         bool          `json:"primary"`
	State           ShardState    `json:"state"`
	Node            string        `json:"node"`
	Size            int64         `json:"size"`
	DocumentCount   int64         `json:"document_count"`
	RelocatingNode  string        `json:"relocating_node,omitempty"`
	UnassignedReason string       `json:"unassigned_reason,omitempty"`
}

// PartitionMetadata 分区元数据
type PartitionMetadata struct {
	ID              string        `json:"id"`
	Name            string        `json:"name"`
	Range           *PartitionRange `json:"range,omitempty"`
	Size            int64         `json:"size"`
	DocumentCount   int64         `json:"document_count"`
	CreatedAt       time.Time     `json:"created_at"`
	LastModified    time.Time     `json:"last_modified"`
}

// ReplicaMetadata 副本元数据
type ReplicaMetadata struct {
	ID              string        `json:"id"`
	Node            string        `json:"node"`
	State           ReplicaState  `json:"state"`
	Size            int64         `json:"size"`
	Lag             time.Duration `json:"lag,omitempty"`
	LastSyncAt      time.Time     `json:"last_sync_at,omitempty"`
}

// PerformanceMetrics 性能指标
type PerformanceMetrics struct {
	QPS             float64       `json:"qps"`
	AvgLatency      time.Duration `json:"avg_latency"`
	P95Latency      time.Duration `json:"p95_latency"`
	P99Latency      time.Duration `json:"p99_latency"`
	ErrorRate       float64       `json:"error_rate"`
	ThroughputMBPS  float64       `json:"throughput_mbps"`
	CPUUsage        float64       `json:"cpu_usage"`
	MemoryUsage     float64       `json:"memory_usage"`
	DiskIOPS        float64       `json:"disk_iops"`
}

// StorageMetadata 存储元数据
type StorageMetadata struct {
	TotalSize       int64         `json:"total_size"`
	UsedSize        int64         `json:"used_size"`
	FreeSize        int64         `json:"free_size"`
	FileCount       int64         `json:"file_count"`
	CompressionRatio float64      `json:"compression_ratio"`
	StorageType     string        `json:"storage_type"`
	Location        string        `json:"location"`
}

// OperationProgress 操作进度
type OperationProgress struct {
	Type            string        `json:"type"`
	Total           int64         `json:"total"`
	Completed       int64         `json:"completed"`
	Percentage      float64       `json:"percentage"`
	EstimatedTime   time.Duration `json:"estimated_time"`
	StartedAt       time.Time     `json:"started_at"`
	Message         string        `json:"message,omitempty"`
}

// Operation 操作信息
type Operation struct {
	Type            string        `json:"type"`
	Status          string        `json:"status"`
	StartedAt       time.Time     `json:"started_at"`
	CompletedAt     *time.Time    `json:"completed_at,omitempty"`
	Duration        time.Duration `json:"duration"`
	Message         string        `json:"message,omitempty"`
	Progress        *OperationProgress `json:"progress,omitempty"`
}

// 枚举类型定义
type (
	IndexType           string
	DatabaseEngine      string
	IndexState          string
	IndexHealth         string
	TokenizerType       string
	DataType            string
	TableType           string
	ColumnEncoding      string
	ColumnCompression   string
	PartitionType       string
	BucketType          string
	CompressionType     string
	BuildMode           string
	ShardState          string
	ReplicaState        string
)

// IndexType 常量
const (
	IndexTypeInverted   IndexType = "inverted"
	IndexTypeBitmap     IndexType = "bitmap"
	IndexTypeBloomFilter IndexType = "bloom_filter"
	IndexTypeNgram      IndexType = "ngram"
	IndexTypeFullText   IndexType = "fulltext"
)

// DatabaseEngine 常量
const (
	EngineStarRocks  DatabaseEngine = "starrocks"
	EngineClickHouse DatabaseEngine = "clickhouse"
	EngineDoris      DatabaseEngine = "doris"
)

// IndexState 常量
const (
	IndexStateCreating   IndexState = "creating"
	IndexStateActive     IndexState = "active"
	IndexStateInactive   IndexState = "inactive"
	IndexStateBuilding   IndexState = "building"
	IndexStateRebuilding IndexState = "rebuilding"
	IndexStateDeleting   IndexState = "deleting"
	IndexStateError      IndexState = "error"
)

// IndexHealth 常量
const (
	IndexHealthGreen  IndexHealth = "green"
	IndexHealthYellow IndexHealth = "yellow"
	IndexHealthRed    IndexHealth = "red"
	IndexHealthGray   IndexHealth = "gray"
)

// TokenizerType 常量
const (
	TokenizerTypeStandard TokenizerType = "standard"
	TokenizerTypeIK       TokenizerType = "ik"
	TokenizerTypeJieba    TokenizerType = "jieba"
	TokenizerTypeCustom   TokenizerType = "custom"
)

// DataType 常量
const (
	DataTypeString    DataType = "string"
	DataTypeText      DataType = "text"
	DataTypeKeyword   DataType = "keyword"
	DataTypeInteger   DataType = "integer"
	DataTypeLong      DataType = "long"
	DataTypeFloat     DataType = "float"
	DataTypeDouble    DataType = "double"
	DataTypeBoolean   DataType = "boolean"
	DataTypeDate      DataType = "date"
	DataTypeDateTime  DataType = "datetime"
	DataTypeTimestamp DataType = "timestamp"
	DataTypeBinary    DataType = "binary"
	DataTypeJSON      DataType = "json"
	DataTypeArray     DataType = "array"
	DataTypeMap       DataType = "map"
	DataTypeStruct    DataType = "struct"
)

// 验证方法

// Validate 验证IndexConfig
func (c *IndexConfig) Validate() error {
	if c.Name == "" {
		return errors.New("index name is required")
	}

	if c.Database == "" {
		return errors.New("database name is required")
	}

	if c.Table == "" {
		return errors.New("table name is required")
	}

	if c.Type == "" {
		return errors.New("index type is required")
	}

	if c.Engine == "" {
		return errors.New("database engine is required")
	}

	if len(c.Columns) == 0 {
		return errors.New("at least one column is required")
	}

	// 验证列配置
	for i, col := range c.Columns {
		if err := col.Validate(); err != nil {
			return fmt.Errorf("invalid column %d: %w", i, err)
		}
	}

	// 验证分词器配置
	if c.Tokenizer != nil {
		if err := c.Tokenizer.Validate(); err != nil {
			return fmt.Errorf("invalid tokenizer config: %w", err)
		}
	}

	return nil
}

// Validate 验证IndexColumn
func (c *IndexColumn) Validate() error {
	if c.Name == "" {
		return errors.New("column name is required")
	}

	if c.Type == "" {
		return errors.New("column type is required")
	}

	return nil
}

// Validate 验证TokenizerConfig
func (c *TokenizerConfig) Validate() error {
	if c.Type == "" {
		return errors.New("tokenizer type is required")
	}

	switch c.Type {
	case TokenizerTypeIK:
		if c.IKConfig == nil {
			return errors.New("IK tokenizer config is required")
		}
	case TokenizerTypeJieba:
		if c.JiebaConfig == nil {
			return errors.New("Jieba tokenizer config is required")
		}
	case TokenizerTypeCustom:
		if c.CustomConfig == nil {
			return errors.New("custom tokenizer config is required")
		}
	}

	return nil
}

// Validate 验证ColumnInfo
func (c *ColumnInfo) Validate() error {
	if c.Name == "" {
		return errors.New("column name is required")
	}

	if c.Type == "" {
		return errors.New("column type is required")
	}

	// 验证数据类型特定的参数
	switch c.Type {
	case DataTypeString, DataTypeText:
		if c.Size <= 0 {
			return errors.New("size must be positive for string/text types")
		}
	case DataTypeFloat, DataTypeDouble:
		if c.Precision < 0 || c.Scale < 0 {
			return errors.New("precision and scale must be non-negative")
		}
	}

	return nil
}

// ToJSON 转换为JSON字符串
func (c *IndexConfig) ToJSON() (string, error) {
	data, err := json.Marshal(c)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// FromJSON 从JSON字符串解析
func (c *IndexConfig) FromJSON(jsonStr string) error {
	return json.Unmarshal([]byte(jsonStr), c)
}

// GetFullName 获取完整索引名称
func (c *IndexConfig) GetFullName() string {
	return fmt.Sprintf("%s.%s.%s", c.Database, c.Table, c.Name)
}

// IsEnabled 检查索引是否启用
func (s *IndexStatus) IsEnabled() bool {
	return s.Enabled && s.State == IndexStateActive
}

// IsHealthy 检查索引是否健康
func (s *IndexStatus) IsHealthy() bool {
	return s.Health == IndexHealthGreen
}

// HasErrors 检查是否有错误
func (s *IndexStatus) HasErrors() bool {
	return len(s.Errors) > 0
}

// GetColumnByName 根据名称获取列信息
func (s *TableSchema) GetColumnByName(name string) *ColumnInfo {
	for _, col := range s.Columns {
		if col.Name == name {
			return col
		}
	}
	return nil
}

// GetPrimaryKeyColumns 获取主键列
func (s *TableSchema) GetPrimaryKeyColumns() []*ColumnInfo {
	var pkColumns []*ColumnInfo
	for _, col := range s.Columns {
		if col.PrimaryKey {
			pkColumns = append(pkColumns, col)
		}
	}
	return pkColumns
}

// GetIndexedColumns 获取有索引的列
func (s *TableSchema) GetIndexedColumns() []*ColumnInfo {
	var indexedColumns []*ColumnInfo
	for _, col := range s.Columns {
		if col.Indexed {
			indexedColumns = append(indexedColumns, col)
		}
	}
	return indexedColumns
}

// IsTextColumn 检查是否为文本类型列
func (c *ColumnInfo) IsTextColumn() bool {
	return c.Type == DataTypeText || c.Type == DataTypeString
}

// IsNumericColumn 检查是否为数值类型列
func (c *ColumnInfo) IsNumericColumn() bool {
	return c.Type == DataTypeInteger || c.Type == DataTypeLong ||
		c.Type == DataTypeFloat || c.Type == DataTypeDouble
}

// IsDateTimeColumn 检查是否为日期时间类型列
func (c *ColumnInfo) IsDateTimeColumn() bool {
	return c.Type == DataTypeDate || c.Type == DataTypeDateTime ||
		c.Type == DataTypeTimestamp
}

// String 返回字符串表示
func (c *IndexConfig) String() string {
	return fmt.Sprintf("IndexConfig{Name:%s, Database:%s, Table:%s, Type:%s, Engine:%s}",
		c.Name, c.Database, c.Table, c.Type, c.Engine)
}

// String 返回字符串表示
func (s *IndexStatus) String() string {
	return fmt.Sprintf("IndexStatus{State:%s, Health:%s, Enabled:%t}",
		s.State, s.Health, s.Enabled)
}

//Personal.AI order the ending
