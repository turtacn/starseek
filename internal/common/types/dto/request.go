package dto

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// BaseRequest 基础请求结构体
type BaseRequest struct {
	RequestID   string            `json:"request_id,omitempty" validate:"omitempty,uuid"`
	Timestamp   time.Time         `json:"timestamp,omitempty"`
	ClientID    string            `json:"client_id,omitempty" validate:"omitempty,min=1,max=100"`
	Version     string            `json:"version,omitempty" validate:"omitempty,semver"`
	TraceID     string            `json:"trace_id,omitempty" validate:"omitempty,min=1,max=100"`
	UserID      string            `json:"user_id,omitempty" validate:"omitempty,min=1,max=100"`
	TenantID    string            `json:"tenant_id,omitempty" validate:"omitempty,min=1,max=100"`
	Headers     map[string]string `json:"headers,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// PaginationRequest 分页请求结构体
type PaginationRequest struct {
	Page     int    `json:"page" form:"page" validate:"min=1" default:"1"`
	Size     int    `json:"size" form:"size" validate:"min=1,max=1000" default:"20"`
	Offset   int    `json:"offset,omitempty" form:"offset" validate:"min=0"`
	Limit    int    `json:"limit,omitempty" form:"limit" validate:"min=1,max=1000"`
	SortBy   string `json:"sort_by,omitempty" form:"sort_by" validate:"omitempty,min=1,max=100"`
	SortOrder string `json:"sort_order,omitempty" form:"sort_order" validate:"omitempty,oneof=asc desc ASC DESC" default:"asc"`
	Cursor   string `json:"cursor,omitempty" form:"cursor" validate:"omitempty"`
}

// FilterRequest 过滤请求结构体
type FilterRequest struct {
	Conditions []*FilterCondition `json:"conditions,omitempty" validate:"omitempty,dive"`
	Logic      FilterLogic        `json:"logic,omitempty" validate:"omitempty,oneof=and or AND OR" default:"and"`
	Groups     []*FilterGroup     `json:"groups,omitempty" validate:"omitempty,dive"`
}

// FilterCondition 过滤条件
type FilterCondition struct {
	Field    string      `json:"field" validate:"required,min=1,max=100"`
	Operator FilterOperator `json:"operator" validate:"required,oneof=eq ne gt ge lt le in nin like ilike regex exists"`
	Value    interface{} `json:"value" validate:"required"`
	Values   []interface{} `json:"values,omitempty" validate:"omitempty"`
	CaseSensitive bool   `json:"case_sensitive,omitempty" default:"false"`
}

// FilterGroup 过滤条件组
type FilterGroup struct {
	Conditions []*FilterCondition `json:"conditions" validate:"required,min=1,dive"`
	Logic      FilterLogic        `json:"logic" validate:"required,oneof=and or AND OR" default:"and"`
}

// SearchRequest 搜索请求结构体
type SearchRequest struct {
	BaseRequest

	// 搜索参数
	Query        string            `json:"query" validate:"required,min=1,max=1000"`
	QueryType    QueryType         `json:"query_type,omitempty" validate:"omitempty,oneof=match match_phrase match_all wildcard regex fuzzy term range"`
	Database     string            `json:"database" validate:"required,min=1,max=100"`
	Table        string            `json:"table" validate:"required,min=1,max=100"`
	Columns      []string          `json:"columns,omitempty" validate:"omitempty,dive,min=1,max=100"`

	// 分页和排序
	Pagination   *PaginationRequest `json:"pagination,omitempty"`

	// 过滤条件
	Filter       *FilterRequest    `json:"filter,omitempty"`

	// 搜索选项
	Options      *SearchOptions    `json:"options,omitempty"`

	// 高亮设置
	Highlight    *HighlightRequest `json:"highlight,omitempty"`

	// 聚合查询
	Aggregations []*AggregationRequest `json:"aggregations,omitempty" validate:"omitempty,dive"`

	// 建议查询
	Suggest      *SuggestRequest   `json:"suggest,omitempty"`

	// 相似度搜索
	Similarity   *SimilarityRequest `json:"similarity,omitempty"`
}

// SearchOptions 搜索选项
type SearchOptions struct {
	Analyzer         string  `json:"analyzer,omitempty" validate:"omitempty,min=1,max=100"`
	Fuzziness        string  `json:"fuzziness,omitempty" validate:"omitempty,oneof=AUTO 0 1 2"`
	PrefixLength     int     `json:"prefix_length,omitempty" validate:"omitempty,min=0,max=10"`
	MaxExpansions    int     `json:"max_expansions,omitempty" validate:"omitempty,min=1,max=1000"`
	MinShouldMatch   string  `json:"min_should_match,omitempty"`
	Boost            float64 `json:"boost,omitempty" validate:"omitempty,min=0"`
	TieBreaker       float64 `json:"tie_breaker,omitempty" validate:"omitempty,min=0,max=1"`
	ExplainScore     bool    `json:"explain_score,omitempty"`
	IncludeScore     bool    `json:"include_score,omitempty" default:"true"`
	TrackTotalHits   bool    `json:"track_total_hits,omitempty" default:"true"`
	Timeout          int     `json:"timeout,omitempty" validate:"omitempty,min=1,max=300000"` // milliseconds
	AllowPartialResults bool `json:"allow_partial_results,omitempty"`
}

// HighlightRequest 高亮请求
type HighlightRequest struct {
	Enabled      bool              `json:"enabled" default:"false"`
	Fields       []string          `json:"fields,omitempty" validate:"omitempty,dive,min=1,max=100"`
	PreTags      []string          `json:"pre_tags,omitempty"`
	PostTags     []string          `json:"post_tags,omitempty"`
	FragmentSize int               `json:"fragment_size,omitempty" validate:"omitempty,min=1,max=1000"`
	NumFragments int               `json:"num_fragments,omitempty" validate:"omitempty,min=1,max=100"`
	Order        string            `json:"order,omitempty" validate:"omitempty,oneof=score none"`
	RequireFieldMatch bool          `json:"require_field_match,omitempty"`
	BoundaryChars string            `json:"boundary_chars,omitempty"`
	BoundaryMaxScan int             `json:"boundary_max_scan,omitempty" validate:"omitempty,min=1,max=1000"`
}

// AggregationRequest 聚合请求
type AggregationRequest struct {
	Name     string           `json:"name" validate:"required,min=1,max=100"`
	Type     AggregationType  `json:"type" validate:"required,oneof=terms date_histogram histogram range stats cardinality avg sum min max"`
	Field    string           `json:"field" validate:"required,min=1,max=100"`
	Size     int              `json:"size,omitempty" validate:"omitempty,min=1,max=1000"`
	Order    *AggregationOrder `json:"order,omitempty"`
	Interval string           `json:"interval,omitempty"`
	Format   string           `json:"format,omitempty"`
	Missing  interface{}      `json:"missing,omitempty"`
	Ranges   []*AggregationRange `json:"ranges,omitempty" validate:"omitempty,dive"`
	SubAggregations []*AggregationRequest `json:"sub_aggregations,omitempty" validate:"omitempty,dive"`
}

// AggregationOrder 聚合排序
type AggregationOrder struct {
	Field string `json:"field" validate:"required"`
	Order string `json:"order" validate:"required,oneof=asc desc ASC DESC"`
}

// AggregationRange 聚合范围
type AggregationRange struct {
	Key  string      `json:"key,omitempty"`
	From interface{} `json:"from,omitempty"`
	To   interface{} `json:"to,omitempty"`
}

// SuggestRequest 建议请求
type SuggestRequest struct {
	Enabled    bool               `json:"enabled" default:"false"`
	Text       string             `json:"text" validate:"required_if=Enabled true,min=1,max=1000"`
	Suggesters []*SuggesterConfig `json:"suggesters,omitempty" validate:"omitempty,dive"`
}

// SuggesterConfig 建议器配置
type SuggesterConfig struct {
	Name     string        `json:"name" validate:"required,min=1,max=100"`
	Type     SuggesterType `json:"type" validate:"required,oneof=term phrase completion"`
	Field    string        `json:"field" validate:"required,min=1,max=100"`
	Size     int           `json:"size,omitempty" validate:"omitempty,min=1,max=100"`
	Analyzer string        `json:"analyzer,omitempty"`
}

// SimilarityRequest 相似度搜索请求
type SimilarityRequest struct {
	Enabled        bool    `json:"enabled" default:"false"`
	Document       string  `json:"document,omitempty" validate:"required_if=Enabled true"`
	Fields         []string `json:"fields,omitempty" validate:"omitempty,dive,min=1,max=100"`
	MinTermFreq    int     `json:"min_term_freq,omitempty" validate:"omitempty,min=1"`
	MaxQueryTerms  int     `json:"max_query_terms,omitempty" validate:"omitempty,min=1,max=1000"`
	MinDocFreq     int     `json:"min_doc_freq,omitempty" validate:"omitempty,min=1"`
	MinWordLength  int     `json:"min_word_length,omitempty" validate:"omitempty,min=1"`
	MaxWordLength  int     `json:"max_word_length,omitempty" validate:"omitempty,min=1"`
	BoostTerms     float64 `json:"boost_terms,omitempty" validate:"omitempty,min=0"`
	Include        bool    `json:"include,omitempty"`
	MinScore       float64 `json:"min_score,omitempty" validate:"omitempty,min=0"`
}

// IndexCreateRequest 索引创建请求
type IndexCreateRequest struct {
	BaseRequest

	Name         string                 `json:"name" validate:"required,min=1,max=100,alphanum"`
	Database     string                 `json:"database" validate:"required,min=1,max=100"`
	Table        string                 `json:"table" validate:"required,min=1,max=100"`
	Type         string                 `json:"type" validate:"required,oneof=inverted bitmap bloom_filter ngram fulltext"`
	Engine       string                 `json:"engine" validate:"required,oneof=starrocks clickhouse doris"`
	Columns      []*IndexColumnRequest  `json:"columns" validate:"required,min=1,dive"`
	Settings     map[string]interface{} `json:"settings,omitempty"`
	Mappings     map[string]interface{} `json:"mappings,omitempty"`
	Aliases      []string               `json:"aliases,omitempty" validate:"omitempty,dive,min=1,max=100"`
	Description  string                 `json:"description,omitempty" validate:"omitempty,max=500"`
	Tags         map[string]string      `json:"tags,omitempty"`

	// 高级配置
	Shards       int                    `json:"shards,omitempty" validate:"omitempty,min=1,max=100"`
	Replicas     int                    `json:"replicas,omitempty" validate:"omitempty,min=0,max=10"`
	Partitions   []*PartitionRequest    `json:"partitions,omitempty" validate:"omitempty,dive"`
	Tokenizer    *TokenizerRequest      `json:"tokenizer,omitempty"`

	// 构建选项
	BuildAsync   bool                   `json:"build_async,omitempty" default:"false"`
	BuildOptions *BuildOptionsRequest   `json:"build_options,omitempty"`
}

// IndexColumnRequest 索引列请求
type IndexColumnRequest struct {
	Name         string                 `json:"name" validate:"required,min=1,max=100"`
	Type         string                 `json:"type" validate:"required,oneof=string text keyword integer long float double boolean date datetime timestamp"`
	Analyzer     string                 `json:"analyzer,omitempty" validate:"omitempty,min=1,max=100"`
	Boost        float64                `json:"boost,omitempty" validate:"omitempty,min=0"`
	Store        bool                   `json:"store,omitempty" default:"true"`
	Index        bool                   `json:"index,omitempty" default:"true"`
	DocValues    bool                   `json:"doc_values,omitempty" default:"true"`
	Nullable     bool                   `json:"nullable,omitempty" default:"true"`
	Properties   map[string]interface{} `json:"properties,omitempty"`
}

// PartitionRequest 分区请求
type PartitionRequest struct {
	Type         string      `json:"type" validate:"required,oneof=range list hash"`
	Columns      []string    `json:"columns" validate:"required,min=1,dive,min=1,max=100"`
	Expression   string      `json:"expression,omitempty"`
	Buckets      int         `json:"buckets,omitempty" validate:"omitempty,min=1,max=1000"`
	Values       []interface{} `json:"values,omitempty"`
	Ranges       []*RangeRequest `json:"ranges,omitempty" validate:"omitempty,dive"`
}

// RangeRequest 范围请求
type RangeRequest struct {
	Name  string      `json:"name" validate:"required,min=1,max=100"`
	Start interface{} `json:"start"`
	End   interface{} `json:"end"`
}

// TokenizerRequest 分词器请求
type TokenizerRequest struct {
	Type         string                 `json:"type" validate:"required,oneof=standard ik jieba custom"`
	Language     string                 `json:"language,omitempty" validate:"omitempty,oneof=chinese english japanese korean"`
	MaxTokenLength int                  `json:"max_token_length,omitempty" validate:"omitempty,min=1,max=1000"`
	Filters      []string               `json:"filters,omitempty" validate:"omitempty,dive,min=1,max=100"`
	StopWords    []string               `json:"stop_words,omitempty"`
	Synonyms     []string               `json:"synonyms,omitempty"`
	UserDict     string                 `json:"user_dict,omitempty"`
	Config       map[string]interface{} `json:"config,omitempty"`
}

// BuildOptionsRequest 构建选项请求
type BuildOptionsRequest struct {
	Priority     int                    `json:"priority,omitempty" validate:"omitempty,min=1,max=10"`
	BatchSize    int                    `json:"batch_size,omitempty" validate:"omitempty,min=1,max=100000"`
	MaxMemory    int64                  `json:"max_memory,omitempty" validate:"omitempty,min=1"`
	Parallelism  int                    `json:"parallelism,omitempty" validate:"omitempty,min=1,max=100"`
	Timeout      int                    `json:"timeout,omitempty" validate:"omitempty,min=1,max=86400"` // seconds
	CheckData    bool                   `json:"check_data,omitempty" default:"true"`
	SkipErrors   bool                   `json:"skip_errors,omitempty" default:"false"`
	Compression  string                 `json:"compression,omitempty" validate:"omitempty,oneof=none gzip lz4 snappy zstd"`
	Config       map[string]interface{} `json:"config,omitempty"`
}

// IndexUpdateRequest 索引更新请求
type IndexUpdateRequest struct {
	BaseRequest

	IndexID      string                 `json:"index_id" validate:"required,uuid"`
	Name         string                 `json:"name,omitempty" validate:"omitempty,min=1,max=100,alphanum"`
	Settings     map[string]interface{} `json:"settings,omitempty"`
	Mappings     map[string]interface{} `json:"mappings,omitempty"`
	Aliases      []string               `json:"aliases,omitempty" validate:"omitempty,dive,min=1,max=100"`
	Description  string                 `json:"description,omitempty" validate:"omitempty,max=500"`
	Tags         map[string]string      `json:"tags,omitempty"`
	Enabled      *bool                  `json:"enabled,omitempty"`

	// 更新选项
	UpdateOptions *UpdateOptionsRequest `json:"update_options,omitempty"`
}

// UpdateOptionsRequest 更新选项请求
type UpdateOptionsRequest struct {
	Force        bool `json:"force,omitempty" default:"false"`
	Async        bool `json:"async,omitempty" default:"false"`
	WaitForCompletion bool `json:"wait_for_completion,omitempty" default:"true"`
	Timeout      int  `json:"timeout,omitempty" validate:"omitempty,min=1,max=3600"` // seconds
}

// IndexDeleteRequest 索引删除请求
type IndexDeleteRequest struct {
	BaseRequest

	IndexID      string   `json:"index_id" validate:"required,uuid"`
	Force        bool     `json:"force,omitempty" default:"false"`
	Async        bool     `json:"async,omitempty" default:"false"`
	IgnoreUnavailable bool `json:"ignore_unavailable,omitempty" default:"false"`
	AllowNoIndices bool   `json:"allow_no_indices,omitempty" default:"true"`
}

// IndexListRequest 索引列表请求
type IndexListRequest struct {
	BaseRequest

	Database     string             `json:"database,omitempty" form:"database" validate:"omitempty,min=1,max=100"`
	Table        string             `json:"table,omitempty" form:"table" validate:"omitempty,min=1,max=100"`
	Type         string             `json:"type,omitempty" form:"type" validate:"omitempty,oneof=inverted bitmap bloom_filter ngram fulltext"`
	Engine       string             `json:"engine,omitempty" form:"engine" validate:"omitempty,oneof=starrocks clickhouse doris"`
	Status       string             `json:"status,omitempty" form:"status" validate:"omitempty,oneof=creating active inactive building rebuilding deleting error"`
	Health       string             `json:"health,omitempty" form:"health" validate:"omitempty,oneof=green yellow red gray"`

	// 分页和过滤
	Pagination   *PaginationRequest `json:"pagination,omitempty"`
	Filter       *FilterRequest     `json:"filter,omitempty"`

	// 查询选项
	IncludeSettings bool            `json:"include_settings,omitempty" form:"include_settings" default:"false"`
	IncludeMappings bool            `json:"include_mappings,omitempty" form:"include_mappings" default:"false"`
	IncludeStats    bool            `json:"include_stats,omitempty" form:"include_stats" default:"false"`
	IncludeHealth   bool            `json:"include_health,omitempty" form:"include_health" default:"true"`
}

// IndexStatusRequest 索引状态请求
type IndexStatusRequest struct {
	BaseRequest

	IndexID      string   `json:"index_id" validate:"required,uuid"`
	IncludeStats bool     `json:"include_stats,omitempty" default:"true"`
	IncludeShards bool    `json:"include_shards,omitempty" default:"false"`
	IncludeSegments bool  `json:"include_segments,omitempty" default:"false"`
}

// IndexRebuildRequest 索引重建请求
type IndexRebuildRequest struct {
	BaseRequest

	IndexID      string                 `json:"index_id" validate:"required,uuid"`
	Force        bool                   `json:"force,omitempty" default:"false"`
	Async        bool                   `json:"async,omitempty" default:"true"`
	OnlyIfEmpty  bool                   `json:"only_if_empty,omitempty" default:"false"`

	// 重建选项
	Options      *BuildOptionsRequest   `json:"options,omitempty"`
}

// ConfigGetRequest 配置获取请求
type ConfigGetRequest struct {
	BaseRequest

	Key          string   `json:"key,omitempty" form:"key" validate:"omitempty,min=1,max=200"`
	Namespace    string   `json:"namespace,omitempty" form:"namespace" validate:"omitempty,min=1,max=100"`
	Version      string   `json:"version,omitempty" form:"version" validate:"omitempty,semver"`
	Environment  string   `json:"environment,omitempty" form:"environment" validate:"omitempty,oneof=development staging production"`
	IncludeDefault bool   `json:"include_default,omitempty" form:"include_default" default:"true"`
	IncludeHistory bool   `json:"include_history,omitempty" form:"include_history" default:"false"`
}

// ConfigUpdateRequest 配置更新请求
type ConfigUpdateRequest struct {
	BaseRequest

	Key          string      `json:"key" validate:"required,min=1,max=200"`
	Value        interface{} `json:"value" validate:"required"`
	Namespace    string      `json:"namespace,omitempty" validate:"omitempty,min=1,max=100"`
	Environment  string      `json:"environment,omitempty" validate:"omitempty,oneof=development staging production"`
	Description  string      `json:"description,omitempty" validate:"omitempty,max=500"`
	Tags         map[string]string `json:"tags,omitempty"`

	// 更新选项
	Validation   bool        `json:"validation,omitempty" default:"true"`
	Backup       bool        `json:"backup,omitempty" default:"true"`
	NotifyChange bool        `json:"notify_change,omitempty" default:"true"`
}

// BulkRequest 批量请求结构体
type BulkRequest struct {
	BaseRequest

	Operations   []*BulkOperation `json:"operations" validate:"required,min=1,max=1000,dive"`

	// 批量选项
	FailOnError  bool `json:"fail_on_error,omitempty" default:"false"`
	Parallel     bool `json:"parallel,omitempty" default:"false"`
	BatchSize    int  `json:"batch_size,omitempty" validate:"omitempty,min=1,max=1000"`
	Timeout      int  `json:"timeout,omitempty" validate:"omitempty,min=1,max=3600"` // seconds
}

// BulkOperation 批量操作
type BulkOperation struct {
	Type         BulkOperationType `json:"type" validate:"required,oneof=create update delete index"`
	IndexID      string            `json:"index_id,omitempty" validate:"omitempty,uuid"`
	ID           string            `json:"id,omitempty"`
	Document     interface{}       `json:"document,omitempty"`
	Upsert       bool              `json:"upsert,omitempty" default:"false"`
	RetryOnConflict int            `json:"retry_on_conflict,omitempty" validate:"omitempty,min=0,max=10"`
}

// 枚举类型定义
type (
	FilterLogic       string
	FilterOperator    string
	QueryType         string
	AggregationType   string
	SuggesterType     string
	BulkOperationType string
)

// 枚举常量定义
const (
	// FilterLogic 常量
	FilterLogicAnd FilterLogic = "and"
	FilterLogicOr  FilterLogic = "or"

	// FilterOperator 常量
	FilterOperatorEq     FilterOperator = "eq"
	FilterOperatorNe     FilterOperator = "ne"
	FilterOperatorGt     FilterOperator = "gt"
	FilterOperatorGe     FilterOperator = "ge"
	FilterOperatorLt     FilterOperator = "lt"
	FilterOperatorLe     FilterOperator = "le"
	FilterOperatorIn     FilterOperator = "in"
	FilterOperatorNin    FilterOperator = "nin"
	FilterOperatorLike   FilterOperator = "like"
	FilterOperatorIlike  FilterOperator = "ilike"
	FilterOperatorRegex  FilterOperator = "regex"
	FilterOperatorExists FilterOperator = "exists"

	// QueryType 常量
	QueryTypeMatch       QueryType = "match"
	QueryTypeMatchPhrase QueryType = "match_phrase"
	QueryTypeMatchAll    QueryType = "match_all"
	QueryTypeWildcard    QueryType = "wildcard"
	QueryTypeRegex       QueryType = "regex"
	QueryTypeFuzzy       QueryType = "fuzzy"
	QueryTypeTerm        QueryType = "term"
	QueryTypeRange       QueryType = "range"

	// AggregationType 常量
	AggregationTypeTerms         AggregationType = "terms"
	AggregationTypeDateHistogram AggregationType = "date_histogram"
	AggregationTypeHistogram     AggregationType = "histogram"
	AggregationTypeRange         AggregationType = "range"
	AggregationTypeStats         AggregationType = "stats"
	AggregationTypeCardinality   AggregationType = "cardinality"
	AggregationTypeAvg           AggregationType = "avg"
	AggregationTypeSum           AggregationType = "sum"
	AggregationTypeMin           AggregationType = "min"
	AggregationTypeMax           AggregationType = "max"

	// SuggesterType 常量
	SuggesterTypeTerm       SuggesterType = "term"
	SuggesterTypePhrase     SuggesterType = "phrase"
	SuggesterTypeCompletion SuggesterType = "completion"

	// BulkOperationType 常量
	BulkOperationTypeCreate BulkOperationType = "create"
	BulkOperationTypeUpdate BulkOperationType = "update"
	BulkOperationTypeDelete BulkOperationType = "delete"
	BulkOperationTypeIndex  BulkOperationType = "index"
)

// 验证方法

// Validate 验证BaseRequest
func (r *BaseRequest) Validate() error {
	if r.RequestID != "" && !isValidUUID(r.RequestID) {
		return errors.New("invalid request ID format")
	}

	if r.Version != "" && !isValidSemver(r.Version) {
		return errors.New("invalid version format")
	}

	return nil
}

// Validate 验证PaginationRequest
func (r *PaginationRequest) Validate() error {
	if r.Page < 1 {
		return errors.New("page must be greater than 0")
	}

	if r.Size < 1 || r.Size > 1000 {
		return errors.New("size must be between 1 and 1000")
	}

	if r.Offset < 0 {
		return errors.New("offset must be non-negative")
	}

	if r.Limit != 0 && (r.Limit < 1 || r.Limit > 1000) {
		return errors.New("limit must be between 1 and 1000")
	}

	if r.SortOrder != "" {
		validOrders := []string{"asc", "desc", "ASC", "DESC"}
		if !contains(validOrders, r.SortOrder) {
			return fmt.Errorf("invalid sort order: %s", r.SortOrder)
		}
	}

	return nil
}

// Validate 验证FilterCondition
func (r *FilterCondition) Validate() error {
	if r.Field == "" {
		return errors.New("filter field is required")
	}

	if r.Operator == "" {
		return errors.New("filter operator is required")
	}

	// 验证操作符特定的值要求
	switch r.Operator {
	case FilterOperatorIn, FilterOperatorNin:
		if len(r.Values) == 0 {
			return errors.New("values are required for in/nin operators")
		}
	case FilterOperatorExists:
		// exists操作符不需要值
	default:
		if r.Value == nil {
			return errors.New("value is required for this operator")
		}
	}

	return nil
}

// Validate 验证SearchRequest
func (r *SearchRequest) Validate() error {
	if err := r.BaseRequest.Validate(); err != nil {
		return fmt.Errorf("base request validation failed: %w", err)
	}

	if r.Query == "" {
		return errors.New("search query is required")
	}

	if r.Database == "" {
		return errors.New("database is required")
	}

	if r.Table == "" {
		return errors.New("table is required")
	}

	if r.Pagination != nil {
		if err := r.Pagination.Validate(); err != nil {
			return fmt.Errorf("pagination validation failed: %w", err)
		}
	}

	if r.Filter != nil {
		if err := r.Filter.Validate(); err != nil {
			return fmt.Errorf("filter validation failed: %w", err)
		}
	}

	return nil
}

// Validate 验证FilterRequest
func (r *FilterRequest) Validate() error {
	for i, condition := range r.Conditions {
		if err := condition.Validate(); err != nil {
			return fmt.Errorf("condition %d validation failed: %w", i, err)
		}
	}

	for i, group := range r.Groups {
		if err := group.Validate(); err != nil {
			return fmt.Errorf("group %d validation failed: %w", i, err)
		}
	}

	return nil
}

// Validate 验证FilterGroup
func (r *FilterGroup) Validate() error {
	if len(r.Conditions) == 0 {
		return errors.New("filter group must have at least one condition")
	}

	for i, condition := range r.Conditions {
		if err := condition.Validate(); err != nil {
			return fmt.Errorf("condition %d validation failed: %w", i, err)
		}
	}

	return nil
}

// Validate 验证IndexCreateRequest
func (r *IndexCreateRequest) Validate() error {
	if err := r.BaseRequest.Validate(); err != nil {
		return fmt.Errorf("base request validation failed: %w", err)
	}

	if r.Name == "" {
		return errors.New("index name is required")
	}

	if !isValidIndexName(r.Name) {
		return errors.New("invalid index name format")
	}

	if r.Database == "" {
		return errors.New("database is required")
	}

	if r.Table == "" {
		return errors.New("table is required")
	}

	if r.Type == "" {
		return errors.New("index type is required")
	}

	if r.Engine == "" {
		return errors.New("database engine is required")
	}

	if len(r.Columns) == 0 {
		return errors.New("at least one column is required")
	}

	for i, column := range r.Columns {
		if err := column.Validate(); err != nil {
			return fmt.Errorf("column %d validation failed: %w", i, err)
		}
	}

	return nil
}

// Validate 验证IndexColumnRequest
func (r *IndexColumnRequest) Validate() error {
	if r.Name == "" {
		return errors.New("column name is required")
	}

	if r.Type == "" {
		return errors.New("column type is required")
	}

	validTypes := []string{"string", "text", "keyword", "integer", "long", "float", "double", "boolean", "date", "datetime", "timestamp"}
	if !contains(validTypes, r.Type) {
		return fmt.Errorf("invalid column type: %s", r.Type)
	}

	if r.Boost < 0 {
		return errors.New("boost must be non-negative")
	}

	return nil
}

// 默认值设置方法

// SetDefaults 设置PaginationRequest默认值
func (r *PaginationRequest) SetDefaults() {
	if r.Page == 0 {
		r.Page = 1
	}
	if r.Size == 0 {
		r.Size = 20
	}
	if r.SortOrder == "" {
		r.SortOrder = "asc"
	}
}

// SetDefaults 设置FilterRequest默认值
func (r *FilterRequest) SetDefaults() {
	if r.Logic == "" {
		r.Logic = FilterLogicAnd
	}
	for _, group := range r.Groups {
		group.SetDefaults()
	}
}

// SetDefaults 设置FilterGroup默认值
func (r *FilterGroup) SetDefaults() {
	if r.Logic == "" {
		r.Logic = FilterLogicAnd
	}
	for _, condition := range r.Conditions {
		condition.SetDefaults()
	}
}

// SetDefaults 设置FilterCondition默认值
func (r *FilterCondition) SetDefaults() {
	// CaseSensitive默认为false，已在结构体标签中设置
}

// SetDefaults 设置SearchOptions默认值
func (r *SearchOptions) SetDefaults() {
	if r.IncludeScore == false {
		r.IncludeScore = true
	}
	if r.TrackTotalHits == false {
		r.TrackTotalHits = true
	}
}

// 类型转换方法

// ToMap 将请求转换为map
func (r *SearchRequest) ToMap() map[string]interface{} {
	result := make(map[string]interface{})

	// 基础字段
	result["query"] = r.Query
	result["query_type"] = r.QueryType
	result["database"] = r.Database
	result["table"] = r.Table

	if len(r.Columns) > 0 {
		result["columns"] = r.Columns
	}

	// 分页
	if r.Pagination != nil {
		result["pagination"] = r.Pagination.ToMap()
	}

	// 过滤
	if r.Filter != nil {
		result["filter"] = r.Filter.ToMap()
	}

	// 选项
	if r.Options != nil {
		result["options"] = r.Options
	}

	return result
}

// ToMap 将分页请求转换为map
func (r *PaginationRequest) ToMap() map[string]interface{} {
	result := make(map[string]interface{})

	result["page"] = r.Page
	result["size"] = r.Size

	if r.Offset > 0 {
		result["offset"] = r.Offset
	}
	if r.Limit > 0 {
		result["limit"] = r.Limit
	}
	if r.SortBy != "" {
		result["sort_by"] = r.SortBy
	}
	if r.SortOrder != "" {
		result["sort_order"] = r.SortOrder
	}
	if r.Cursor != "" {
		result["cursor"] = r.Cursor
	}

	return result
}

// ToMap 将过滤请求转换为map
func (r *FilterRequest) ToMap() map[string]interface{} {
	result := make(map[string]interface{})

	if len(r.Conditions) > 0 {
		conditions := make([]map[string]interface{}, len(r.Conditions))
		for i, condition := range r.Conditions {
			conditions[i] = condition.ToMap()
		}
		result["conditions"] = conditions
	}

	if r.Logic != "" {
		result["logic"] = r.Logic
	}

	if len(r.Groups) > 0 {
		groups := make([]map[string]interface{}, len(r.Groups))
		for i, group := range r.Groups {
			groups[i] = group.ToMap()
		}
		result["groups"] = groups
	}

	return result
}

// ToMap 将过滤条件转换为map
func (r *FilterCondition) ToMap() map[string]interface{} {
	result := make(map[string]interface{})

	result["field"] = r.Field
	result["operator"] = r.Operator

	if r.Value != nil {
		result["value"] = r.Value
	}
	if len(r.Values) > 0 {
		result["values"] = r.Values
	}
	if r.CaseSensitive {
		result["case_sensitive"] = r.CaseSensitive
	}

	return result
}

// ToMap 将过滤组转换为map
func (r *FilterGroup) ToMap() map[string]interface{} {
	result := make(map[string]interface{})

	conditions := make([]map[string]interface{}, len(r.Conditions))
	for i, condition := range r.Conditions {
		conditions[i] = condition.ToMap()
	}
	result["conditions"] = conditions
	result["logic"] = r.Logic

	return result
}

// FromQueryParams 从查询参数解析PaginationRequest
func (r *PaginationRequest) FromQueryParams(params url.Values) error {
	if page := params.Get("page"); page != "" {
		if p, err := strconv.Atoi(page); err == nil && p > 0 {
			r.Page = p
		}
	}

	if size := params.Get("size"); size != "" {
		if s, err := strconv.Atoi(size); err == nil && s > 0 && s <= 1000 {
			r.Size = s
		}
	}

	if offset := params.Get("offset"); offset != "" {
		if o, err := strconv.Atoi(offset); err == nil && o >= 0 {
			r.Offset = o
		}
	}

	if limit := params.Get("limit"); limit != "" {
		if l, err := strconv.Atoi(limit); err == nil && l > 0 && l <= 1000 {
			r.Limit = l
		}
	}

	if sortBy := params.Get("sort_by"); sortBy != "" {
		r.SortBy = sortBy
	}

	if sortOrder := params.Get("sort_order"); sortOrder != "" {
		validOrders := []string{"asc", "desc", "ASC", "DESC"}
		if contains(validOrders, sortOrder) {
			r.SortOrder = sortOrder
		}
	}

	if cursor := params.Get("cursor"); cursor != "" {
		r.Cursor = cursor
	}

	// 设置默认值
	r.SetDefaults()

	return r.Validate()
}

// 工具函数

// contains 检查字符串切片是否包含指定值
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// isValidUUID 验证UUID格式
func isValidUUID(u string) bool {
	r := regexp.MustCompile("^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$")
	return r.MatchString(strings.ToLower(u))
}

// isValidSemver 验证语义版本格式
func isValidSemver(v string) bool {
	r := regexp.MustCompile(`^(0|[1-9]\d*)\.(0|[1-9]\d*)\.(0|[1-9]\d*)(?:-((?:0|[1-9]\d*|\d*[a-zA-Z-][0-9a-zA-Z-]*)(?:\.(?:0|[1-9]\d*|\d*[a-zA-Z-][0-9a-zA-Z-]*))*))?(?:\+([0-9a-zA-Z-]+(?:\.[0-9a-zA-Z-]+)*))?$`)
	return r.MatchString(v)
}

// isValidIndexName 验证索引名称格式
func isValidIndexName(name string) bool {
	// 索引名称只能包含字母、数字和下划线，不能以数字开头
	r := regexp.MustCompile("^[a-zA-Z_][a-zA-Z0-9_]*$")
	return r.MatchString(name) && len(name) <= 100
}

// ToJSON 转换为JSON字符串
func (r *SearchRequest) ToJSON() (string, error) {
	data, err := json.Marshal(r)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// String 返回字符串表示
func (r *SearchRequest) String() string {
	return fmt.Sprintf("SearchRequest{Query:%s, Database:%s, Table:%s}", r.Query, r.Database, r.Table)
}

// String 返回字符串表示
func (r *IndexCreateRequest) String() string {
	return fmt.Sprintf("IndexCreateRequest{Name:%s, Database:%s, Table:%s, Type:%s}", r.Name, r.Database, r.Table, r.Type)
}

//Personal.AI order the ending
