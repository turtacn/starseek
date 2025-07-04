package search

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
	"unicode"

	"github.com/google/uuid"
)

// SearchRepository 搜索数据访问接口
type SearchRepository interface {
	// 搜索操作
	SaveSearch(ctx context.Context, search *Search) error
	GetSearchByID(ctx context.Context, id string) (*Search, error)
	GetSearchesByUserID(ctx context.Context, userID string, limit int) ([]*Search, error)

	// 查询操作
	SaveQuery(ctx context.Context, query *Query) error
	GetQueryByHash(ctx context.Context, hash string) (*Query, error)
	GetPopularQueries(ctx context.Context, limit int) ([]*Query, error)

	// 文档操作
	SaveDocument(ctx context.Context, document *Document) error
	GetDocumentByID(ctx context.Context, id string) (*Document, error)
	SearchDocuments(ctx context.Context, query *Query, options *SearchOptions) ([]*Document, error)

	// 结果操作
	SaveSearchResults(ctx context.Context, results []*SearchResult) error
	GetResultsBySearchID(ctx context.Context, searchID string) ([]*SearchResult, error)

	// 统计操作
	GetSearchStatistics(ctx context.Context, timeRange *TimeRange) (*SearchStatistics, error)
	GetQueryStatistics(ctx context.Context, queryID string) (*QueryStatistics, error)
}

// MetricsCollector 指标收集器接口
type MetricsCollector interface {
	RecordSearchDuration(duration time.Duration)
	RecordQueryCount(queryType string)
	RecordCacheHit(hit bool)
	RecordErrorCount(errorType string)
	RecordResultCount(count int)
}

// CacheService 缓存服务接口
type CacheService interface {
	Get(ctx context.Context, key string) (interface{}, error)
	Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error
	Delete(ctx context.Context, key string) error
	Exists(ctx context.Context, key string) bool
}

// SearchDomainService 搜索领域服务
type SearchDomainService struct {
	repository       SearchRepository
	metricsCollector MetricsCollector
	cacheService     CacheService

	// 配置参数
	config *SearchConfig

	// 缓存
	queryCache       sync.Map // map[string]*Query
	documentCache    sync.Map // map[string]*Document

	// 性能监控
	performanceStats *PerformanceStats

	// 同步锁
	mu sync.RWMutex
}

// SearchConfig 搜索配置
type SearchConfig struct {
	// 查询配置
	MaxQueryLength    int           `json:"max_query_length"`
	MaxResults        int           `json:"max_results"`
	DefaultPageSize   int           `json:"default_page_size"`
	QueryTimeout      time.Duration `json:"query_timeout"`

	// 缓存配置
	CacheEnabled      bool          `json:"cache_enabled"`
	CacheTTL          time.Duration `json:"cache_ttl"`
	CacheMaxSize      int           `json:"cache_max_size"`

	// 评分配置
	TFIDFEnabled      bool          `json:"tfidf_enabled"`
	BoostFactors      map[string]float64 `json:"boost_factors"`
	MinScore          float64       `json:"min_score"`

	// 高亮配置
	HighlightEnabled  bool          `json:"highlight_enabled"`
	HighlightTags     HighlightTags `json:"highlight_tags"`
	MaxFragments      int           `json:"max_fragments"`
	FragmentSize      int           `json:"fragment_size"`

	// 聚合配置
	AggregationEnabled bool         `json:"aggregation_enabled"`
	MaxAggregations   int           `json:"max_aggregations"`

	// 性能配置
	MonitoringEnabled bool          `json:"monitoring_enabled"`
	SlowQueryThreshold time.Duration `json:"slow_query_threshold"`
}

// HighlightTags 高亮标签配置
type HighlightTags struct {
	PreTag  string `json:"pre_tag"`
	PostTag string `json:"post_tag"`
}

// PerformanceStats 性能统计
type PerformanceStats struct {
	TotalQueries     int64         `json:"total_queries"`
	SuccessfulQueries int64        `json:"successful_queries"`
	FailedQueries    int64         `json:"failed_queries"`
	AverageLatency   time.Duration `json:"average_latency"`
	CacheHitRate     float64       `json:"cache_hit_rate"`
	SlowQueries      int64         `json:"slow_queries"`
	LastUpdated      time.Time     `json:"last_updated"`

	mu sync.RWMutex
}

// SearchStatistics 搜索统计
type SearchStatistics struct {
	QueryCount       int64     `json:"query_count"`
	DocumentCount    int64     `json:"document_count"`
	AverageLatency   int64     `json:"average_latency"` // milliseconds
	CacheHitRate     float64   `json:"cache_hit_rate"`
	PopularQueries   []string  `json:"popular_queries"`
	ErrorRate        float64   `json:"error_rate"`
	Period           TimeRange `json:"period"`
}

// QueryStatistics 查询统计
type QueryStatistics struct {
	QueryID         string        `json:"query_id"`
	ExecutionCount  int64         `json:"execution_count"`
	AverageLatency  time.Duration `json:"average_latency"`
	SuccessRate     float64       `json:"success_rate"`
	LastExecuted    time.Time     `json:"last_executed"`
	ResultStats     *ResultStats  `json:"result_stats"`
}

// ResultStats 结果统计
type ResultStats struct {
	AverageResultCount float64 `json:"average_result_count"`
	MaxResultCount     int     `json:"max_result_count"`
	MinResultCount     int     `json:"min_result_count"`
	AverageScore       float64 `json:"average_score"`
	MaxScore           float64 `json:"max_score"`
}

// TimeRange 时间范围
type TimeRange struct {
	Start time.Time `json:"start"`
	End   time.Time `json:"end"`
}

// NewSearchDomainService 创建搜索领域服务
func NewSearchDomainService(
	repository SearchRepository,
	metricsCollector MetricsCollector,
	cacheService CacheService,
	config *SearchConfig,
) *SearchDomainService {
	if config == nil {
		config = getDefaultSearchConfig()
	}

	return &SearchDomainService{
		repository:       repository,
		metricsCollector: metricsCollector,
		cacheService:     cacheService,
		config:           config,
		performanceStats: &PerformanceStats{
			LastUpdated: time.Now(),
		},
	}
}

// ExecuteSearch 执行搜索
func (s *SearchDomainService) ExecuteSearch(ctx context.Context, req *SearchRequest) (*SearchResponse, error) {
	startTime := time.Now()

	// 创建搜索实体
	search := NewSearch(req.UserID, req.TenantID, req.Query, req.Database, req.Table)
	search.TraceID = req.TraceID
	search.SessionID = req.SessionID
	search.Priority = req.Priority

	// 设置搜索参数
	if req.Filters != nil {
		search.Filters = req.Filters
	}
	if req.SortBy != nil {
		search.SortBy = req.SortBy
	}
	if req.Pagination != nil {
		search.Pagination = req.Pagination
	}
	if req.Options != nil {
		search.Options = req.Options
	}

	// 验证搜索请求
	if err := s.validateSearchRequest(req); err != nil {
		return nil, fmt.Errorf("invalid search request: %w", err)
	}

	// 开始搜索
	if err := search.Start(); err != nil {
		return nil, fmt.Errorf("failed to start search: %w", err)
	}

	// 保存搜索记录
	if err := s.repository.SaveSearch(ctx, search); err != nil {
		s.recordError("save_search_failed")
		return nil, fmt.Errorf("failed to save search: %w", err)
	}

	defer func() {
		// 记录性能指标
		duration := time.Since(startTime)
		s.recordSearchDuration(duration)

		// 更新搜索记录
		search.Duration = duration
		s.repository.SaveSearch(ctx, search)
	}()

	// 解析和优化查询
	query, err := s.parseAndOptimizeQuery(ctx, req.Query, req.QueryType, req.UserID, req.TenantID)
	if err != nil {
		search.Fail(err.Error())
		s.recordError("query_parse_failed")
		return nil, fmt.Errorf("failed to parse query: %w", err)
	}

	search.SetQuery(query)
	search.ProcessedQuery = query.OptimizedQuery
	search.QueryType = query.QueryType

	// 检查缓存
	var results []*SearchResult
	var documents []*Document
	cacheKey := s.generateCacheKey(req)

	if s.config.CacheEnabled {
		if cachedResults, err := s.getCachedResults(ctx, cacheKey); err == nil {
			results = cachedResults.Results
			documents = cachedResults.Documents
			search.CacheHit = true
			search.CacheKey = cacheKey
		}
	}

	// 如果缓存未命中，执行搜索
	if results == nil {
		documents, err = s.searchDocuments(ctx, query, req.Options)
		if err != nil {
			search.Fail(err.Error())
			s.recordError("document_search_failed")
			return nil, fmt.Errorf("failed to search documents: %w", err)
		}

		// 计算相关度和排序
		results, err = s.calculateRelevanceAndSort(ctx, documents, query, req.Options)
		if err != nil {
			search.Fail(err.Error())
			s.recordError("relevance_calculation_failed")
			return nil, fmt.Errorf("failed to calculate relevance: %w", err)
		}

		// 缓存结果
		if s.config.CacheEnabled && len(results) > 0 {
			s.cacheResults(ctx, cacheKey, &CachedResults{
				Results:   results,
				Documents: documents,
			})
		}
	}

	// 应用过滤和分页
	filteredResults, totalHits := s.applyFiltersAndPagination(results, req.Filters, req.Pagination)

	// 处理高亮
	if s.config.HighlightEnabled && req.Options != nil && req.Options.IncludeHighlight {
		s.processHighlights(filteredResults, query, req.Options)
	}

	// 聚合结果
	var aggregations map[string]*AggregationResult
	if s.config.AggregationEnabled && req.Aggregations != nil {
		aggregations = s.aggregateResults(documents, req.Aggregations)
	}

	// 计算最大分数
	var maxScore float64
	if len(filteredResults) > 0 {
		maxScore = filteredResults[0].Score
	}

	// 完成搜索
	search.Complete(totalHits, maxScore, len(filteredResults))

	// 更新查询统计
	query.UpdateStats(search.Duration, search.IsCompleted())
	s.repository.SaveQuery(ctx, query)

	// 保存搜索结果
	for _, result := range filteredResults {
		search.AddResult(result)
	}
	s.repository.SaveSearchResults(ctx, filteredResults)

	// 构建响应
	response := &SearchResponse{
		SearchID:     search.ID,
		Results:      filteredResults,
		TotalHits:    totalHits,
		MaxScore:     maxScore,
		Aggregations: aggregations,
		Performance:  s.calculatePerformanceMetrics(search),
		CacheHit:     search.CacheHit,
		Query:        query,
	}

	// 记录指标
	s.recordSearchMetrics(search, response)

	return response, nil
}

// parseAndOptimizeQuery 解析和优化查询
func (s *SearchDomainService) parseAndOptimizeQuery(ctx context.Context, rawQuery string, queryType QueryType, userID, tenantID string) (*Query, error) {
	// 检查查询缓存
	queryHash := generateQueryHash(rawQuery)
	if cachedQuery, exists := s.queryCache.Load(queryHash); exists {
		query := cachedQuery.(*Query)
		return query, nil
	}

	// 从数据库查找现有查询
	if existingQuery, err := s.repository.GetQueryByHash(ctx, queryHash); err == nil {
		s.queryCache.Store(queryHash, existingQuery)
		return existingQuery, nil
	}

	// 创建新查询
	query := NewQuery(userID, tenantID, rawQuery, queryType)

	// 验证查询
	if err := query.Validate(); err != nil {
		return nil, fmt.Errorf("query validation failed: %w", err)
	}

	// 解析查询
	if err := query.Parse(); err != nil {
		return nil, fmt.Errorf("query parsing failed: %w", err)
	}

	// 优化查询
	if err := s.optimizeQuery(query); err != nil {
		return nil, fmt.Errorf("query optimization failed: %w", err)
	}

	// 生成执行计划
	if err := query.GenerateExecutionPlan(); err != nil {
		return nil, fmt.Errorf("execution plan generation failed: %w", err)
	}

	// 保存查询
	if err := s.repository.SaveQuery(ctx, query); err != nil {
		return nil, fmt.Errorf("failed to save query: %w", err)
	}

	// 缓存查询
	s.queryCache.Store(queryHash, query)

	return query, nil
}

// optimizeQuery 优化查询
func (s *SearchDomainService) optimizeQuery(query *Query) error {
	optimizations := []string{}

	// 查询重写
	rewrittenQuery := s.rewriteQuery(query.RawQuery)
	if rewrittenQuery != query.RawQuery {
		query.OptimizedQuery = rewrittenQuery
		optimizations = append(optimizations, "query_rewrite")
	} else {
		query.OptimizedQuery = query.RawQuery
	}

	// 词干提取和同义词扩展
	if s.shouldExpandSynonyms(query) {
		expandedQuery := s.expandSynonyms(query.OptimizedQuery)
		query.OptimizedQuery = expandedQuery
		optimizations = append(optimizations, "synonym_expansion")
	}

	// 字段权重优化
	if s.shouldOptimizeFieldWeights(query) {
		s.optimizeFieldWeights(query)
		optimizations = append(optimizations, "field_weight_optimization")
	}

	// 设置优化信息
	if query.ExecutionPlan == nil {
		query.ExecutionPlan = &ExecutionPlan{}
	}
	query.ExecutionPlan.Optimizations = optimizations

	return nil
}

// rewriteQuery 查询重写
func (s *SearchDomainService) rewriteQuery(rawQuery string) string {
	// 标准化空格
	rewritten := regexp.MustCompile(`\s+`).ReplaceAllString(strings.TrimSpace(rawQuery), " ")

	// 移除多余的引号
	rewritten = strings.ReplaceAll(rewritten, `""`, `"`)

	// 处理通配符
	rewritten = s.optimizeWildcards(rewritten)

	// 处理布尔操作符
	rewritten = s.optimizeBooleanOperators(rewritten)

	return rewritten
}

// optimizeWildcards 优化通配符
func (s *SearchDomainService) optimizeWildcards(query string) string {
	// 移除不必要的通配符
	optimized := regexp.MustCompile(`\*+`).ReplaceAllString(query, "*")
	optimized = regexp.MustCompile(`\?+`).ReplaceAllString(optimized, "?")

	// 优化通配符位置
	optimized = regexp.MustCompile(`^\*`).ReplaceAllString(optimized, "")

	return optimized
}

// optimizeBooleanOperators 优化布尔操作符
func (s *SearchDomainService) optimizeBooleanOperators(query string) string {
	// 标准化布尔操作符
	optimized := regexp.MustCompile(`(?i)\s+and\s+`).ReplaceAllString(query, " AND ")
	optimized = regexp.MustCompile(`(?i)\s+or\s+`).ReplaceAllString(optimized, " OR ")
	optimized = regexp.MustCompile(`(?i)\s+not\s+`).ReplaceAllString(optimized, " NOT ")

	// 移除多余的操作符
	optimized = regexp.MustCompile(`(AND|OR|NOT)\s+(AND|OR|NOT)`).ReplaceAllString(optimized, "$2")

	return optimized
}

// expandSynonyms 扩展同义词
func (s *SearchDomainService) expandSynonyms(query string) string {
	// 简化的同义词扩展实现
	synonyms := map[string][]string{
		"search": {"find", "lookup", "query"},
		"fast":   {"quick", "rapid", "speedy"},
		"good":   {"excellent", "great", "nice"},
	}

	words := strings.Fields(query)
	expanded := make([]string, 0, len(words))

	for _, word := range words {
		lowerWord := strings.ToLower(word)
		if syns, exists := synonyms[lowerWord]; exists {
			// 创建同义词组
			synGroup := "(" + word
			for _, syn := range syns {
				synGroup += " OR " + syn
			}
			synGroup += ")"
			expanded = append(expanded, synGroup)
		} else {
			expanded = append(expanded, word)
		}
	}

	return strings.Join(expanded, " ")
}

// shouldExpandSynonyms 判断是否应该扩展同义词
func (s *SearchDomainService) shouldExpandSynonyms(query *Query) bool {
	// 简单的启发式规则
	return len(strings.Fields(query.RawQuery)) <= 5 && query.QueryType == QueryTypeMatch
}

// shouldOptimizeFieldWeights 判断是否应该优化字段权重
func (s *SearchDomainService) shouldOptimizeFieldWeights(query *Query) bool {
	return query.QueryType == QueryTypeMatch || query.QueryType == QueryTypeMatchPhrase
}

// optimizeFieldWeights 优化字段权重
func (s *SearchDomainService) optimizeFieldWeights(query *Query) {
	if query.ParsedQuery == nil {
		return
	}

	// 根据配置设置字段权重
	if query.ParsedQuery.Boosts == nil {
		query.ParsedQuery.Boosts = make(map[string]float64)
	}

	// 应用默认权重
	for field, boost := range s.config.BoostFactors {
		query.ParsedQuery.Boosts[field] = boost
	}
}

// searchDocuments 搜索文档
func (s *SearchDomainService) searchDocuments(ctx context.Context, query *Query, options *SearchOptions) ([]*Document, error) {
	// 设置搜索选项
	if options == nil {
		options = s.getDefaultSearchOptions()
	}

	// 应用超时
	searchCtx, cancel := context.WithTimeout(ctx, s.config.QueryTimeout)
	defer cancel()

	// 执行搜索
	documents, err := s.repository.SearchDocuments(searchCtx, query, options)
	if err != nil {
		return nil, fmt.Errorf("document search failed: %w", err)
	}

	return documents, nil
}

// calculateRelevanceAndSort 计算相关度并排序
func (s *SearchDomainService) calculateRelevanceAndSort(ctx context.Context, documents []*Document, query *Query, options *SearchOptions) ([]*SearchResult, error) {
	results := make([]*SearchResult, 0, len(documents))

	// 计算文档频率统计
	docStats := s.calculateDocumentStatistics(documents, query)

	for i, doc := range documents {
		// 创建搜索结果
		result := NewSearchResult("", doc.ID, doc.IndexName, i+1, 0.0)
		result.Source = doc.Source
		result.Document = doc

		// 计算相关度评分
		score := s.calculateRelevanceScore(doc, query, docStats)
		result.Score = score
		result.Relevance = score
		result.Confidence = s.calculateConfidence(score, docStats)

		// 应用最小分数过滤
		if score >= s.config.MinScore {
			results = append(results, result)
		}
	}

	// 排序结果
	s.sortResults(results, options)

	return results, nil
}

// calculateRelevanceScore 计算相关度评分
func (s *SearchDomainService) calculateRelevanceScore(doc *Document, query *Query, docStats *DocumentStatistics) float64 {
	if !s.config.TFIDFEnabled {
		return s.calculateSimpleScore(doc, query)
	}

	return s.calculateTFIDFScore(doc, query, docStats)
}

// calculateSimpleScore 计算简单评分
func (s *SearchDomainService) calculateSimpleScore(doc *Document, query *Query) float64 {
	if query.ParsedQuery == nil || len(query.ParsedQuery.Terms) == 0 {
		return 1.0
	}

	score := 0.0
	totalTerms := float64(len(query.ParsedQuery.Terms))

	// 在文档中查找查询词
	docText := s.extractDocumentText(doc)
	docTextLower := strings.ToLower(docText)

	for _, term := range query.ParsedQuery.Terms {
		termLower := strings.ToLower(term)

		// 计算词频
		count := float64(strings.Count(docTextLower, termLower))
		if count > 0 {
			// 简单的词频评分
			termScore := math.Log(1.0 + count)
			score += termScore / totalTerms
		}
	}

	// 应用字段权重
	score = s.applyFieldBoosts(score, doc, query)

	return math.Min(score, 10.0) // 限制最大分数
}

// calculateTFIDFScore 计算TF-IDF评分
func (s *SearchDomainService) calculateTFIDFScore(doc *Document, query *Query, docStats *DocumentStatistics) float64 {
	if query.ParsedQuery == nil || len(query.ParsedQuery.Terms) == 0 {
		return 1.0
	}

	score := 0.0
	docText := s.extractDocumentText(doc)
	docWords := strings.Fields(strings.ToLower(docText))
	docWordCount := len(docWords)

	if docWordCount == 0 {
		return 0.0
	}

	for _, term := range query.ParsedQuery.Terms {
		termLower := strings.ToLower(term)

		// 计算词频 (TF)
		tf := s.calculateTermFrequency(termLower, docWords)

		// 计算逆文档频率 (IDF)
		idf := s.calculateInverseDocumentFrequency(termLower, docStats)

		// TF-IDF 评分
		tfidf := tf * idf

		// 应用查询词权重
		if query.ParsedQuery.Boosts != nil {
			if boost, exists := query.ParsedQuery.Boosts[termLower]; exists {
				tfidf *= boost
			}
		}

		score += tfidf
	}

	// 标准化评分
	score = score / float64(len(query.ParsedQuery.Terms))

	// 应用字段权重
	score = s.applyFieldBoosts(score, doc, query)

	return score
}

// calculateTermFrequency 计算词频
func (s *SearchDomainService) calculateTermFrequency(term string, docWords []string) float64 {
	count := 0
	for _, word := range docWords {
		if word == term {
			count++
		}
	}

	if count == 0 {
		return 0.0
	}

	// 使用对数标准化
	return 1.0 + math.Log(float64(count))
}

// calculateInverseDocumentFrequency 计算逆文档频率
func (s *SearchDomainService) calculateInverseDocumentFrequency(term string, docStats *DocumentStatistics) float64 {
	totalDocs := float64(docStats.TotalDocuments)
	docsWithTerm := float64(docStats.TermDocumentFrequency[term])

	if docsWithTerm == 0 {
		return 0.0
	}

	return math.Log(totalDocs / docsWithTerm)
}

// calculateDocumentStatistics 计算文档统计信息
func (s *SearchDomainService) calculateDocumentStatistics(documents []*Document, query *Query) *DocumentStatistics {
	stats := &DocumentStatistics{
		TotalDocuments:        len(documents),
		TermDocumentFrequency: make(map[string]int),
		AverageDocumentLength: 0.0,
	}

	totalLength := 0

	if query.ParsedQuery != nil {
		for _, doc := range documents {
			docText := s.extractDocumentText(doc)
			docWords := strings.Fields(strings.ToLower(docText))
			totalLength += len(docWords)

			// 计算包含查询词的文档数
			termsSeen := make(map[string]bool)
			for _, word := range docWords {
				for _, term := range query.ParsedQuery.Terms {
					termLower := strings.ToLower(term)
					if word == termLower && !termsSeen[termLower] {
						stats.TermDocumentFrequency[termLower]++
						termsSeen[termLower] = true
					}
				}
			}
		}
	}

	if len(documents) > 0 {
		stats.AverageDocumentLength = float64(totalLength) / float64(len(documents))
	}

	return stats
}

// DocumentStatistics 文档统计信息
type DocumentStatistics struct {
	TotalDocuments        int            `json:"total_documents"`
	TermDocumentFrequency map[string]int `json:"term_document_frequency"`
	AverageDocumentLength float64        `json:"average_document_length"`
}

// extractDocumentText 提取文档文本
func (s *SearchDomainService) extractDocumentText(doc *Document) string {
	var textParts []string

	// 从source中提取文本字段
	for key, value := range doc.Source {
		if str, ok := value.(string); ok && s.isTextualField(key) {
			textParts = append(textParts, str)
		}
	}

	return strings.Join(textParts, " ")
}

// isTextualField 判断是否为文本字段
func (s *SearchDomainService) isTextualField(fieldName string) bool {
	textualFields := []string{"title", "content", "description", "body", "text", "name"}
	fieldLower := strings.ToLower(fieldName)

	for _, field := range textualFields {
		if strings.Contains(fieldLower, field) {
			return true
		}
	}

	return true // 默认认为是文本字段
}

// applyFieldBoosts 应用字段权重
func (s *SearchDomainService) applyFieldBoosts(score float64, doc *Document, query *Query) float64 {
	if query.ParsedQuery == nil || query.ParsedQuery.Boosts == nil {
		return score
	}

	boost := 1.0

	// 检查文档字段是否匹配权重配置
	for field, fieldBoost := range query.ParsedQuery.Boosts {
		if _, exists := doc.Source[field]; exists {
			boost *= fieldBoost
		}
	}

	return score * boost
}

// calculateConfidence 计算置信度
func (s *SearchDomainService) calculateConfidence(score float64, docStats *DocumentStatistics) float64 {
	// 简单的置信度计算
	maxPossibleScore := 10.0
	confidence := math.Min(score/maxPossibleScore, 1.0)

	// 基于文档统计调整置信度
	if docStats.TotalDocuments > 0 {
		adjustment := math.Log(float64(docStats.TotalDocuments)) / 10.0
		confidence = math.Min(confidence+adjustment, 1.0)
	}

	return confidence
}

// sortResults 排序结果
func (s *SearchDomainService) sortResults(results []*SearchResult, options *SearchOptions) {
	if options != nil && len(options.SortBy) > 0 {
		s.sortByCustomFields(results, options.SortBy)
	} else {
		// 默认按分数排序
		sort.Slice(results, func(i, j int) bool {
			return results[i].Score > results[j].Score
		})
	}

	// 更新排名
	for i, result := range results {
		result.Rank = i + 1
	}
}

// sortByCustomFields 按自定义字段排序
func (s *SearchDomainService) sortByCustomFields(results []*SearchResult, sortFields []SortField) {
	sort.Slice(results, func(i, j int) bool {
		for _, sortField := range sortFields {
			val1 := s.getFieldValue(results[i], sortField.Field)
			val2 := s.getFieldValue(results[j], sortField.Field)

			cmp := s.compareValues(val1, val2)
			if cmp != 0 {
				if sortField.Order == SortOrderAsc {
					return cmp < 0
				} else {
					return cmp > 0
				}
			}
		}

		// 如果所有排序字段相等，按分数排序
		return results[i].Score > results[j].Score
	})
}

// getFieldValue 获取字段值
func (s *SearchDomainService) getFieldValue(result *SearchResult, fieldName string) interface{} {
	// 特殊字段处理
	switch fieldName {
	case "_score":
		return result.Score
	case "_rank":
		return result.Rank
	}

	// 从source中获取
	if result.Source != nil {
		if value, exists := result.Source[fieldName]; exists {
			return value
		}
	}

	// 从fields中获取
	if result.Fields != nil {
		if value, exists := result.Fields[fieldName]; exists {
			return value
		}
	}

	return nil
}

// compareValues 比较值
func (s *SearchDomainService) compareValues(val1, val2 interface{}) int {
	if val1 == nil && val2 == nil {
		return 0
	}
	if val1 == nil {
		return -1
	}
	if val2 == nil {
		return 1
	}

	// 尝试转换为数值比较
	if num1, ok1 := s.toFloat64(val1); ok1 {
		if num2, ok2 := s.toFloat64(val2); ok2 {
			if num1 < num2 {
				return -1
			} else if num1 > num2 {
				return 1
			}
			return 0
		}
	}

	// 字符串比较
	str1 := fmt.Sprintf("%v", val1)
	str2 := fmt.Sprintf("%v", val2)

	if str1 < str2 {
		return -1
	} else if str1 > str2 {
		return 1
	}
	return 0
}

// toFloat64 转换为float64
func (s *SearchDomainService) toFloat64(val interface{}) (float64, bool) {
	switch v := val.(type) {
	case float64:
		return v, true
	case float32:
		return float64(v), true
	case int:
		return float64(v), true
	case int32:
		return float64(v), true
	case int64:
		return float64(v), true
	case string:
		if f, err := strconv.ParseFloat(v, 64); err == nil {
			return f, true
		}
	}
	return 0, false
}

// applyFiltersAndPagination 应用过滤和分页
func (s *SearchDomainService) applyFiltersAndPagination(results []*SearchResult, filters map[string]interface{}, pagination *PaginationParams) ([]*SearchResult, int64) {
	totalHits := int64(len(results))

	// 应用过滤器
	if filters != nil && len(filters) > 0 {
		filteredResults := make([]*SearchResult, 0, len(results))
		for _, result := range results {
			if s.matchesFilters(result, filters) {
				filteredResults = append(filteredResults, result)
			}
		}
		results = filteredResults
		totalHits = int64(len(results))
	}

	// 应用分页
	if pagination != nil {
		offset := pagination.Offset
		if offset < 0 {
			offset = (pagination.Page - 1) * pagination.Size
		}

		if offset >= len(results) {
			return []*SearchResult{}, totalHits
		}

		end := offset + pagination.Size
		if end > len(results) {
			end = len(results)
		}

		results = results[offset:end]
	}

	return results, totalHits
}

// matchesFilters 检查是否匹配过滤条件
func (s *SearchDomainService) matchesFilters(result *SearchResult, filters map[string]interface{}) bool {
	for field, filterValue := range filters {
		resultValue := s.getFieldValue(result, field)
		if !s.matchesFilter(resultValue, filterValue) {
			return false
		}
	}
	return true
}

// matchesFilter 检查单个过滤条件
func (s *SearchDomainService) matchesFilter(resultValue, filterValue interface{}) bool {
	if resultValue == nil {
		return filterValue == nil
	}

	// 直接相等比较
	if resultValue == filterValue {
		return true
	}

	// 字符串包含检查
	if resultStr, ok := resultValue.(string); ok {
		if filterStr, ok := filterValue.(string); ok {
			return strings.Contains(strings.ToLower(resultStr), strings.ToLower(filterStr))
		}
	}

	// 数值范围检查（如果filterValue是map类型）
	if filterMap, ok := filterValue.(map[string]interface{}); ok {
		return s.matchesRangeFilter(resultValue, filterMap)
	}

	return false
}

// matchesRangeFilter 检查范围过滤条件
func (s *SearchDomainService) matchesRangeFilter(resultValue interface{}, rangeFilter map[string]interface{}) bool {
	resultNum, ok := s.toFloat64(resultValue)
	if !ok {
		return false
	}

	if gte, exists := rangeFilter["gte"]; exists {
		if gteNum, ok := s.toFloat64(gte); ok && resultNum < gteNum {
			return false
		}
	}

	if lte, exists := rangeFilter["lte"]; exists {
		if lteNum, ok := s.toFloat64(lte); ok && resultNum > lteNum {
			return false
		}
	}

	if gt, exists := rangeFilter["gt"]; exists {
		if gtNum, ok := s.toFloat64(gt); ok && resultNum <= gtNum {
			return false
		}
	}

	if lt, exists := rangeFilter["lt"]; exists {
		if ltNum, ok := s.toFloat64(lt); ok && resultNum >= ltNum {
			return false
		}
	}

	return true
}

// processHighlights 处理高亮
func (s *SearchDomainService) processHighlights(results []*SearchResult, query *Query, options *SearchOptions) {
	if query.ParsedQuery == nil || len(query.ParsedQuery.Terms) == 0 {
		return
	}

	for _, result := range results {
		s.addHighlightsToResult(result, query.ParsedQuery.Terms, s.config.HighlightTags)
	}
}

// addHighlightsToResult 为结果添加高亮
func (s *SearchDomainService) addHighlightsToResult(result *SearchResult, terms []string, tags HighlightTags) {
	if result.Source == nil {
		return
	}

	highlights := make(map[string][]string)

	for field, value := range result.Source {
		if str, ok := value.(string); ok && s.isTextualField(field) {
			fragments := s.generateHighlightFragments(str, terms, tags)
			if len(fragments) > 0 {
				highlights[field] = fragments
			}
		}
	}

	if len(highlights) > 0 {
		result.SetHighlight(highlights)
	}
}

// generateHighlightFragments 生成高亮片段
func (s *SearchDomainService) generateHighlightFragments(text string, terms []string, tags HighlightTags) []string {
	if text == "" || len(terms) == 0 {
		return nil
	}

	fragments := make([]string, 0)
	textLower := strings.ToLower(text)

	// 查找所有匹配位置
	matches := make([]HighlightMatch, 0)
	for _, term := range terms {
		termLower := strings.ToLower(term)
		start := 0
		for {
			index := strings.Index(textLower[start:], termLower)
			if index == -1 {
				break
			}

			actualIndex := start + index
			matches = append(matches, HighlightMatch{
				Start:  actualIndex,
				End:    actualIndex + len(term),
				Term:   term,
				Length: len(term),
			})
			start = actualIndex + len(term)
		}
	}

	if len(matches) == 0 {
		return fragments
	}

	// 排序匹配项
	sort.Slice(matches, func(i, j int) bool {
		return matches[i].Start < matches[j].Start
	})

	// 生成片段
	fragmentSize := s.config.FragmentSize
	if fragmentSize <= 0 {
		fragmentSize = 150
	}

	maxFragments := s.config.MaxFragments
	if maxFragments <= 0 {
		maxFragments = 3
	}

	processedRanges := make([]bool, len(text))

	for _, match := range matches {
		if len(fragments) >= maxFragments {
			break
		}

		// 检查是否已处理
		if processedRanges[match.Start] {
			continue
		}

		// 计算片段范围
		fragmentStart := match.Start - fragmentSize/2
		if fragmentStart < 0 {
			fragmentStart = 0
		}

		fragmentEnd := fragmentStart + fragmentSize
		if fragmentEnd > len(text) {
			fragmentEnd = len(text)
			fragmentStart = fragmentEnd - fragmentSize
			if fragmentStart < 0 {
				fragmentStart = 0
			}
		}

		// 调整到词边界
		fragmentStart = s.adjustToWordBoundary(text, fragmentStart, false)
		fragmentEnd = s.adjustToWordBoundary(text, fragmentEnd, true)

		// 生成高亮片段
		fragment := s.highlightFragment(text[fragmentStart:fragmentEnd], terms, tags)
		if fragment != "" {
			fragments = append(fragments, fragment)

			// 标记已处理范围
			for i := fragmentStart; i < fragmentEnd && i < len(processedRanges); i++ {
				processedRanges[i] = true
			}
		}
	}

	return fragments
}

// HighlightMatch 高亮匹配
type HighlightMatch struct {
	Start  int    `json:"start"`
	End    int    `json:"end"`
	Term   string `json:"term"`
	Length int    `json:"length"`
}

// adjustToWordBoundary 调整到词边界
func (s *SearchDomainService) adjustToWordBoundary(text string, pos int, forward bool) int {
	if pos <= 0 {
		return 0
	}
	if pos >= len(text) {
		return len(text)
	}

	if forward {
		// 向前查找词边界
		for pos < len(text) && !unicode.IsSpace(rune(text[pos])) {
			pos++
		}
	} else {
		// 向后查找词边界
		for pos > 0 && !unicode.IsSpace(rune(text[pos-1])) {
			pos--
		}
	}

	return pos
}

// highlightFragment 高亮片段
func (s *SearchDomainService) highlightFragment(fragment string, terms []string, tags HighlightTags) string {
	if fragment == "" {
		return ""
	}

	highlighted := fragment
	fragmentLower := strings.ToLower(fragment)

	// 为每个词添加高亮标签
	for _, term := range terms {
		termLower := strings.ToLower(term)

		// 使用正则表达式进行替换，保持原始大小写
		pattern := `(?i)\b` + regexp.QuoteMeta(term) + `\b`
		re := regexp.MustCompile(pattern)

		highlighted = re.R
