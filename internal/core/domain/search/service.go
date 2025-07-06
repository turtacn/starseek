package search

import (
	"context"
	"fmt"
	"math"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/turtacn/starseek/internal/common/logger"
	"github.com/turtacn/starseek/internal/core/interfaces"
)

// =============================================================================
// Search Domain Service Interface
// =============================================================================

// SearchDomainService defines the interface for search domain operations
type SearchDomainService interface {
	// Core search operations
	Search(ctx context.Context, req *SearchRequest) (*SearchResult, error)
	MultiSearch(ctx context.Context, req *MultiSearchRequest) (*MultiSearchResult, error)
	SearchWithAggregation(ctx context.Context, req *AggregatedSearchRequest) (*AggregatedSearchResult, error)

	// Query processing
	ParseQuery(ctx context.Context, query string) (*ParsedQuery, error)
	OptimizeQuery(ctx context.Context, query *ParsedQuery) (*OptimizedQuery, error)
	ValidateQuery(ctx context.Context, query *ParsedQuery) error

	// Result processing
	ProcessResults(ctx context.Context, results *RawSearchResults) (*ProcessedSearchResults, error)
	AggregateResults(ctx context.Context, results []*SearchResult) (*AggregatedResults, error)
	RankResults(ctx context.Context, results []*SearchHit, query *ParsedQuery) ([]*RankedSearchHit, error)

	// Highlighting
	HighlightResults(ctx context.Context, results []*SearchHit, query *ParsedQuery, options *HighlightOptions) error
	GenerateSnippets(ctx context.Context, content string, terms []string, options *SnippetOptions) ([]string, error)

	// Suggestions and auto-complete
	GetSearchSuggestions(ctx context.Context, req *SuggestionRequest) (*SuggestionResult, error)
	GetAutoComplete(ctx context.Context, req *AutoCompleteRequest) (*AutoCompleteResult, error)

	// Analytics and monitoring
	RecordSearchMetrics(ctx context.Context, metrics *SearchMetrics) error
	GetSearchAnalytics(ctx context.Context, req *AnalyticsRequest) (*SearchAnalytics, error)

	// Index optimization
	GetOptimalIndexes(ctx context.Context, query *ParsedQuery) ([]string, error)
	EstimateQueryCost(ctx context.Context, query *ParsedQuery, indexes []string) (*QueryCostEstimate, error)
}

// =============================================================================
// Search Domain Service Implementation
// =============================================================================

// searchDomainService implements SearchDomainService
type searchDomainService struct {
	searchRepo interfaces.SearchRepository
	cacheRepo  interfaces.CacheRepository
	logger     logger.Logger

	// Query optimization
	queryOptimizer *QueryOptimizer
	queryParser    *QueryParser

	// Scoring and ranking
	scoringEngine *ScoringEngine
	rankingEngine *RankingEngine

	// Result processing
	resultProcessor *ResultProcessor
	highlighter     *Highlighter

	// Analytics
	metricsCollector *MetricsCollector

	// Configuration
	config *SearchServiceConfig

	// Thread safety
	mu sync.RWMutex
}

// NewSearchDomainService creates a new search domain service
func NewSearchDomainService(
	searchRepo interfaces.SearchRepository,
	cacheRepo interfaces.CacheRepository,
	config *SearchServiceConfig,
) SearchDomainService {
	if config == nil {
		config = DefaultSearchServiceConfig()
	}

	logger := logger.GetLogger()

	service := &searchDomainService{
		searchRepo:       searchRepo,
		cacheRepo:        cacheRepo,
		logger:           logger,
		config:           config,
		queryOptimizer:   NewQueryOptimizer(config.OptimizationConfig),
		queryParser:      NewQueryParser(config.ParserConfig),
		scoringEngine:    NewScoringEngine(config.ScoringConfig),
		rankingEngine:    NewRankingEngine(config.RankingConfig),
		resultProcessor:  NewResultProcessor(config.ProcessingConfig),
		highlighter:      NewHighlighter(config.HighlightConfig),
		metricsCollector: NewMetricsCollector(config.MetricsConfig),
	}

	return service
}

// Search performs a search operation
func (s *searchDomainService) Search(ctx context.Context, req *SearchRequest) (*SearchResult, error) {
	startTime := time.Now()

	// Input validation
	if err := s.validateSearchRequest(req); err != nil {
		return nil, fmt.Errorf("invalid search request: %w", err)
	}

	// Parse query
	parsedQuery, err := s.ParseQuery(ctx, req.Query)
	if err != nil {
		return nil, fmt.Errorf("failed to parse query: %w", err)
	}

	// Optimize query
	optimizedQuery, err := s.OptimizeQuery(ctx, parsedQuery)
	if err != nil {
		s.logger.WarnContext(ctx, "Query optimization failed, using original query",
			logger.ErrorField(err))
		optimizedQuery = &OptimizedQuery{ParsedQuery: parsedQuery}
	}

	// Get optimal indexes
	indexes, err := s.GetOptimalIndexes(ctx, parsedQuery)
	if err != nil {
		s.logger.WarnContext(ctx, "Failed to get optimal indexes",
			logger.ErrorField(err))
		indexes = req.Indexes
	}

	// Execute search
	searchReq := s.buildRepositorySearchRequest(req, optimizedQuery, indexes)
	rawResult, err := s.searchRepo.Search(ctx, searchReq)
	if err != nil {
		return nil, fmt.Errorf("repository search failed: %w", err)
	}

	// Process results
	processedResults, err := s.processSearchResults(ctx, rawResult, parsedQuery, req)
	if err != nil {
		return nil, fmt.Errorf("failed to process search results: %w", err)
	}

	// Record metrics
	duration := time.Since(startTime)
	metrics := &SearchMetrics{
		QueryText:        req.Query,
		IndexNames:       indexes,
		ResultCount:      int64(len(processedResults.Hits)),
		Duration:         duration,
		CacheHit:         false,
		OptimizationUsed: optimizedQuery.IsOptimized,
		Timestamp:        startTime,
	}

	if err := s.RecordSearchMetrics(ctx, metrics); err != nil {
		s.logger.WarnContext(ctx, "Failed to record search metrics",
			logger.ErrorField(err))
	}

	result := &SearchResult{
		Query:          req.Query,
		ParsedQuery:    parsedQuery,
		OptimizedQuery: optimizedQuery,
		Hits:           processedResults.Hits,
		Total:          processedResults.Total,
		MaxScore:       processedResults.MaxScore,
		Aggregations:   processedResults.Aggregations,
		Duration:       duration,
		IndexesUsed:    indexes,
		Suggestions:    processedResults.Suggestions,
		Metadata:       processedResults.Metadata,
	}

	return result, nil
}

// MultiSearch performs multiple search operations
func (s *searchDomainService) MultiSearch(ctx context.Context, req *MultiSearchRequest) (*MultiSearchResult, error) {
	startTime := time.Now()

	if len(req.Searches) == 0 {
		return &MultiSearchResult{}, nil
	}

	results := make([]*SearchResult, len(req.Searches))
	errors := make([]error, len(req.Searches))

	// Execute searches concurrently if enabled
	if req.Concurrent && len(req.Searches) > 1 {
		var wg sync.WaitGroup
		for i, searchReq := range req.Searches {
			wg.Add(1)
			go func(idx int, sreq *SearchRequest) {
				defer wg.Done()
				result, err := s.Search(ctx, sreq)
				results[idx] = result
				errors[idx] = err
			}(i, searchReq)
		}
		wg.Wait()
	} else {
		// Execute searches sequentially
		for i, searchReq := range req.Searches {
			result, err := s.Search(ctx, searchReq)
			results[i] = result
			errors[idx] = err
		}
	}

	// Collect successful results
	successfulResults := make([]*SearchResult, 0, len(results))
	searchErrors := make([]error, 0)

	for i, result := range results {
		if errors[i] != nil {
			searchErrors = append(searchErrors, errors[i])
		} else if result != nil {
			successfulResults = append(successfulResults, result)
		}
	}

	duration := time.Since(startTime)

	multiResult := &MultiSearchResult{
		Results:    successfulResults,
		Errors:     searchErrors,
		Duration:   duration,
		Total:      len(req.Searches),
		Successful: len(successfulResults),
		Failed:     len(searchErrors),
	}

	return multiResult, nil
}

// SearchWithAggregation performs search with aggregations
func (s *searchDomainService) SearchWithAggregation(ctx context.Context, req *AggregatedSearchRequest) (*AggregatedSearchResult, error) {
	startTime := time.Now()

	// Perform base search
	searchResult, err := s.Search(ctx, &req.SearchRequest)
	if err != nil {
		return nil, fmt.Errorf("base search failed: %w", err)
	}

	// Execute aggregations
	aggregationResults := make(map[string]*AggregationResult)

	for name, aggConfig := range req.Aggregations {
		aggResult, err := s.executeAggregation(ctx, aggConfig, searchResult)
		if err != nil {
			s.logger.WarnContext(ctx, "Aggregation failed",
				logger.String("aggregation", name),
				logger.ErrorField(err))
			continue
		}
		aggregationResults[name] = aggResult
	}

	duration := time.Since(startTime)

	result := &AggregatedSearchResult{
		SearchResult: searchResult,
		Aggregations: aggregationResults,
		Duration:     duration,
	}

	return result, nil
}

// ParseQuery parses a query string into a structured query
func (s *searchDomainService) ParseQuery(ctx context.Context, query string) (*ParsedQuery, error) {
	if strings.TrimSpace(query) == "" {
		return &ParsedQuery{
			Original: query,
			Type:     QueryTypeMatchAll,
			Terms:    []string{},
		}, nil
	}

	return s.queryParser.Parse(query)
}

// OptimizeQuery optimizes a parsed query
func (s *searchDomainService) OptimizeQuery(ctx context.Context, query *ParsedQuery) (*OptimizedQuery, error) {
	return s.queryOptimizer.Optimize(query)
}

// ValidateQuery validates a parsed query
func (s *searchDomainService) ValidateQuery(ctx context.Context, query *ParsedQuery) error {
	if query == nil {
		return fmt.Errorf("query cannot be nil")
	}

	// Check query complexity
	if len(query.Terms) > s.config.MaxQueryTerms {
		return fmt.Errorf("query has too many terms: %d (max: %d)",
			len(query.Terms), s.config.MaxQueryTerms)
	}

	// Check boolean complexity
	if query.BooleanClauses != nil && len(query.BooleanClauses) > s.config.MaxBooleanClauses {
		return fmt.Errorf("query has too many boolean clauses: %d (max: %d)",
			len(query.BooleanClauses), s.config.MaxBooleanClauses)
	}

	// Validate field restrictions
	for _, field := range query.Fields {
		if !s.isValidField(field) {
			return fmt.Errorf("invalid field: %s", field)
		}
	}

	return nil
}

// ProcessResults processes raw search results
func (s *searchDomainService) ProcessResults(ctx context.Context, results *RawSearchResults) (*ProcessedSearchResults, error) {
	return s.resultProcessor.Process(results)
}

// AggregateResults aggregates multiple search results
func (s *searchDomainService) AggregateResults(ctx context.Context, results []*SearchResult) (*AggregatedResults, error) {
	if len(results) == 0 {
		return &AggregatedResults{}, nil
	}

	// Collect all hits
	allHits := make([]*SearchHit, 0)
	totalResults := int64(0)
	maxScore := 0.0

	for _, result := range results {
		allHits = append(allHits, result.Hits...)
		totalResults += result.Total
		if result.MaxScore > maxScore {
			maxScore = result.MaxScore
		}
	}

	// Remove duplicates based on document ID
	deduplicatedHits := s.deduplicateHits(allHits)

	// Sort by score
	sort.Slice(deduplicatedHits, func(i, j int) bool {
		return deduplicatedHits[i].Score > deduplicatedHits[j].Score
	})

	return &AggregatedResults{
		Hits:         deduplicatedHits,
		Total:        totalResults,
		MaxScore:     maxScore,
		Deduplicated: len(allHits) - len(deduplicatedHits),
	}, nil
}

// RankResults ranks search results using scoring algorithms
func (s *searchDomainService) RankResults(ctx context.Context, hits []*SearchHit, query *ParsedQuery) ([]*RankedSearchHit, error) {
	rankedHits := make([]*RankedSearchHit, len(hits))

	for i, hit := range hits {
		// Calculate relevance score
		relevanceScore := s.scoringEngine.CalculateRelevanceScore(hit, query)

		// Calculate popularity score
		popularityScore := s.scoringEngine.CalculatePopularityScore(hit)

		// Calculate freshness score
		freshnessScore := s.scoringEngine.CalculateFreshnessScore(hit)

		// Combine scores
		finalScore := s.scoringEngine.CombineScores(relevanceScore, popularityScore, freshnessScore)

		rankedHits[i] = &RankedSearchHit{
			SearchHit:       hit,
			RelevanceScore:  relevanceScore,
			PopularityScore: popularityScore,
			FreshnessScore:  freshnessScore,
			FinalScore:      finalScore,
			RankingFactors:  s.scoringEngine.GetRankingFactors(hit, query),
		}
	}

	// Sort by final score
	sort.Slice(rankedHits, func(i, j int) bool {
		return rankedHits[i].FinalScore > rankedHits[j].FinalScore
	})

	// Update rankings
	for i, hit := range rankedHits {
		hit.Rank = i + 1
	}

	return rankedHits, nil
}

// HighlightResults adds highlighting to search results
func (s *searchDomainService) HighlightResults(ctx context.Context, hits []*SearchHit, query *ParsedQuery, options *HighlightOptions) error {
	if options == nil {
		options = DefaultHighlightOptions()
	}

	for _, hit := range hits {
		highlights := make(map[string][]string)

		for field, content := range hit.Source {
			if contentStr, ok := content.(string); ok && s.shouldHighlightField(field, options) {
				highlighted := s.highlighter.HighlightText(contentStr, query.Terms, options)
				if len(highlighted) > 0 {
					highlights[field] = highlighted
				}
			}
		}

		hit.Highlight = highlights
	}

	return nil
}

// GenerateSnippets generates text snippets with highlighted terms
func (s *searchDomainService) GenerateSnippets(ctx context.Context, content string, terms []string, options *SnippetOptions) ([]string, error) {
	if options == nil {
		options = DefaultSnippetOptions()
	}

	return s.highlighter.GenerateSnippets(content, terms, options), nil
}

// GetSearchSuggestions provides search suggestions
func (s *searchDomainService) GetSearchSuggestions(ctx context.Context, req *SuggestionRequest) (*SuggestionResult, error) {
	// Check cache first
	cacheKey := s.buildSuggestionCacheKey(req)
	if cached, err := s.getCachedSuggestions(ctx, cacheKey); err == nil && cached != nil {
		return cached, nil
	}

	// Generate suggestions
	suggestions := make([]*Suggestion, 0)

	// Term suggestions
	termSuggestions, err := s.generateTermSuggestions(ctx, req)
	if err == nil {
		suggestions = append(suggestions, termSuggestions...)
	}

	// Phrase suggestions
	phraseSuggestions, err := s.generatePhraseSuggestions(ctx, req)
	if err == nil {
		suggestions = append(suggestions, phraseSuggestions...)
	}

	// Completion suggestions
	completionSuggestions, err := s.generateCompletionSuggestions(ctx, req)
	if err == nil {
		suggestions = append(suggestions, completionSuggestions...)
	}

	// Sort by score
	sort.Slice(suggestions, func(i, j int) bool {
		return suggestions[i].Score > suggestions[j].Score
	})

	// Limit results
	if len(suggestions) > req.Size {
		suggestions = suggestions[:req.Size]
	}

	result := &SuggestionResult{
		Suggestions: suggestions,
		Total:       len(suggestions),
	}

	// Cache result
	s.cacheSuggestions(ctx, cacheKey, result)

	return result, nil
}

// GetAutoComplete provides auto-completion suggestions
func (s *searchDomainService) GetAutoComplete(ctx context.Context, req *AutoCompleteRequest) (*AutoCompleteResult, error) {
	// Implementation for auto-completion
	completions := make([]*AutoCompletion, 0)

	// Generate completions based on partial input
	if len(req.Prefix) >= req.MinLength {
		indexCompletions, err := s.generateIndexBasedCompletions(ctx, req)
		if err == nil {
			completions = append(completions, indexCompletions...)
		}

		// Add popular queries
		popularCompletions, err := s.generatePopularCompletions(ctx, req)
		if err == nil {
			completions = append(completions, popularCompletions...)
		}
	}

	// Sort and limit
	sort.Slice(completions, func(i, j int) bool {
		return completions[i].Score > completions[j].Score
	})

	if len(completions) > req.Size {
		completions = completions[:req.Size]
	}

	return &AutoCompleteResult{
		Completions: completions,
		Total:       len(completions),
	}, nil
}

// RecordSearchMetrics records search performance metrics
func (s *searchDomainService) RecordSearchMetrics(ctx context.Context, metrics *SearchMetrics) error {
	return s.metricsCollector.Record(metrics)
}

// GetSearchAnalytics retrieves search analytics
func (s *searchDomainService) GetSearchAnalytics(ctx context.Context, req *AnalyticsRequest) (*SearchAnalytics, error) {
	return s.metricsCollector.GetAnalytics(req)
}

// GetOptimalIndexes determines the best indexes for a query
func (s *searchDomainService) GetOptimalIndexes(ctx context.Context, query *ParsedQuery) ([]string, error) {
	return s.queryOptimizer.SelectOptimalIndexes(query)
}

// EstimateQueryCost estimates the cost of executing a query
func (s *searchDomainService) EstimateQueryCost(ctx context.Context, query *ParsedQuery, indexes []string) (*QueryCostEstimate, error) {
	return s.queryOptimizer.EstimateCost(query, indexes)
}

// =============================================================================
// Private Helper Methods
// =============================================================================

// validateSearchRequest validates a search request
func (s *searchDomainService) validateSearchRequest(req *SearchRequest) error {
	if req == nil {
		return fmt.Errorf("search request cannot be nil")
	}

	if len(req.Query) > s.config.MaxQueryLength {
		return fmt.Errorf("query too long: %d characters (max: %d)",
			len(req.Query), s.config.MaxQueryLength)
	}

	if req.Size > s.config.MaxResultSize {
		return fmt.Errorf("result size too large: %d (max: %d)",
			req.Size, s.config.MaxResultSize)
	}

	return nil
}

// buildRepositorySearchRequest builds a repository search request
func (s *searchDomainService) buildRepositorySearchRequest(
	req *SearchRequest,
	optimizedQuery *OptimizedQuery,
	indexes []string,
) *interfaces.SearchRequest {
	return &interfaces.SearchRequest{
		Index:   strings.Join(indexes, ","),
		Query:   optimizedQuery.RepositoryQuery,
		From:    req.From,
		Size:    req.Size,
		Sort:    req.Sort,
		Source:  req.Fields,
		Timeout: req.Timeout,
	}
}

// processSearchResults processes search results from repository
func (s *searchDomainService) processSearchResults(
	ctx context.Context,
	rawResult *interfaces.SearchResponse,
	query *ParsedQuery,
	req *SearchRequest,
) (*ProcessedSearchResults, error) {
	// Convert repository hits to domain hits
	hits := make([]*SearchHit, len(rawResult.Hits.Hits))
	for i, hit := range rawResult.Hits.Hits {
		hits[i] = &SearchHit{
			ID:        hit.ID,
			Index:     hit.Index,
			Score:     *hit.Score,
			Source:    hit.Source,
			Highlight: make(map[string][]string),
		}
	}

	// Apply highlighting if requested
	if req.Highlight {
		s.HighlightResults(ctx, hits, query, req.HighlightOptions)
	}

	// Rank results if custom ranking is enabled
	if req.EnableRanking {
		rankedHits, err := s.RankResults(ctx, hits, query)
		if err != nil {
			s.logger.WarnContext(ctx, "Failed to rank results", logger.ErrorField(err))
		} else {
			// Convert back to SearchHit
			for i, rankedHit := range rankedHits {
				hits[i] = rankedHit.SearchHit
				hits[i].Score = rankedHit.FinalScore
			}
		}
	}

	return &ProcessedSearchResults{
		Hits:         hits,
		Total:        rawResult.Hits.Total.Value,
		MaxScore:     *rawResult.Hits.MaxScore,
		Aggregations: rawResult.Aggregations,
		Suggestions:  make([]*Suggestion, 0),
		Metadata:     make(map[string]interface{}),
	}, nil
}

// executeAggregation executes an aggregation
func (s *searchDomainService) executeAggregation(
	ctx context.Context,
	config *AggregationConfig,
	searchResult *SearchResult,
) (*AggregationResult, error) {
	// Implementation depends on aggregation type
	switch config.Type {
	case AggregationTypeTerms:
		return s.executeTermsAggregation(config, searchResult)
	case AggregationTypeRange:
		return s.executeRangeAggregation(config, searchResult)
	case AggregationTypeHistogram:
		return s.executeHistogramAggregation(config, searchResult)
	case AggregationTypeStats:
		return s.executeStatsAggregation(config, searchResult)
	default:
		return nil, fmt.Errorf("unsupported aggregation type: %s", config.Type)
	}
}

// executeTermsAggregation executes a terms aggregation
func (s *searchDomainService) executeTermsAggregation(
	config *AggregationConfig,
	searchResult *SearchResult,
) (*AggregationResult, error) {
	termCounts := make(map[string]int64)

	for _, hit := range searchResult.Hits {
		if value, exists := hit.Source[config.Field]; exists {
			if strValue, ok := value.(string); ok {
				termCounts[strValue]++
			}
		}
	}

	// Convert to buckets
	buckets := make([]*AggregationBucket, 0, len(termCounts))
	for term, count := range termCounts {
		buckets = append(buckets, &AggregationBucket{
			Key:      term,
			DocCount: count,
		})
	}

	// Sort by count
	sort.Slice(buckets, func(i, j int) bool {
		return buckets[i].DocCount > buckets[j].DocCount
	})

	// Limit results
	if config.Size > 0 && len(buckets) > config.Size {
		buckets = buckets[:config.Size]
	}

	return &AggregationResult{
		Type:    AggregationTypeTerms,
		Buckets: buckets,
	}, nil
}

// executeRangeAggregation executes a range aggregation
func (s *searchDomainService) executeRangeAggregation(
	config *AggregationConfig,
	searchResult *SearchResult,
) (*AggregationResult, error) {
	// Implementation for range aggregation
	buckets := make([]*AggregationBucket, len(config.Ranges))

	for i, rangeConfig := range config.Ranges {
		count := int64(0)
		for _, hit := range searchResult.Hits {
			if value, exists := hit.Source[config.Field]; exists {
				if numValue, ok := value.(float64); ok {
					if numValue >= rangeConfig.From && numValue < rangeConfig.To {
						count++
					}
				}
			}
		}

		buckets[i] = &AggregationBucket{
			Key:      fmt.Sprintf("%.2f-%.2f", rangeConfig.From, rangeConfig.To),
			DocCount: count,
			From:     &rangeConfig.From,
			To:       &rangeConfig.To,
		}
	}

	return &AggregationResult{
		Type:    AggregationTypeRange,
		Buckets: buckets,
	}, nil
}

// executeHistogramAggregation executes a histogram aggregation
func (s *searchDomainService) executeHistogramAggregation(
	config *AggregationConfig,
	searchResult *SearchResult,
) (*AggregationResult, error) {
	// Implementation for histogram aggregation
	histogramBuckets := make(map[int64]int64)

	for _, hit := range searchResult.Hits {
		if value, exists := hit.Source[config.Field]; exists {
			if numValue, ok := value.(float64); ok {
				bucket := int64(numValue/config.Interval) * int64(config.Interval)
				histogramBuckets[bucket]++
			}
		}
	}

	// Convert to buckets
	buckets := make([]*AggregationBucket, 0, len(histogramBuckets))
	for bucketKey, count := range histogramBuckets {
		buckets = append(buckets, &AggregationBucket{
			Key:      fmt.Sprintf("%d", bucketKey),
			DocCount: count,
		})
	}

	// Sort by key
	sort.Slice(buckets, func(i, j int) bool {
		return buckets[i].Key < buckets[j].Key
	})

	return &AggregationResult{
		Type:    AggregationTypeHistogram,
		Buckets: buckets,
	}, nil
}

// executeStatsAggregation executes a statistics aggregation
func (s *searchDomainService) executeStatsAggregation(
	config *AggregationConfig,
	searchResult *SearchResult,
) (*AggregationResult, error) {
	values := make([]float64, 0)

	for _, hit := range searchResult.Hits {
		if value, exists := hit.Source[config.Field]; exists {
			if numValue, ok := value.(float64); ok {
				values = append(values, numValue)
			}
		}
	}

	if len(values) == 0 {
		return &AggregationResult{
			Type: AggregationTypeStats,
			Stats: &AggregationStats{
				Count: 0,
			},
		}, nil
	}

	// Calculate statistics
	sum := 0.0
	min := values[0]
	max := values[0]

	for _, v := range values {
		sum += v
		if v < min {
			min = v
		}
		if v > max {
			max = v
		}
	}

	avg := sum / float64(len(values))

	return &AggregationResult{
		Type: AggregationTypeStats,
		Stats: &AggregationStats{
			Count: int64(len(values)),
			Min:   min,
			Max:   max,
			Avg:   avg,
			Sum:   sum,
		},
	}, nil
}

// deduplicateHits removes duplicate hits based on document ID
func (s *searchDomainService) deduplicateHits(hits []*SearchHit) []*SearchHit {
	seen := make(map[string]bool)
	result := make([]*SearchHit, 0, len(hits))

	for _, hit := range hits {
		key := fmt.Sprintf("%s:%s", hit.Index, hit.ID)
		if !seen[key] {
			seen[key] = true
			result = append(result, hit)
		}
	}

	return result
}

// isValidField checks if a field is valid for searching
func (s *searchDomainService) isValidField(field string) bool {
	// Implementation depends on your field validation rules
	if len(field) == 0 || len(field) > 100 {
		return false
	}

	// Check for invalid characters
	matched, _ := regexp.MatchString("^[a-zA-Z0-9._-]+$", field)
	return matched
}

// shouldHighlightField determines if a field should be highlighted
func (s *searchDomainService) shouldHighlightField(field string, options *HighlightOptions) bool {
	if len(options.Fields) == 0 {
		return true // Highlight all fields if none specified
	}

	for _, highlightField := range options.Fields {
		if field == highlightField {
			return true
		}
	}

	return false
}

// buildSuggestionCacheKey builds a cache key for suggestions
func (s *searchDomainService) buildSuggestionCacheKey(req *SuggestionRequest) string {
	return fmt.Sprintf("suggestions:%s:%s:%d", req.Index, req.Text, req.Size)
}

// getCachedSuggestions retrieves cached suggestions
func (s *searchDomainService) getCachedSuggestions(ctx context.Context, key string) (*SuggestionResult, error) {
	// Implementation for cache retrieval
	return nil, fmt.Errorf("not implemented")
}

// cacheSuggestions caches suggestion results
func (s *searchDomainService) cacheSuggestions(ctx context.Context, key string, result *SuggestionResult) {
	// Implementation for caching suggestions
}

// generateTermSuggestions generates term-based suggestions
func (s *searchDomainService) generateTermSuggestions(ctx context.Context, req *SuggestionRequest) ([]*Suggestion, error) {
	// Implementation for term suggestions
	return make([]*Suggestion, 0), nil
}

// generatePhraseSuggestions generates phrase-based suggestions
func (s *searchDomainService) generatePhraseSuggestions(ctx context.Context, req *SuggestionRequest) ([]*Suggestion, error) {
	// Implementation for phrase suggestions
	return make([]*Suggestion, 0), nil
}

// generateCompletionSuggestions generates completion suggestions
func (s *searchDomainService) generateCompletionSuggestions(ctx context.Context, req *SuggestionRequest) ([]*Suggestion, error) {
	// Implementation for completion suggestions
	return make([]*Suggestion, 0), nil
}

// generateIndexBasedCompletions generates completions from index data
func (s *searchDomainService) generateIndexBasedCompletions(ctx context.Context, req *AutoCompleteRequest) ([]*AutoCompletion, error) {
	// Implementation for index-based completions
	return make([]*AutoCompletion, 0), nil
}

// generatePopularCompletions generates popular query completions
func (s *searchDomainService) generatePopularCompletions(ctx context.Context, req *AutoCompleteRequest) ([]*AutoCompletion, error) {
	// Implementation for popular completions
	return make([]*AutoCompletion, 0), nil
}

//Personal.AI order the ending
