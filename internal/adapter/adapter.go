package adapter

import (
	"context"

	"github.com/turtacn/starseek/internal/domain/model"
)

// DBAdapter defines the common interface for different OLAP database adapters.
// This allows Starseek to interact with various data warehouses like StarRocks, Doris, ClickHouse.
type DBAdapter interface {
	// BuildSearchSQL generates the database-specific SQL query string and its arguments
	// based on the parsed search query. This query typically focuses on retrieving RowIDs
	// that match the search criteria.
	BuildSearchSQL(tableName string, columnMetadata *model.IndexMetadata, parsedQuery *model.ParsedQuery) (sql string, args []interface{}, err error)

	// ExecuteSearch executes a given SQL query and returns the raw results.
	// The results are expected to be a slice of maps, where each map represents a row
	// and keys are column names.
	ExecuteSearch(ctx context.Context, sql string, args []interface{}) ([]map[string]interface{}, error)

	// FetchMetadata attempts to discover existing full-text index metadata from the database.
	// This can be used for automatic index registration or synchronization.
	FetchMetadata(ctx context.Context) ([]*model.IndexMetadata, error)

	// GetRowIDColumn returns the name of the column that serves as the unique row identifier
	// for a given table in this database system (e.g., StarRocks might have `_row_id` or a primary key).
	GetRowIDColumn(tableName string) (string, error)

	// Name returns the name of the adapter (e.g., "starrocks", "doris").
	Name() string
}

//Personal.AI order the ending
