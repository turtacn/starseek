package index

import (
	"context"
)

// IndexMetadataRepository 定义了管理索引元数据持久化的接口。
// 这个接口是一个端口，其具体的实现（适配器）将负责与底层数据库或存储系统交互。
type IndexMetadataRepository interface {
	// Save 将索引元数据保存到存储中。
	// 如果 metadata.ID 为空，则认为是创建新记录，并由存储层生成ID。
	// 如果 metadata.ID 不为空，则认为是更新现有记录。
	//
	// ctx: 用于操作的上下文，可用于传递追踪信息、取消信号或超时。
	// metadata: 要保存的 IndexMetadata 结构体指针。
	//
	// 返回错误，如果保存操作失败。
	Save(ctx context.Context, metadata *IndexMetadata) error

	// FindByID 根据索引的唯一ID查找并返回 IndexMetadata。
	//
	// ctx: 用于操作的上下文。
	// id: 索引的唯一标识符。
	//
	// 返回找到的 IndexMetadata 结构体指针和错误。
	// 如果未找到，应返回 nil 和一个特定的错误（例如 ferrors.NewNotFoundError）。
	FindByID(ctx context.Context, id string) (*IndexMetadata, error)

	// FindByTableColumn 根据表名和列名查找并返回 IndexMetadata。
	// 假设一个表的一列最多只有一个特定类型的索引元数据（例如，一个列只有一个全文索引）。
	//
	// ctx: 用于操作的上下文。
	// tableName: 索引所属的数据库表名。
	// columnName: 索引所基于的列名。
	//
	// 返回找到的 IndexMetadata 结构体指针和错误。
	// 如果未找到，应返回 nil 和一个特定的错误（例如 ferrors.NewNotFoundError）。
	FindByTableColumn(ctx context.Context, tableName, columnName string) (*IndexMetadata, error)

	// ListAll 返回所有已注册的索引元数据列表。
	//
	// ctx: 用于操作的上下文。
	//
	// 返回 IndexMetadata 结构体指针的切片和错误。
	// 如果没有索引，返回空切片和 nil 错误。
	ListAll(ctx context.Context) ([]*IndexMetadata, error)

	// ListByIndexedColumns 根据提供的表名和列名列表查找并返回匹配的索引元数据列表。
	// 例如，可以用于批量检查给定多个表和列是否存在索引。
	// 注意：这里的匹配逻辑可以是“AND”或“OR”，通常由具体实现决定，
	// 但如果需要复杂查询，最好传入一个更通用的查询结构体（如 ListIndexesQuery）。
	// 这里为了简洁，直接使用切片。
	//
	// ctx: 用于操作的上下文。
	// tableNames: 要匹配的表名列表。如果为空，则不按表名过滤。
	// columnNames: 要匹配的列名列表。如果为空，则不按列名过滤。
	//
	// 返回匹配的 IndexMetadata 结构体指针的切片和错误。
	ListByIndexedColumns(ctx context.Context, tableNames []string, columnNames []string) ([]*IndexMetadata, error)

	// DeleteByID 根据索引的唯一ID删除 IndexMetadata。
	//
	// ctx: 用于操作的上下文。
	// id: 索引的唯一标识符。
	//
	// 返回错误，如果删除操作失败。即使ID不存在，通常也不应返回错误。
	DeleteByID(ctx context.Context, id string) error

	// Close 关闭存储层连接（如果需要）。
	// 例如，对于数据库连接池或内存存储，可能需要清理资源。
	Close(ctx context.Context) error
}
