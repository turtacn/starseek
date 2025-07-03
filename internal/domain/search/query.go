package search

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"go.uber.org/zap"

	"github.com/turtacn/starseek/internal/common/enum" // 引入 enum 包，包含 DatabaseType, IndexType
	ferrors "github.com/turtacn/starseek/internal/common/errors"
	"github.com/turtacn/starseek/internal/domain/index"            // 引入 IndexMetadata, IndexService
	"github.com/turtacn/starseek/internal/domain/tokenizer"        // 引入 TokenizerService
	"github.com/turtacn/starseek/internal/infrastructure/database" // 引入 DBClientFactory
	"github.com/turtacn/starseek/internal/infrastructure/logger"   // 引入日志接口
)

// QueryBuilder 接口定义了根据 SearchQuery 和索引元数据构建数据库查询（SQL）的能力。
type QueryBuilder interface {
	// BuildSQL 方法根据解析后的 SearchQuery 和索引元数据生成适合底层数据库的 SQL 语句。
	// 它考虑查询类型、关键词、字段、表、过滤器和分页参数。
	BuildSQL(ctx context.Context, query *SearchQuery, indexedColumns []*index.IndexMetadata) (string, error)
}

// QueryExecutor 接口定义了执行数据库查询并返回原始结果的能力。
type QueryExecutor interface {
	// Execute 方法执行给定的 SQL 语句，并根据指定的数据库类型从对应的 DBClient 获取结果。
	// 它返回一个 map 切片，每个 map 代表一行数据。
	Execute(ctx context.Context, sql string, dbType enum.DatabaseType) ([]map[string]interface{}, error)
}

// ==============================================================================
// QueryBuilder 实现
// ==============================================================================

// queryBuilder 是 QueryBuilder 接口的实现。
// 它负责根据 SearchQuery 构建 SQL 语句。
type queryBuilder struct {
	// 目前 BuildSQL 方法直接接收 tokenized keywords 和 indexedColumns，
	// 因此 tokenizerService 和 indexService 在这里作为未来扩展的依赖，
	// 例如用于动态获取索引元数据或在 BuildSQL 之前进行额外的分词校验。
	// 对于当前需求，主要逻辑基于传入的 *SearchQuery 和 []*index.IndexMetadata。
	tokenizerService tokenizer.TokenizerService
	indexService     index.IndexService
	log              logger.Logger
}

// NewQueryBuilder 创建并返回一个新的 QueryBuilder 实例。
func NewQueryBuilder(
	tokenizerSvc tokenizer.TokenizerService,
	indexSvc index.IndexService,
	log logger.Logger,
) (QueryBuilder, error) {
	if log == nil {
		return nil, ferrors.NewInternalError("logger is nil", nil)
	}
	if tokenizerSvc == nil {
		return nil, ferrors.NewInternalError("tokenizer service is nil", nil)
	}
	if indexSvc == nil {
		return nil, ferrors.NewInternalError("index service is nil", nil)
	}
	log.Info("QueryBuilder initialized.")
	return &queryBuilder{
		tokenizerService: tokenizerSvc,
		indexService:     indexSvc,
		log:              log,
	}, nil
}

// BuildSQL 方法根据 SearchQuery 和索引元数据生成 SQL 语句。
// 该方法需要处理复杂的逻辑，包括：
// 1. 根据 SearchQuery.Tables 确定要查询的表。
// 2. 根据 SearchQuery.Fields 和 indexedColumns 确定要搜索的列。
// 3. 根据 SearchQuery.QueryType (MATCH_ALL, MATCH_ANY, FIELD_SPECIFIC) 构建 WHERE 子句。
// 4. 处理 SearchQuery.Filters。
// 5. 添加分页 (LIMIT/OFFSET)。
// 6. 考虑排序（尽管当前模型中未明确定义，但可以在此扩展）。
// 注意：本示例提供一个简化版，以展示结构和主要思想。
// 实际生产环境的全文搜索 SQL 会非常复杂，通常需要专门的全文索引引擎如 ClickHouse 的 BM25/Okapi BM25, FTS5, Elasticsearch, Lucene等。
func (qb *queryBuilder) BuildSQL(ctx context.Context, query *SearchQuery, indexedColumns []*index.IndexMetadata) (string, error) {
	if query == nil {
		return "", ferrors.NewInternalError("search query is nil", nil)
	}
	if len(query.Tables) == 0 {
		return "", ferrors.NewBadRequestError("no tables specified in search query", nil)
	}
	if len(query.Keywords) == 0 {
		// 如果没有关键词，可能返回空结果或根据过滤条件返回
		// 这里简单处理为只返回过滤结果或空
		if len(query.Filters) == 0 {
			return "", ferrors.NewBadRequestError("no keywords and no filters specified in search query", nil)
		}
	}

	var sqlParts []string
	var whereClauses []string
	var selectedColumns []string // Store columns that are actually selected for a more refined query

	// 1. Select Columns: Select all columns from data.
	// In a real system, you might select only indexed columns + common metadata.
	// For simplicity, we assume we need to retrieve all original columns for `Data` field in SearchResult.
	// Or, more efficiently, select '*' and filter/project later.
	// Let's assume we select * for now, and filter/project original columns later.
	// For highlighting, we might need specific column content, but that's a post-processing step.
	selectedColumns = append(selectedColumns, "*") // Select all columns for simplicity

	// 2. Build WHERE clause based on keywords and fields
	// Group indexed columns by table name
	tableIndexedColumns := make(map[string][]*index.IndexMetadata)
	for _, ic := range indexedColumns {
		tableIndexedColumns[ic.TableName] = append(tableIndexedColumns[ic.TableName], ic)
	}

	var tableConditions []string
	for _, tableName := range query.Tables {
		cols := tableIndexedColumns[tableName]
		if len(cols) == 0 {
			qb.log.Warn("No indexed columns found for table", zap.String("table", tableName))
			// Continue, or potentially return error if table must have indexed columns
			continue
		}

		var fieldConditions []string
		for _, keyword := range query.Keywords {
			var keywordFieldConditions []string
			// If specific fields are requested, only search in those, otherwise search all indexed columns for the table.
			targetCols := cols
			if len(query.Fields) > 0 {
				var filteredCols []*index.IndexMetadata
				for _, col := range cols {
					for _, field := range query.Fields {
						if col.ColumnName == field {
							filteredCols = append(filteredCols, col)
							break
						}
					}
				}
				targetCols = filteredCols
			}

			if len(targetCols) == 0 && len(query.Fields) > 0 {
				qb.log.Warn("No matching indexed columns found for specified fields in table",
					zap.String("table", tableName), zap.Strings("fields", query.Fields))
				// If no target columns for specific fields, this keyword won't find anything in this table.
				continue
			}

			for _, col := range targetCols {
				// ClickHouse SQL for text search: `positionCaseInsensitive(column_name, 'keyword') > 0`
				// Or using `LIKE` for simple substring search, but it's slow.
				// For ClickHouse's full-text search, it would involve more complex functions or materialized views.
				// For this example, we use `LIKE` for simplicity.
				keywordFieldConditions = append(keywordFieldConditions, fmt.Sprintf("%s LIKE '%%%s%%'", col.ColumnName, keyword))
			}
			if len(keywordFieldConditions) > 0 {
				fieldConditions = append(fieldConditions, "("+strings.Join(keywordFieldConditions, " OR ")+")")
			}
		}

		if len(fieldConditions) > 0 {
			// Combine field conditions based on query type
			var keywordCombineOp string
			switch query.QueryType {
			case MATCH_ALL:
				keywordCombineOp = " AND "
			case MATCH_ANY:
				keywordCombineOp = " OR "
			default: // Default to MATCH_ANY if not specified
				qb.log.Warn("Unsupported query type, defaulting to MATCH_ANY", zap.String("query_type", query.QueryType.String()))
				keywordCombineOp = " OR "
			}
			tableConditions = append(tableConditions, "("+strings.Join(fieldConditions, keywordCombineOp)+")")
		}
	}

	// Combine all table conditions (if multiple tables, typically OR them)
	if len(tableConditions) > 0 {
		whereClauses = append(whereClauses, "("+strings.Join(tableConditions, " OR ")+")")
	}

	// 3. Add Filter clauses
	var filterClauses []string
	for field, value := range query.Filters {
		filterClauses = append(filterClauses, fmt.Sprintf("%s = '%s'", field, value))
	}
	if len(filterClauses) > 0 {
		whereClauses = append(whereClauses, strings.Join(filterClauses, " AND "))
	}

	// 4. Construct final WHERE clause
	if len(whereClauses) > 0 {
		sqlParts = append(sqlParts, "WHERE "+strings.Join(whereClauses, " AND "))
	}

	// 5. Add Pagination
	limit := query.PageSize
	offset := (query.Page - 1) * query.PageSize // Page is 1-indexed
	if limit <= 0 {                             // Default limit if not set or invalid
		limit = 20
	}
	if offset < 0 { // Ensure offset is not negative
		offset = 0
	}

	// SQL for fetching data and pagination
	// Note: Score calculation and MinScore filtering would typically happen within the database
	// using complex UDFs or materialized views for full-text search engines.
	// For simple LIKE queries, scoring is usually done in application layer.
	// We will omit score for now in the SQL directly.
	mainSQL := fmt.Sprintf("SELECT %s FROM %s %s LIMIT %d OFFSET %d",
		strings.Join(selectedColumns, ", "),
		strings.Join(query.Tables, ", "), // Simple comma-separated tables. For JOINs, would be more complex.
		strings.Join(sqlParts, " "),
		limit,
		offset,
	)

	qb.log.Debug("Generated SQL for search query", zap.String("sql", mainSQL), zap.Any("query", query))
	return mainSQL, nil
}

// ==============================================================================
// QueryExecutor 实现
// ==============================================================================

// sqlQueryExecutor 是 QueryExecutor 接口的实现。
// 它通过 DBClientFactory 管理的 DBClient 实例来执行 SQL。
type sqlQueryExecutor struct {
	dbClientFactory database.DBClientFactory
	log             logger.Logger
}

// NewSQLQueryExecutor 创建并返回一个新的 QueryExecutor 实例。
func NewSQLQueryExecutor(dbClientFactory database.DBClientFactory, log logger.Logger) (QueryExecutor, error) {
	if log == nil {
		return nil, ferrors.NewInternalError("logger is nil", nil)
	}
	if dbClientFactory == nil {
		return nil, ferrors.NewInternalError("database client factory is nil", nil)
	}
	log.Info("SQLQueryExecutor initialized.")
	return &sqlQueryExecutor{
		dbClientFactory: dbClientFactory,
		log:             log,
	}, nil
}

// Execute 方法执行给定的 SQL 语句。
func (sqe *sqlQueryExecutor) Execute(ctx context.Context, sql string, dbType enum.DatabaseType) ([]map[string]interface{}, error) {
	if sql == "" {
		return nil, ferrors.NewBadRequestError("SQL string is empty", nil)
	}

	dbClient, err := sqe.dbClientFactory.GetDBClient(dbType)
	if err != nil {
		sqe.log.Error("Failed to get DBClient for database type",
			zap.String("db_type", dbType.String()), zap.Error(err))
		return nil, ferrors.NewExternalServiceError(fmt.Sprintf("failed to get DBClient for %s: %v", dbType.String(), err), err)
	}

	sqe.log.Debug("Executing SQL query", zap.String("sql", sql), zap.String("db_type", dbType.String()))
	rows, err := dbClient.Query(ctx, sql)
	if err != nil {
		sqe.log.Error("Failed to execute SQL query",
			zap.String("sql", sql), zap.String("db_type", dbType.String()), zap.Error(err))
		return nil, ferrors.NewExternalServiceError(fmt.Sprintf("failed to execute query: %v", err), err)
	}
	defer rows.Close()

	columns, err := rows.Columns()
	if err != nil {
		sqe.log.Error("Failed to get columns from query result", zap.Error(err))
		return nil, ferrors.NewInternalError("failed to get columns from query result", err)
	}

	var results []map[string]interface{}
	for rows.Next() {
		row := make(map[string]interface{})
		columnPointers := make([]interface{}, len(columns))
		for i := range columns {
			var val interface{}
			columnPointers[i] = &val
		}

		if err := rows.Scan(columnPointers...); err != nil {
			sqe.log.Error("Failed to scan row from query result", zap.Error(err))
			return nil, ferrors.NewInternalError("failed to scan row from query result", err)
		}

		for i, colName := range columns {
			val := *(columnPointers[i].(*interface{}))
			row[colName] = val
		}
		results = append(results, row)
	}

	if err = rows.Err(); err != nil {
		sqe.log.Error("Error iterating over query results", zap.Error(err))
		return nil, ferrors.NewExternalServiceError("error iterating over query results", err)
	}

	sqe.log.Debug("SQL query executed successfully", zap.Int("row_count", len(results)))
	return results, nil
}
