package service

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/turtacn/starseek/internal/common/constant"
	"github.com/turtacn/starseek/internal/common/errorx"
	"github.com/turtacn/starseek/internal/domain/model"
	"github.com/turtacn/starseek/internal/domain/repository"
	"github.com/turtacn/starseek/internal/infrastructure/cache"
	"go.uber.org/zap"
)

// TaskScheduler is a domain service that orchestrates concurrent data warehouse queries
// and manages caching of intermediate results (like RowIDs for keywords).
type TaskScheduler struct {
	searchRepo  repository.SearchRepository
	indexRepo   repository.IndexMetadataRepository
	cache       cache.Cache
	logger      *zap.Logger
	concurrency chan struct{} // Controls concurrent database queries
}

// NewTaskScheduler creates a new TaskScheduler instance.
func NewTaskScheduler(
	searchRepo repository.SearchRepository,
	indexRepo repository.IndexMetadataRepository,
	cache cache.Cache,
	logger *zap.Logger,
) *TaskScheduler {
	return &TaskScheduler{
		searchRepo:  searchRepo,
		indexRepo:   indexRepo,
		cache:       cache,
		logger:      logger.Named("TaskScheduler"),
		concurrency: make(chan struct{}, constant.MaxConcurrentDBQueries), // Initialize semaphore
	}
}

// Schedule orchestrates the execution of a parsed search query.
// It performs concurrent lookups for keywords and aggregates RowIDs.
// The result is a consolidated list of `SearchResultItem` containing only `RowID` and potentially a preliminary `Score`.
func (ts *TaskScheduler) Schedule(ctx context.Context, query *model.ParsedQuery) ([]*model.SearchResultItem, error) {
	// For each target field, we need to get its index metadata to pass to the adapter.
	// This can be cached in a real system.
	allIndexMetadata, err := ts.indexRepo.ListAll(ctx) // Fetch all for lookup efficiency
	if err != nil {
		return nil, errorx.NewError(errorx.ErrDatabase.Code, "failed to get all index metadata for scheduling", err)
	}
	indexMetadataMap := make(map[string]map[string]*model.IndexMetadata) // table -> column -> metadata
	for _, meta := range allIndexMetadata {
		if _, ok := indexMetadataMap[meta.TableName]; !ok {
			indexMetadataMap[meta.TableName] = make(map[string]*model.IndexMetadata)
		}
		indexMetadataMap[meta.TableName][meta.ColumnName] = meta
	}

	// Prepare concurrent tasks for each relevant field and query type
	var (
		wg               sync.WaitGroup
		resultsChan      = make(chan map[string]interface{}, len(query.TargetFields)*2) // Buffered channel for results
		errChan          = make(chan error, 1)
		mu               sync.Mutex              // Protects final aggregated results
		aggregatedRowIDs = make(map[string]bool) // Using map to store unique RowIDs
	)

	// Combine keywords from various parsed query parts for initial RowID fetching
	// For simplicity, we'll run one search per table that has *any* of the target fields.
	// A more refined approach would query each field independently and merge results.

	// Determine unique tables involved
	tablesToQuery := make(map[string]bool)
	for _, field := range query.TargetFields {
		parts := strings.SplitN(field, ".", 2) // Assuming "table.column" or "column"
		tableName := parts[0]
		if len(parts) == 2 {
			tableName = parts[0]
		} else {
			// If only column is given, try to find a table from available metadata
			for tbl, colMap := range indexMetadataMap {
				if _, ok := colMap[field]; ok {
					tableName = tbl
					break
				}
			}
			if tableName == "" { // If still no table found, skip this field
				ts.logger.Warn("Could not determine table for field, skipping", zap.String("field", field))
				continue
			}
		}
		tablesToQuery[tableName] = true
	}

	if len(tablesToQuery) == 0 {
		return []*model.SearchResultItem{}, nil // No tables to query
	}

	// For each table, build a composite search query
	for tableName := range tablesToQuery {
		wg.Add(1)
		ts.concurrency <- struct{}{} // Acquire token
		go func(tbl string) {
			defer wg.Done()
			defer func() { <-ts.concurrency }() // Release token

			tableMeta := make([]*model.IndexMetadata, 0)
			if colMetas, ok := indexMetadataMap[tbl]; ok {
				for _, meta := range colMetas {
					// Only include metadata for the target fields
					if contains(query.TargetFields, meta.ColumnName) {
						tableMeta = append(tableMeta, meta)
					}
				}
			}

			if len(tableMeta) == 0 {
				ts.logger.Info("No relevant index metadata for table, skipping search", zap.String("table", tbl))
				return
			}

			// Call SearchRepository.Search for the table.
			// This will internally use the DBAdapter to build and execute the SQL.
			// The SearchRepository should aggregate results for the entire table.
			results, err := ts.searchRepo.Search(ctx, query, tableMeta)
			if err != nil {
				select {
				case errChan <- errorx.NewError(errorx.ErrDatabase.Code, fmt.Sprintf("failed to search table %s", tbl), err):
				default:
					// Error already sent
				}
				return
			}

			mu.Lock()
			for _, item := range results.Items {
				aggregatedRowIDs[item.RowID] = true
			}
			mu.Unlock()
			ts.logger.Debug("Fetched RowIDs for table", zap.String("table", tbl), zap.Int("count", len(results.Items)))

		}(tableName)
	}

	wg.Wait()
	close(resultsChan) // Close channel after all goroutines are done

	select {
	case err := <-errChan:
		return nil, err // Return the first error encountered
	default:
		// No errors
	}

	finalItems := make([]*model.SearchResultItem, 0, len(aggregatedRowIDs))
	for rowID := range aggregatedRowIDs {
		finalItems = append(finalItems, &model.SearchResultItem{RowID: rowID})
	}

	ts.logger.Info("Task scheduling complete", zap.Int("total_unique_row_ids", len(finalItems)))
	return finalItems, nil
}

// contains helper function
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

//Personal.AI order the ending
