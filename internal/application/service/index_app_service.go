package service

import (
	"context"
	"fmt"
	"time"

	"github.com/turtacn/starseek/internal/common/errorx"
	"github.com/turtacn/starseek/internal/domain/model"
	"github.com/turtacn/starseek/internal/domain/repository"
	"github.com/turtacn/starseek/internal/infrastructure/cache"
	"go.uber.org/zap"
)

// IndexAppService is the application service responsible for managing index metadata.
type IndexAppService struct {
	indexRepo repository.IndexMetadataRepository
	cache     cache.Cache // For caching index metadata or related lookup data
	logger    *zap.Logger
}

// NewIndexAppService creates a new IndexAppService instance.
func NewIndexAppService(
	indexRepo repository.IndexMetadataRepository,
	cache cache.Cache,
	logger *zap.Logger,
) *IndexAppService {
	return &IndexAppService{
		indexRepo: indexRepo,
		cache:     cache,
		logger:    logger.Named("IndexAppService"),
	}
}

// RegisterIndex registers a new index metadata entry.
func (s *IndexAppService) RegisterIndex(ctx context.Context, metadata *model.IndexMetadata) (*model.IndexMetadata, error) {
	// Basic validation
	if metadata.TableName == "" || metadata.ColumnName == "" {
		return nil, errorx.ErrBadRequest.Wrap(fmt.Errorf("table_name and column_name cannot be empty"))
	}

	// Check if an index already exists for this table and column
	existing, err := s.indexRepo.FindByTableAndColumn(ctx, metadata.TableName, metadata.ColumnName)
	if err == nil {
		// If found and it's the same ID, it's an update. If different ID, it's a conflict.
		if metadata.ID != 0 && existing.ID == metadata.ID {
			s.logger.Info("Updating existing index metadata", zap.Uint("id", metadata.ID), zap.String("table", metadata.TableName), zap.String("column", metadata.ColumnName))
		} else {
			return nil, errorx.ErrIndexAlreadyExists.Wrap(
				fmt.Errorf("an index for table '%s' column '%s' already exists with ID %d", metadata.TableName, metadata.ColumnName, existing.ID),
			)
		}
	} else if !errorx.IsNotFound(err) { // If it's not a "not found" error, it's a real DB error
		s.logger.Error("Failed to check for existing index metadata", zap.Error(err), zap.String("table", metadata.TableName), zap.String("column", metadata.ColumnName))
		return nil, errorx.NewError(errorx.ErrInternalServer.Code, "Failed to register index due to database error", err)
	}

	// Set creation/update timestamps
	now := time.Now()
	if metadata.ID == 0 {
		metadata.CreatedAt = now
	}
	metadata.UpdatedAt = now

	if err := s.indexRepo.Save(ctx, metadata); err != nil {
		s.logger.Error("Failed to save index metadata", zap.Error(err), zap.Any("metadata", metadata))
		return nil, errorx.NewError(errorx.ErrInternalServer.Code, "Failed to save index metadata", err)
	}

	s.logger.Info("Index metadata registered/updated successfully", zap.Any("metadata", metadata))
	// Invalidate relevant cache entries if any, e.g., s.cache.Delete(constant.CacheKeyPrefix + "all_index_metadata")
	return metadata, nil
}

// DeleteIndex deletes an index metadata entry by ID.
func (s *IndexAppService) DeleteIndex(ctx context.Context, id uint) error {
	s.logger.Info("Attempting to delete index metadata", zap.Uint("id", id))
	if err := s.indexRepo.DeleteByID(ctx, id); err != nil {
		s.logger.Error("Failed to delete index metadata", zap.Error(err), zap.Uint("id", id))
		return errorx.NewError(errorx.ErrInternalServer.Code, "Failed to delete index metadata", err)
	}
	s.logger.Info("Index metadata deleted successfully", zap.Uint("id", id))
	// Invalidate relevant cache entries
	return nil
}

// ListIndexes retrieves all registered index metadata entries.
func (s *IndexAppService) ListIndexes(ctx context.Context) ([]*model.IndexMetadata, error) {
	// Optionally, implement caching here
	// cacheKey := constant.CacheKeyPrefix + "all_index_metadata"
	// if cachedData, err := s.cache.Get(ctx, cacheKey); err == nil {
	// 	var metadata []*model.IndexMetadata
	// 	if err := json.Unmarshal([]byte(cachedData), &metadata); err == nil {
	// 		s.logger.Debug("Retrieved index metadata from cache")
	// 		return metadata, nil
	// 	}
	// 	s.logger.Warn("Failed to unmarshal cached index metadata, fetching from DB", zap.Error(err))
	// }

	metadata, err := s.indexRepo.ListAll(ctx)
	if err != nil {
		s.logger.Error("Failed to list all index metadata", zap.Error(err))
		return nil, errorx.NewError(errorx.ErrInternalServer.Code, "Failed to list index metadata", err)
	}

	// Cache the results
	// if jsonBytes, err := json.Marshal(metadata); err == nil {
	// 	s.cache.Set(ctx, cacheKey, string(jsonBytes), constant.DefaultCacheTTL)
	// } else {
	// 	s.logger.Error("Failed to marshal index metadata for caching", zap.Error(err))
	// }

	s.logger.Info("Listed all index metadata", zap.Int("count", len(metadata)))
	return metadata, nil
}

// SyncFromDB is a placeholder for a method that would synchronize index metadata
// from the underlying data warehouse (e.g., by calling DBAdapter.FetchMetadata).
// This could be run as a periodic background task.
func (s *IndexAppService) SyncFromDB(ctx context.Context) error {
	// This would typically involve iterating through configured DB adapters
	// and calling their FetchMetadata method.
	s.logger.Info("Starting index metadata synchronization from databases (Not implemented fully)")
	// Example:
	// starrocksAdapter := ... // Get a configured StarRocks adapter
	// srMeta, err := starrocksAdapter.FetchMetadata(ctx)
	// if err != nil {
	// 	s.logger.Error("Failed to sync from StarRocks", zap.Error(err))
	// } else {
	// 	for _, meta := range srMeta {
	// 		// Save or update each discovered metadata
	// 		s.RegisterIndex(ctx, meta) // This handles upsert logic
	// 	}
	// }
	return nil
}

//Personal.AI order the ending
