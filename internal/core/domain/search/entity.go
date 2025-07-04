package search

import (
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
)

// Search 搜索实体，表示一次搜索操作
type Search struct {
	// 基础字段
	ID        string    `json:"id" db:"id"`
	QueryID   string    `json:"query_id" db:"query_id"`
	UserID    string    `json:"user_id" db:"user_id"`
	TenantID  string    `json:"tenant_id" db:"tenant_id"`
	SessionID string    `json:"session_id" db:"session_id"`
	TraceID   string    `json:"trace_id" db:"trace_id"`

	// 搜索内容
	RawQuery     string            `json:"raw_query" db:"raw_query"`
	ProcessedQuery string          `json:"processed_query" db:"processed_query"`
	QueryType    QueryType         `json:"query_type" db:"query_type"`
	Database     string            `json:"database" db:"database"`
	Table        string            `json:"table" db:"table_name"`
	Columns      []string          `json:"columns" db:"columns"`

	// 搜索参数
	Filters      map[string]interface{} `json:"filters" db:"filters"`
	SortBy       []SortField            `json:"sort_by" db:"sort_by"`
	Pagination   *PaginationParams      `json:"pagination" db:"pagination"`
	Options      *SearchOptions         `json:"options" db:"options"`

	// 执行信息
	Status       SearchStatus      `json:"status" db:"status"`
	StartTime    time.Time         `json:"start_time" db:"start_time"`
	EndTime      *time.Time        `json:"end_time" db:"end_time"`
	Duration     time.Duration     `json:"duration" db:"duration"`
	ErrorMessage string            `json:"error_message" db:"error_message"`

	// 结果统计
	TotalHits    int64             `json:"total_hits" db:"total_hits"`
	MaxScore     float64           `json:"max_score" db:"max_score"`
	ResultCount  int               `json:"result_count" db:"result_count"`

	// 性能指标
	Performance  *PerformanceMetrics `json:"performance" db:"performance"`

	// 缓存信息
	CacheKey     string            `json:"cache_key" db:"cache_key"`
	CacheHit     bool              `json:"cache_hit" db:"cache_hit"`
	CacheExpiry  *time.Time        `json:"cache_expiry" db:"cache_expiry"`

	// 元数据
	Metadata     map[string]interface{} `json:"metadata" db:"metadata"`
	Tags         []string              `json:"tags" db:"tags"`
	Priority     SearchPriority        `json:"priority" db:"priority"`

	// 审计字段
	CreatedAt    time.Time         `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time         `json:"updated_at" db:"updated_at"`
	CreatedBy    string            `json:"created_by" db:"created_by"`

	// 关联关系
	Query        *Query            `json:"query,omitempty"`
	Results      []*SearchResult   `json:"results,omitempty"`
	Document     []*Document       `json:"documents,omitempty"`
}

// SearchResult 搜索结果实体
type SearchResult struct {
	// 基础字段
	ID         string    `json:"id" db:"id"`
	SearchID   string    `json:"search_id" db:"search_id"`
	DocumentID string    `json:"document_id" db:"document_id"`
	IndexName  string    `json:"index_name" db:"index_name"`

	// 排序和评分
	Rank       int       `json:"rank" db:"rank"`
	Score      float64   `json:"score" db:"score"`
	Relevance  float64   `json:"relevance" db:"relevance"`
	Confidence float64   `json:"confidence" db:"confidence"`

	// 文档内容
	Source     map[string]interface{} `json:"source" db:"source"`
	Fields     map[string]interface{} `json:"fields" db:"fields"`
	Highlight  map[string][]string    `json:"highlight" db:"highlight"`

	// 排序值
	SortValues []interface{}          `json:"sort_values" db:"sort_values"`

	// 解释信息
	Explanation *ScoreExplanation     `json:"explanation" db:"explanation"`

	// 版本信息
	Version     int64     `json:"version" db:"version"`
	SeqNo       int64     `json:"seq_no" db:"seq_no"`
	PrimaryTerm int64     `json:"primary_term" db:"primary_term"`

	// 元数据
	Metadata    map[string]interface{} `json:"metadata" db:"metadata"`
	Tags        []string               `json:"tags" db:"tags"`

	// 审计字段
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`

	// 关联关系
	Search      *Search   `json:"search,omitempty"`
	Document    *Document `json:"document,omitempty"`
}

// Query 查询实体
type Query struct {
	// 基础字段
	ID          string    `json:"id" db:"id"`
	Hash        string    `json:"hash" db:"hash"`
	UserID      string    `json:"user_id" db:"user_id"`
	TenantID    string    `json:"tenant_id" db:"tenant_id"`

	// 查询内容
	RawQuery    string            `json:"raw_query" db:"raw_query"`
	ParsedQuery *ParsedQuery      `json:"parsed_query" db:"parsed_query"`
	QueryType   QueryType         `json:"query_type" db:"query_type"`
	Language    string            `json:"language" db:"language"`

	// 查询结构
	Conditions  []*QueryCondition `json:"conditions" db:"conditions"`
	Filters     []*FilterCondition `json:"filters" db:"filters"`

	// 执行计划
	ExecutionPlan *ExecutionPlan   `json:"execution_plan" db:"execution_plan"`
	OptimizedQuery string          `json:"optimized_query" db:"optimized_query"`

	// 统计信息
	UseCount    int64             `json:"use_count" db:"use_count"`
	LastUsed    *time.Time        `json:"last_used" db:"last_used"`
	AvgDuration time.Duration     `json:"avg_duration" db:"avg_duration"`
	SuccessRate float64           `json:"success_rate" db:"success_rate"`

	// 缓存配置
	Cacheable   bool              `json:"cacheable" db:"cacheable"`
	CacheTTL    time.Duration     `json:"cache_ttl" db:"cache_ttl"`

	// 验证信息
	IsValid     bool              `json:"is_valid" db:"is_valid"`
	Errors      []string          `json:"errors" db:"errors"`
	Warnings    []string          `json:"warnings" db:"warnings"`

	// 元数据
	Metadata    map[string]interface{} `json:"metadata" db:"metadata"`
	Tags        []string              `json:"tags" db:"tags"`

	// 审计字段
	CreatedAt   time.Time         `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time         `json:"updated_at" db:"updated_at"`

	// 关联关系
	Searches    []*Search         `json:"searches,omitempty"`
}

// Document 文档实体
type Document struct {
	// 基础字段
	ID          string    `json:"id" db:"id"`
	ExternalID  string    `json:"external_id" db:"external_id"`
	IndexName   string    `json:"index_name" db:"index_name"`
	Database    string    `json:"database" db:"database"`
	Table       string    `json:"table_name" db:"table_name"`

	// 文档内容
	Source      map[string]interface{} `json:"source" db:"source"`
	Fields      map[string]*DocumentField `json:"fields" db:"fields"`

	// 索引信息
	IndexedAt   time.Time         `json:"indexed_at" db:"indexed_at"`
	IndexStatus DocumentStatus    `json:"index_status" db:"index_status"`
	IndexError  string            `json:"index_error" db:"index_error"`

	// 版本控制
	Version     int64             `json:"version" db:"version"`
	SeqNo       int64             `json:"seq_no" db:"seq_no"`
	PrimaryTerm int64             `json:"primary_term" db:"primary_term"`
	ParentID    string            `json:"parent_id" db:"parent_id"`
	RoutingValue string           `json:"routing_value" db:"routing_value"`

	// 文档属性
	Size        int64             `json:"size" db:"size"`
	Language    string            `json:"language" db:"language"`
	ContentType string            `json:"content_type" db:"content_type"`
	Encoding    string            `json:"encoding" db:"encoding"`

	// 处理信息
	Pipeline    string            `json:"pipeline" db:"pipeline"`
	Processor   []string          `json:"processor" db:"processor"`

	// 统计信息
	AccessCount int64             `json:"access_count" db:"access_count"`
	LastAccessed *time.Time       `json:"last_accessed" db:"last_accessed"`

	// 元数据
	Metadata    map[string]interface{} `json:"metadata" db:"metadata"`
	Tags        []string              `json:"tags" db:"tags"`

	// 审计字段
	CreatedAt   time.Time         `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time         `json:"updated_at" db:"updated_at"`
	CreatedBy   string            `json:"created_by" db:"created_by"`
	UpdatedBy   string            `json:"updated_by" db:"updated_by"`

	// 关联关系
	SearchResults []*SearchResult `json:"search_results,omitempty"`
}

// 支持结构体定义

// SortField 排序字段
type SortField struct {
	Field string    `json:"field"`
	Order SortOrder `json:"order"`
	Mode  SortMode  `json:"mode,omitempty"`
}

// PaginationParams 分页参数
type PaginationParams struct {
	Page   int    `json:"page"`
	Size   int    `json:"size"`
	Offset int    `json:"offset"`
	Cursor string `json:"cursor,omitempty"`
}

// SearchOptions 搜索选项
type SearchOptions struct {
	Timeout          time.Duration `json:"timeout"`
	IncludeScore     bool          `json:"include_score"`
	IncludeSource    bool          `json:"include_source"`
	IncludeHighlight bool          `json:"include_highlight"`
	TrackTotalHits   bool          `json:"track_total_hits"`
	AllowPartialResults bool       `json:"allow_partial_results"`
	Analyzer         string        `json:"analyzer,omitempty"`
	Fuzziness        string        `json:"fuzziness,omitempty"`
	MinScore         float64       `json:"min_score,omitempty"`
}

// PerformanceMetrics 性能指标
type PerformanceMetrics struct {
	QueryTime     time.Duration `json:"query_time"`
	FetchTime     time.Duration `json:"fetch_time"`
	ProcessTime   time.Duration `json:"process_time"`
	SerializeTime time.Duration `json:"serialize_time"`
	TotalTime     time.Duration `json:"total_time"`
	MemoryUsage   int64         `json:"memory_usage"`
	CPUUsage      float64       `json:"cpu_usage"`
	CacheHitRate  float64       `json:"cache_hit_rate"`
}

// ParsedQuery 解析后的查询
type ParsedQuery struct {
	Type       QueryType                  `json:"type"`
	Terms      []string                   `json:"terms"`
	Phrases    []string                   `json:"phrases"`
	Wildcards  []string                   `json:"wildcards"`
	Operators  []string                   `json:"operators"`
	Fields     []string                   `json:"fields"`
	Boosts     map[string]float64         `json:"boosts"`
	Conditions map[string]interface{}     `json:"conditions"`
}

// QueryCondition 查询条件
type QueryCondition struct {
	Field     string      `json:"field"`
	Operator  string      `json:"operator"`
	Value     interface{} `json:"value"`
	Boost     float64     `json:"boost,omitempty"`
	Required  bool        `json:"required"`
}

// FilterCondition 过滤条件
type FilterCondition struct {
	Field     string      `json:"field"`
	Operator  string      `json:"operator"`
	Value     interface{} `json:"value"`
	Values    []interface{} `json:"values,omitempty"`
}

// ExecutionPlan 执行计划
type ExecutionPlan struct {
	Stages       []*ExecutionStage      `json:"stages"`
	EstimatedCost float64               `json:"estimated_cost"`
	Indexes      []string               `json:"indexes"`
	Optimizations []string              `json:"optimizations"`
	Warnings     []string               `json:"warnings"`
}

// ExecutionStage 执行阶段
type ExecutionStage struct {
	Type        string                 `json:"type"`
	Operation   string                 `json:"operation"`
	Cost        float64                `json:"cost"`
	EstimatedRows int64                `json:"estimated_rows"`
	Details     map[string]interface{} `json:"details"`
}

// ScoreExplanation 得分解释
type ScoreExplanation struct {
	Value       float64              `json:"value"`
	Description string               `json:"description"`
	Details     []*ScoreExplanation  `json:"details,omitempty"`
}

// DocumentField 文档字段
type DocumentField struct {
	Name         string      `json:"name"`
	Type         string      `json:"type"`
	Value        interface{} `json:"value"`
	Indexed      bool        `json:"indexed"`
	Stored       bool        `json:"stored"`
	Analyzed     bool        `json:"analyzed"`
	Boost        float64     `json:"boost,omitempty"`
	Analyzer     string      `json:"analyzer,omitempty"`
}

// 枚举类型定义
type (
	QueryType       string
	SearchStatus    string
	SearchPriority  string
	SortOrder       string
	SortMode        string
	DocumentStatus  string
)

// 枚举常量定义
const (
	// QueryType 常量
	QueryTypeMatch       QueryType = "match"
	QueryTypeMatchPhrase QueryType = "match_phrase"
	QueryTypeMatchAll    QueryType = "match_all"
	QueryTypeWildcard    QueryType = "wildcard"
	QueryTypeRegex       QueryType = "regex"
	QueryTypeFuzzy       QueryType = "fuzzy"
	QueryTypeTerm        QueryType = "term"
	QueryTypeRange       QueryType = "range"
	QueryTypeBool        QueryType = "bool"
	QueryTypePrefix      QueryType = "prefix"

	// SearchStatus 常量
	SearchStatusPending   SearchStatus = "pending"
	SearchStatusRunning   SearchStatus = "running"
	SearchStatusCompleted SearchStatus = "completed"
	SearchStatusFailed    SearchStatus = "failed"
	SearchStatusCancelled SearchStatus = "cancelled"
	SearchStatusTimeout   SearchStatus = "timeout"

	// SearchPriority 常量
	SearchPriorityLow    SearchPriority = "low"
	SearchPriorityNormal SearchPriority = "normal"
	SearchPriorityHigh   SearchPriority = "high"
	SearchPriorityUrgent SearchPriority = "urgent"

	// SortOrder 常量
	SortOrderAsc  SortOrder = "asc"
	SortOrderDesc SortOrder = "desc"

	// SortMode 常量
	SortModeMin SortMode = "min"
	SortModeMax SortMode = "max"
	SortModeSum SortMode = "sum"
	SortModeAvg SortMode = "avg"

	// DocumentStatus 常量
	DocumentStatusIndexed   DocumentStatus = "indexed"
	DocumentStatusIndexing  DocumentStatus = "indexing"
	DocumentStatusFailed    DocumentStatus = "failed"
	DocumentStatusDeleted   DocumentStatus = "deleted"
	DocumentStatusUpdating  DocumentStatus = "updating"
)

// Search 实体方法

// NewSearch 创建新的搜索实体
func NewSearch(userID, tenantID, rawQuery, database, table string) *Search {
	now := time.Now()
	id := uuid.New().String()

	return &Search{
		ID:           id,
		QueryID:      generateQueryID(rawQuery),
		UserID:       userID,
		TenantID:     tenantID,
		RawQuery:     rawQuery,
		Database:     database,
		Table:        table,
		Status:       SearchStatusPending,
		Priority:     SearchPriorityNormal,
		StartTime:    now,
		CreatedAt:    now,
		UpdatedAt:    now,
		Metadata:     make(map[string]interface{}),
		Results:      make([]*SearchResult, 0),
	}
}

// Validate 验证搜索实体
func (s *Search) Validate() error {
	if s.ID == "" {
		return errors.New("search ID is required")
	}

	if s.UserID == "" {
		return errors.New("user ID is required")
	}

	if s.TenantID == "" {
		return errors.New("tenant ID is required")
	}

	if s.RawQuery == "" {
		return errors.New("raw query is required")
	}

	if len(s.RawQuery) > 10000 {
		return errors.New("query too long (max 10000 characters)")
	}

	if s.Database == "" {
		return errors.New("database is required")
	}

	if s.Table == "" {
		return errors.New("table is required")
	}

	if !isValidQueryType(s.QueryType) {
		return fmt.Errorf("invalid query type: %s", s.QueryType)
	}

	if !isValidSearchStatus(s.Status) {
		return fmt.Errorf("invalid search status: %s", s.Status)
	}

	if !isValidSearchPriority(s.Priority) {
		return fmt.Errorf("invalid search priority: %s", s.Priority)
	}

	// 验证分页参数
	if s.Pagination != nil {
		if err := s.Pagination.Validate(); err != nil {
			return fmt.Errorf("pagination validation failed: %w", err)
		}
	}

	return nil
}

// Start 开始搜索
func (s *Search) Start() error {
	if s.Status != SearchStatusPending {
		return fmt.Errorf("cannot start search in status: %s", s.Status)
	}

	s.Status = SearchStatusRunning
	s.StartTime = time.Now()
	s.UpdatedAt = time.Now()

	return nil
}

// Complete 完成搜索
func (s *Search) Complete(totalHits int64, maxScore float64, resultCount int) error {
	if s.Status != SearchStatusRunning {
		return fmt.Errorf("cannot complete search in status: %s", s.Status)
	}

	now := time.Now()
	s.Status = SearchStatusCompleted
	s.EndTime = &now
	s.Duration = now.Sub(s.StartTime)
	s.TotalHits = totalHits
	s.MaxScore = maxScore
	s.ResultCount = resultCount
	s.UpdatedAt = now

	return nil
}

// Fail 搜索失败
func (s *Search) Fail(errorMessage string) error {
	if s.Status == SearchStatusCompleted {
		return errors.New("cannot fail completed search")
	}

	now := time.Now()
	s.Status = SearchStatusFailed
	s.EndTime = &now
	s.Duration = now.Sub(s.StartTime)
	s.ErrorMessage = errorMessage
	s.UpdatedAt = now

	return nil
}

// Cancel 取消搜索
func (s *Search) Cancel() error {
	if s.Status == SearchStatusCompleted || s.Status == SearchStatusFailed {
		return fmt.Errorf("cannot cancel search in status: %s", s.Status)
	}

	now := time.Now()
	s.Status = SearchStatusCancelled
	s.EndTime = &now
	s.Duration = now.Sub(s.StartTime)
	s.UpdatedAt = now

	return nil
}

// AddResult 添加搜索结果
func (s *Search) AddResult(result *SearchResult) error {
	if result == nil {
		return errors.New("result cannot be nil")
	}

	if result.SearchID != s.ID {
		result.SearchID = s.ID
	}

	result.Search = s
	s.Results = append(s.Results, result)

	return nil
}

// SetQuery 设置查询对象
func (s *Search) SetQuery(query *Query) {
	s.Query = query
	s.QueryID = query.ID
}

// SetPerformance 设置性能指标
func (s *Search) SetPerformance(performance *PerformanceMetrics) {
	s.Performance = performance
}

// AddTag 添加标签
func (s *Search) AddTag(tag string) {
	if !contains(s.Tags, tag) {
		s.Tags = append(s.Tags, tag)
	}
}

// RemoveTag 移除标签
func (s *Search) RemoveTag(tag string) {
	for i, t := range s.Tags {
		if t == tag {
			s.Tags = append(s.Tags[:i], s.Tags[i+1:]...)
			break
		}
	}
}

// HasTag 检查是否有标签
func (s *Search) HasTag(tag string) bool {
	return contains(s.Tags, tag)
}

// SetMetadata 设置元数据
func (s *Search) SetMetadata(key string, value interface{}) {
	if s.Metadata == nil {
		s.Metadata = make(map[string]interface{})
	}
	s.Metadata[key] = value
}

// GetMetadata 获取元数据
func (s *Search) GetMetadata(key string) (interface{}, bool) {
	if s.Metadata == nil {
		return nil, false
	}
	value, exists := s.Metadata[key]
	return value, exists
}

// IsCompleted 检查是否完成
func (s *Search) IsCompleted() bool {
	return s.Status == SearchStatusCompleted
}

// IsFailed 检查是否失败
func (s *Search) IsFailed() bool {
	return s.Status == SearchStatusFailed
}

// IsRunning 检查是否运行中
func (s *Search) IsRunning() bool {
	return s.Status == SearchStatusRunning
}

// GetDurationMs 获取持续时间（毫秒）
func (s *Search) GetDurationMs() int64 {
	return s.Duration.Milliseconds()
}

// SearchResult 实体方法

// NewSearchResult 创建新的搜索结果
func NewSearchResult(searchID, documentID, indexName string, rank int, score float64) *SearchResult {
	now := time.Now()

	return &SearchResult{
		ID:         uuid.New().String(),
		SearchID:   searchID,
		DocumentID: documentID,
		IndexName:  indexName,
		Rank:       rank,
		Score:      score,
		Source:     make(map[string]interface{}),
		Fields:     make(map[string]interface{}),
		Highlight:  make(map[string][]string),
		Metadata:   make(map[string]interface{}),
		CreatedAt:  now,
		UpdatedAt:  now,
	}
}

// Validate 验证搜索结果
func (sr *SearchResult) Validate() error {
	if sr.ID == "" {
		return errors.New("search result ID is required")
	}

	if sr.SearchID == "" {
		return errors.New("search ID is required")
	}

	if sr.DocumentID == "" {
		return errors.New("document ID is required")
	}

	if sr.IndexName == "" {
		return errors.New("index name is required")
	}

	if sr.Rank < 1 {
		return errors.New("rank must be greater than 0")
	}

	if sr.Score < 0 {
		return errors.New("score cannot be negative")
	}

	return nil
}

// SetSource 设置源数据
func (sr *SearchResult) SetSource(source map[string]interface{}) {
	sr.Source = source
	sr.UpdatedAt = time.Now()
}

// SetHighlight 设置高亮信息
func (sr *SearchResult) SetHighlight(highlight map[string][]string) {
	sr.Highlight = highlight
	sr.UpdatedAt = time.Now()
}

// SetExplanation 设置得分解释
func (sr *SearchResult) SetExplanation(explanation *ScoreExplanation) {
	sr.Explanation = explanation
	sr.UpdatedAt = time.Now()
}

// AddHighlight 添加高亮字段
func (sr *SearchResult) AddHighlight(field string, fragments []string) {
	if sr.Highlight == nil {
		sr.Highlight = make(map[string][]string)
	}
	sr.Highlight[field] = fragments
}

// GetHighlight 获取高亮信息
func (sr *SearchResult) GetHighlight(field string) ([]string, bool) {
	if sr.Highlight == nil {
		return nil, false
	}
	fragments, exists := sr.Highlight[field]
	return fragments, exists
}

// Query 实体方法

// NewQuery 创建新的查询实体
func NewQuery(userID, tenantID, rawQuery string, queryType QueryType) *Query {
	now := time.Now()

	return &Query{
		ID:         uuid.New().String(),
		Hash:       generateQueryHash(rawQuery),
		UserID:     userID,
		TenantID:   tenantID,
		RawQuery:   rawQuery,
		QueryType:  queryType,
		IsValid:    false,
		Cacheable:  true,
		CacheTTL:   time.Hour,
		Metadata:   make(map[string]interface{}),
		CreatedAt:  now,
		UpdatedAt:  now,
		Searches:   make([]*Search, 0),
	}
}

// Validate 验证查询实体
func (q *Query) Validate() error {
	q.Errors = make([]string, 0)
	q.Warnings = make([]string, 0)

	if q.ID == "" {
		q.Errors = append(q.Errors, "query ID is required")
	}

	if q.UserID == "" {
		q.Errors = append(q.Errors, "user ID is required")
	}

	if q.TenantID == "" {
		q.Errors = append(q.Errors, "tenant ID is required")
	}

	if q.RawQuery == "" {
		q.Errors = append(q.Errors, "raw query is required")
	}

	if len(q.RawQuery) > 10000 {
		q.Errors = append(q.Errors, "query too long (max 10000 characters)")
	}

	if !isValidQueryType(q.QueryType) {
		q.Errors = append(q.Errors, fmt.Sprintf("invalid query type: %s", q.QueryType))
	}

	// 验证查询语法
	if err := q.validateSyntax(); err != nil {
		q.Errors = append(q.Errors, err.Error())
	}

	// 设置验证状态
	q.IsValid = len(q.Errors) == 0
	q.UpdatedAt = time.Now()

	if len(q.Errors) > 0 {
		return fmt.Errorf("query validation failed: %v", q.Errors)
	}

	return nil
}

// validateSyntax 验证查询语法
func (q *Query) validateSyntax() error {
	// 基础语法检查
	if strings.TrimSpace(q.RawQuery) == "" {
		return errors.New("query cannot be empty")
	}

	// 检查特殊字符
	if containsInvalidChars(q.RawQuery) {
		q.Warnings = append(q.Warnings, "query contains potentially unsafe characters")
	}

	// 检查查询长度
	if len(strings.Fields(q.RawQuery)) > 100 {
		q.Warnings = append(q.Warnings, "query is very long and may impact performance")
	}

	return nil
}

// Parse 解析查询
func (q *Query) Parse() error {
	parsedQuery, err := parseQuery(q.RawQuery, q.QueryType)
	if err != nil {
		q.IsValid = false
		q.Errors = append(q.Errors, err.Error())
		return err
	}

	q.ParsedQuery = parsedQuery
	q.IsValid = true
	q.UpdatedAt = time.Now()

	return nil
}

// Optimize 优化查询
func (q *Query) Optimize() error {
	if !q.IsValid {
		return errors.New("cannot optimize invalid query")
	}

	// 实现查询优化逻辑
	optimized := optimizeQuery(q.RawQuery)
	q.OptimizedQuery = optimized
	q.UpdatedAt = time.Now()

	return nil
}

// GenerateExecutionPlan 生成执行计划
func (q *Query) GenerateExecutionPlan() error {
	if !q.IsValid {
		return errors.New("cannot generate execution plan for invalid query")
	}

	plan, err := generateExecutionPlan(q)
	if err != nil {
		return err
	}

	q.ExecutionPlan = plan
	q.UpdatedAt = time.Now()

	return nil
}

// AddSearch 添加搜索记录
func (q *Query) AddSearch(search *Search) {
	search.Query = q
	search.QueryID = q.ID
	q.Searches = append(q.Searches, search)
	q.UseCount++
	q.LastUsed = &search.StartTime
	q.UpdatedAt = time.Now()
}

// UpdateStats 更新统计信息
func (q *Query) UpdateStats(duration time.Duration, success bool) {
	q.UseCount++
	now := time.Now()
	q.LastUsed = &now

	// 更新平均执行时间
	if q.AvgDuration == 0 {
		q.AvgDuration = duration
	} else {
		q.AvgDuration = (q.AvgDuration + duration) / 2
	}

	// 更新成功率
	if success {
		q.SuccessRate = (q.SuccessRate*float64(q.UseCount-1) + 1.0) / float64(q.UseCount)
	} else {
		q.SuccessRate = (q.SuccessRate * float64(q.UseCount-1)) / float64(q.UseCount)
	}

	q.UpdatedAt = time.Now()
}

// Document 实体方法

// NewDocument 创建新的文档实体
func NewDocument(externalID, indexName, database, table string) *Document {
	now := time.Now()

	return &Document{
		ID:          uuid.New().String(),
		ExternalID:  externalID,
		IndexName:   indexName,
		Database:    database,
		Table:       table,
		Source:      make(map[string]interface{}),
		Fields:      make(map[string]*DocumentField),
		IndexStatus: DocumentStatusIndexing,
		Version:     1,
		Metadata:    make(map[string]interface{}),
		CreatedAt:   now,
		UpdatedAt:   now,
		IndexedAt:   now,
	}
}

// Validate 验证文档实体
func (d *Document) Validate() error {
	if d.ID == "" {
		return errors.New("document ID is required")
	}

	if d.ExternalID == "" {
		return errors.New("external ID is required")
	}

	if d.IndexName == "" {
		return errors.New("index name is required")
	}

	if d.Database == "" {
		return errors.New("database is required")
	}

	if d.Table == "" {
		return errors.New("table is required")
	}

	if !isValidDocumentStatus(d.IndexStatus) {
		return fmt.Errorf("invalid document status: %s", d.IndexStatus)
	}

	if d.Version < 1 {
		return errors.New("version must be greater than 0")
	}

	return nil
}

// SetSource 设置源数据
func (d *Document) SetSource(source map[string]interface{}) {
	d.Source = source
	d.Size = int64(calculateSourceSize(source))
	d.UpdatedAt = time.Now()
}

// AddField 添加字段
func (d *Document) AddField(name, fieldType string, value interface{}) {
	field := &DocumentField{
		Name:    name,
		Type:    fieldType,
		Value:   value,
		Indexed: true,
		Stored:  true,
	}

	if d.Fields == nil {
		d.Fields = make(map[string]*DocumentField)
	}

	d.Fields[name] = field
	d.UpdatedAt = time.Now()
}

// GetField 获取字段
func (d *Document) GetField(name string) (*DocumentField, bool) {
	if d.Fields == nil {
		return nil, false
	}
	field, exists := d.Fields[name]
	return field, exists
}

// MarkIndexed 标记为已索引
func (d *Document) MarkIndexed() {
	d.IndexStatus = DocumentStatusIndexed
	d.IndexError = ""
	d.IndexedAt = time.Now()
	d.UpdatedAt = time.Now()
}

// MarkIndexFailed 标记索引失败
func (d *Document) MarkIndexFailed(errorMessage string) {
	d.IndexStatus = DocumentStatusFailed
	d.IndexError = errorMessage
	d.UpdatedAt = time.Now()
}

// UpdateVersion 更新版本
func (d *Document) UpdateVersion() {
	d.Version++
	d.SeqNo++
	d.UpdatedAt = time.Now()
}

// MarkDeleted 标记为已删除
func (d *Document) MarkDeleted() {
	d.IndexStatus = DocumentStatusDeleted
	d.UpdatedAt = time.Now()
}

// IncrementAccess 增加访问计数
func (d *Document) IncrementAccess() {
	d.AccessCount++
	now := time.Now()
	d.LastAccessed = &now
}

// 分页参数验证方法

// Validate 验证分页参数
func (p *PaginationParams) Validate() error {
	if p.Page < 1 {
		return errors.New("page must be greater than 0")
	}

	if p.Size < 1 || p.Size > 10000 {
		return errors.New("size must be between 1 and 10000")
	}

	if p.Offset < 0 {
		return errors.New("offset cannot be negative")
	}

	return nil
}

// 工具函数

// generateQueryID 生成查询ID
func generateQueryID(rawQuery string) string {
	return fmt.Sprintf("query_%s_%d", generateHash(rawQuery)[:8], time.Now().Unix())
}

// generateQueryHash 生成查询哈希
func generateQueryHash(rawQuery string) string {
	return generateHash(rawQuery)
}

// generateHash 生成哈希值
func generateHash(input string) string {
	// 简化的哈希实现，实际应使用更强的哈希算法
	return fmt.Sprintf("%x", len(input)*31+strings.Count(input, " "))
}

// isValidQueryType 验证查询类型
func isValidQueryType(queryType QueryType) bool {
	validTypes := []QueryType{
		QueryTypeMatch, QueryTypeMatchPhrase, QueryTypeMatchAll,
		QueryTypeWildcard, QueryTypeRegex, QueryTypeFuzzy,
		QueryTypeTerm, QueryTypeRange, QueryTypeBool, QueryTypePrefix,
	}

	for _, t := range validTypes {
		if t == queryType {
			return true
		}
	}
	return false
}

// isValidSearchStatus 验证搜索状态
func isValidSearchStatus(status SearchStatus) bool {
	validStatuses := []SearchStatus{
		SearchStatusPending, SearchStatusRunning, SearchStatusCompleted,
		SearchStatusFailed, SearchStatusCancelled, SearchStatusTimeout,
	}

	for _, s := range validStatuses {
		if s == status {
			return true
		}
	}
	return false
}

// isValidSearchPriority 验证搜索优先级
func isValidSearchPriority(priority SearchPriority) bool {
	validPriorities := []SearchPriority{
		SearchPriorityLow, SearchPriorityNormal,
		SearchPriorityHigh, SearchPriorityUrgent,
	}

	for _, p := range validPriorities {
		if p == priority {
			return true
		}
	}
	return false
}

// isValidDocumentStatus 验证文档状态
func isValidDocumentStatus(status DocumentStatus) bool {
	validStatuses := []DocumentStatus{
		DocumentStatusIndexed, DocumentStatusIndexing, DocumentStatusFailed,
		DocumentStatusDeleted, DocumentStatusUpdating,
	}

	for _, s := range validStatuses {
		if s == status {
			return true
		}
	}
	return false
}

// contains 检查字符串数组是否包含指定值
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// containsInvalidChars 检查是否包含无效字符
func containsInvalidChars(query string) bool {
	// 检查潜在的SQL注入或其他危险字符
	dangerousPatterns := []string{
		"<script", "</script>", "javascript:", "eval(",
		"DROP TABLE", "DELETE FROM", "UPDATE SET",
	}

	lowerQuery := strings.ToLower(query)
	for _, pattern := range dangerousPatterns {
		if strings.Contains(lowerQuery, strings.ToLower(pattern)) {
			return true
		}
	}

	return false
}

// parseQuery 解析查询（简化实现）
func parseQuery(rawQuery string, queryType QueryType) (*ParsedQuery, error) {
	parsed := &ParsedQuery{
		Type:       queryType,
		Terms:      strings.Fields(rawQuery),
		Conditions: make(map[string]interface{}),
	}

	// 基础解析逻辑
	if queryType == QueryTypeMatchPhrase {
		parsed.Phrases = []string{rawQuery}
	}

	// 提取通配符
	wildcardPattern := regexp.MustCompile(`\*|\?`)
	if wildcardPattern.MatchString(rawQuery) {
		parsed.Wildcards = wildcardPattern.FindAllString(rawQuery, -1)
	}

	return parsed, nil
}

// optimizeQuery 优化查询（简化实现）
func optimizeQuery(rawQuery string) string {
	// 基础优化：去除多余空格
	optimized := strings.TrimSpace(rawQuery)
	optimized = regexp.MustCompile(`\s+`).ReplaceAllString(optimized, " ")

	return optimized
}

// generateExecutionPlan 生成执行计划（简化实现）
func generateExecutionPlan(query *Query) (*ExecutionPlan, error) {
	plan := &ExecutionPlan{
		Stages:        make([]*ExecutionStage, 0),
		EstimatedCost: 1.0,
		Indexes:       []string{},
		Optimizations: []string{},
	}

	// 添加基础执行阶段
	stage := &ExecutionStage{
		Type:          "search",
		Operation:     "full_text_search",
		Cost:          1.0,
		EstimatedRows: 1000,
		Details:       make(map[string]interface{}),
	}

	plan.Stages = append(plan.Stages, stage)

	return plan, nil
}

// calculateSourceSize 计算源数据大小（简化实现）
func calculateSourceSize(source map[string]interface{}) int {
	data, _ := json.Marshal(source)
	return len(data)
}

//Personal.AI order the ending
