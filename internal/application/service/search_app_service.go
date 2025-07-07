package service

import (
	"context"
	"fmt"
	"strings"

	"github.com/turtacn/starseek/internal/common/constant"
	"github.com/turtacn/starseek/internal/common/errorx"
	"github.com/turtacn/starseek/internal/domain/model"
	"github.com/turtacn/starseek/internal/domain/repository"
	"github.com/turtacn/starseek/internal/infrastructure/cache"
	"go.uber.org/zap"
)

// SearchAppService is the application service responsible for the end-to-end search workflow.
type SearchAppService struct {
	queryProcessor *QueryProcessor
	taskScheduler  *TaskScheduler
	rankingService *RankingService
	searchRepo     repository.SearchRepository
	indexRepo      repository.IndexMetadataRepository // To retrieve index details needed by SearchRepository
	cache          cache.Cache
	logger         *zap.Logger
}

// NewSearchAppService creates a new SearchAppService instance.
func NewSearchAppService(
	qp *QueryProcessor,
	ts *TaskScheduler,
	rs *RankingService,
	sr repository.SearchRepository,
	ir repository.IndexMetadataRepository,
	cache cache.Cache,
	logger *zap.Logger,
) *SearchAppService {
	return &SearchAppService{
		queryProcessor: qp,
		taskScheduler:  ts,
		rankingService: rs,
		searchRepo:     sr,
		indexRepo:      ir,
		cache:          cache,
		logger:         logger.Named("SearchAppService"),
	}
}

// Search executes a full-text search based on the raw query and target fields.
func (s *SearchAppService) Search(ctx context.Context, rawQuery string, fields []string, page, pageSize int) (*model.SearchResult, error) {
	if rawQuery == "" {
		return nil, errorx.ErrBadRequest.Wrap(fmt.Errorf("search query cannot be empty"))
	}
	if len(fields) == 0 {
		return nil, errorx.ErrBadRequest.Wrap(fmt.Errorf("target fields cannot be empty"))
	}
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = constant.DefaultPageSize
	}

	// 1. Parse the raw query into a structured object
	parsedQuery, err := s.queryProcessor.Parse(rawQuery, fields)
	if err != nil {
		s.logger.Error("Failed to parse query", zap.Error(err), zap.String("query", rawQuery))
		return nil, errorx.NewError(errorx.ErrBadRequest.Code, "Failed to parse search query", err)
	}
	parsedQuery.Page = page
	parsedQuery.PageSize = pageSize

	s.logger.Info("Query parsed", zap.Any("parsed_query", parsedQuery))

	// 2. Schedule and execute initial search across relevant tables/fields
	// This step fetches RowIDs that match the criteria.
	initialResults, err := s.taskScheduler.Schedule(ctx, parsedQuery)
	if err != nil {
		s.logger.Error("Failed during task scheduling for initial search", zap.Error(err))
		return nil, errorx.NewError(errorx.ErrInternalServer.Code, "Failed to perform initial search", err)
	}

	if len(initialResults) == 0 {
		return &model.SearchResult{
			Items:     []*model.SearchResultItem{},
			TotalHits: 0,
			Page:      page,
			PageSize:  pageSize,
		}, nil
	}

	// Group RowIDs by table for efficient batch fetching of original data
	rowIDsByTable := make(map[string][]string)
	// We need to know which table each RowID belongs to.
	// For simplicity, for now, we assume all results from `taskScheduler.Schedule` are from the same "logical" table
	// or that FetchOriginalData can handle cross-table fetches (which it doesn't currently, it takes one table name).
	// A more robust solution would return table info from taskScheduler, or query for it.

	// For demonstration, let's assume the first field's table is the target table for fetching.
	// In a real multi-table scenario, you'd need to identify the source table for each RowID.
	// For simplicity, let's just pick the first relevant table from metadata.
	var targetTable string
	allMeta, err := s.indexRepo.ListAll(ctx)
	if err != nil {
		s.logger.Error("Failed to list all index metadata to determine target table", zap.Error(err))
		return nil, errorx.NewError(errorx.ErrInternalServer.Code, "Failed to determine target table", err)
	}
	for _, meta := range allMeta {
		if contains(fields, meta.ColumnName) { // Check if this table/column is relevant
			targetTable = meta.TableName
			break
		}
	}

	if targetTable == "" {
		s.logger.Warn("Could not determine a target table for fetching original data based on fields", zap.Strings("fields", fields))
		return &model.SearchResult{Items: []*model.SearchResultItem{}, TotalHits: 0, Page: page, PageSize: pageSize}, nil
	}

	// Extract RowIDs for the target table
	var currentTableRowIDs []string
	for _, item := range initialResults {
		currentTableRowIDs = append(currentTableRowIDs, item.RowID)
	}

	// 3. Fetch original data for the matched RowIDs
	// Apply pagination here for fetching original data if initialResults is large.
	// However, fetching all data first, then paginating and highlighting is also an option.
	// For simplicity, let's fetch all matched then paginate.
	// For large datasets, this might be inefficient.
	// It's better to push limit/offset down to the DB when retrieving by RowIDs if possible.

	// Apply pagination to initialResults *before* fetching original data for efficiency
	startIndex := (page - 1) * pageSize
	endIndex := startIndex + pageSize
	if startIndex >= len(initialResults) {
		return &model.SearchResult{Items: []*model.SearchResultItem{}, TotalHits: int64(len(initialResults)), Page: page, PageSize: pageSize}, nil
	}
	if endIndex > len(initialResults) {
		endIndex = len(initialResults)
	}
	paginatedRowIDs := make([]string, 0, endIndex-startIndex)
	for _, item := range initialResults[startIndex:endIndex] {
		paginatedRowIDs = append(paginatedRowIDs, item.RowID)
	}

	originalDataList, err := s.searchRepo.FetchOriginalData(ctx, targetTable, paginatedRowIDs)
	if err != nil {
		s.logger.Error("Failed to fetch original data", zap.Error(err), zap.String("table", targetTable), zap.Any("row_ids", paginatedRowIDs))
		return nil, errorx.NewError(errorx.ErrInternalServer.Code, "Failed to retrieve original data for search results", err)
	}

	// Map fetched data back to SearchResultItems
	resultsMap := make(map[string]map[string]interface{})
	for _, data := range originalDataList {
		if rowID, ok := data[s.getActualRowIDColumnName(targetTable)]; ok { // Assuming RowID is 'id' or other PK
			if strRowID, isStr := rowID.(string); isStr { // Ensure it's string
				resultsMap[strRowID] = data
			}
		}
	}

	// Update items with full original data and prepare for ranking/highlighting
	finalResults := make([]*model.SearchResultItem, 0, len(paginatedRowIDs))
	for _, item := range initialResults[startIndex:endIndex] { // Iterate over the paginated set
		if data, ok := resultsMap[item.RowID]; ok {
			item.OriginalData = data
			item.Highlights = s.generateHighlights(data, parsedQuery) // Generate highlights here
			finalResults = append(finalResults, item)
		}
	}

	// 4. Rank the results (now with full original data)
	rankedResults := s.rankingService.Rank(finalResults, parsedQuery)

	s.logger.Info("Search completed successfully", zap.Int("total_hits", len(initialResults)), zap.Int("returned_items", len(rankedResults)))

	return &model.SearchResult{
		Items:     rankedResults,
		TotalHits: int64(len(initialResults)), // Total hits before pagination
		Page:      page,
		PageSize:  pageSize,
	}, nil
}

// getActualRowIDColumnName is a helper to get the column name used as RowID.
// This needs to be dynamic based on the DBAdapter or index metadata.
func (s *SearchAppService) getActualRowIDColumnName(tableName string) string {
	// This should come from the DBAdapter or be configured for each table.
	// For StarRocks, it's typically the primary key column or 'id'.
	return "id" // Placeholder
}

// generateHighlights is a placeholder for highlighting logic.
// In a real system, this would use a text processing library to find and wrap
// matching keywords in original text with highlight tags (e.g., <em>).
func (s *SearchAppService) generateHighlights(data map[string]interface{}, query *model.ParsedQuery) map[string]string {
	highlights := make(map[string]string)
	combinedKeywords := append(query.Keywords, query.MustWords...)
	combinedKeywords = append(combinedKeywords, query.ShouldWords...) // Also consider should words for highlighting

	for _, field := range query.TargetFields {
		if val, ok := data[field]; ok {
			if text, isStr := val.(string); isStr {
				highlightedText := text
				// Simple highlighting: replace keywords with <em>keyword</em>
				for _, kw := range combinedKeywords {
					highlightedText = strings.ReplaceAll(highlightedText, kw, fmt.Sprintf("<em>%s</em>", kw))
					highlightedText = strings.ReplaceAll(highlightedText, strings.ToLower(kw), fmt.Sprintf("<em>%s</em>", strings.ToLower(kw)))
					highlightedText = strings.ReplaceAll(highlightedText, strings.ToUpper(kw), fmt.Sprintf("<em>%s</em>", strings.ToUpper(kw)))
				}
				// Also highlight field-specific keywords
				for _, fq := range query.FieldQueries {
					if fq.FieldName == field {
						for _, kw := range fq.Keywords {
							highlightedText = strings.ReplaceAll(highlightedText, kw, fmt.Sprintf("<em>%s</em>", kw))
							highlightedText = strings.ReplaceAll(highlightedText, strings.ToLower(kw), fmt.Sprintf("<em>%s</em>", strings.ToLower(kw)))
							highlightedText = strings.ReplaceAll(highlightedText, strings.ToUpper(kw), fmt.Sprintf("<em>%s</em>", strings.ToUpper(kw)))
						}
						if fq.IsPhrase {
							phrase := strings.Join(fq.Keywords, " ")
							highlightedText = strings.ReplaceAll(highlightedText, phrase, fmt.Sprintf("<em>%s</em>", phrase))
						}
					}
				}
				highlights[field] = highlightedText
			}
		}
	}
	return highlights
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

//Personal.AI order the ending
