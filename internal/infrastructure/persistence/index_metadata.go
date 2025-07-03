package persistence

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql" // MySQL 驱动
	"go.uber.org/zap"                  // 用于日志字段

	ferrors "github.com/turtacn/starseek/internal/common/errors"
	"github.com/turtacn/starseek/internal/domain/enum"           // 引入枚举类型
	"github.com/turtacn/starseek/internal/domain/index"          // 引入 IndexMetadata 和 IndexMetadataRepository 接口
	"github.com/turtacn/starseek/internal/infrastructure/logger" // 引入日志接口
)

// sqlIndexMetadataRepository 是 IndexMetadataRepository 接口的 SQL 数据库实现。
type sqlIndexMetadataRepository struct {
	db  *sql.DB
	log logger.Logger
}

// NewSQLIndexMetadataRepository 创建并返回一个新的 sqlIndexMetadataRepository 实例。
// 它接收一个 *sql.DB 数据库连接和 Logger 接口。
func NewSQLIndexMetadataRepository(db *sql.DB, log logger.Logger) (index.IndexMetadataRepository, error) {
	if db == nil {
		return nil, ferrors.NewInternalError("database client is nil", nil)
	}

	repo := &sqlIndexMetadataRepository{
		db:  db,
		log: log,
	}

	// 确保表结构存在 (通常在生产环境通过数据库迁移工具管理)
	// 这里为了方便开发和测试，直接在初始化时尝试创建表
	if err := repo.ensureTableExists(context.Background()); err != nil {
		return nil, ferrors.NewInternalError("failed to ensure index_metadata table exists", err)
	}

	return repo, nil
}

// ensureTableExists 检查并创建 index_metadata 表。
func (r *sqlIndexMetadataRepository) ensureTableExists(ctx context.Context) error {
	createTableSQL := `
	CREATE TABLE IF NOT EXISTS index_metadata (
		id VARCHAR(36) PRIMARY KEY,
		table_name VARCHAR(255) NOT NULL,
		column_name VARCHAR(255) NOT NULL,
		index_type VARCHAR(50) NOT NULL,
		tokenizer VARCHAR(100),
		data_type VARCHAR(100) NOT NULL,
		description TEXT,
		created_at DATETIME NOT NULL,
		updated_at DATETIME NOT NULL,
		UNIQUE KEY uk_table_column (table_name, column_name)
	);`

	_, err := r.db.ExecContext(ctx, createTableSQL)
	if err != nil {
		r.log.Error("Failed to create index_metadata table", zap.Error(err))
		return err
	}
	r.log.Info("Ensured index_metadata table exists.")
	return nil
}

// Save 将索引元数据保存到数据库。如果ID存在则更新，否则插入。
func (r *sqlIndexMetadataRepository) Save(ctx context.Context, metadata *index.IndexMetadata) error {
	if metadata.ID == "" {
		return ferrors.NewInternalError("metadata ID cannot be empty for save operation", nil)
	}

	// 检查记录是否存在以决定是 INSERT 还是 UPDATE
	var exists bool
	query := `SELECT COUNT(*) FROM index_metadata WHERE id = ?`
	err := r.db.QueryRowContext(ctx, query, metadata.ID).Scan(&exists)
	if err != nil {
		r.log.Error(fmt.Sprintf("Failed to check existence of index metadata %s: %v", metadata.ID, err))
		return ferrors.NewExternalServiceError("failed to check existing index metadata", err)
	}

	metadata.UpdatedAt = time.Now() // 每次保存都更新更新时间

	if exists {
		// UPDATE
		updateSQL := `
			UPDATE index_metadata SET
				table_name = ?,
				column_name = ?,
				index_type = ?,
				tokenizer = ?,
				data_type = ?,
				description = ?,
				updated_at = ?
			WHERE id = ?`
		_, err := r.db.ExecContext(ctx, updateSQL,
			metadata.TableName, metadata.ColumnName, string(metadata.IndexType), metadata.Tokenizer,
			metadata.DataType, metadata.Description, metadata.UpdatedAt, metadata.ID)
		if err != nil {
			r.log.Error(fmt.Sprintf("Failed to update index metadata %s: %v", metadata.ID, err))
			return ferrors.NewExternalServiceError("failed to update index metadata", err)
		}
		r.log.Debug(fmt.Sprintf("Updated index metadata: %s", metadata.ID))
	} else {
		// INSERT
		metadata.CreatedAt = time.Now() // 仅在创建时设置创建时间
		insertSQL := `
			INSERT INTO index_metadata (
				id, table_name, column_name, index_type, tokenizer, data_type, description, created_at, updated_at
			) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`
		_, err := r.db.ExecContext(ctx, insertSQL,
			metadata.ID, metadata.TableName, metadata.ColumnName, string(metadata.IndexType), metadata.Tokenizer,
			metadata.DataType, metadata.Description, metadata.CreatedAt, metadata.UpdatedAt)
		if err != nil {
			r.log.Error(fmt.Sprintf("Failed to insert index metadata %s: %v", metadata.ID, err))
			// 检查是否是唯一约束冲突 (MySQL错误码1062)
			if strings.Contains(err.Error(), "Duplicate entry") && strings.Contains(err.Error(), "uk_table_column") {
				return ferrors.NewAlreadyExistsError(fmt.Sprintf("index for table %s, column %s already exists", metadata.TableName, metadata.ColumnName), err)
			}
			return ferrors.NewExternalServiceError("failed to insert index metadata", err)
		}
		r.log.Debug(fmt.Sprintf("Inserted new index metadata: %s", metadata.ID))
	}

	return nil
}

// FindByID 根据索引的唯一ID查找并返回 IndexMetadata。
func (r *sqlIndexMetadataRepository) FindByID(ctx context.Context, id string) (*index.IndexMetadata, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT id, table_name, column_name, index_type, tokenizer, data_type, description, created_at, updated_at
		FROM index_metadata WHERE id = ?`, id)

	metadata := &index.IndexMetadata{}
	var indexTypeStr string
	var tokenizer sql.NullString // 使用 sql.NullString 处理可空字段

	err := row.Scan(
		&metadata.ID, &metadata.TableName, &metadata.ColumnName, &indexTypeStr, &tokenizer,
		&metadata.DataType, &metadata.Description, &metadata.CreatedAt, &metadata.UpdatedAt)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ferrors.NewNotFoundError(fmt.Sprintf("index metadata with ID %s not found", id), err)
		}
		r.log.Error(fmt.Sprintf("Failed to query index metadata by ID %s: %v", id, err))
		return nil, ferrors.NewExternalServiceError("failed to get index metadata by ID", err)
	}

	metadata.IndexType = enum.IndexType(indexTypeStr)
	metadata.Tokenizer = tokenizer.String // 如果为NULL，则为""

	r.log.Debug(fmt.Sprintf("Found index metadata by ID %s", id))
	return metadata, nil
}

// FindByTableColumn 根据表名和列名查找并返回 IndexMetadata。
func (r *sqlIndexMetadataRepository) FindByTableColumn(ctx context.Context, tableName, columnName string) (*index.IndexMetadata, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT id, table_name, column_name, index_type, tokenizer, data_type, description, created_at, updated_at
		FROM index_metadata WHERE table_name = ? AND column_name = ?`, tableName, columnName)

	metadata := &index.IndexMetadata{}
	var indexTypeStr string
	var tokenizer sql.NullString

	err := row.Scan(
		&metadata.ID, &metadata.TableName, &metadata.ColumnName, &indexTypeStr, &tokenizer,
		&metadata.DataType, &metadata.Description, &metadata.CreatedAt, &metadata.UpdatedAt)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ferrors.NewNotFoundError(fmt.Sprintf("index metadata for table %s, column %s not found", tableName, columnName), err)
		}
		r.log.Error(fmt.Sprintf("Failed to query index metadata by table %s, column %s: %v", tableName, columnName, err))
		return nil, ferrors.NewExternalServiceError("failed to get index metadata by table/column", err)
	}

	metadata.IndexType = enum.IndexType(indexTypeStr)
	metadata.Tokenizer = tokenizer.String

	r.log.Debug(fmt.Sprintf("Found index metadata for table %s, column %s", tableName, columnName))
	return metadata, nil
}

// ListAll 返回所有已注册的索引元数据列表。
func (r *sqlIndexMetadataRepository) ListAll(ctx context.Context) ([]*index.IndexMetadata, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, table_name, column_name, index_type, tokenizer, data_type, description, created_at, updated_at
		FROM index_metadata`)
	if err != nil {
		r.log.Error(fmt.Sprintf("Failed to query all index metadata: %v", err))
		return nil, ferrors.NewExternalServiceError("failed to list all index metadata", err)
	}
	defer rows.Close()

	var indexes []*index.IndexMetadata
	for rows.Next() {
		metadata := &index.IndexMetadata{}
		var indexTypeStr string
		var tokenizer sql.NullString
		err := rows.Scan(
			&metadata.ID, &metadata.TableName, &metadata.ColumnName, &indexTypeStr, &tokenizer,
			&metadata.DataType, &metadata.Description, &metadata.CreatedAt, &metadata.UpdatedAt)
		if err != nil {
			r.log.Error(fmt.Sprintf("Failed to scan index metadata row: %v", err))
			return nil, ferrors.NewInternalError("failed to scan index metadata row", err)
		}
		metadata.IndexType = enum.IndexType(indexTypeStr)
		metadata.Tokenizer = tokenizer.String
		indexes = append(indexes, metadata)
	}

	if err = rows.Err(); err != nil {
		r.log.Error(fmt.Sprintf("Error during rows iteration for ListAll: %v", err))
		return nil, ferrors.NewExternalServiceError("error during rows iteration", err)
	}

	r.log.Debug(fmt.Sprintf("Listed %d index metadata records", len(indexes)))
	return indexes, nil
}

// ListByIndexedColumns 根据提供的表名和列名列表查找并返回匹配的索引元数据列表。
func (r *sqlIndexMetadataRepository) ListByIndexedColumns(ctx context.Context, tableNames []string, columnNames []string) ([]*index.IndexMetadata, error) {
	var (
		conditions []string
		args       []interface{}
	)

	if len(tableNames) > 0 {
		placeholders := make([]string, len(tableNames))
		for i := range tableNames {
			placeholders[i] = "?"
			args = append(args, tableNames[i])
		}
		conditions = append(conditions, fmt.Sprintf("table_name IN (%s)", strings.Join(placeholders, ",")))
	}

	if len(columnNames) > 0 {
		placeholders := make([]string, len(columnNames))
		for i := range columnNames {
			placeholders[i] = "?"
			args = append(args, columnNames[i])
		}
		conditions = append(conditions, fmt.Sprintf("column_name IN (%s)", strings.Join(placeholders, ",")))
	}

	whereClause := ""
	if len(conditions) > 0 {
		whereClause = " WHERE " + strings.Join(conditions, " AND ")
	}

	querySQL := fmt.Sprintf(`
		SELECT id, table_name, column_name, index_type, tokenizer, data_type, description, created_at, updated_at
		FROM index_metadata%s`, whereClause)

	rows, err := r.db.QueryContext(ctx, querySQL, args...)
	if err != nil {
		r.log.Error(fmt.Sprintf("Failed to query index metadata by indexed columns: %v", err),
			zap.String("query", querySQL), zap.Any("args", args))
		return nil, ferrors.NewExternalServiceError("failed to list index metadata by columns", err)
	}
	defer rows.Close()

	var indexes []*index.IndexMetadata
	for rows.Next() {
		metadata := &index.IndexMetadata{}
		var indexTypeStr string
		var tokenizer sql.NullString
		err := rows.Scan(
			&metadata.ID, &metadata.TableName, &metadata.ColumnName, &indexTypeStr, &tokenizer,
			&metadata.DataType, &metadata.Description, &metadata.CreatedAt, &metadata.UpdatedAt)
		if err != nil {
			r.log.Error(fmt.Sprintf("Failed to scan index metadata row for ListByIndexedColumns: %v", err))
			return nil, ferrors.NewInternalError("failed to scan index metadata row", err)
		}
		metadata.IndexType = enum.IndexType(indexTypeStr)
		metadata.Tokenizer = tokenizer.String
		indexes = append(indexes, metadata)
	}

	if err = rows.Err(); err != nil {
		r.log.Error(fmt.Sprintf("Error during rows iteration for ListByIndexedColumns: %v", err))
		return nil, ferrors.NewExternalServiceError("error during rows iteration", err)
	}

	r.log.Debug(fmt.Sprintf("Listed %d index metadata records by indexed columns", len(indexes)))
	return indexes, nil
}

// DeleteByID 根据索引的唯一ID删除 IndexMetadata。
func (r *sqlIndexMetadataRepository) DeleteByID(ctx context.Context, id string) error {
	result, err := r.db.ExecContext(ctx, `DELETE FROM index_metadata WHERE id = ?`, id)
	if err != nil {
		r.log.Error(fmt.Sprintf("Failed to delete index metadata %s: %v", id, err))
		return ferrors.NewExternalServiceError("failed to delete index metadata", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		r.log.Error(fmt.Sprintf("Failed to get rows affected for delete of %s: %v", id, err))
		return ferrors.NewInternalError("failed to check rows affected after delete", err)
	}
	if rowsAffected == 0 {
		r.log.Debug(fmt.Sprintf("Delete operation for index metadata %s, no rows affected (might not exist)", id))
		// 根据接口约定，删除不存在的ID不返回错误
	} else {
		r.log.Debug(fmt.Sprintf("Deleted %d row(s) for index metadata %s", rowsAffected, id))
	}

	return nil
}

// Close 关闭数据库连接。
func (r *sqlIndexMetadataRepository) Close(ctx context.Context) error {
	r.log.Info("Closing database connection for index metadata repository...")
	return r.db.Close()
}
