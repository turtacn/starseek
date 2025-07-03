package index

import (
	"time"

	"github.com/turtacn/starseek/internal/domain/enum" // 引入枚举包
)

// IndexMetadata 表示一个已注册的搜索索引的元数据领域模型。
// 它是索引模块的核心数据结构。
type IndexMetadata struct {
	ID          string         `json:"id"`          // 索引的唯一标识符 (例如 UUID)
	TableName   string         `json:"table_name"`  // 索引所基于的数据库表名
	ColumnName  string         `json:"column_name"` // 索引所基于的表中的列名
	IndexType   enum.IndexType `json:"index_type"`  // 索引类型 (例如 FULLTEXT, BTREE, HASH)
	Tokenizer   string         `json:"tokenizer"`   // 文本索引使用的分词器 (例如 "standard", "whitespace", "ngram")
	DataType    string         `json:"data_type"`   // 被索引列的数据类型 (例如 "TEXT", "VARCHAR", "INT", "TIMESTAMP")
	Description string         `json:"description"` // 索引的描述信息
	CreatedAt   time.Time      `json:"created_at"`  // 索引创建时间
	UpdatedAt   time.Time      `json:"updated_at"`  // 索引最后更新时间
}

// RegisterIndexCommand 定义了注册新索引的请求参数。
// 这是一个输入模型，用于从 API 或其他输入源接收创建索引所需的信息。
type RegisterIndexCommand struct {
	TableName   string         `json:"table_name" validate:"required,min=1"`                             // 数据库表名，必填
	ColumnName  string         `json:"column_name" validate:"required,min=1"`                            // 列名，必填
	IndexType   enum.IndexType `json:"index_type" validate:"required,oneof=FULLTEXT BTREE HASH SPATIAL"` // 索引类型，必填且为预定义值之一
	Tokenizer   string         `json:"tokenizer"`                                                        // 可选，分词器，主要用于 FULLTEXT 索引
	DataType    string         `json:"data_type" validate:"required,min=1"`                              // 被索引列的数据类型，必填
	Description string         `json:"description"`                                                      // 可选，描述信息
}

// ListIndexesQuery 定义了查询索引列表的参数。
// 这是一个输入模型，用于从 API 或其他输入源接收筛选、分页信息。
type ListIndexesQuery struct {
	TableName  string         `json:"table_name"`  // 可选，按表名过滤
	ColumnName string         `json:"column_name"` // 可选，按列名过滤
	IndexType  enum.IndexType `json:"index_type"`  // 可选，按索引类型过滤
	Limit      int            `json:"limit"`       // 可选，分页限制 (默认为 10 或 20)
	Offset     int            `json:"offset"`      // 可选，分页偏移
}

// Validate 方法可以为 ListIndexesQuery 添加默认值和基本校验
func (q *ListIndexesQuery) Validate() {
	if q.Limit <= 0 {
		q.Limit = 10 // 默认每页10条
	}
	if q.Limit > 100 {
		q.Limit = 100 // 最大每页100条
	}
	if q.Offset < 0 {
		q.Offset = 0
	}
}
