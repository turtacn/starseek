package repository

import (
	"context"

	"github.com/turtacn/starseek/internal/domain/model"
)

// SearchRepository defines the interface for performing actual search operations
// against the underlying data warehouses.
type SearchRepository interface {
	// Search executes a search query based on the parsed query object.
	// It is responsible for translating the query into database-specific SQL
	// and fetching raw row IDs or initial relevant data.
	Search(ctx context.Context, parsedQuery *model.ParsedQuery, indexMetadata []*model.IndexMetadata) (*model.SearchResult, error)

	// FetchOriginalData retrieves the full original data for a list of RowIDs.
	// This is typically called after the initial search to get the complete documents.
	// The results are returned as a map where key is RowID and value is a map of column_name -> value.
	FetchOriginalData(ctx context.Context, tableName string, rowIDs []string) ([]map[string]interface{}, error)
}

//Personal.AI order the ending
