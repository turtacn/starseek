package search

import (
	"context"
	"time"
)

// =============================================================================
// Search Domain Repository Interface
// =============================================================================

// SearchDomainRepository defines the interface for search domain data access operations
type SearchDomainRepository interface {
	// Full-text search operations
	ExecuteFullTextSearch(ctx context.Context, req *FullTextSearchRequest) (*FullTextSearchResult, error)
	ExecuteComplexSearch(ctx context.Context, req *ComplexSearchRequest) (*ComplexSearchResult, error)
	ExecuteAdvancedSearch(ctx context.Context, req *AdvancedSearchRequest) (*AdvancedSearchResult, error)

	// Batch search operations
	ExecuteBatchSearch(ctx context.Context, req *BatchSearchRequest) (*BatchSearchResult, error)
	ExecuteParallelSearch(ctx context.Context, req *ParallelSearchRequest) (*ParallelSearchResult, error)
	ExecuteMultiIndexSearch(ctx context.Context, req *MultiIndexSearchRequest) (*MultiIndexSearchResult, error)

	// Search statistics and analytics
	GetSearchStatistics(ctx context.Context, req *SearchStatisticsRequest) (*SearchStatistics, error)
	GetIndexStatistics(ctx context.Context, indexName string) (*IndexStatistics, error)
	GetQueryPerformanceMetrics(ctx context.Context, req *QueryPerformanceRequest) (*QueryPerformanceMetrics, error)
	GetSearchTrends(ctx context.Context, req *SearchTrendsRequest) (*SearchTrends, error)

	// Document operations
	GetDocumentCount(ctx context.Context, indexName string, filter *DocumentFilter) (int64, error)
	GetDocumentDistribution(ctx context.Context, req *DocumentDistributionRequest) (*DocumentDistribution, error)
	GetDocumentSample(ctx context.Context, req *DocumentSampleRequest) (*DocumentSample, error)

	// Index health and status
	GetIndexHealth(ctx context.Context, indexName string) (*IndexHealth, error)
	GetIndexStatus(ctx context.Context, indexName string) (*IndexStatus, error)
	GetIndexMetadata(ctx context.Context, indexName string) (*IndexMetadata, error)
	GetClusterHealth(ctx context.Context) (*ClusterHealth, error)

	// Query analysis and optimization
	AnalyzeQuery(ctx context.Context, req *QueryAnalysisRequest) (*QueryAnalysisResult, error)
	ExplainQuery(ctx context.Context, req *QueryExplanationRequest) (*QueryExplanation, error)
	OptimizeQuery(ctx context.Context, req *QueryOptimizationRequest) (*QueryOptimizationResult, error)
	ValidateQuery(ctx context.Context, req *QueryValidationRequest) (*QueryValidationResult, error)

	// Search result caching
	GetCachedSearchResult(ctx context.Context, cacheKey string) (*CachedSearchResult, error)
	SetCachedSearchResult(ctx context.Context, cacheKey string, result *CachedSearchResult, ttl time.Duration) error
	InvalidateCachedResults(ctx context.Context, pattern string) error
	GetCacheStatistics(ctx context.Context) (*CacheStatistics, error)

	// Search suggestions and auto-complete
	GetSearchSuggestions(ctx context.Context, req *SuggestionRequest) (*SuggestionResult, error)
	GetAutoCompleteSuggestions(ctx context.Context, req *AutoCompleteRequest) (*AutoCompleteResult, error)
	IndexSuggestionTerms(ctx context.Context, req *IndexSuggestionRequest) error
	UpdateSuggestionFrequency(ctx context.Context, req *SuggestionFrequencyRequest) error

	// Faceted search operations
	ExecuteFacetedSearch(ctx context.Context, req *FacetedSearchRequest) (*FacetedSearchResult, error)
	GetFacetValues(ctx context.Context, req *FacetValuesRequest) (*FacetValues, error)
	GetFacetCounts(ctx context.Context, req *FacetCountsRequest) (*FacetCounts, error)

	// Aggregation operations
	ExecuteAggregation(ctx context.Context, req *AggregationRequest) (*AggregationResult, error)
	ExecuteMultiAggregation(ctx context.Context, req *MultiAggregationRequest) (*MultiAggregationResult, error)
	GetAggregationStatistics(ctx context.Context, req *AggregationStatisticsRequest) (*AggregationStatistics, error)

	// Scroll and pagination
	InitializeScroll(ctx context.Context, req *ScrollInitRequest) (*ScrollContext, error)
	ContinueScroll(ctx context.Context, scrollID string) (*ScrollResult, error)
	ClearScroll(ctx context.Context, scrollID string) error
	GetScrollStatistics(ctx context.Context) (*ScrollStatistics, error)

	// Search monitoring and logging
	LogSearchQuery(ctx context.Context, req *SearchQueryLog) error
	GetSearchQueryHistory(ctx context.Context, req *SearchHistoryRequest) (*SearchHistory, error)
	GetSearchMetrics(ctx context.Context, req *SearchMetricsRequest) (*SearchMetrics, error)

	// Index maintenance and optimization
	RefreshIndex(ctx context.Context, indexName string) (*RefreshResult, error)
	OptimizeIndex(ctx context.Context, indexName string, options *OptimizeOptions) (*OptimizeResult, error)
	FlushIndex(ctx context.Context, indexName string) (*FlushResult, error)

	// Transaction support
	BeginSearchTransaction(ctx context.Context, options *TransactionOptions) (SearchTransaction, error)

	// Health check and monitoring
	HealthCheck(ctx context.Context) (*HealthCheckResult, error)
	GetPerformanceMetrics(ctx context.Context) (*PerformanceMetrics, error)
	GetResourceUsage(ctx context.Context) (*ResourceUsage, error)
}

// =============================================================================
// Search Transaction Interface
// =============================================================================

// SearchTransaction defines the interface for search transaction operations
type SearchTransaction interface {
	// Transaction control
	Commit(ctx context.Context) error
	Rollback(ctx context.Context) error
	IsActive() bool

	// Search operations within transaction
	ExecuteSearch(ctx context.Context, req *TransactionalSearchRequest) (*TransactionalSearchResult, error)
	ExecuteBatchSearch(ctx context.Context, req *TransactionalBatchSearchRequest) (*TransactionalBatchSearchResult, error)

	// State management
	GetTransactionID() string
	GetStartTime() time.Time
	GetTimeout() time.Duration
}

// =============================================================================
// Full-Text Search Request/Response Types
// =============================================================================

// FullTextSearchRequest represents a full-text search request
type FullTextSearchRequest struct {
	// Basic search parameters
	Query    string   `json:"query"`
	Indexes  []string `json:"indexes"`
	Fields   []string `json:"fields,omitempty"`
	Analyzer string   `json:"analyzer,omitempty"`

	// Pagination
	From int `json:"from,omitempty"`
	Size int `json:"size,omitempty"`

	// Sorting
	Sort []*SortField `json:"sort,omitempty"`

	// Filtering
	Filters []*SearchFilter `json:"filters,omitempty"`

	// Highlighting
	Highlight *HighlightOptions `json:"highlight,omitempty"`

	// Performance options
	Timeout    time.Duration `json:"timeout,omitempty"`
	Preference string        `json:"preference,omitempty"`
	Routing    string        `json:"routing,omitempty"`

	// Search options
	MinScore    *float64 `json:"min_score,omitempty"`
	TrackScores bool     `json:"track_scores,omitempty"`

	// Metadata
	RequestID string `json:"request_id,omitempty"`
	UserID    string `json:"user_id,omitempty"`
	SessionID string `json:"session_id,omitempty"`
}

// FullTextSearchResult represents a full-text search result
type FullTextSearchResult struct {
	// Search results
	Hits     []*SearchHit `json:"hits"`
	Total    int64        `json:"total"`
	MaxScore float64      `json:"max_score"`

	// Performance metrics
	TimeTaken  time.Duration `json:"time_taken"`
	ShardsInfo *ShardsInfo   `json:"shards_info"`

	// Aggregations
	Aggregations map[string]*AggregationResult `json:"aggregations,omitempty"`

	// Search metadata
	SearchID  string    `json:"search_id"`
	Timestamp time.Time `json:"timestamp"`

	// Suggestions
	Suggestions []*Suggestion `json:"suggestions,omitempty"`

	// Debug information
	DebugInfo *SearchDebugInfo `json:"debug_info,omitempty"`
}

// ComplexSearchRequest represents a complex search request with multiple conditions
type ComplexSearchRequest struct {
	// Query composition
	BoolQuery     *BoolQuery     `json:"bool_query,omitempty"`
	MatchQuery    *MatchQuery    `json:"match_query,omitempty"`
	RangeQuery    *RangeQuery    `json:"range_query,omitempty"`
	TermsQuery    *TermsQuery    `json:"terms_query,omitempty"`
	WildcardQuery *WildcardQuery `json:"wildcard_query,omitempty"`
	FuzzyQuery    *FuzzyQuery    `json:"fuzzy_query,omitempty"`

	// Search context
	Indexes []string     `json:"indexes"`
	From    int          `json:"from,omitempty"`
	Size    int          `json:"size,omitempty"`
	Sort    []*SortField `json:"sort,omitempty"`

	// Advanced options
	PostFilter *PostFilter   `json:"post_filter,omitempty"`
	Rescore    *RescoreQuery `json:"rescore,omitempty"`
	MinScore   *float64      `json:"min_score,omitempty"`

	// Performance
	Timeout    time.Duration `json:"timeout,omitempty"`
	SearchType string        `json:"search_type,omitempty"`

	// Metadata
	RequestID  string `json:"request_id,omitempty"`
	TrackingID string `json:"tracking_id,omitempty"`
}

// ComplexSearchResult represents a complex search result
type ComplexSearchResult struct {
	// Results
	Hits     []*SearchHit `json:"hits"`
	Total    int64        `json:"total"`
	MaxScore float64      `json:"max_score"`

	// Performance
	TimeTaken  time.Duration `json:"time_taken"`
	ShardsInfo *ShardsInfo   `json:"shards_info"`

	// Query analysis
	QueryProfile *QueryProfile `json:"query_profile,omitempty"`

	// Aggregations
	Aggregations map[string]*AggregationResult `json:"aggregations,omitempty"`

	// Metadata
	SearchID  string    `json:"search_id"`
	Timestamp time.Time `json:"timestamp"`
}

// AdvancedSearchRequest represents an advanced search request
type AdvancedSearchRequest struct {
	// Multi-query support
	Queries []*QueryClause `json:"queries"`

	// Advanced filters
	GeoBoundingBox *GeoBoundingBox `json:"geo_bounding_box,omitempty"`
	GeoDistance    *GeoDistance    `json:"geo_distance,omitempty"`
	DateRange      *DateRange      `json:"date_range,omitempty"`

	// Faceted search
	Facets []*FacetRequest `json:"facets,omitempty"`

	// Result processing
	Grouping   *GroupingOptions   `json:"grouping,omitempty"`
	Collapsing *CollapsingOptions `json:"collapsing,omitempty"`

	// Performance tuning
	CacheKey string        `json:"cache_key,omitempty"`
	CacheTTL time.Duration `json:"cache_ttl,omitempty"`

	// Context
	Indexes []string      `json:"indexes"`
	From    int           `json:"from,omitempty"`
	Size    int           `json:"size,omitempty"`
	Timeout time.Duration `json:"timeout,omitempty"`

	// Metadata
	RequestID string            `json:"request_id,omitempty"`
	Analytics *AnalyticsOptions `json:"analytics,omitempty"`
}

// AdvancedSearchResult represents an advanced search result
type AdvancedSearchResult struct {
	// Results
	Hits     []*SearchHit `json:"hits"`
	Total    int64        `json:"total"`
	MaxScore float64      `json:"max_score"`

	// Facets
	Facets []*FacetResult `json:"facets,omitempty"`

	// Grouping
	Groups []*GroupResult `json:"groups,omitempty"`

	// Performance
	TimeTaken time.Duration `json:"time_taken"`
	CacheHit  bool          `json:"cache_hit"`

	// Analytics
	Analytics *SearchAnalytics `json:"analytics,omitempty"`

	// Metadata
	SearchID  string    `json:"search_id"`
	Timestamp time.Time `json:"timestamp"`
}

// =============================================================================
// Batch Search Request/Response Types
// =============================================================================

// BatchSearchRequest represents a batch search request
type BatchSearchRequest struct {
	// Batch configuration
	Searches       []*BatchSearchItem `json:"searches"`
	Concurrent     bool               `json:"concurrent,omitempty"`
	MaxConcurrency int                `json:"max_concurrency,omitempty"`

	// Global options
	Timeout         time.Duration `json:"timeout,omitempty"`
	ContinueOnError bool          `json:"continue_on_error,omitempty"`

	// Performance
	Priority int `json:"priority,omitempty"`

	// Metadata
	BatchID   string `json:"batch_id,omitempty"`
	RequestID string `json:"request_id,omitempty"`
}

// BatchSearchItem represents a single search in a batch
type BatchSearchItem struct {
	// Search parameters
	Query   string   `json:"query"`
	Indexes []string `json:"indexes"`
	From    int      `json:"from,omitempty"`
	Size    int      `json:"size,omitempty"`

	// Item identification
	ItemID string `json:"item_id,omitempty"`
	Name   string `json:"name,omitempty"`

	// Options
	Priority int           `json:"priority,omitempty"`
	Timeout  time.Duration `json:"timeout,omitempty"`
}

// BatchSearchResult represents a batch search result
type BatchSearchResult struct {
	// Results
	Results []*BatchSearchItemResult `json:"results"`

	// Summary
	Total      int `json:"total"`
	Successful int `json:"successful"`
	Failed     int `json:"failed"`

	// Performance
	TimeTaken time.Duration `json:"time_taken"`

	// Metadata
	BatchID   string    `json:"batch_id"`
	Timestamp time.Time `json:"timestamp"`

	// Errors
	Errors []*BatchSearchError `json:"errors,omitempty"`
}

// BatchSearchItemResult represents a single search result in a batch
type BatchSearchItemResult struct {
	// Item identification
	ItemID string `json:"item_id"`
	Name   string `json:"name,omitempty"`

	// Results
	Hits     []*SearchHit `json:"hits"`
	Total    int64        `json:"total"`
	MaxScore float64      `json:"max_score"`

	// Performance
	TimeTaken time.Duration `json:"time_taken"`

	// Status
	Success bool   `json:"success"`
	Error   string `json:"error,omitempty"`
}

// BatchSearchError represents an error in batch search
type BatchSearchError struct {
	ItemID string `json:"item_id"`
	Name   string `json:"name,omitempty"`
	Error  string `json:"error"`
	Code   string `json:"code,omitempty"`
}

// =============================================================================
// Statistics and Analytics Types
// =============================================================================

// SearchStatisticsRequest represents a search statistics request
type SearchStatisticsRequest struct {
	// Time range
	StartTime time.Time `json:"start_time,omitempty"`
	EndTime   time.Time `json:"end_time,omitempty"`

	// Filters
	Indexes []string `json:"indexes,omitempty"`
	UserIDs []string `json:"user_ids,omitempty"`

	// Grouping
	GroupBy []string `json:"group_by,omitempty"`

	// Options
	IncludeDetails bool `json:"include_details,omitempty"`

	// Pagination
	From int `json:"from,omitempty"`
	Size int `json:"size,omitempty"`
}

// SearchStatistics represents search statistics
type SearchStatistics struct {
	// Basic metrics
	TotalSearches      int64 `json:"total_searches"`
	SuccessfulSearches int64 `json:"successful_searches"`
	FailedSearches     int64 `json:"failed_searches"`

	// Performance metrics
	AverageResponseTime time.Duration `json:"average_response_time"`
	MinResponseTime     time.Duration `json:"min_response_time"`
	MaxResponseTime     time.Duration `json:"max_response_time"`

	// Result metrics
	AverageResultCount   float64 `json:"average_result_count"`
	TotalResultsReturned int64   `json:"total_results_returned"`

	// Query patterns
	TopQueries []*QueryStats `json:"top_queries,omitempty"`
	TopIndexes []*IndexStats `json:"top_indexes,omitempty"`

	// Time-based metrics
	SearchesByHour map[int]int64    `json:"searches_by_hour,omitempty"`
	SearchesByDay  map[string]int64 `json:"searches_by_day,omitempty"`

	// Error analysis
	ErrorsByType map[string]int64 `json:"errors_by_type,omitempty"`

	// Metadata
	GeneratedAt time.Time  `json:"generated_at"`
	TimeRange   *TimeRange `json:"time_range"`
}

// QueryStats represents statistics for a specific query
type QueryStats struct {
	Query          string        `json:"query"`
	Count          int64         `json:"count"`
	AverageTime    time.Duration `json:"average_time"`
	SuccessRate    float64       `json:"success_rate"`
	AverageResults float64       `json:"average_results"`
	LastExecuted   time.Time     `json:"last_executed"`
}

// IndexStats represents statistics for a specific index
type IndexStats struct {
	IndexName     string        `json:"index_name"`
	SearchCount   int64         `json:"search_count"`
	AverageTime   time.Duration `json:"average_time"`
	DocumentCount int64         `json:"document_count"`
	Size          int64         `json:"size"`
	LastAccessed  time.Time     `json:"last_accessed"`
}

// =============================================================================
// Caching Types
// =============================================================================

// CachedSearchResult represents a cached search result
type CachedSearchResult struct {
	// Original request
	RequestHash string   `json:"request_hash"`
	Query       string   `json:"query"`
	Indexes     []string `json:"indexes"`

	// Cached results
	Hits         []*SearchHit                  `json:"hits"`
	Total        int64                         `json:"total"`
	MaxScore     float64                       `json:"max_score"`
	Aggregations map[string]*AggregationResult `json:"aggregations,omitempty"`

	// Cache metadata
	CachedAt     time.Time `json:"cached_at"`
	ExpiresAt    time.Time `json:"expires_at"`
	HitCount     int64     `json:"hit_count"`
	LastAccessed time.Time `json:"last_accessed"`

	// Performance
	OriginalTimeTaken time.Duration `json:"original_time_taken"`
	CompressionRatio  float64       `json:"compression_ratio,omitempty"`
}

// CacheStatistics represents cache statistics
type CacheStatistics struct {
	// Basic metrics
	TotalKeys int64 `json:"total_keys"`
	TotalSize int64 `json:"total_size"`

	// Hit/miss metrics
	HitCount  int64   `json:"hit_count"`
	MissCount int64   `json:"miss_count"`
	HitRate   float64 `json:"hit_rate"`

	// Performance metrics
	AverageHitTime  time.Duration `json:"average_hit_time"`
	AverageMissTime time.Duration `json:"average_miss_time"`

	// Eviction metrics
	EvictionCount int64 `json:"eviction_count"`
	ExpiredCount  int64 `json:"expired_count"`

	// Memory usage
	MemoryUsage    int64 `json:"memory_usage"`
	MaxMemoryUsage int64 `json:"max_memory_usage"`

	// Metadata
	GeneratedAt time.Time     `json:"generated_at"`
	Uptime      time.Duration `json:"uptime"`
}

// =============================================================================
// Supporting Types
// =============================================================================

// SearchHit represents a single search result
type SearchHit struct {
	ID        string                 `json:"id"`
	Index     string                 `json:"index"`
	Score     float64                `json:"score"`
	Source    map[string]interface{} `json:"source"`
	Highlight map[string][]string    `json:"highlight,omitempty"`
	Sort      []interface{}          `json:"sort,omitempty"`
	Fields    map[string]interface{} `json:"fields,omitempty"`
	Version   *int64                 `json:"version,omitempty"`

	// Additional metadata
	Explanation    *ScoreExplanation     `json:"explanation,omitempty"`
	MatchedQueries []string              `json:"matched_queries,omitempty"`
	InnerHits      map[string]*InnerHits `json:"inner_hits,omitempty"`
}

// ShardsInfo represents information about shards
type ShardsInfo struct {
	Total      int             `json:"total"`
	Successful int             `json:"successful"`
	Skipped    int             `json:"skipped"`
	Failed     int             `json:"failed"`
	Failures   []*ShardFailure `json:"failures,omitempty"`
}

// ShardFailure represents a shard failure
type ShardFailure struct {
	Index  string `json:"index"`
	Shard  int    `json:"shard"`
	Node   string `json:"node"`
	Reason string `json:"reason"`
	Status int    `json:"status"`
}

// SortField represents a sorting field
type SortField struct {
	Field        string      `json:"field"`
	Order        string      `json:"order"` // "asc" or "desc"
	Mode         string      `json:"mode,omitempty"`
	Missing      interface{} `json:"missing,omitempty"`
	UnmappedType string      `json:"unmapped_type,omitempty"`
}

// SearchFilter represents a search filter
type SearchFilter struct {
	Field    string        `json:"field"`
	Operator string        `json:"operator"` // "eq", "ne", "gt", "gte", "lt", "lte", "in", "nin"
	Value    interface{}   `json:"value"`
	Values   []interface{} `json:"values,omitempty"`
}

// HighlightOptions represents highlighting options
type HighlightOptions struct {
	Fields            []string `json:"fields,omitempty"`
	PreTags           []string `json:"pre_tags,omitempty"`
	PostTags          []string `json:"post_tags,omitempty"`
	FragmentSize      int      `json:"fragment_size,omitempty"`
	NumberOfFragments int      `json:"number_of_fragments,omitempty"`
	RequireFieldMatch bool     `json:"require_field_match,omitempty"`
	Type              string   `json:"type,omitempty"`
	BoundaryChars     string   `json:"boundary_chars,omitempty"`
	BoundaryMaxScan   int      `json:"boundary_max_scan,omitempty"`
}

// TimeRange represents a time range
type TimeRange struct {
	StartTime time.Time     `json:"start_time"`
	EndTime   time.Time     `json:"end_time"`
	Duration  time.Duration `json:"duration"`
}

// ScoreExplanation represents score explanation
type ScoreExplanation struct {
	Value       float64             `json:"value"`
	Description string              `json:"description"`
	Details     []*ScoreExplanation `json:"details,omitempty"`
}

// InnerHits represents inner hits
type InnerHits struct {
	Hits     []*SearchHit `json:"hits"`
	Total    int64        `json:"total"`
	MaxScore float64      `json:"max_score"`
}

// SearchDebugInfo represents debug information
type SearchDebugInfo struct {
	QueryParsed        bool          `json:"query_parsed"`
	QueryOptimized     bool          `json:"query_optimized"`
	CacheUsed          bool          `json:"cache_used"`
	IndexesQueried     []string      `json:"indexes_queried"`
	ShardsQueried      int           `json:"shards_queried"`
	RewriteCount       int           `json:"rewrite_count"`
	OptimizationTime   time.Duration `json:"optimization_time"`
	ExecutionTime      time.Duration `json:"execution_time"`
	PostProcessingTime time.Duration `json:"post_processing_time"`
}

// QueryProfile represents query profiling information
type QueryProfile struct {
	Type        string                   `json:"type"`
	Description string                   `json:"description"`
	Time        time.Duration            `json:"time"`
	Breakdown   map[string]time.Duration `json:"breakdown,omitempty"`
	Children    []*QueryProfile          `json:"children,omitempty"`
}

// BoolQuery represents a boolean query
type BoolQuery struct {
	Must               []*QueryClause `json:"must,omitempty"`
	Filter             []*QueryClause `json:"filter,omitempty"`
	Should             []*QueryClause `json:"should,omitempty"`
	MustNot            []*QueryClause `json:"must_not,omitempty"`
	MinimumShouldMatch int            `json:"minimum_should_match,omitempty"`
	Boost              *float64       `json:"boost,omitempty"`
}

// QueryClause represents a query clause
type QueryClause struct {
	Type     string                 `json:"type"`
	Field    string                 `json:"field,omitempty"`
	Value    interface{}            `json:"value,omitempty"`
	Operator string                 `json:"operator,omitempty"`
	Boost    *float64               `json:"boost,omitempty"`
	Options  map[string]interface{} `json:"options,omitempty"`
}

// MatchQuery represents a match query
type MatchQuery struct {
	Field              string   `json:"field"`
	Query              string   `json:"query"`
	Operator           string   `json:"operator,omitempty"`
	MinimumShouldMatch string   `json:"minimum_should_match,omitempty"`
	Analyzer           string   `json:"analyzer,omitempty"`
	Boost              *float64 `json:"boost,omitempty"`
}

// RangeQuery represents a range query
type RangeQuery struct {
	Field    string      `json:"field"`
	Gte      interface{} `json:"gte,omitempty"`
	Gt       interface{} `json:"gt,omitempty"`
	Lte      interface{} `json:"lte,omitempty"`
	Lt       interface{} `json:"lt,omitempty"`
	Format   string      `json:"format,omitempty"`
	TimeZone string      `json:"time_zone,omitempty"`
	Boost    *float64    `json:"boost,omitempty"`
}

// TermsQuery represents a terms query
type TermsQuery struct {
	Field  string        `json:"field"`
	Values []interface{} `json:"values"`
	Boost  *float64      `json:"boost,omitempty"`
}

// WildcardQuery represents a wildcard query
type WildcardQuery struct {
	Field           string   `json:"field"`
	Value           string   `json:"value"`
	Boost           *float64 `json:"boost,omitempty"`
	CaseInsensitive bool     `json:"case_insensitive,omitempty"`
}

// FuzzyQuery represents a fuzzy query
type FuzzyQuery struct {
	Field          string      `json:"field"`
	Value          string      `json:"value"`
	Fuzziness      interface{} `json:"fuzziness,omitempty"`
	PrefixLength   int         `json:"prefix_length,omitempty"`
	MaxExpansions  int         `json:"max_expansions,omitempty"`
	Transpositions bool        `json:"transpositions,omitempty"`
	Boost          *float64    `json:"boost,omitempty"`
}

// PostFilter represents a post filter
type PostFilter struct {
	Type    string                 `json:"type"`
	Field   string                 `json:"field"`
	Value   interface{}            `json:"value"`
	Options map[string]interface{} `json:"options,omitempty"`
}

// RescoreQuery represents a rescore query
type RescoreQuery struct {
	Query              *QueryClause `json:"query"`
	QueryWeight        *float64     `json:"query_weight,omitempty"`
	RescoreQueryWeight *float64     `json:"rescore_query_weight,omitempty"`
	ScoreMode          string       `json:"score_mode,omitempty"`
	WindowSize         int          `json:"window_size,omitempty"`
}

// Default options and configurations
func DefaultHighlightOptions() *HighlightOptions {
	return &HighlightOptions{
		PreTags:           []string{"<em>"},
		PostTags:          []string{"</em>"},
		FragmentSize:      150,
		NumberOfFragments: 5,
		RequireFieldMatch: false,
		Type:              "unified",
		BoundaryChars:     ".,!? \t\n",
		BoundaryMaxScan:   20,
	}
}

//Personal.AI order the ending
