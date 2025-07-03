package index

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid" // 用于生成唯一ID
	"go.uber.org/zap"        // 用于日志字段

	ferrors "github.com/turtacn/starseek/internal/common/errors"
	"github.com/turtacn/starseek/internal/infrastructure/logger" // 日志接口
)

// IndexService 定义了索引领域的业务逻辑接口。
// 它是应用层与索引领域模型交互的入口。
type IndexService interface {
	// RegisterIndex 注册一个新的索引元数据。
	// 它包含业务规则，例如检查索引是否已存在。
	//
	// ctx: 上下文。
	// cmd: 注册索引的命令参数。
	//
	// 返回错误，如果注册失败（例如，索引已存在、参数无效或持久化错误）。
	RegisterIndex(ctx context.Context, cmd RegisterIndexCommand) error

	// GetIndexByTableColumn 根据表名和列名获取索引元数据。
	//
	// ctx: 上下文。
	// tableName: 数据库表名。
	// columnName: 列名。
	//
	// 返回 IndexMetadata 指针和错误。如果未找到，返回 ferrors.NewNotFoundError。
	GetIndexByTableColumn(ctx context.Context, tableName, columnName string) (*IndexMetadata, error)

	// ListIndexes 根据查询参数列出索引元数据。
	//
	// ctx: 上下文。
	// query: 查询参数，可包含过滤和分页信息。
	//
	// 返回 IndexMetadata 列表和错误。
	ListIndexes(ctx context.Context, query ListIndexesQuery) ([]*IndexMetadata, error)

	// DeleteIndexByID 根据ID删除索引元数据。
	// 此方法是为管理目的而添加，通常不会直接由查询模块使用。
	//
	// ctx: 上下文。
	// id: 索引的唯一标识符。
	//
	// 返回错误，如果删除失败。
	DeleteIndexByID(ctx context.Context, id string) error
}

// indexService 实现了 IndexService 接口。
// 它依赖 IndexMetadataRepository 进行数据持久化，并使用 Logger 进行日志记录。
type indexService struct {
	repo IndexMetadataRepository
	log  logger.Logger
}

// NewIndexService 创建并返回 IndexService 的新实例。
// 它是构造函数，通过依赖注入接收 IndexMetadataRepository 和 Logger。
func NewIndexService(repo IndexMetadataRepository, log logger.Logger) IndexService {
	return &indexService{
		repo: repo,
		log:  log,
	}
}

// RegisterIndex 实现了 IndexService 接口的 RegisterIndex 方法。
func (s *indexService) RegisterIndex(ctx context.Context, cmd RegisterIndexCommand) error {
	s.log.Info("Attempting to register new index",
		zap.String("tableName", cmd.TableName),
		zap.String("columnName", cmd.ColumnName),
		zap.String("indexType", string(cmd.IndexType)))

	// 1. 业务规则：检查是否已存在相同表名和列名的索引
	existingIndex, err := s.repo.FindByTableColumn(ctx, cmd.TableName, cmd.ColumnName)
	if err != nil && !ferrors.IsNotFoundError(err) {
		// 如果是其他错误，而不是找不到，则返回
		s.log.Error(fmt.Sprintf("Failed to check existing index for %s.%s: %v", cmd.TableName, cmd.ColumnName, err))
		return ferrors.NewInternalError("failed to check existing index", err)
	}
	if existingIndex != nil {
		s.log.Warn(fmt.Sprintf("Index for %s.%s already exists: %s", cmd.TableName, cmd.ColumnName, existingIndex.ID))
		return ferrors.NewAlreadyExistsError(fmt.Sprintf("index for table %s, column %s already exists with ID %s",
			cmd.TableName, cmd.ColumnName, existingIndex.ID), nil)
	}

	// 2. 创建 IndexMetadata 领域模型实例
	metadata := &IndexMetadata{
		ID:          uuid.New().String(), // 生成唯一ID
		TableName:   cmd.TableName,
		ColumnName:  cmd.ColumnName,
		IndexType:   cmd.IndexType,
		Tokenizer:   cmd.Tokenizer,
		DataType:    cmd.DataType,
		Description: cmd.Description,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	// 3. 持久化元数据
	if err := s.repo.Save(ctx, metadata); err != nil {
		s.log.Error(fmt.Sprintf("Failed to save new index metadata for %s.%s: %v", cmd.TableName, cmd.ColumnName, err))
		return ferrors.NewInternalError("failed to save index metadata", err)
	}

	s.log.Info(fmt.Sprintf("Successfully registered index %s for %s.%s", metadata.ID, cmd.TableName, cmd.ColumnName))
	return nil
}

// GetIndexByTableColumn 实现了 IndexService 接口的 GetIndexByTableColumn 方法。
func (s *indexService) GetIndexByTableColumn(ctx context.Context, tableName, columnName string) (*IndexMetadata, error) {
	s.log.Debug(fmt.Sprintf("Fetching index for %s.%s", tableName, columnName))

	metadata, err := s.repo.FindByTableColumn(ctx, tableName, columnName)
	if err != nil {
		if ferrors.IsNotFoundError(err) {
			s.log.Info(fmt.Sprintf("Index for %s.%s not found", tableName, columnName))
			return nil, ferrors.NewNotFoundError(fmt.Sprintf("index for table %s, column %s not found", tableName, columnName), err)
		}
		s.log.Error(fmt.Sprintf("Failed to retrieve index for %s.%s: %v", tableName, columnName, err))
		return nil, ferrors.NewInternalError("failed to retrieve index metadata", err)
	}

	s.log.Debug(fmt.Sprintf("Successfully fetched index %s for %s.%s", metadata.ID, tableName, columnName))
	return metadata, nil
}

// ListIndexes 实现了 IndexService 接口的 ListIndexes 方法。
func (s *indexService) ListIndexes(ctx context.Context, query ListIndexesQuery) ([]*IndexMetadata, error) {
	s.log.Debug("Listing indexes with query",
		zap.String("tableName", query.TableName),
		zap.String("columnName", query.ColumnName),
		zap.String("indexType", string(query.IndexType)),
		zap.Int("limit", query.Limit),
		zap.Int("offset", query.Offset))

	// 应用默认值和基本校验
	query.Validate()

	// 对于复杂的 List 查询，IndexMetadataRepository 可能需要一个 ListIndexesQuery 参数，
	// 但当前 IndexMetadataRepository 接口没有直接支持 ListIndexesQuery 的方法。
	// 这里我们模拟其行为，如果 repo 有更细粒度的查询方法，应调用它。
	// 为了适配当前 Repository 接口，我们先获取所有，然后进行内存过滤（效率较低，仅作示例）。
	// 在实际生产中，Repository 应支持基于 ListIndexesQuery 的数据库层过滤和分页。
	allIndexes, err := s.repo.ListAll(ctx) // 假设 repository 提供 ListAll
	if err != nil {
		s.log.Error(fmt.Sprintf("Failed to list all indexes from repository: %v", err))
		return nil, ferrors.NewInternalError("failed to list indexes", err)
	}

	filteredIndexes := make([]*IndexMetadata, 0)
	for _, idx := range allIndexes {
		match := true
		if query.TableName != "" && idx.TableName != query.TableName {
			match = false
		}
		if match && query.ColumnName != "" && idx.ColumnName != query.ColumnName {
			match = false
		}
		if match && query.IndexType != "" && idx.IndexType != query.IndexType {
			match = false
		}

		if match {
			filteredIndexes = append(filteredIndexes, idx)
		}
	}

	// 模拟分页
	start := query.Offset
	end := query.Offset + query.Limit
	if start > len(filteredIndexes) {
		return []*IndexMetadata{}, nil // 偏移量超出范围，返回空
	}
	if end > len(filteredIndexes) {
		end = len(filteredIndexes)
	}

	s.log.Info(fmt.Sprintf("Successfully listed %d indexes (total %d filtered)", end-start, len(filteredIndexes)))
	return filteredIndexes[start:end], nil
}

// DeleteIndexByID 实现了 IndexService 接口的 DeleteIndexByID 方法。
func (s *indexService) DeleteIndexByID(ctx context.Context, id string) error {
	s.log.Info(fmt.Sprintf("Attempting to delete index with ID: %s", id))

	// 可以在这里添加业务规则，例如：
	// 1. 检查索引是否正在被使用 (如果索引有状态或引用)
	// 2. 权限检查等

	err := s.repo.DeleteByID(ctx, id)
	if err != nil {
		s.log.Error(fmt.Sprintf("Failed to delete index %s from repository: %v", id, err))
		return ferrors.NewInternalError(fmt.Sprintf("failed to delete index %s", id), err)
	}

	s.log.Info(fmt.Sprintf("Successfully deleted index with ID: %s", id))
	return nil
}
