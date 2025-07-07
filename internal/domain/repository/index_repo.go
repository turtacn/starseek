package repository

import (
	"context"

	"github.com/turtacn/starseek/internal/domain/model"
)

// IndexMetadataRepository defines the interface for CRUD operations on IndexMetadata.
type IndexMetadataRepository interface {
	// Save persists a new or updates an existing IndexMetadata.
	Save(ctx context.Context, metadata *model.IndexMetadata) error

	// FindByID retrieves an IndexMetadata by its unique ID.
	FindByID(ctx context.Context, id uint) (*model.IndexMetadata, error)

	// FindByTableAndColumn retrieves an IndexMetadata by its table and column names.
	FindByTableAndColumn(ctx context.Context, tableName, columnName string) (*model.IndexMetadata, error)

	// ListByTable retrieves all IndexMetadata entries for a given table name.
	ListByTable(ctx context.Context, tableName string) ([]*model.IndexMetadata, error)

	// ListAll retrieves all IndexMetadata entries in the system.
	ListAll(ctx context.Context) ([]*model.IndexMetadata, error)

	// DeleteByID deletes an IndexMetadata by its unique ID.
	DeleteByID(ctx context.Context, id uint) error
}

//Personal.AI order the ending
