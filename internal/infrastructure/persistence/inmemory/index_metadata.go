package inmemory

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time" // 用于模拟时间字段

	ferrors "github.com/turtacn/starseek/internal/common/errors"
	"github.com/turtacn/starseek/internal/domain/index" // 引入 IndexMetadata 和 IndexMetadataRepository 接口
)

// inmemoryIndexMetadataRepository 是 IndexMetadataRepository 接口的内存实现。
// 它主要用于单元测试和简单的本地开发场景。
type inmemoryIndexMetadataRepository struct {
	mu    sync.RWMutex                    // 读写互斥锁，确保并发安全
	store map[string]*index.IndexMetadata // Key: ID, Value: IndexMetadata
	tcMap map[string]string               // Key: TableName|ColumnName, Value: ID (用于快速通过表名和列名查找)
}

// NewInMemoryIndexMetadataRepository 创建并返回一个新的内存 IndexMetadataRepository 实例。
func NewInMemoryIndexMetadataRepository() index.IndexMetadataRepository {
	return &inmemoryIndexMetadataRepository{
		store: make(map[string]*index.IndexMetadata),
		tcMap: make(map[string]string),
	}
}

// Save 将索引元数据保存到内存中。
// 如果 metadata.ID 存在，则更新；否则，插入。
// （注意：ID通常由领域服务生成，不应为空）
func (r *inmemoryIndexMetadataRepository) Save(ctx context.Context, metadata *index.IndexMetadata) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if metadata.ID == "" {
		return ferrors.NewInternalError("metadata ID cannot be empty for in-memory save operation", nil)
	}

	// 检查是否已存在相同 TableName 和 ColumnName 的索引（唯一约束模拟）
	tcKey := fmt.Sprintf("%s|%s", metadata.TableName, metadata.ColumnName)
	existingIDForTC, foundTC := r.tcMap[tcKey]

	if _, exists := r.store[metadata.ID]; exists {
		// 更新现有记录
		// 如果ID存在，但其表列组合与现有tcMap中的不同，这表示逻辑错误或数据不一致
		if foundTC && existingIDForTC != metadata.ID {
			// This scenario should ideally not happen if IDs are truly unique and
			// TableName|ColumnName is also unique across the system.
			// It implies trying to update an index to conflict with another existing one.
			return ferrors.NewAlreadyExistsError(
				fmt.Sprintf("another index with table %s, column %s already exists with ID %s",
					metadata.TableName, metadata.ColumnName, existingIDForTC), nil)
		}
		// 确保更新时间
		metadata.UpdatedAt = time.Now()
		r.store[metadata.ID] = metadata
	} else {
		// 插入新记录
		if foundTC {
			return ferrors.NewAlreadyExistsError(
				fmt.Sprintf("index for table %s, column %s already exists with ID %s",
					metadata.TableName, metadata.ColumnName, existingIDForTC), nil)
		}
		// 确保创建和更新时间
		now := time.Now()
		metadata.CreatedAt = now
		metadata.UpdatedAt = now
		r.store[metadata.ID] = metadata
		r.tcMap[tcKey] = metadata.ID // 添加到tcMap
	}
	return nil
}

// FindByID 根据索引的唯一ID查找并返回 IndexMetadata。
func (r *inmemoryIndexMetadataRepository) FindByID(ctx context.Context, id string) (*index.IndexMetadata, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	metadata, found := r.store[id]
	if !found {
		return nil, ferrors.NewNotFoundError(fmt.Sprintf("index metadata with ID %s not found in memory", id), nil)
	}
	// 返回副本以防止外部修改内部存储
	copiedMetadata := *metadata
	return &copiedMetadata, nil
}

// FindByTableColumn 根据表名和列名查找并返回 IndexMetadata。
func (r *inmemoryIndexMetadataRepository) FindByTableColumn(ctx context.Context, tableName, columnName string) (*index.IndexMetadata, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	tcKey := fmt.Sprintf("%s|%s", tableName, columnName)
	id, foundID := r.tcMap[tcKey]
	if !foundID {
		return nil, ferrors.NewNotFoundError(fmt.Sprintf("index metadata for table %s, column %s not found in memory", tableName, columnName), nil)
	}

	metadata, found := r.store[id]
	if !found {
		// 理论上tcMap和store应该同步，如果不同步则表示内部不一致
		return nil, ferrors.NewInternalError(fmt.Sprintf("inconsistency: ID %s from tcMap not found in store", id), nil)
	}
	// 返回副本以防止外部修改内部存储
	copiedMetadata := *metadata
	return &copiedMetadata, nil
}

// ListAll 返回所有已注册的索引元数据列表。
func (r *inmemoryIndexMetadataRepository) ListAll(ctx context.Context) ([]*index.IndexMetadata, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var indexes []*index.IndexMetadata
	for _, metadata := range r.store {
		// 返回副本以防止外部修改内部存储
		copiedMetadata := *metadata
		indexes = append(indexes, &copiedMetadata)
	}
	return indexes, nil
}

// ListByIndexedColumns 根据提供的表名和列名列表查找并返回匹配的索引元数据列表。
// 这里的匹配逻辑是“AND”：如果提供了表名列表和列名列表，则必须同时满足两者。
func (r *inmemoryIndexMetadataRepository) ListByIndexedColumns(ctx context.Context, tableNames []string, columnNames []string) ([]*index.IndexMetadata, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var filteredIndexes []*index.IndexMetadata
	for _, metadata := range r.store {
		matchTableName := true
		if len(tableNames) > 0 {
			matchTableName = false
			for _, tn := range tableNames {
				if metadata.TableName == tn {
					matchTableName = true
					break
				}
			}
		}

		matchColumnName := true
		if len(columnNames) > 0 {
			matchColumnName = false
			for _, cn := range columnNames {
				if metadata.ColumnName == cn {
					matchColumnName = true
					break
				}
			}
		}

		if matchTableName && matchColumnName {
			// 返回副本以防止外部修改内部存储
			copiedMetadata := *metadata
			filteredIndexes = append(filteredIndexes, &copiedMetadata)
		}
	}
	return filteredIndexes, nil
}

// DeleteByID 根据索引的唯一ID删除 IndexMetadata。
// 即使ID不存在，也不返回错误。
func (r *inmemoryIndexMetadataRepository) DeleteByID(ctx context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if metadata, found := r.store[id]; found {
		// 从tcMap中移除
		tcKey := fmt.Sprintf("%s|%s", metadata.TableName, metadata.ColumnName)
		delete(r.tcMap, tcKey)
		// 从store中移除
		delete(r.store, id)
	}
	return nil // 删除不存在的ID不报错
}

// Close 对于内存仓库，Close 方法通常不做任何事情，因为没有资源需要释放。
func (r *inmemoryIndexMetadataRepository) Close(ctx context.Context) error {
	// 内存存储无需额外关闭操作
	return nil
}
