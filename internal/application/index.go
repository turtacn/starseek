package application

import (
	"context"
	"fmt"

	"go.uber.org/zap"

	"github.com/turtacn/starseek/internal/common/enum" // 用于 IndexType, DatabaseType 等枚举
	ferrors "github.com/turtacn/starseek/internal/common/errors"
	"github.com/turtacn/starseek/internal/domain/index" // 引入 IndexService 接口和相关命令/查询/实体
	"github.com/turtacn/starseek/internal/infrastructure/logger"
)

// IndexApplicationService 接口定义了索引管理的应用层操作。
// 它是 API 层与领域层之间的桥梁，负责处理请求的初步校验和业务流程协调。
type IndexApplicationService interface {
	// RegisterIndex 用于注册或更新一个索引的元数据。
	RegisterIndex(ctx context.Context, cmd index.RegisterIndexCommand) error

	// ListIndexes 用于列出符合条件的索引元数据。
	ListIndexes(ctx context.Context, query index.ListIndexesQuery) ([]*index.IndexMetadata, error)

	// DeleteIndex 用于删除一个索引的元数据。
	// 实际操作中可能还会触发底层索引的删除。
	DeleteIndex(ctx context.Context, cmd index.DeleteIndexCommand) error
}

// indexApplicationService 是 IndexApplicationService 接口的实现。
// 它封装了对领域服务 index.IndexService 的调用，并处理应用层的校验逻辑。
type indexApplicationService struct {
	indexService index.IndexService // 依赖领域层的 IndexService
	log          logger.Logger
}

// NewIndexApplicationService 创建并返回一个新的 IndexApplicationService 实例。
func NewIndexApplicationService(indexService index.IndexService, log logger.Logger) (IndexApplicationService, error) {
	if indexService == nil {
		return nil, ferrors.NewInternalError("index service is nil", nil)
	}
	if log == nil {
		return nil, ferrors.NewInternalError("logger is nil", nil)
	}
	log.Info("IndexApplicationService initialized.")
	return &indexApplicationService{
		indexService: indexService,
		log:          log,
	}, nil
}

// RegisterIndex 方法处理注册新索引或更新现有索引的请求。
func (s *indexApplicationService) RegisterIndex(ctx context.Context, cmd index.RegisterIndexCommand) error {
	s.log.Info("RegisterIndex request received",
		zap.String("table_name", cmd.TableName),
		zap.String("column_name", cmd.ColumnName),
		zap.String("index_type", cmd.IndexType.String()))

	// 1. 参数校验 (应用层职责)
	if cmd.TableName == "" {
		return ferrors.NewBadRequestError("table name cannot be empty", nil)
	}
	if cmd.ColumnName == "" {
		return ferrors.NewBadRequestError("column name cannot be empty", nil)
	}
	if cmd.IndexType == enum.UNKNOWN_INDEX_TYPE {
		return ferrors.NewBadRequestError("invalid index type provided", nil)
	}

	// 对于全文索引，需要分词器名称
	if cmd.IndexType == enum.FULL_TEXT_INDEX {
		if cmd.TokenizerName == "" {
			return ferrors.NewBadRequestError("tokenizer name is required for full-text index", nil)
		}
		// 还可以进一步校验 tokenizerName 是否是系统支持的有效分词器
		// 例如：_, err := s.tokenizerService.GetTokenizer(cmd.TokenizerName)
		// if err != nil { return ferrors.NewBadRequestError("unsupported tokenizer name", err) }
	}

	// 对于结构化索引，可以有额外的校验，比如 ensure DatabaseType is valid
	if cmd.DatabaseType == enum.UNKNOWN_DATABASE_TYPE {
		// 可以根据实际情况，如果 DatabaseType是可选的，这里不报错
		// 如果是必需的，这里报错
	}

	// 2. 构造领域层命令或实体 (IndexMetadata)
	// 在此处，IndexMetadata 通常由领域服务内部根据传入的Command创建，
	// 因为领域实体应由领域逻辑控制其生命周期和一致性。
	// 这里将 RegisterIndexCommand 直接传递给领域服务。
	err := s.indexService.RegisterIndex(ctx, cmd)
	if err != nil {
		s.log.Error("Failed to register index in domain service",
			zap.String("table_name", cmd.TableName),
			zap.String("column_name", cmd.ColumnName),
			zap.Error(err))
		return ferrors.WrapInternalError(err, "failed to register index") // Wrap domain errors
	}

	s.log.Info("Index registered successfully",
		zap.String("table_name", cmd.TableName),
		zap.String("column_name", cmd.ColumnName))
	return nil
}

// ListIndexes 方法处理列出索引元数据的请求。
func (s *indexApplicationService) ListIndexes(ctx context.Context, query index.ListIndexesQuery) ([]*index.IndexMetadata, error) {
	s.log.Info("ListIndexes request received",
		zap.String("table_name", query.TableName),
		zap.String("column_name", query.ColumnName),
		zap.String("index_type", query.IndexType.String()))

	// 1. 参数校验 (可选，如果查询参数有复杂逻辑)
	// 例如，如果 IndexType 传入了，确保它是有效的
	if query.IndexType != enum.UNKNOWN_INDEX_TYPE {
		_, err := enum.ParseIndexType(query.IndexType.String()) // Re-parse to validate enum string if needed
		if err != nil {
			return nil, ferrors.NewBadRequestError("invalid index type filter", nil)
		}
	}

	// 2. 调用领域服务获取数据
	var metadata []*index.IndexMetadata
	var err error

	if query.TableName != "" || query.ColumnName != "" || query.IndexType != enum.UNKNOWN_INDEX_TYPE {
		// If specific filters are provided, call specialized domain method (if exists)
		// For now, indexService.GetIndexMetadata only takes tableName
		// If we need column/indexType specific filtering, the domain service method signature
		// or its internal logic needs to support it.
		// For simplicity, we'll call GetIndexMetadata (which currently only takes tableName)
		// and filter in application layer if domain layer doesn't fully support it.
		// Better: domain service exposes a method like GetFilteredIndexMetadata(ctx, query)
		if query.TableName != "" {
			metadata, err = s.indexService.GetIndexMetadata(ctx, query.TableName)
		} else {
			// If tableName is empty but other filters exist, fetch all and filter
			metadata, err = s.indexService.GetAllIndexMetadata(ctx)
		}

		if err != nil {
			s.log.Error("Failed to list indexes from domain service", zap.Error(err))
			return nil, ferrors.WrapInternalError(err, "failed to retrieve index metadata")
		}

		// Apply additional filters if domain service doesn't do it
		filteredMetadata := []*index.IndexMetadata{}
		for _, md := range metadata {
			match := true
			if query.ColumnName != "" && md.ColumnName != query.ColumnName {
				match = false
			}
			if query.IndexType != enum.UNKNOWN_INDEX_TYPE && md.IndexType != query.IndexType {
				match = false
			}
			if match {
				filteredMetadata = append(filteredMetadata, md)
			}
		}
		metadata = filteredMetadata

	} else {
		// No specific filters, list all
		metadata, err = s.indexService.GetAllIndexMetadata(ctx)
		if err != nil {
			s.log.Error("Failed to list all indexes from domain service", zap.Error(err))
			return nil, ferrors.WrapInternalError(err, "failed to retrieve all index metadata")
		}
	}

	s.log.Info("Indexes listed successfully", zap.Int("count", len(metadata)))
	return metadata, nil
}

// DeleteIndex 方法处理删除索引元数据的请求。
func (s *indexApplicationService) DeleteIndex(ctx context.Context, cmd index.DeleteIndexCommand) error {
	s.log.Info("DeleteIndex request received",
		zap.String("table_name", cmd.TableName),
		zap.String("column_name", cmd.ColumnName))

	// 1. 参数校验
	if cmd.TableName == "" {
		return ferrors.NewBadRequestError("table name cannot be empty for delete", nil)
	}
	if cmd.ColumnName == "" {
		return ferrors.NewBadRequestError("column name cannot be empty for delete", nil)
	}

	// 2. 调用领域服务执行删除
	err := s.indexService.DeleteIndex(ctx, cmd)
	if err != nil {
		s.log.Error("Failed to delete index in domain service",
			zap.String("table_name", cmd.TableName),
			zap.String("column_name", cmd.ColumnName),
			zap.Error(err))
		return ferrors.WrapInternalError(err, "failed to delete index")
	}

	s.log.Info("Index deleted successfully",
		zap.String("table_name", cmd.TableName),
		zap.String("column_name", cmd.ColumnName))
	return nil
}
