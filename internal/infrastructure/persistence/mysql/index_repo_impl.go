package mysql

import (
	"context"
	"errors"

	"github.com/turtacn/starseek/internal/common/errorx"
	"github.com/turtacn/starseek/internal/domain/model"
	"github.com/turtacn/starseek/internal/domain/repository"
	"github.com/turtacn/starseek/internal/infrastructure/log"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// MySQLIndexRepo implements the IndexMetadataRepository interface using GORM and MySQL.
type MySQLIndexRepo struct {
	db  *gorm.DB
	log log.Logger
}

// NewMySQLIndexRepo creates a new MySQLIndexRepo instance.
func NewMySQLIndexRepo(db *gorm.DB, logger log.Logger) repository.IndexMetadataRepository {
	return &MySQLIndexRepo{
		db:  db,
		log: logger,
	}
}

// Save creates or updates an IndexMetadata record.
func (repo *MySQLIndexRepo) Save(ctx context.Context, metadata *model.IndexMetadata) (*model.IndexMetadata, error) {
	if metadata == nil {
		return nil, errorx.ErrBadRequest.Wrap(errors.New("metadata cannot be nil"))
	}

	// Try to find an existing record by TableName and ColumnName
	var existing model.IndexMetadata
	result := repo.db.WithContext(ctx).
		Where("table_name = ? AND column_name = ?", metadata.TableName, metadata.ColumnName).
		First(&existing)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			// Record not found, create a new one
			repo.log.Info("Creating new index metadata record", zap.Any("metadata", metadata))
			if createResult := repo.db.WithContext(ctx).Create(metadata); createResult.Error != nil {
				repo.log.Error("Failed to create index metadata", zap.Error(createResult.Error), zap.Any("metadata", metadata))
				return nil, errorx.Wrap(createResult.Error, errorx.ErrDatabase)
			}
		} else {
			// Other database error
			repo.log.Error("Failed to query existing index metadata", zap.Error(result.Error), zap.Any("metadata", metadata))
			return nil, errorx.Wrap(result.Error, errorx.ErrDatabase)
		}
	} else {
		// Record found, update it
		metadata.ID = existing.ID // Ensure the ID is set for update
		repo.log.Info("Updating existing index metadata record", zap.Uint("id", metadata.ID), zap.Any("metadata", metadata))
		if updateResult := repo.db.WithContext(ctx).Save(metadata); updateResult.Error != nil {
			repo.log.Error("Failed to update index metadata", zap.Error(updateResult.Error), zap.Uint("id", metadata.ID))
			return nil, errorx.Wrap(updateResult.Error, errorx.ErrDatabase)
		}
	}

	repo.log.Info("Index metadata saved successfully", zap.Uint("id", metadata.ID))
	return metadata, nil
}

// FindByID retrieves an IndexMetadata record by its ID.
func (repo *MySQLIndexRepo) FindByID(ctx context.Context, id uint) (*model.IndexMetadata, error) {
	var metadata model.IndexMetadata
	result := repo.db.WithContext(ctx).First(&metadata, id)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			repo.log.Warn("Index metadata not found by ID", zap.Uint("id", id))
			return nil, errorx.Wrap(result.Error, errorx.ErrIndexNotFound)
		}
		repo.log.Error("Failed to find index metadata by ID", zap.Error(result.Error), zap.Uint("id", id))
		return nil, errorx.Wrap(result.Error, errorx.ErrDatabase)
	}
	return &metadata, nil
}

// FindByTableAndColumn retrieves an IndexMetadata by table and column name.
func (repo *MySQLIndexRepo) FindByTableAndColumn(ctx context.Context, tableName, columnName string) (*model.IndexMetadata, error) {
	var metadata model.IndexMetadata
	result := repo.db.WithContext(ctx).
		Where("table_name = ? AND column_name = ?", tableName, columnName).
		First(&metadata)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			repo.log.Warn("Index metadata not found by table and column", zap.String("table", tableName), zap.String("column", columnName))
			return nil, errorx.Wrap(result.Error, errorx.ErrIndexNotFound)
		}
		repo.log.Error("Failed to find index metadata by table and column", zap.Error(result.Error), zap.String("table", tableName), zap.String("column", columnName))
		return nil, errorx.Wrap(result.Error, errorx.ErrDatabase)
	}
	return &metadata, nil
}

// ListByTable retrieves all IndexMetadata records for a given table.
func (repo *MySQLIndexRepo) ListByTable(ctx context.Context, tableName string) ([]*model.IndexMetadata, error) {
	var metadatas []*model.IndexMetadata
	result := repo.db.WithContext(ctx).Where("table_name = ?", tableName).Find(&metadatas)
	if result.Error != nil {
		repo.log.Error("Failed to list index metadata by table", zap.Error(result.Error), zap.String("table", tableName))
		return nil, errorx.Wrap(result.Error, errorx.ErrDatabase)
	}
	return metadatas, nil
}

// ListAll retrieves all IndexMetadata records.
func (repo *MySQLIndexRepo) ListAll(ctx context.Context) ([]*model.IndexMetadata, error) {
	var metadatas []*model.IndexMetadata
	result := repo.db.WithContext(ctx).Find(&metadatas)
	if result.Error != nil {
		repo.log.Error("Failed to list all index metadata", zap.Error(result.Error))
		return nil, errorx.Wrap(result.Error, errorx.ErrDatabase)
	}
	return metadatas, nil
}

// DeleteByID deletes an IndexMetadata record by its ID.
func (repo *MySQLIndexRepo) DeleteByID(ctx context.Context, id uint) error {
	result := repo.db.WithContext(ctx).Delete(&model.IndexMetadata{}, id)
	if result.Error != nil {
		repo.log.Error("Failed to delete index metadata by ID", zap.Error(result.Error), zap.Uint("id", id))
		return errorx.Wrap(result.Error, errorx.ErrDatabase)
	}
	if result.RowsAffected == 0 {
		repo.log.Warn("Attempted to delete non-existent index metadata by ID", zap.Uint("id", id))
		return errorx.Wrap(gorm.ErrRecordNotFound, errorx.ErrIndexNotFound) // Indicate nothing was deleted
	}
	repo.log.Info("Index metadata deleted successfully by ID", zap.Uint("id", id))
	return nil
}

// DeleteByTableAndColumn deletes an IndexMetadata record by table and column name.
func (repo *MySQLIndexRepo) DeleteByTableAndColumn(ctx context.Context, tableName, columnName string) error {
	result := repo.db.WithContext(ctx).
		Where("table_name = ? AND column_name = ?", tableName, columnName).
		Delete(&model.IndexMetadata{})
	if result.Error != nil {
		repo.log.Error("Failed to delete index metadata by table and column", zap.Error(result.Error), zap.String("table", tableName), zap.String("column", columnName))
		return errorx.Wrap(result.Error, errorx.ErrDatabase)
	}
	if result.RowsAffected == 0 {
		repo.log.Warn("Attempted to delete non-existent index metadata by table and column", zap.String("table", tableName), zap.String("column", columnName))
		return errorx.Wrap(gorm.ErrRecordNotFound, errorx.ErrIndexNotFound)
	}
	repo.log.Info("Index metadata deleted successfully by table and column", zap.String("table", tableName), zap.String("column", columnName))
	return nil
}

//Personal.AI order the ending
