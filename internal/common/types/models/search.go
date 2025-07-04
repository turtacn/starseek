package models

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"
)

// SearchRequest 搜索请求结构体
type SearchRequest struct {
	// 基础查询参数
	Query       string            `json:"query" validate:"required,min=1,max=1000"`
	QueryType   QueryType         `json:"query_type,omitempty" validate:"oneof=match phrase prefix wildcard fuzzy"`
	Fields      []string          `json:"fields,omitempty" validate:"dive,required"`
	Analyzer    string            `json:"analyzer,omitempty"`

	// 索引参数
	Index       string            `json:"index" validate:"required,min=1,max=100"`
	Indices     []string          `json:"indices,omitempty" validate:"dive,required"`

	// 分页参数
	From        int               `json:"from" validate:"min=0,max=10000"`
	Size        int               `json:"size" validate:"min=1,max=1000"`

	// 过滤条件
	Filters     []*QueryCondition `json:"filters,omitempty"`
	PostFilters []*QueryCondition `json:"post_filters,omitempty"`

	// 排序选项
	Sort        []*SortOption     `json:"sort,omitempty"`

	// 高亮配置
	Highlight   *HighlightConfig  `json:"highlight,omitempty"`

	// 聚合查询
	Aggregations map[string]*Aggregation `json:"aggregations,omitempty"`

	// 搜索选项
	Options     *SearchOptions    `json:"options,omitempty"`

	// 元数据
	TraceID     string            `json:"trace_id,omitempty"`
	UserID      string            `json:"user_id,omitempty"`
	SessionID   string            `json:"session_id,omitempty"`
	Timestamp   time.Time         `json:"timestamp,omitempty"`

	// 缓存控制
	UseCache    bool              `json:"use_cache,omitempty"`
	CacheTTL    time.Duration     `json:"cache_ttl,omitempty"`

	// 调试信息
	Debug       bool              `json:"debug,omitempty"`
	Explain     bool              `json:"explain,omitempty"`
}

// SearchResponse 搜索响应结构体
type SearchResponse struct {
	// 搜索结果
	Hits        *SearchHits       `json:"hits"`

	// 聚合结果
	Aggregations map[string]*AggregationResult `json:"aggregations,omitempty"`

	// 建议结果
	Suggestions map[string]*SuggestionResult `json:"suggestions,omitempty"`

	// 元数据
	Took        time.Duration     `json:"took"`
	TimedOut    bool              `json:"timed_out"`
	Shards      *ShardInfo        `json:"_shards,omitempty"`

	// 分页信息
	ScrollID    string            `json:"scroll_id,omitempty"`

	// 调试信息
	DebugInfo   *DebugInfo        `json:"debug_info,omitempty"`

	// 错误信息
	Errors      []string          `json:"errors,omitempty"`
	Warnings    []string          `json:"warnings,omitempty"`

	// 追踪信息
	TraceID     string            `json:"trace_id,omitempty"`
	RequestID   string            `json:"request_id,omitempty"`
}

// SearchHits 搜索结果集合
type SearchHits struct {
	Total       *TotalHits        `json:"total"`
	MaxScore    *float64          `json:"max_score"`
	Hits        []*SearchHit      `json:"hits"`
}

// SearchHit 单个搜索结果
type SearchHit struct {
	// 基础信息
	Index       string            `json:"_index"`
	Type        string            `json:"_type,omitempty"`
	ID          string            `json:"_id"`
	Score       *float64          `json:"_score"`

	// 源数据
	Source      map[string]interface{} `json:"_source,omitempty"`
	Fields      map[string]interface{} `json:"fields,omitempty"`

	// 高亮信息
	Highlight   map[string][]string    `json:"highlight,omitempty"`

	// 排序值
	Sort        []interface{}     `json:"sort,omitempty"`

	// 匹配信息
	MatchedQueries []string       `json:"matched_queries,omitempty"`

	// 解释信息
	Explanation *Explanation      `json:"_explanation,omitempty"`

	// 版本信息
	Version     *int64            `json:"_version,omitempty"`
	SeqNo       *int64            `json:"_seq_no,omitempty"`
	PrimaryTerm *int64            `json:"_primary_term,omitempty"`

	// 路由信息
	Routing     string            `json:"_routing,omitempty"`

	// 嵌套信息
	InnerHits   map[string]*SearchHits `json:"inner_hits,omitempty"`
}

// QueryCondition 查询条件结构体
type QueryCondition struct {
	// 条件类型
	Type        ConditionType     `json:"type" validate:"required,oneof=term terms range match bool wildcard prefix fuzzy"`

	// 字段名
	Field       string            `json:"field" validate:"required_unless=Type bool"`

	// 查询值
	Value       interface{}       `json:"value,omitempty"`
	Values      []interface{}     `json:"values,omitempty"`

	// 范围查询参数
	From        interface{}       `json:"from,omitempty"`
	To          interface{}       `json:"to,omitempty"`
	IncludeLower bool             `json:"include_lower,omitempty"`
	IncludeUpper bool             `json:"include_upper,omitempty"`

	// 模糊查询参数
	Fuzziness   string            `json:"fuzziness,omitempty"`
	PrefixLength int              `json:"prefix_length,omitempty"`
	MaxExpansions int             `json:"max_expansions,omitempty"`

	// 匹配查询参数
	Operator    string            `json:"operator,omitempty"`
	MinimumShouldMatch string     `json:"minimum_should_match,omitempty"`
	Analyzer    string            `json:"analyzer,omitempty"`

	// 布尔查询参数
	Must        []*QueryCondition `json:"must,omitempty"`
	MustNot     []*QueryCondition `json:"must_not,omitempty"`
	Should      []*QueryCondition `json:"should,omitempty"`
	Filter      []*QueryCondition `json:"filter,omitempty"`

	// 权重
	Boost       *float64          `json:"boost,omitempty"`

	// 查询名称
	QueryName   string            `json:"_name,omitempty"`
}

// SortOption 排序选项结构体
type SortOption struct {
	// 字段名
	Field       string            `json:"field" validate:"required"`

	// 排序方向
	Order       SortOrder         `json:"order" validate:"oneof=asc desc"`

	// 排序模式
	Mode        SortMode          `json:"mode,omitempty" validate:"oneof=min max sum avg median"`

	// 缺失值处理
	Missing     interface{}       `json:"missing,omitempty"`

	// 嵌套排序
	NestedPath  string            `json:"nested_path,omitempty"`
	NestedFilter *QueryCondition  `json:"nested_filter,omitempty"`

	// 脚本排序
	Script      *ScriptSort       `json:"script,omitempty"`

	// 地理距离排序
	GeoDistance *GeoDistanceSort  `json:"geo_distance,omitempty"`
}

// HighlightConfig 高亮配置
type HighlightConfig struct {
	// 全局设置
	PreTags     []string          `json:"pre_tags,omitempty"`
	PostTags    []string          `json:"post_tags,omitempty"`
	FragmentSize int              `json:"fragment_size,omitempty"`
	NumberOfFragments int          `json:"number_of_fragments,omitempty"`

	// 字段级设置
	Fields      map[string]*HighlightField `json:"fields" validate:"required,min=1"`

	// 高亮器类型
	Type        string            `json:"type,omitempty"`
	Fragmenter  string            `json:"fragmenter,omitempty"`

	// 边界设置
	BoundaryChars string           `json:"boundary_chars,omitempty"`
	BoundaryMaxScan int            `json:"boundary_max_scan,omitempty"`

	// 编码器
	Encoder     string            `json:"encoder,omitempty"`

	// 标签模式
	TagsSchema  string            `json:"tags_schema,omitempty"`

	// 是否要求字段匹配
	RequireFieldMatch bool         `json:"require_field_match,omitempty"`
}

// HighlightField 高亮字段配置
type HighlightField struct {
	Type              string            `json:"type,omitempty"`
	FragmentSize      int               `json:"fragment_size,omitempty"`
	NumberOfFragments int               `json:"number_of_fragments,omitempty"`
	FragmentOffset    int               `json:"fragment_offset,omitempty"`
	PreTags           []string          `json:"pre_tags,omitempty"`
	PostTags          []string          `json:"post_tags,omitempty"`
	MatchedFields     []string          `json:"matched_fields,omitempty"`
	PhraseLimit       int               `json:"phrase_limit,omitempty"`
	RequireFieldMatch bool              `json:"require_field_match,omitempty"`
	BoundaryChars     string            `json:"boundary_chars,omitempty"`
	BoundaryMaxScan   int               `json:"boundary_max_scan,omitempty"`
	MaxAnalyzedOffset int               `json:"max_analyzed_offset,omitempty"`
}

// Aggregation 聚合查询配置
type Aggregation struct {
	// 聚合类型
	Type        AggregationType   `json:"type" validate:"required"`

	// 字段名
	Field       string            `json:"field,omitempty"`
	Script      *Script           `json:"script,omitempty"`

	// 通用参数
	Size        int               `json:"size,omitempty"`
	ShardSize   int               `json:"shard_size,omitempty"`

	// 条件聚合参数
	Ranges      []*RangeParam     `json:"ranges,omitempty"`
	Keyed       bool              `json:"keyed,omitempty"`

	// 直方图参数
	Interval    interface{}       `json:"interval,omitempty"`
	MinDocCount int               `json:"min_doc_count,omitempty"`

	// 日期直方图参数
	CalendarInterval string         `json:"calendar_interval,omitempty"`
	FixedInterval    string         `json:"fixed_interval,omitempty"`
	TimeZone         string         `json:"time_zone,omitempty"`

	// 地理网格参数
	Precision    int               `json:"precision,omitempty"`

	// 嵌套聚合
	Aggregations map[string]*Aggregation `json:"aggregations,omitempty"`

	// 过滤器
	Filter      *QueryCondition   `json:"filter,omitempty"`
	Filters     map[string]*QueryCondition `json:"filters,omitempty"`

	// 排序
	Order       map[string]string `json:"order,omitempty"`

	// 包含/排除模式
	Include     interface{}       `json:"include,omitempty"`
	Exclude     interface{}       `json:"exclude,omitempty"`
}

// AggregationResult 聚合查询结果
type AggregationResult struct {
	// 度量聚合结果
	Value       *float64          `json:"value,omitempty"`
	ValueAsString string          `json:"value_as_string,omitempty"`

	// 桶聚合结果
	Buckets     interface{}       `json:"buckets,omitempty"` // []Bucket or map[string]Bucket
	DocCountError int64           `json:"doc_count_error_upper_bound,omitempty"`
	SumOtherDocCount int64        `json:"sum_other_doc_count,omitempty"`

	// 统计聚合结果
	Count       int64             `json:"count,omitempty"`
	Min         *float64          `json:"min,omitempty"`
	Max         *float64          `json:"max,omitempty"`
	Avg         *float64          `json:"avg,omitempty"`
	Sum         *float64          `json:"sum,omitempty"`

	// 百分位聚合结果
	Values      map[string]float64 `json:"values,omitempty"`

	// 嵌套聚合结果
	Aggregations map[string]*AggregationResult `json:"aggregations,omitempty"`

	// 元数据
	Meta        map[string]interface{} `json:"meta,omitempty"`
}

// Bucket 聚合桶
type Bucket struct {
	Key         interface{}       `json:"key"`
	KeyAsString string           `json:"key_as_string,omitempty"`
	DocCount    int64            `json:"doc_count"`
	From        *float64         `json:"from,omitempty"`
	To          *float64         `json:"to,omitempty"`

	// 嵌套聚合
	Aggregations map[string]*AggregationResult `json:"aggregations,omitempty"`
}

// SearchOptions 搜索选项
type SearchOptions struct {
	// 超时设置
	Timeout              time.Duration `json:"timeout,omitempty"`
	TerminateAfter       int           `json:"terminate_after,omitempty"`

	// 路由设置
	Routing              string        `json:"routing,omitempty"`
	Preference           string        `json:"preference,omitempty"`

	// 搜索类型
	SearchType           SearchType    `json:"search_type,omitempty"`

	// 滚动搜索
	Scroll               time.Duration `json:"scroll,omitempty"`

	// 版本控制
	Version              bool          `json:"version,omitempty"`
	VersionType          string        `json:"version_type,omitempty"`

	// 序列号
	SeqNoPrimaryTerm     bool          `json:"seq_no_primary_term,omitempty"`

	// 允许部分结果
	AllowPartialSearchResults bool     `json:"allow_partial_search_results,omitempty"`

	// 批处理大小
	BatchedReduceSize    int           `json:"batched_reduce_size,omitempty"`

	// 请求缓存
	RequestCache         bool          `json:"request_cache,omitempty"`

	// 统计组
	Stats                []string      `json:"stats,omitempty"`

	// 最小评分
	MinScore             *float64      `json:"min_score,omitempty"`

	// 索引增强
	IndicesBoost         map[string]float64 `json:"indices_boost,omitempty"`
}

// SuggestionResult 建议结果
type SuggestionResult struct {
	Text    string                `json:"text"`
	Offset  int                   `json:"offset"`
	Length  int                   `json:"length"`
	Options []*SuggestionOption   `json:"options"`
}

// SuggestionOption 建议选项
type SuggestionOption struct {
	Text      string              `json:"text"`
	Score     float64             `json:"score"`
	Freq      int                 `json:"freq,omitempty"`
	Highlight map[string][]string `json:"highlight,omitempty"`
	CollateMatch bool             `json:"collate_match,omitempty"`
}

// TotalHits 总命中数
type TotalHits struct {
	Value    int64        `json:"value"`
	Relation HitsRelation `json:"relation"`
}

// ShardInfo 分片信息
type ShardInfo struct {
	Total      int `json:"total"`
	Successful int `json:"successful"`
	Skipped    int `json:"skipped"`
	Failed     int `json:"failed"`
}

// DebugInfo 调试信息
type DebugInfo struct {
	QueryTime    time.Duration          `json:"query_time"`
	ParseTime    time.Duration          `json:"parse_time"`
	RewriteTime  time.Duration          `json:"rewrite_time"`

	ParsedQuery  map[string]interface{} `json:"parsed_query,omitempty"`
	Explain      map[string]interface{} `json:"explain,omitempty"`

	ShardStats   map[string]*ShardDebugInfo `json:"shard_stats,omitempty"`
}

// ShardDebugInfo 分片调试信息
type ShardDebugInfo struct {
	QueryTime   time.Duration `json:"query_time"`
	FetchTime   time.Duration `json:"fetch_time"`
	ResultCount int           `json:"result_count"`
}

// Explanation 评分解释
type Explanation struct {
	Value       float64        `json:"value"`
	Description string         `json:"description"`
	Details     []*Explanation `json:"details,omitempty"`
}

// Script 脚本配置
type Script struct {
	Source string                 `json:"source,omitempty"`
	ID     string                 `json:"id,omitempty"`
	Lang   string                 `json:"lang,omitempty"`
	Params map[string]interface{} `json:"params,omitempty"`
}

// ScriptSort 脚本排序
type ScriptSort struct {
	Script *Script   `json:"script"`
	Type   string    `json:"type,omitempty"`
	Order  SortOrder `json:"order,omitempty"`
}

// GeoDistanceSort 地理距离排序
type GeoDistanceSort struct {
	Field      string      `json:"field"`
	Location   interface{} `json:"location"` // can be string, array, or object
	Unit       string      `json:"unit,omitempty"`
	Mode       SortMode    `json:"mode,omitempty"`
	Order      SortOrder   `json:"order,omitempty"`
	IgnoreUnmapped bool    `json:"ignore_unmapped,omitempty"`
}

// RangeParam 范围参数
type RangeParam struct {
	Key  string      `json:"key,omitempty"`
	From interface{} `json:"from,omitempty"`
	To   interface{} `json:"to,omitempty"`
}

// 枚举类型定义
type (
	QueryType       string
	ConditionType   string
	SortOrder       string
	SortMode        string
	AggregationType string
	SearchType      string
	HitsRelation    string
)

// QueryType 常量
const (
	QueryTypeMatch    QueryType = "match"
	QueryTypePhrase   QueryType = "phrase"
	QueryTypePrefix   QueryType = "prefix"
	QueryTypeWildcard QueryType = "wildcard"
	QueryTypeFuzzy    QueryType = "fuzzy"
	QueryTypeTerm     QueryType = "term"
	QueryTypeTerms    QueryType = "terms"
	QueryTypeRange    QueryType = "range"
	QueryTypeBool     QueryType = "bool"
)

// ConditionType 常量
const (
	ConditionTypeTerm     ConditionType = "term"
	ConditionTypeTerms    ConditionType = "terms"
	ConditionTypeRange    ConditionType = "range"
	ConditionTypeMatch    ConditionType = "match"
	ConditionTypeBool     ConditionType = "bool"
	ConditionTypeWildcard ConditionType = "wildcard"
	ConditionTypePrefix   ConditionType = "prefix"
	ConditionTypeFuzzy    ConditionType = "fuzzy"
	ConditionTypeExists   ConditionType = "exists"
)

// SortOrder 常量
const (
	SortOrderAsc  SortOrder = "asc"
	SortOrderDesc SortOrder = "desc"
)

// SortMode 常量
const (
	SortModeMin    SortMode = "min"
	SortModeMax    SortMode = "max"
	SortModeSum    SortMode = "sum"
	SortModeAvg    SortMode = "avg"
	SortModeMedian SortMode = "median"
)

// AggregationType 常量
const (
	AggregationTypeTerms         AggregationType = "terms"
	AggregationTypeHistogram     AggregationType = "histogram"
	AggregationTypeDateHistogram AggregationType = "date_histogram"
	AggregationTypeRange         AggregationType = "range"
	AggregationTypeDateRange     AggregationType = "date_range"
	AggregationTypeSum           AggregationType = "sum"
	AggregationTypeAvg           AggregationType = "avg"
	AggregationTypeMax           AggregationType = "max"
	AggregationTypeMin           AggregationType = "min"
	AggregationTypeCount         AggregationType = "value_count"
	AggregationTypeStats         AggregationType = "stats"
	AggregationTypeExtendedStats AggregationType = "extended_stats"
	AggregationTypePercentiles   AggregationType = "percentiles"
	AggregationTypeCardinality   AggregationType = "cardinality"
	AggregationTypeFilters       AggregationType = "filters"
	AggregationTypeNested        AggregationType = "nested"
	AggregationTypeGeohashGrid   AggregationType = "geohash_grid"
)

// SearchType 常量
const (
	SearchTypeQueryThenFetch SearchType = "query_then_fetch"
	SearchTypeDfsQueryThenFetch SearchType = "dfs_query_then_fetch"
)

// HitsRelation 常量
const (
	HitsRelationEqualTo    HitsRelation = "eq"
	HitsRelationGreaterThan HitsRelation = "gte"
)

// 验证方法

// Validate 验证SearchRequest
func (r *SearchRequest) Validate() error {
	if r.Query == "" {
		return errors.New("query is required")
	}

	if len(r.Query) > 1000 {
		return errors.New("query is too long")
	}

	if r.Index == "" && len(r.Indices) == 0 {
		return errors.New("index or indices is required")
	}

	if r.From < 0 || r.From > 10000 {
		return errors.New("from must be between 0 and 10000")
	}

	if r.Size < 1 || r.Size > 1000 {
		return errors.New("size must be between 1 and 1000")
	}

	// 验证过滤条件
	for _, filter := range r.Filters {
		if err := filter.Validate(); err != nil {
			return fmt.Errorf("invalid filter: %w", err)
		}
	}

	// 验证排序选项
	for _, sort := range r.Sort {
		if err := sort.Validate(); err != nil {
			return fmt.Errorf("invalid sort: %w", err)
		}
	}

	// 验证高亮配置
	if r.Highlight != nil {
		if err := r.Highlight.Validate(); err != nil {
			return fmt.Errorf("invalid highlight: %w", err)
		}
	}

	return nil
}

// Validate 验证QueryCondition
func (q *QueryCondition) Validate() error {
	if q.Type == "" {
		return errors.New("condition type is required")
	}

	// 布尔查询不需要字段名
	if q.Type != ConditionTypeBool && q.Field == "" {
		return errors.New("field is required for non-bool conditions")
	}

	switch q.Type {
	case ConditionTypeTerm:
		if q.Value == nil {
			return errors.New("value is required for term condition")
		}
	case ConditionTypeTerms:
		if len(q.Values) == 0 {
			return errors.New("values are required for terms condition")
		}
	case ConditionTypeRange:
		if q.From == nil && q.To == nil {
			return errors.New("from or to is required for range condition")
		}
	case ConditionTypeBool:
		if len(q.Must) == 0 && len(q.MustNot) == 0 && len(q.Should) == 0 && len(q.Filter) == 0 {
			return errors.New("bool condition must have at least one clause")
		}
		// 递归验证子条件
		for _, cond := range q.Must {
			if err := cond.Validate(); err != nil {
				return err
			}
		}
		for _, cond := range q.MustNot {
			if err := cond.Validate(); err != nil {
				return err
			}
		}
		for _, cond := range q.Should {
			if err := cond.Validate(); err != nil {
				return err
			}
		}
		for _, cond := range q.Filter {
			if err := cond.Validate(); err != nil {
				return err
			}
		}
	}

	return nil
}

// Validate 验证SortOption
func (s *SortOption) Validate() error {
	if s.Field == "" && s.Script == nil && s.GeoDistance == nil {
		return errors.New("field, script, or geo_distance is required")
	}

	if s.Order != "" && s.Order != SortOrderAsc && s.Order != SortOrderDesc {
		return errors.New("order must be 'asc' or 'desc'")
	}

	return nil
}

// Validate 验证HighlightConfig
func (h *HighlightConfig) Validate() error {
	if len(h.Fields) == 0 {
		return errors.New("highlight fields are required")
	}

	for fieldName, field := range h.Fields {
		if fieldName == "" {
			return errors.New("highlight field name cannot be empty")
		}
		if field == nil {
			return errors.New("highlight field configuration cannot be nil")
		}
	}

	return nil
}

// ToJSON 转换为JSON字符串
func (r *SearchRequest) ToJSON() (string, error) {
	data, err := json.Marshal(r)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// FromJSON 从JSON字符串解析
func (r *SearchRequest) FromJSON(jsonStr string) error {
	return json.Unmarshal([]byte(jsonStr), r)
}

// ToJSON 转换为JSON字符串
func (r *SearchResponse) ToJSON() (string, error) {
	data, err := json.Marshal(r)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// FromJSON 从JSON字符串解析
func (r *SearchResponse) FromJSON(jsonStr string) error {
	return json.Unmarshal([]byte(jsonStr), r)
}

// GetTotalHits 获取总命中数
func (r *SearchResponse) GetTotalHits() int64 {
	if r.Hits != nil && r.Hits.Total != nil {
		return r.Hits.Total.Value
	}
	return 0
}

// GetHitsCount 获取返回的结果数
func (r *SearchResponse) GetHitsCount() int {
	if r.Hits != nil {
		return len(r.Hits.Hits)
	}
	return 0
}

// IsEmpty 检查是否为空结果
func (r *SearchResponse) IsEmpty() bool {
	return r.GetHitsCount() == 0
}

// GetMaxScore 获取最大评分
func (r *SearchResponse) GetMaxScore() float64 {
	if r.Hits != nil && r.Hits.MaxScore != nil {
		return *r.Hits.MaxScore
	}
	return 0.0
}

// HasAggregations 检查是否有聚合结果
func (r *SearchResponse) HasAggregations() bool {
	return len(r.Aggregations) > 0
}

// GetSourceAsMap 获取源数据映射
func (h *SearchHit) GetSourceAsMap() map[string]interface{} {
	if h.Source != nil {
		return h.Source
	}
	return make(map[string]interface{})
}

// GetScore 获取评分，如果为空则返回0
func (h *SearchHit) GetScore() float64 {
	if h.Score != nil {
		return *h.Score
	}
	return 0.0
}

// HasHighlight 检查是否有高亮信息
func (h *SearchHit) HasHighlight() bool {
	return len(h.Highlight) > 0
}

// GetHighlightForField 获取指定字段的高亮信息
func (h *SearchHit) GetHighlightForField(field string) []string {
	if h.Highlight != nil {
		return h.Highlight[field]
	}
	return nil
}

// AddFilter 添加过滤条件
func (r *SearchRequest) AddFilter(condition *QueryCondition) {
	if r.Filters == nil {
		r.Filters = make([]*QueryCondition, 0)
	}
	r.Filters = append(r.Filters, condition)
}

// AddSort 添加排序选项
func (r *SearchRequest) AddSort(option *SortOption) {
	if r.Sort == nil {
		r.Sort = make([]*SortOption, 0)
	}
	r.Sort = append(r.Sort, option)
}

// SetPagination 设置分页参数
func (r *SearchRequest) SetPagination(from, size int) {
	r.From = from
	r.Size = size
}

// String 返回字符串表示
func (r *SearchRequest) String() string {
	return fmt.Sprintf("SearchRequest{Query:%s, Index:%s, From:%d, Size:%d}",
		r.Query, r.Index, r.From, r.Size)
}

// String 返回字符串表示
func (r *SearchResponse) String() string {
	return fmt.Sprintf("SearchResponse{Total:%d, Hits:%d, Took:%v}",
		r.GetTotalHits(), r.GetHitsCount(), r.Took)
}

//Personal.AI order the ending
