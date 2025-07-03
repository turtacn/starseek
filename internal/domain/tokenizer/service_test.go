package search_test

import (
	"context"
	"fmt"
	"math"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	ferrors "github.com/turtacn/starseek/internal/common/errors"
	"github.com/turtacn/starseek/internal/common/enum"
	"github.com/turtacn/starseek/internal/domain/index"
	"github.com/turtacn/starseek/internal/domain/search"      // 被测试的 search 领域包
	"github.com/turtacn/starseek/internal/domain/tokenizer"    // 引入 tokenizer 接口
	"github.com/turtacn/starseek/internal/infrastructure/cache" // 引入 cache 接口
)

// MockLogger (from previous files, duplicated here for self-containment of test file)
// Using the same MockLogger from previous test files.
// For brevity, not including its definition again, assume it's available.
// If you're copying this directly, include the MockLogger definition.
// var mockLog *MockLogger // Assumed to be initialized in TestMain or init()

func init() {
	if mockLog == nil { // Ensure it's initialized only once if TestMain isn't running it
		mockLog = NewMockLogger()
	}
}

// Mocks for SearchService dependencies

// MockQueryBuilder for searchService testing
type MockQueryBuilder struct {
	mock.Mock
}

func (m *MockQueryBuilder) BuildSQL(ctx context.Context, query *search.SearchQuery, indexedColumns []*index.IndexMetadata) (string, error) {
	args := m.Called(ctx, query, indexedColumns)
	return args.String(0), args.Error(1)
}

// MockQueryExecutor for searchService testing
type MockQueryExecutor struct {
	mock.Mock
}

func (m *MockQueryExecutor) Execute(ctx context.Context, sql string, dbType enum.DatabaseType) ([]map[string]interface{}, error) {
	args := m.Called(ctx, sql, dbType)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]map[string]interface{}), args.Error(1)
}

// MockRanker for searchService testing
type MockRanker struct {
	mock.Mock
}

func (m *MockRanker) Rank(ctx context.Context, results []*search.SearchResult, query *search.SearchQuery) error {
	args := m.Called(ctx, results, query)
	// Simulate in-place sorting for ranker:
	// For testing, we just set arbitrary scores and then sort.
	// In real tests, BM25Ranker would handle actual scoring.
	if len(results) > 0 {
		// Assign mock scores (e.g., reverse order of RowID)
		for i, res := range results {
			// This is a very simple scoring for testing purposes.
			// It ensures results are re-ordered predictably.
			if res.RowID == "doc1" {
				res.Score = 100.0
			} else if res.RowID == "doc2" {
				res.Score = 50.0
			} else if res.RowID == "doc3" {
				res.Score = 150.0 // Make doc3 highest
			} else {
				res.Score = float64(len(results) - i) // Default fallback
			}
		}
		sort.Slice(results, func(i, j int) bool {
			return results[i].Score > results[j].Score
		})
	}
	return args.Error(0)
}

// MockTokenizerService for searchService testing
type MockTokenizerService struct {
	mock.Mock
}

func (m *MockTokenizerService) GetTokenizer(indexType enum.IndexType, tokenizerName string) (tokenizer.Tokenizer, error) {
	args := m.Called(indexType, tokenizerName)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(tokenizer.Tokenizer), args.Error(1)
}

// MockTokenizer for the MockTokenizerService
type TestTokenizer struct{}

func (t *TestTokenizer) Tokenize(text string) ([]string, error) {
	// Simple tokenizer for tests: split by space and convert to lowercase
	tokens := strings.Fields(strings.ToLower(text))
	return tokens, nil
}

// MockIndexService for searchService testing
type MockIndexService struct {
	mock.Mock
}

func (m *MockIndexService) GetIndexMetadata(ctx context.Context, tableName string) ([]*index.IndexMetadata, error) {
	args := m.Called(ctx, tableName)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*index.IndexMetadata), args.Error(1)
}
func (m *MockIndexService) GetAllIndexMetadata(ctx context.Context) ([]*index.IndexMetadata, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*index.IndexMetadata), args.Error(1)
}
func (m *MockIndexService) UpdateIndexMetadata(ctx context.Context, metadata []*index.IndexMetadata) error {
	args := m.Called(ctx, metadata)
	return args.Error(0)
}

// MockCacheClient (from previous files, duplicated here for self-containment of test file)
// Using the same MockCacheClient from previous test files.
// For brevity, not including its definition again, assume it's available.
// If you're copying this directly, include the MockCacheClient definition.

// ==============================================================================
// SearchService Tests
// ==============================================================================

func setupSearchService(t *testing.T) (*MockQueryBuilder, *MockQueryExecutor, *MockRanker, *MockTokenizerService, *MockIndexService, *MockCacheClient, search.SearchService) {
	mockQueryBuilder := new(MockQueryBuilder)
	mockQueryExecutor := new(MockQueryExecutor)
	mockRanker := new(MockRanker)
	mockTokenizerService := new(MockTokenizerService)
	mockIndexService := new(MockIndexService)
	mockCacheClient := new(MockCacheClient)

	svc, err := search.NewSearchService(
		mockQueryBuilder,
		mockQueryExecutor,
		mockRanker,
		mockTokenizerService,
		mockIndexService,
		mockLog,
		mockCacheClient,
	)
	require.NoError(t, err)
	assert.NotNil(t, svc)
	mockLog.Logs = nil // Clear logs after service initialization log

	return mockQueryBuilder, mockQueryExecutor, mockRanker, mockTokenizerService, mockIndexService, mockCacheClient, svc
}

func TestNewSearchService_Success(t *testing.T) {
	_, _, _, _, _, _, _ = setupSearchService(t)
	assert.Contains(t, mockLog.Logs, "INFO: SearchService initialized.")
	mockLog.Logs = nil
}

func TestNewSearchService_NilDependencies(t *testing.T) {
	mockQueryBuilder := new(MockQueryBuilder)
	mockQueryExecutor := new(MockQueryExecutor)
	mockRanker := new(MockRanker)
	mockTokenizerService := new(MockTokenizerService)
	mockIndexService := new(MockIndexService)
	mockCacheClient := new(MockCacheClient)

	tests := []struct {
		name        string
		qb          search.QueryBuilder
		qe          search.QueryExecutor
		rank        search.Ranker
		ts          tokenizer.TokenizerService
		is          index.IndexService
		log         *MockLogger // Using concrete type here
		cache       cache.CacheClient
		expectedErr string
	}{
		{"nil queryBuilder", nil, mockQE, mockRank, mockTS, mockIS, mockLog, mockCache, "query builder is nil"},
		{"nil queryExecutor", mockQB, nil, mockRank, mockTS, mockIS, mockLog, mockCache, "query executor is nil"},
		{"nil ranker", mockQB, mockQE, nil, mockTS, mockIS, mockLog, mockCache, "ranker is nil"},
		{"nil tokenizerService", mockQB, mockQE, mockRank, nil, mockIS, mockLog, mockCache, "tokenizer service is nil"},
		{"nil indexService", mockQB, mockQE, mockRank, mockTS, nil, mockLog, mockCache, "index service is nil"},
		{"nil logger", mockQB, mockQE, mockRank, mockTS, mockIS, nil, mockCache, "logger is nil"},
		{"nil cacheClient", mockQB, mockQE, mockRank, mockTS, mockIS, mockLog, nil, "cache client is nil"},
	}

	// Make copies of mock objects to avoid interference between tests
	mockQB := new(MockQueryBuilder)
	mockQE := new(MockQueryExecutor)
	mockRank := new(MockRanker)
	mockTS := new(MockTokenizerService)
	mockIS := new(MockIndexService)
	mockCache := new(MockCacheClient)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := search.NewSearchService(tt.qb, tt.qe, tt.rank, tt.ts, tt.is, tt.log, tt.cache)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), tt.expectedErr)
		})
	}
}

func TestSearchService_Search_FullFlowSuccess(t *testing.T) {
	mockQueryBuilder, mockQueryExecutor, mockRanker, mockTokenizerService, mockIndexService, _, svc := setupSearchService(t)

	ctx := context.Background()
	req := &search.SearchRequest{
		Query:        "go lang microservices",
		Tables:       []string{"products"},
		Fields:       []string{"name", "description"},
		Page:         1,
		PageSize:     10,
		QueryType:    search.MATCH_ANY,
		DatabaseType: "clickhouse",
	}

	// 1. Mock TokenizerService.GetTokenizer
	mockTokenizerService.On("GetTokenizer", enum.FULL_TEXT_INDEX, enum.ENGLISH_TOKENIZER.String()).Return(&TestTokenizer{}, nil).Once()

	// 2. Mock IndexService.GetIndexMetadata
	indexedCols := []*index.IndexMetadata{
		{TableName: "products", ColumnName: "name", IndexType: enum.FULL_TEXT_INDEX},
		{TableName: "products", ColumnName: "description", IndexType: enum.FULL_TEXT_INDEX},
	}
	mockIndexService.On("GetIndexMetadata", ctx, "products").Return(indexedCols, nil).Once()

	// 3. Mock QueryBuilder.BuildSQL
	expectedSearchQuery := &search.SearchQuery{
		Tables:    []string{"products"},
		Keywords:  []string{"go", "lang", "microservices"},
		Fields:    []string{"name", "description"},
		Filters:   nil, // No filters in this request
		Page:      1,
		PageSize:  10,
		QueryType: search.MATCH_ANY,
	}
	mockSQL := "SELECT * FROM products WHERE ... LIMIT 10 OFFSET 0"
	mockQueryBuilder.On("BuildSQL", ctx, mock.AnythingOfType("*search.SearchQuery"), indexedCols).
		Return(mockSQL, nil).
		Run(func(args mock.Arguments) {
			// Deep check the SearchQuery argument passed to BuildSQL
			argQuery := args.Get(1).(*search.SearchQuery)
			assert.Equal(t, expectedSearchQuery.Tables, argQuery.Tables)
			assert.Equal(t, expectedSearchQuery.Keywords, argQuery.Keywords)
			assert.Equal(t, expectedSearchQuery.Fields, argQuery.Fields)
			assert.Equal(t, expectedSearchQuery.Page, argQuery.Page)
			assert.Equal(t, expectedSearchQuery.PageSize, argQuery.PageSize)
			assert.Equal(t, expectedSearchQuery.QueryType, argQuery.QueryType)
		}).Once()

	// 4. Mock QueryExecutor.Execute
	rawResults := []map[string]interface{}{
		{"id": uint64(1), "name": "GoLang Book", "description": "Learn Go programming language from scratch.", "category": "Programming"},
		{"id": uint64(2), "name": "ClickHouse Basics", "description": "A guide to ClickHouse database fundamentals.", "category": "Database"},
		{"id": uint64(3), "name": "Microservices in Go", "description": "Building scalable microservices with Go.", "category": "Programming"},
	}
	mockQueryExecutor.On("Execute", ctx, mockSQL, enum.CLICKHOUSE).Return(rawResults, nil).Once()

	// 5. Mock Ranker.Rank (it modifies results in place, our mock does arbitrary re-sorting)
	// The mockRanker will assign scores and sort, making doc3 -> doc1 -> doc2
	mockRanker.On("Rank", ctx, mock.AnythingOfType("[]*search.SearchResult"), mock.AnythingOfType("*search.SearchQuery")).
		Return(nil).Once()

	// Call the service
	resp, err := svc.Search(ctx, req)
	require.NoError(t, err)
	assert.NotNil(t, resp)

	// Assertions on the final response
	assert.Equal(t, int64(3), resp.TotalHits)
	assert.Len(t, resp.Results, 3)
	assert.Equal(t, req.Page, resp.Page)
	assert.Equal(t, req.PageSize, resp.PageSize)

	// Verify ordering by mockRanker logic: doc3, doc1, doc2
	assert.Equal(t, "doc3", resp.Results[0].RowID)
	assert.Equal(t, "doc1", resp.Results[1].RowID)
	assert.Equal(t, "doc2", resp.Results[2].RowID)

	// Verify highlighting (simple string replace)
	assert.Contains(t, resp.Results[0].HighlightedData["name"], "<b>Go</b>")
	assert.Contains(t, resp.Results[0].HighlightedData["name"], "<b>Microservices</b>")
	assert.Contains(t, resp.Results[0].HighlightedData["description"], "<b>Go</b>")
	assert.Contains(t, resp.Results[0].HighlightedData["description"], "<b>microservices</b>") // lower case as tokenized
	assert.NotContains(t, resp.Results[0].HighlightedData["category"], "<b>") // category not in Fields

	assert.Contains(t, mockLog.Logs, "INFO: Search completed successfully total_hits=3 page=1")

	// Verify all mocks were called
	mockTokenizerService.AssertExpectations(t)
	mockIndexService.AssertExpectations(t)
	mockQueryBuilder.AssertExpectations(t)
	mockQueryExecutor.AssertExpectations(t)
	mockRanker.AssertExpectations(t)
	mockLog.Logs = nil
}

func TestSearchService_Search_TokenizerError(t *testing.T) {
	mockQueryBuilder, mockQueryExecutor, mockRanker, mockTokenizerService, mockIndexService, _, svc := setupSearchService(t)

	ctx := context.Background()
	req := &search.SearchRequest{
		Query:  "test",
		Tables: []string{"products"},
	}

	// Simulate tokenizer service error
	mockTokenizerService.On("GetTokenizer", mock.Anything, mock.Anything).Return(nil, fmt.Errorf("tokenizer init error")).Once()

	_, err := svc.Search(ctx, req)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to initialize tokenizer for search")
	mockTokenizerService.AssertExpectations(t)
	mockQueryBuilder.AssertNotCalled(t, "BuildSQL", mock.Anything, mock.Anything, mock.Anything)
	mockQueryExecutor.AssertNotCalled(t, "Execute", mock.Anything, mock.Anything, mock.Anything)
	mockRanker.AssertNotCalled(t, "Rank", mock.Anything, mock.Anything, mock.Anything)
	assert.Contains(t, mockLog.Logs, "ERROR: Failed to get tokenizer")
	mockLog.Logs = nil
}

func TestSearchService_Search_EmptyQueryAndNoFilters(t *testing.T) {
	mockQueryBuilder, mockQueryExecutor, mockRanker, mockTokenizerService, mockIndexService, _, svc := setupSearchService(t)

	ctx := context.Background()
	req := &search.SearchRequest{
		Query:  "   ", // Empty query after trimming/tokenizing
		Tables: []string{"products"},
	}

	mockTokenizerService.On("GetTokenizer", mock.Anything, mock.Anything).Return(&TestTokenizer{}, nil).Once()

	_, err := svc.Search(ctx, req)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "search query is empty and no filters provided")
	mockTokenizerService.AssertExpectations(t)
	mockQueryBuilder.AssertNotCalled(t, "BuildSQL", mock.Anything, mock.Anything, mock.Anything)
	mockQueryExecutor.AssertNotCalled(t, "Execute", mock.Anything, mock.Anything, mock.Anything)
	mockRanker.AssertNotCalled(t, "Rank", mock.Anything, mock.Anything, mock.Anything)
	mockLog.Logs = nil
}

func TestSearchService_Search_QueryBuilderError(t *testing.T) {
	mockQueryBuilder, mockQueryExecutor, mockRanker, mockTokenizerService, mockIndexService, _, svc := setupSearchService(t)

	ctx := context.Background()
	req := &search.SearchRequest{
		Query:  "test",
		Tables: []string{"products"},
	}

	mockTokenizerService.On("GetTokenizer", mock.Anything, mock.Anything).Return(&TestTokenizer{}, nil).Once()
	mockIndexService.On("GetIndexMetadata", mock.Anything, mock.Anything).Return([]*index.IndexMetadata{{TableName: "products", ColumnName: "col1"}}, nil).Once()
	mockQueryBuilder.On("BuildSQL", mock.Anything, mock.Anything, mock.Anything).Return("", fmt.Errorf("SQL build error")).Once()

	_, err := svc.Search(ctx, req)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to build SQL query")
	mockTokenizerService.AssertExpectations(t)
	mockIndexService.AssertExpectations(t)
	mockQueryBuilder.AssertExpectations(t)
	mockQueryExecutor.AssertNotCalled(t, "Execute", mock.Anything, mock.Anything, mock.Anything)
	mockRanker.AssertNotCalled(t, "Rank", mock.Anything, mock.Anything, mock.Anything)
	assert.Contains(t, mockLog.Logs, "ERROR: Failed to build SQL query")
	mockLog.Logs = nil
}

func TestSearchService_Search_QueryExecutorError(t *testing.T) {
	mockQueryBuilder, mockQueryExecutor, mockRanker, mockTokenizerService, mockIndexService, _, svc := setupSearchService(t)

	ctx := context.Background()
	req := &search.SearchRequest{
		Query:  "test",
		Tables: []string{"products"},
	}

	mockTokenizerService.On("GetTokenizer", mock.Anything, mock.Anything).Return(&TestTokenizer{}, nil).Once()
	mockIndexService.On("GetIndexMetadata", mock.Anything, mock.Anything).Return([]*index.IndexMetadata{{TableName: "products", ColumnName: "col1"}}, nil).Once()
	mockQueryBuilder.On("BuildSQL", mock.Anything, mock.Anything, mock.Anything).Return("SELECT * FROM products", nil).Once()
	mockQueryExecutor.On("Execute", mock.Anything, mock.Anything, mock.Anything).Return(nil, fmt.Errorf("DB connection error")).Once()

	_, err := svc.Search(ctx, req)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to execute database query")
	mockTokenizerService.AssertExpectations(t)
	mockIndexService.AssertExpectations(t)
	mockQueryBuilder.AssertExpectations(t)
	mockQueryExecutor.AssertExpectations(t)
	mockRanker.AssertNotCalled(t, "Rank", mock.Anything, mock.Anything, mock.Anything)
	assert.Contains(t, mockLog.Logs, "ERROR: Failed to execute SQL query")
	mockLog.Logs = nil
}

func TestSearchService_Search_RankerError(t *testing.T) {
	mockQueryBuilder, mockQueryExecutor, mockRanker, mockTokenizerService, mockIndexService, _, svc := setupSearchService(t)

	ctx := context.Background()
	req := &search.SearchRequest{
		Query:  "test",
		Tables: []string{"products"},
	}

	mockTokenizerService.On("GetTokenizer", mock.Anything, mock.Anything).Return(&TestTokenizer{}, nil).Once()
	mockIndexService.On("GetIndexMetadata", mock.Anything, mock.Anything).Return([]*index.IndexMetadata{{TableName: "products", ColumnName: "col1"}}, nil).Once()
	mockQueryBuilder.On("BuildSQL", mock.Anything, mock.Anything, mock.Anything).Return("SELECT * FROM products", nil).Once()
	mockQueryExecutor.On("Execute", mock.Anything, mock.Anything, mock.Anything).Return([]map[string]interface{}{{"id": uint64(1), "data": "test"}}, nil).Once()
	mockRanker.On("Rank", mock.Anything, mock.Anything, mock.Anything).Return(fmt.Errorf("ranking algorithm error")).Once()

	_, err := svc.Search(ctx, req)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to rank search results")
	mockTokenizerService.AssertExpectations(t)
	mockIndexService.AssertExpectations(t)
	mockQueryBuilder.AssertExpectations(t)
	mockQueryExecutor.AssertExpectations(t)
	mockRanker.AssertExpectations(t)
	assert.Contains(t, mockLog.Logs, "ERROR: Failed to rank search results")
	mockLog.Logs = nil
}
