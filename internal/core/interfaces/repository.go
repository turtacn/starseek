package interfaces

import (
	"context"
	"time"

	"github.com/turtacn/starseek/internal/core/domain"
)

// SearchRepository 定义搜索数据存储操作接口
type SearchRepository interface {
	// 文档操作
	IndexDocument(ctx context.Context, indexName string, doc *domain.Document) error
	IndexDocuments(ctx context.Context, indexName string, docs []*domain.Document) error
	UpdateDocument(ctx context.Context, indexName string, docID string, doc *domain.Document) error
	DeleteDocument(ctx context.Context, indexName string, docID string) error
	DeleteDocuments(ctx context.Context, indexName string, docIDs []string) error
	GetDocument(ctx context.Context, indexName string, docID string) (*domain.Document, error)
	GetDocuments(ctx context.Context, indexName string, docIDs []string) ([]*domain.Document, error)

	// 搜索操作
	Search(ctx context.Context, req *SearchQuery) (*SearchResults, error)
	SearchWithScroll(ctx context.Context, req *SearchQuery, scrollID string) (*ScrollResults, error)
	MultiSearch(ctx context.Context, queries []*SearchQuery) ([]*SearchResults, error)
	SearchSuggestions(ctx context.Context, indexName string, query string, limit int) ([]string, error)

	// 聚合操作
	Aggregate(ctx context.Context, req *AggregationQuery) (*AggregationResults, error)
	Count(ctx context.Context, indexName string, query string) (int64, error)

	// 批量操作
	BulkIndex(ctx context.Context, operations []*BulkOperation) (*BulkResponse, error)
	BulkUpdate(ctx context.Context, operations []*BulkOperation) (*BulkResponse, error)
	BulkDelete(ctx context.Context, operations []*BulkOperation) (*BulkResponse, error)

	// 事务操作
	BeginTransaction(ctx context.Context) (Transaction, error)
	ExecuteInTransaction(ctx context.Context, tx Transaction, fn func(ctx context.Context, tx Transaction) error) error
}

// IndexRepository 定义索引元数据存储操作接口
type IndexRepository interface {
	// 索引生命周期管理
	CreateIndex(ctx context.Context, config *IndexConfiguration) error
	UpdateIndex(ctx context.Context, indexName string, config *IndexConfiguration) error
	DeleteIndex(ctx context.Context, indexName string) error
	IndexExists(ctx context.Context, indexName string) (bool, error)

	// 索引配置管理
	GetIndexConfig(ctx context.Context, indexName string) (*IndexConfiguration, error)
	SetIndexConfig(ctx context.Context, indexName string, config *IndexConfiguration) error
	ListIndexes(ctx context.Context, filter *IndexFilter) ([]*IndexInfo, error)

	// 索引元数据管理
	GetIndexMetadata(ctx context.Context, indexName string) (*IndexMetadata, error)
	UpdateIndexMetadata(ctx context.Context, indexName string, metadata *IndexMetadata) error
	GetIndexStats(ctx context.Context, indexName string) (*IndexStatistics, error)

	// 索引映射管理
	GetIndexMapping(ctx context.Context, indexName string) (*IndexMapping, error)
	UpdateIndexMapping(ctx context.Context, indexName string, mapping *IndexMapping) error

	// 索引设置管理
	GetIndexSettings(ctx context.Context, indexName string) (*IndexSettings, error)
	UpdateIndexSettings(ctx context.Context, indexName string, settings *IndexSettings) error

	// 索引别名管理
	CreateIndexAlias(ctx context.Context, aliasName string, indexNames []string) error
	UpdateIndexAlias(ctx context.Context, aliasName string, indexNames []string) error
	DeleteIndexAlias(ctx context.Context, aliasName string) error
	GetIndexAliases(ctx context.Context, indexName string) ([]string, error)

	// 索引模板管理
	CreateIndexTemplate(ctx context.Context, template *IndexTemplate) error
	UpdateIndexTemplate(ctx context.Context, templateName string, template *IndexTemplate) error
	DeleteIndexTemplate(ctx context.Context, templateName string) error
	GetIndexTemplate(ctx context.Context, templateName string) (*IndexTemplate, error)
	ListIndexTemplates(ctx context.Context) ([]*IndexTemplate, error)

	// 事务操作
	BeginTransaction(ctx context.Context) (Transaction, error)
	ExecuteInTransaction(ctx context.Context, tx Transaction, fn func(ctx context.Context, tx Transaction) error) error
}

// CacheRepository 定义缓存操作接口
type CacheRepository interface {
	// 基础操作
	Get(ctx context.Context, key string) ([]byte, error)
	Set(ctx context.Context, key string, value []byte, expiration time.Duration) error
	Delete(ctx context.Context, key string) error
	Exists(ctx context.Context, key string) (bool, error)

	// 批量操作
	MGet(ctx context.Context, keys []string) (map[string][]byte, error)
	MSet(ctx context.Context, items map[string][]byte, expiration time.Duration) error
	MDelete(ctx context.Context, keys []string) error

	// 高级操作
	GetSet(ctx context.Context, key string, value []byte) ([]byte, error)
	Increment(ctx context.Context, key string, delta int64) (int64, error)
	Decrement(ctx context.Context, key string, delta int64) (int64, error)
	Expire(ctx context.Context, key string, expiration time.Duration) error
	TTL(ctx context.Context, key string) (time.Duration, error)

	// 模式操作
	Keys(ctx context.Context, pattern string) ([]string, error)
	Scan(ctx context.Context, cursor uint64, match string, count int64) ([]string, uint64, error)
	DeleteByPattern(ctx context.Context, pattern string) error

	// 哈希操作
	HGet(ctx context.Context, key, field string) ([]byte, error)
	HSet(ctx context.Context, key, field string, value []byte) error
	HDelete(ctx context.Context, key string, fields ...string) error
	HGetAll(ctx context.Context, key string) (map[string][]byte, error)
	HExists(ctx context.Context, key, field string) (bool, error)
	HLen(ctx context.Context, key string) (int64, error)

	// 列表操作
	LPush(ctx context.Context, key string, values ...[]byte) error
	RPush(ctx context.Context, key string, values ...[]byte) error
	LPop(ctx context.Context, key string) ([]byte, error)
	RPop(ctx context.Context, key string) ([]byte, error)
	LLen(ctx context.Context, key string) (int64, error)
	LRange(ctx context.Context, key string, start, stop int64) ([][]byte, error)

	// 集合操作
	SAdd(ctx context.Context, key string, members ...[]byte) error
	SRemove(ctx context.Context, key string, members ...[]byte) error
	SMembers(ctx context.Context, key string) ([][]byte, error)
	SIsMember(ctx context.Context, key string, member []byte) (bool, error)
	SCard(ctx context.Context, key string) (int64, error)

	// 有序集合操作
	ZAdd(ctx context.Context, key string, score float64, member []byte) error
	ZRemove(ctx context.Context, key string, members ...[]byte) error
	ZRange(ctx context.Context, key string, start, stop int64) ([][]byte, error)
	ZRangeByScore(ctx context.Context, key string, min, max float64) ([][]byte, error)
	ZScore(ctx context.Context, key string, member []byte) (float64, error)
	ZCard(ctx context.Context, key string) (int64, error)

	// 缓存统计
	GetStats(ctx context.Context) (*CacheStats, error)
	FlushAll(ctx context.Context) error
	FlushDB(ctx context.Context, db int) error

	// 事务操作
	BeginTransaction(ctx context.Context) (CacheTransaction, error)
	ExecuteInTransaction(ctx context.Context, tx CacheTransaction, fn func(ctx context.Context, tx CacheTransaction) error) error
}

// Transaction 定义通用事务接口
type Transaction interface {
	Commit(ctx context.Context) error
	Rollback(ctx context.Context) error
	IsActive() bool
	GetID() string
}

// CacheTransaction 定义缓存事务接口
type CacheTransaction interface {
	Transaction

	// 事务内操作
	Get(ctx context.Context, key string) ([]byte, error)
	Set(ctx context.Context, key string, value []byte, expiration time.Duration) error
	Delete(ctx context.Context, key string) error
	MGet(ctx context.Context, keys []string) (map[string][]byte, error)
	MSet(ctx context.Context, items map[string][]byte, expiration time.Duration) error

	// 监视键
	Watch(ctx context.Context, keys ...string) error
	Unwatch(ctx context.Context) error
}

// 数据结构定义

// SearchQuery 搜索查询结构
type SearchQuery struct {
	Index        string                  `json:"index"`
	Query        string                  `json:"query"`
	Filters      map[string]interface{}  `json:"filters,omitempty"`
	SortBy       []SortField             `json:"sort_by,omitempty"`
	From         int                     `json:"from,omitempty"`
	Size         int                     `json:"size,omitempty"`
	Highlight    *HighlightConfig        `json:"highlight,omitempty"`
	Aggregations map[string]*Aggregation `json:"aggregations,omitempty"`
	Source       []string                `json:"source,omitempty"`
	Timeout      time.Duration           `json:"timeout,omitempty"`
}

// SearchResults 搜索结果结构
type SearchResults struct {
	Total        int64                         `json:"total"`
	MaxScore     float64                       `json:"max_score"`
	Documents    []*SearchDocument             `json:"documents"`
	Aggregations map[string]*AggregationResult `json:"aggregations,omitempty"`
	ScrollID     string                        `json:"scroll_id,omitempty"`
	Took         time.Duration                 `json:"took"`
}

// ScrollResults 滚动搜索结果
type ScrollResults struct {
	SearchResults
	HasMore  bool   `json:"has_more"`
	ScrollID string `json:"scroll_id"`
}

// SearchDocument 搜索文档结构
type SearchDocument struct {
	ID         string                 `json:"id"`
	Index      string                 `json:"index"`
	Score      float64                `json:"score"`
	Source     map[string]interface{} `json:"source"`
	Highlights map[string][]string    `json:"highlights,omitempty"`
}

// SortField 排序字段
type SortField struct {
	Field string    `json:"field"`
	Order SortOrder `json:"order"`
}

// HighlightConfig 高亮配置
type HighlightConfig struct {
	Fields    map[string]*HighlightField `json:"fields"`
	PreTags   []string                   `json:"pre_tags,omitempty"`
	PostTags  []string                   `json:"post_tags,omitempty"`
	MaxLength int                        `json:"max_length,omitempty"`
}

// HighlightField 高亮字段配置
type HighlightField struct {
	FragmentSize      int      `json:"fragment_size,omitempty"`
	NumberOfFragments int      `json:"number_of_fragments,omitempty"`
	PreTags           []string `json:"pre_tags,omitempty"`
	PostTags          []string `json:"post_tags,omitempty"`
}

// Aggregation 聚合查询
type Aggregation struct {
	Type    AggregationType         `json:"type"`
	Field   string                  `json:"field,omitempty"`
	Size    int                     `json:"size,omitempty"`
	Config  map[string]interface{}  `json:"config,omitempty"`
	SubAggs map[string]*Aggregation `json:"sub_aggregations,omitempty"`
}

// AggregationQuery 聚合查询结构
type AggregationQuery struct {
	Index        string                  `json:"index"`
	Query        string                  `json:"query,omitempty"`
	Filters      map[string]interface{}  `json:"filters,omitempty"`
	Aggregations map[string]*Aggregation `json:"aggregations"`
	Timeout      time.Duration           `json:"timeout,omitempty"`
}

// AggregationResults 聚合结果
type AggregationResults struct {
	Results map[string]*AggregationResult `json:"results"`
	Took    time.Duration                 `json:"took"`
}

// AggregationResult 聚合结果项
type AggregationResult struct {
	Type    AggregationType               `json:"type"`
	Buckets []*AggregationBucket          `json:"buckets,omitempty"`
	Value   interface{}                   `json:"value,omitempty"`
	SubAggs map[string]*AggregationResult `json:"sub_aggregations,omitempty"`
}

// AggregationBucket 聚合桶
type AggregationBucket struct {
	Key      interface{}                   `json:"key"`
	DocCount int64                         `json:"doc_count"`
	SubAggs  map[string]*AggregationResult `json:"sub_aggregations,omitempty"`
}

// BulkOperation 批量操作
type BulkOperation struct {
	Type     BulkOperationType      `json:"type"`
	Index    string                 `json:"index"`
	ID       string                 `json:"id,omitempty"`
	Document *domain.Document       `json:"document,omitempty"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// BulkResponse 批量操作响应
type BulkResponse struct {
	Took   time.Duration       `json:"took"`
	Errors bool                `json:"errors"`
	Items  []*BulkResponseItem `json:"items"`
}

// BulkResponseItem 批量操作响应项
type BulkResponseItem struct {
	Type   BulkOperationType `json:"type"`
	Index  string            `json:"index"`
	ID     string            `json:"id"`
	Status int               `json:"status"`
	Error  string            `json:"error,omitempty"`
}

// IndexConfiguration 索引配置
type IndexConfiguration struct {
	Name     string         `json:"name"`
	Settings *IndexSettings `json:"settings"`
	Mappings *IndexMapping  `json:"mappings"`
	Aliases  []string       `json:"aliases,omitempty"`
}

// IndexInfo 索引信息
type IndexInfo struct {
	Name      string                 `json:"name"`
	Status    IndexStatus            `json:"status"`
	Health    IndexHealth            `json:"health"`
	Primary   int                    `json:"primary"`
	Replicas  int                    `json:"replicas"`
	DocsCount int64                  `json:"docs_count"`
	StoreSize int64                  `json:"store_size"`
	CreatedAt time.Time              `json:"created_at"`
	UpdatedAt time.Time              `json:"updated_at"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

// IndexFilter 索引过滤器
type IndexFilter struct {
	Names  []string      `json:"names,omitempty"`
	Status []IndexStatus `json:"status,omitempty"`
	Health []IndexHealth `json:"health,omitempty"`
	Limit  int           `json:"limit,omitempty"`
	Offset int           `json:"offset,omitempty"`
}

// IndexStatistics 索引统计信息
type IndexStatistics struct {
	DocsCount     int64                `json:"docs_count"`
	DocsDeleted   int64                `json:"docs_deleted"`
	StoreSize     int64                `json:"store_size"`
	IndexingTotal int64                `json:"indexing_total"`
	IndexingTime  time.Duration        `json:"indexing_time"`
	SearchTotal   int64                `json:"search_total"`
	SearchTime    time.Duration        `json:"search_time"`
	QueryCache    *CacheStatistics     `json:"query_cache,omitempty"`
	FieldData     *FieldDataStatistics `json:"field_data,omitempty"`
	Segments      *SegmentStatistics   `json:"segments,omitempty"`
	UpdatedAt     time.Time            `json:"updated_at"`
}

// IndexMapping 索引映射
type IndexMapping struct {
	Properties map[string]*FieldMapping `json:"properties"`
	Dynamic    DynamicMapping           `json:"dynamic,omitempty"`
	Meta       map[string]interface{}   `json:"_meta,omitempty"`
}

// FieldMapping 字段映射
type FieldMapping struct {
	Type       FieldType                `json:"type"`
	Index      bool                     `json:"index,omitempty"`
	Store      bool                     `json:"store,omitempty"`
	Analyzer   string                   `json:"analyzer,omitempty"`
	Format     string                   `json:"format,omitempty"`
	Properties map[string]*FieldMapping `json:"properties,omitempty"`
	Fields     map[string]*FieldMapping `json:"fields,omitempty"`
}

// IndexSettings 索引设置
type IndexSettings struct {
	NumberOfShards   int                    `json:"number_of_shards"`
	NumberOfReplicas int                    `json:"number_of_replicas"`
	RefreshInterval  time.Duration          `json:"refresh_interval"`
	MaxResultWindow  int                    `json:"max_result_window"`
	Analysis         *AnalysisSettings      `json:"analysis,omitempty"`
	Similarity       map[string]interface{} `json:"similarity,omitempty"`
	Routing          *RoutingSettings       `json:"routing,omitempty"`
}

// AnalysisSettings 分析设置
type AnalysisSettings struct {
	Analyzers   map[string]*AnalyzerConfig   `json:"analyzers,omitempty"`
	Tokenizers  map[string]*TokenizerConfig  `json:"tokenizers,omitempty"`
	Filters     map[string]*FilterConfig     `json:"filters,omitempty"`
	CharFilters map[string]*CharFilterConfig `json:"char_filters,omitempty"`
}

// AnalyzerConfig 分析器配置
type AnalyzerConfig struct {
	Type        string   `json:"type"`
	Tokenizer   string   `json:"tokenizer,omitempty"`
	Filters     []string `json:"filters,omitempty"`
	CharFilters []string `json:"char_filters,omitempty"`
}

// TokenizerConfig 分词器配置
type TokenizerConfig struct {
	Type   string                 `json:"type"`
	Config map[string]interface{} `json:"config,omitempty"`
}

// FilterConfig 过滤器配置
type FilterConfig struct {
	Type   string                 `json:"type"`
	Config map[string]interface{} `json:"config,omitempty"`
}

// CharFilterConfig 字符过滤器配置
type CharFilterConfig struct {
	Type   string                 `json:"type"`
	Config map[string]interface{} `json:"config,omitempty"`
}

// RoutingSettings 路由设置
type RoutingSettings struct {
	Required bool `json:"required"`
}

// IndexTemplate 索引模板
type IndexTemplate struct {
	Name          string                 `json:"name"`
	IndexPatterns []string               `json:"index_patterns"`
	Settings      *IndexSettings         `json:"settings,omitempty"`
	Mappings      *IndexMapping          `json:"mappings,omitempty"`
	Aliases       map[string]interface{} `json:"aliases,omitempty"`
	Priority      int                    `json:"priority,omitempty"`
	Version       int64                  `json:"version,omitempty"`
	Meta          map[string]interface{} `json:"_meta,omitempty"`
}

// CacheStats 缓存统计信息
type CacheStats struct {
	Hits           int64     `json:"hits"`
	Misses         int64     `json:"misses"`
	HitRate        float64   `json:"hit_rate"`
	KeysCount      int64     `json:"keys_count"`
	UsedMemory     int64     `json:"used_memory"`
	MaxMemory      int64     `json:"max_memory"`
	Evictions      int64     `json:"evictions"`
	Connections    int       `json:"connections"`
	CommandsTotal  int64     `json:"commands_total"`
	CommandsPerSec float64   `json:"commands_per_sec"`
	UpdatedAt      time.Time `json:"updated_at"`
}

// CacheStatistics 缓存统计
type CacheStatistics struct {
	MemorySize int64 `json:"memory_size"`
	Evictions  int64 `json:"evictions"`
	HitCount   int64 `json:"hit_count"`
	MissCount  int64 `json:"miss_count"`
}

// FieldDataStatistics 字段数据统计
type FieldDataStatistics struct {
	MemorySize int64                     `json:"memory_size"`
	Evictions  int64                     `json:"evictions"`
	Fields     map[string]*FieldDataStat `json:"fields,omitempty"`
}

// FieldDataStat 字段数据统计
type FieldDataStat struct {
	MemorySize int64 `json:"memory_size"`
}

// SegmentStatistics 段统计
type SegmentStatistics struct {
	Count              int64 `json:"count"`
	MemoryInBytes      int64 `json:"memory_in_bytes"`
	TermsMemoryInBytes int64 `json:"terms_memory_in_bytes"`
	IndexWriterMemory  int64 `json:"index_writer_memory"`
	VersionMapMemory   int64 `json:"version_map_memory"`
}

// 枚举类型定义
type (
	SortOrder         string
	AggregationType   string
	BulkOperationType string
	IndexStatus       string
	IndexHealth       string
	DynamicMapping    string
	FieldType         string
)

// SortOrder 常量
const (
	SortOrderAsc  SortOrder = "asc"
	SortOrderDesc SortOrder = "desc"
)

// AggregationType 常量
const (
	AggregationTypeTerms       AggregationType = "terms"
	AggregationTypeHistogram   AggregationType = "histogram"
	AggregationTypeRange       AggregationType = "range"
	AggregationTypeSum         AggregationType = "sum"
	AggregationTypeAvg         AggregationType = "avg"
	AggregationTypeMax         AggregationType = "max"
	AggregationTypeMin         AggregationType = "min"
	AggregationTypeCount       AggregationType = "count"
	AggregationTypeCardinality AggregationType = "cardinality"
)

// BulkOperationType 常量
const (
	BulkOperationTypeIndex  BulkOperationType = "index"
	BulkOperationTypeCreate BulkOperationType = "create"
	BulkOperationTypeUpdate BulkOperationType = "update"
	BulkOperationTypeDelete BulkOperationType = "delete"
)

// IndexStatus 常量
const (
	IndexStatusOpen    IndexStatus = "open"
	IndexStatusClose   IndexStatus = "close"
	IndexStatusDeleted IndexStatus = "deleted"
)

// IndexHealth 常量
const (
	IndexHealthGreen  IndexHealth = "green"
	IndexHealthYellow IndexHealth = "yellow"
	IndexHealthRed    IndexHealth = "red"
)

// DynamicMapping 常量
const (
	DynamicMappingTrue   DynamicMapping = "true"
	DynamicMappingFalse  DynamicMapping = "false"
	DynamicMappingStrict DynamicMapping = "strict"
)

// FieldType 常量
const (
	FieldTypeText     FieldType = "text"
	FieldTypeKeyword  FieldType = "keyword"
	FieldTypeInteger  FieldType = "integer"
	FieldTypeLong     FieldType = "long"
	FieldTypeFloat    FieldType = "float"
	FieldTypeDouble   FieldType = "double"
	FieldTypeBoolean  FieldType = "boolean"
	FieldTypeDate     FieldType = "date"
	FieldTypeBinary   FieldType = "binary"
	FieldTypeObject   FieldType = "object"
	FieldTypeNested   FieldType = "nested"
	FieldTypeGeoPoint FieldType = "geo_point"
	FieldTypeIP       FieldType = "ip"
)

//Personal.AI order the ending
