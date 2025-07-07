package starrocks

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/turtacn/starseek/internal/adapter"
	"github.com/turtacn/starseek/internal/common/enum"
	"github.com/turtacn/starseek/internal/common/errorx"
	"github.com/turtacn/starseek/internal/domain/model"
)

// StarRocksAdapter implements the DBAdapter interface for StarRocks.
type StarRocksAdapter struct {
	db *sql.DB
}

// NewStarRocksAdapter creates a new StarRocksAdapter.
func NewStarRocksAdapter(db *sql.DB) adapter.DBAdapter {
	return &StarRocksAdapter{db: db}
}

// Name returns the adapter name.
func (s *StarRocksAdapter) Name() string {
	return "starrocks"
}

// GetRowIDColumn returns the typical row ID column for StarRocks.
// In StarRocks, `_row_id` is an internal ID, but for user-facing primary keys,
// it's usually best to use the actual primary key. For this example, we'll assume
// the first column listed as primary key or just a general `id` column.
// For simplicity, we assume `id` or `rowid` for now, or the first column in the schema.
func (s *StarRocksAdapter) GetRowIDColumn(tableName string) (string, error) {
	// In a real scenario, you'd query information_schema or similar for PK.
	// For now, let's assume 'id' is a common primary key or rely on the application to specify.
	// StarRocks internal _row_id is not directly exposed for MATCH queries.
	// Often, users will have their own unique ID column.
	// Let's assume a column named 'id' exists.
	return "id", nil
}

// BuildSearchSQL generates StarRocks-specific SQL for full-text search.
// StarRocks uses the `MATCH` function for inverted indexes.
func (s *StarRocksAdapter) BuildSearchSQL(tableName string, columnMetadata *model.IndexMetadata, parsedQuery *model.ParsedQuery) (sqlStr string, args []interface{}, err error) {
	if columnMetadata == nil || columnMetadata.IndexType == enum.None {
		return "", nil, errorx.NewError(
			errorx.ErrBadRequest.Code,
			fmt.Sprintf("no full-text index configured for table %s, column %s", tableName, columnMetadata.ColumnName),
		)
	}

	var matchClauses []string
	var queryParts []string // For building complex match queries

	// Combine keywords for simple MATCH queries
	allKeywords := append(parsedQuery.Keywords, parsedQuery.MustWords...)
	allKeywords = append(allKeywords, parsedQuery.ShouldWords...) // ShouldWords might be used differently for ranking

	if len(allKeywords) > 0 {
		// Basic MATCH(column, 'keyword1 | keyword2') for OR logic, or 'keyword1 & keyword2' for AND.
		// We'll use simple OR logic for now and rely on ranking for AND preference.
		keywordQuery := strings.Join(allKeywords, " | ") // Default to OR for broad matching
		matchClauses = append(matchClauses, fmt.Sprintf("MATCH_ANY(%s, ?)", columnMetadata.ColumnName))
		args = append(args, keywordQuery)
	}

	// Handle field-specific queries (if the current column is one of them)
	for _, fq := range parsedQuery.FieldQueries {
		if fq.FieldName == columnMetadata.ColumnName && len(fq.Keywords) > 0 {
			fieldKeywordQuery := strings.Join(fq.Keywords, " | ")
			// For phrase search, StarRocks MATCH can use `"phrase"` syntax
			if fq.IsPhrase {
				fieldKeywordQuery = fmt.Sprintf(`"%s"`, strings.Join(fq.Keywords, " "))
			}
			matchClauses = append(matchClauses, fmt.Sprintf("MATCH_ANY(%s, ?)", fq.FieldName))
			args = append(args, fieldKeywordQuery)
		}
	}

	if len(matchClauses) == 0 {
		return "", nil, fmt.Errorf("no valid search terms for column %s", columnMetadata.ColumnName)
	}

	// For `NOT` words, StarRocks `MATCH` doesn't have direct exclusion, often handled by `WHERE NOT MATCH`.
	// For simplicity, we omit NOT logic in `MATCH` for now or assume it's handled post-query.
	// A more advanced solution would involve subqueries or complex WHERE clauses.

	// Select the row ID column. Assume "id" as the primary key.
	// In some cases, `_row_id` might be used for internal joining, but it's not always stable or directly queryable.
	// We recommend users have a stable primary key for joining.
	rowIDCol, err := s.GetRowIDColumn(tableName)
	if err != nil {
		return "", nil, err
	}

	sqlStr = fmt.Sprintf("SELECT %s FROM %s WHERE %s",
		rowIDCol,
		tableName,
		strings.Join(matchClauses, " OR "), // Combine multiple MATCH clauses with OR
	)

	return sqlStr, args, nil
}

// ExecuteSearch executes a given SQL query against StarRocks and returns the results.
func (s *StarRocksAdapter) ExecuteSearch(ctx context.Context, sqlQuery string, args []interface{}) ([]map[string]interface{}, error) {
	rows, err := s.db.QueryContext(ctx, sqlQuery, args...)
	if err != nil {
		return nil, errorx.NewError(errorx.ErrDatabase.Code, "failed to execute search query", err)
	}
	defer rows.Close()

	columns, err := rows.Columns()
	if err != nil {
		return nil, errorx.NewError(errorx.ErrDatabase.Code, "failed to get columns from query result", err)
	}

	var results []map[string]interface{}
	for rows.Next() {
		values := make([]interface{}, len(columns))
		valuePtrs := make([]interface{}, len(columns))
		for i := range values {
			valuePtrs[i] = &values[i]
		}

		if err := rows.Scan(valuePtrs...); err != nil {
			return nil, errorx.NewError(errorx.ErrDatabase.Code, "failed to scan row", err)
		}

		row := make(map[string]interface{})
		for i, col := range columns {
			val := values[i]
			// Handle NULL values, convert byte slices to string for text
			if b, ok := val.([]byte); ok {
				row[col] = string(b)
			} else {
				row[col] = val
			}
		}
		results = append(results, row)
	}

	if err = rows.Err(); err != nil {
		return nil, errorx.NewError(errorx.ErrDatabase.Code, "error iterating rows", err)
	}

	return results, nil
}

// FetchMetadata attempts to discover existing full-text index metadata from StarRocks.
// This is typically done by querying `information_schema.columns` or similar.
// For a production-ready system, this would involve parsing `SHOW CREATE TABLE` output or `information_schema.columns`
// to identify columns with `FULLTEXT` or `NGRAM` indexes.
// This example provides a simplistic placeholder.
func (s *StarRocksAdapter) FetchMetadata(ctx context.Context) ([]*model.IndexMetadata, error) {
	// A more robust implementation would query `information_schema.columns` or `SHOW CREATE TABLE`
	// and parse the column definitions for `INDEX TYPE` or `WITH PROPERTIES` clauses.
	// StarRocks: `SHOW CREATE TABLE tbl_name;` looks for `COMMENT 'fulltext index'`.
	// For this example, let's just return a hardcoded dummy index for illustration.
	// In real-world, it's complex and might involve direct parsing of DDL or specific system tables.

	// Example: Query to get table and column names (not full index metadata)
	// rows, err := s.db.QueryContext(ctx, "SELECT TABLE_NAME, COLUMN_NAME, COLUMN_TYPE FROM information_schema.columns WHERE TABLE_SCHEMA = DATABASE();")
	// For actual index type, StarRocks requires parsing SHOW CREATE TABLE or using internal tables not readily exposed.

	// For demonstration, assume we know some columns are indexed.
	// In a real scenario, this would involve complex schema introspection.
	// Let's assume a common table 'articles' with 'title' and 'content' indexed.
	// This part is highly dependent on how StarRocks exposes its index metadata.
	// Currently, it's not as straightforward as querying `information_schema.indexes` for fulltext.

	// Placeholder: This would need a sophisticated parser for `SHOW CREATE TABLE` output or
	// direct queries to BE/FE internal tables, which is out of scope for a basic example.
	// It's often easier for users to register indexes manually or from a known schema.
	// Returning an empty slice for now or a very basic predefined one.
	return []*model.IndexMetadata{
		{
			TableName:  "articles",
			ColumnName: "title",
			IndexType:  enum.Inverted,
			Tokenizer:  enum.Chinese, // Or English, based on table definition
			DataType:   "TEXT",
		},
		{
			TableName:  "articles",
			ColumnName: "content",
			IndexType:  enum.Inverted,
			Tokenizer:  enum.Chinese,
			DataType:   "TEXT",
		},
		{
			TableName:  "products",
			ColumnName: "name",
			IndexType:  enum.Inverted,
			Tokenizer:  enum.English,
			DataType:   "VARCHAR",
		},
	}, nil
}

//Personal.AI order the ending
