package search

import (
	"context"
	"fmt"
	"strings"
	"sync" // For concurrent tokenization if multiple fields/documents need it

	"go.uber.org/zap"

	"github.com/turtacn/starseek/internal/common/enum" // 引入 enum 包，包含 DatabaseType, IndexType
	ferrors "github.com/turtacn/starseek/internal/common/errors"
	"github.com/turtacn/starseek/internal/domain/index"          // 引入 IndexService, IndexMetadata
	"github.com/turtacn/starseek/internal/domain/tokenizer"      // 引入 TokenizerService, Tokenizer
	"github.com/turtacn/starseek/internal/infrastructure/cache"  // 引入 CacheClient
	"github.com/turtacn/starseek/internal/infrastructure/logger" // 引入 Logger 接口
)

// SearchService 接口定义了执行搜索操作的核心业务逻辑。
type SearchService interface {
	// Search 方法接收一个 SearchRequest，执行完整的搜索流程（查询解析、SQL生成、执行、排名、高亮），
	// 并返回 SearchResponse。
	Search(ctx context.Context, req *SearchRequest) (*SearchResponse, error)
}

// searchService 是 SearchService 接口的实现。
// 它协调各个组件来完成搜索流程。
type searchService struct {
	queryBuilder     QueryBuilder
	queryExecutor    QueryExecutor
	ranker           Ranker
	tokenizerService tokenizer.TokenizerService // 用于查询分词和文档内容分词
	indexService     index.IndexService         // 用于获取索引元数据
	log              logger.Logger
	cacheClient      cache.CacheClient // 作为 IDFProvider 的依赖，也可能用于其他缓存
}

// NewSearchService 创建并返回一个新的 SearchService 实例。
// 它需要所有搜索流程所需的依赖。
func NewSearchService(
	queryBuilder QueryBuilder,
	queryExecutor QueryExecutor,
	ranker Ranker,
	tokenizerService tokenizer.TokenizerService,
	indexService index.IndexService,
	log logger.Logger,
	cacheClient cache.CacheClient, // Not directly used by service, but passed to sub-components
) (SearchService, error) {
	if queryBuilder == nil {
		return nil, ferrors.NewInternalError("query builder is nil", nil)
	}
	if queryExecutor == nil {
		return nil, ferrors.NewInternalError("query executor is nil", nil)
	}
	if ranker == nil {
		return nil, ferrors.NewInternalError("ranker is nil", nil)
	}
	if tokenizerService == nil {
		return nil, ferrors.NewInternalError("tokenizer service is nil", nil)
	}
	if indexService == nil {
		return nil, ferrors.NewInternalError("index service is nil", nil)
	}
	if log == nil {
		return nil, ferrors.NewInternalError("logger is nil", nil)
	}
	if cacheClient == nil {
		return nil, ferrors.NewInternalError("cache client is nil", nil)
	}

	log.Info("SearchService initialized.")
	return &searchService{
		queryBuilder:     queryBuilder,
		queryExecutor:    queryExecutor,
		ranker:           ranker,
		tokenizerService: tokenizerService,
		indexService:     indexService,
		log:              log,
		cacheClient:      cacheClient,
	}, nil
}

// Search 方法实现了完整的搜索业务流程。
func (s *searchService) Search(ctx context.Context, req *SearchRequest) (*SearchResponse, error) {
	s.log.Info("Received search request", zap.Any("request", req))

	// 1. 获取查询使用的分词器 (假设根据请求语言或默认设置)
	// 这里可以根据 req.Language 或其他配置选择合适的 tokenizer
	// 为了简化，我们使用一个默认的英语分词器
	tokenizer, err := s.tokenizerService.GetTokenizer(enum.FULL_TEXT_INDEX, enum.ENGLISH_TOKENIZER.String()) // Or dynamically from config/req
	if err != nil {
		s.log.Error("Failed to get tokenizer", zap.Error(err))
		return nil, ferrors.NewInternalError("failed to initialize tokenizer for search", err)
	}

	// 2. 解析查询并分词，构建内部 SearchQuery 对象
	// 将用户输入的查询字符串进行分词
	keywords, err := tokenizer.Tokenize(req.Query)
	if err != nil {
		s.log.Error("Failed to tokenize search query", zap.String("query", req.Query), zap.Error(err))
		return nil, ferrors.NewInternalError("failed to tokenize search query", err)
	}
	if len(keywords) == 0 && len(req.Filters) == 0 {
		return nil, ferrors.NewBadRequestError("search query is empty and no filters provided", nil)
	}

	// 从 IndexService 获取相关的索引元数据
	// 假设这里获取所有表中所有全文本索引列的元数据
	// 实际应用中会根据 req.Tables 过滤
	var indexedColumns []*index.IndexMetadata
	for _, tableName := range req.Tables {
		tableIndexedCols, err := s.indexService.GetIndexMetadata(ctx, tableName)
		if err != nil {
			s.log.Warn("Could not get index metadata for table, skipping", zap.String("table", tableName), zap.Error(err))
			continue // Or return an error depending on strictness
		}
		indexedColumns = append(indexedColumns, tableIndexedCols...)
	}

	// 构建 SearchQuery
	searchQuery := &SearchQuery{
		Tables:    req.Tables,
		Keywords:  keywords,
		Fields:    req.Fields,
		Filters:   req.Filters,
		Page:      req.Page,
		PageSize:  req.PageSize,
		QueryType: req.QueryType,
		// MinScore, Language, SortOrder等可以在SearchRequest中定义并传递
	}
	s.log.Debug("Parsed search query", zap.Any("search_query", searchQuery))

	// 3. 调用 QueryBuilder 生成 SQL
	sqlQuery, err := s.queryBuilder.BuildSQL(ctx, searchQuery, indexedColumns)
	if err != nil {
		s.log.Error("Failed to build SQL query", zap.Any("search_query", searchQuery), zap.Error(err))
		return nil, ferrors.NewInternalError("failed to build SQL query", err)
	}
	s.log.Debug("Generated SQL query", zap.String("sql", sqlQuery))

	// 4. 调用 QueryExecutor 执行 SQL 查询获取原始结果
	// 假设 req.DatabaseType 指定了要查询的数据库类型
	// 或者从配置中读取默认数据库类型
	dbType := enum.CLICKHOUSE // Example: default to ClickHouse
	if req.DatabaseType != "" {
		parsedDBType, err := enum.ParseDatabaseType(req.DatabaseType)
		if err != nil {
			s.log.Warn("Invalid database type in request, using default", zap.String("requested_type", req.DatabaseType))
		} else {
			dbType = parsedDBType
		}
	}

	rawResults, err := s.queryExecutor.Execute(ctx, sqlQuery, dbType)
	if err != nil {
		s.log.Error("Failed to execute SQL query", zap.String("sql", sqlQuery), zap.String("db_type", dbType.String()), zap.Error(err))
		return nil, ferrors.NewExternalServiceError("failed to execute database query", err)
	}
	s.log.Debug("Raw query results obtained", zap.Int("count", len(rawResults)))

	// 转换 rawResults 为 []*SearchResult
	searchResults := make([]*SearchResult, len(rawResults))
	for i, rawRow := range rawResults {
		// Assume 'id' column exists for RowID
		rowID, ok := rawRow["id"]
		if !ok {
			// Fallback or error if 'id' is not present
			s.log.Warn("Row 'id' not found in result, using placeholder", zap.Any("row", rawRow))
			rowID = fmt.Sprintf("row_%d", i) // Placeholder
		}
		searchResults[i] = &SearchResult{
			RowID: fmt.Sprintf("%v", rowID),
			Data:  rawRow,
			Score: 0.0, // Initial score before ranking
		}
	}

	// 5. 调用 Ranker 对结果进行相关度排序
	err = s.ranker.Rank(ctx, searchResults, searchQuery)
	if err != nil {
		s.log.Error("Failed to rank search results", zap.Error(err))
		// Continue without ranking or return error based on business logic
		// For now, let's return an error, as ranking is critical.
		return nil, ferrors.NewInternalError("failed to rank search results", err)
	}
	s.log.Debug("Search results ranked")

	// 6. 调用 Highlighting Engine 对结果进行高亮 (待实现)
	// 高亮通常在搜索结果处理的最后阶段进行。
	// 这里可以定义一个 Highlighting 模块或服务。
	// For now, we'll just iterate and highlight if needed.
	// This would involve re-tokenizing original text fields and finding keyword matches.
	highlightedResults := make([]*SearchResult, len(searchResults))
	var wg sync.WaitGroup
	// Potentially limit concurrency for highlighting
	highlightConcurrencyLimit := 5 // Example limit
	sem := make(chan struct{}, highlightConcurrencyLimit)

	for i, res := range searchResults {
		wg.Add(1)
		sem <- struct{}{} // Acquire token
		go func(idx int, currentRes *SearchResult) {
			defer wg.Done()
			defer func() { <-sem }() // Release token

			highlightedRes := *currentRes // Create a copy
			highlightedRes.HighlightedData = make(map[string]string)

			for field, val := range currentRes.Data {
				strVal, ok := val.(string)
				if !ok || strVal == "" {
					continue
				}

				// Only highlight fields that were part of the original query fields
				// Or all string fields if no specific fields were given
				shouldHighlightField := false
				if len(searchQuery.Fields) == 0 {
					shouldHighlightField = true // Highlight all string fields if no fields specified
				} else {
					for _, qField := range searchQuery.Fields {
						if field == qField {
							shouldHighlightField = true
							break
						}
					}
				}

				if shouldHighlightField {
					// This is a simplified in-memory highlighting.
					// A real highlighting engine would be more robust (e.g., handling HTML, different markers).
					// It would also use the `tokenizer` to match tokens, not just substrings.
					highlightedText := strVal
					for _, keyword := range searchQuery.Keywords {
						// Simple substring replacement. Not robust for actual token-based highlighting.
						// For production, use a dedicated highlighting library/service.
						highlightedText = strings.ReplaceAll(highlightedText, keyword, fmt.Sprintf("<b>%s</b>", keyword))
						highlightedText = strings.ReplaceAll(highlightedText, strings.Title(keyword), fmt.Sprintf("<b>%s</b>", strings.Title(keyword))) // rudimentary case
					}
					highlightedRes.HighlightedData[field] = highlightedText
				}
			}
			highlightedResults[idx] = &highlightedRes
		}(i, res)
	}
	wg.Wait()
	s.log.Debug("Search results highlighted")

	// 7. 返回 SearchResponse
	response := &SearchResponse{
		TotalHits:    int64(len(searchResults)), // In a real system, this would be from DB count query
		Results:      highlightedResults,
		Page:         req.Page,
		PageSize:     req.PageSize,
		QueryLatency: 0, // Placeholder: measure actual latency
	}

	s.log.Info("Search completed successfully", zap.Int("total_hits", len(searchResults)), zap.Int("page", req.Page))
	return response, nil
}
