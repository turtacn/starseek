package search

import (
	"fmt"
)

// SearchResult 表示一个搜索结果文档。
// 它是从数据库中检索到并经过处理后，准备返回给前端的单个条目。
type SearchResult struct {
	Table     string                 `json:"table"`     // 来源表名
	RowID     string                 `json:"row_id"`    // 唯一行ID (通常是主键值，可以是字符串化的数字)
	Score     float64                `json:"score"`     // 搜索相关性得分
	Highlight map[string]string      `json:"highlight"` // 包含匹配关键词的字段高亮片段 (字段名 -> 高亮文本)
	Data      map[string]interface{} `json:"data"`      // 原始行数据，可以是任意字段及其值
}

// SearchQueryType 定义了搜索查询的类型，用于指导查询构建和执行策略。
type SearchQueryType int

const (
	// UNKNOWN_QUERY 未知查询类型
	UNKNOWN_QUERY SearchQueryType = iota

	// MATCH_ALL 表示查询中的所有关键词都必须匹配。
	// (逻辑 AND)
	MATCH_ALL

	// MATCH_ANY 表示查询中的任意关键词匹配即可。
	// (逻辑 OR)
	MATCH_ANY

	// FIELD_SPECIFIC 表示查询针对特定字段进行。
	// 例如：title:"golang" AND content:"microservices"
	FIELD_SPECIFIC

	// FUZZY_MATCH 表示模糊匹配，允许一定程度的拼写错误或相似性。
	FUZZY_MATCH

	// PHRASE_MATCH 表示短语匹配，关键词必须以特定顺序出现且连续。
	// 例如："microservices architecture"
	PHRASE_MATCH
)

// String 方法返回 SearchQueryType 的字符串表示。
func (qt SearchQueryType) String() string {
	switch qt {
	case MATCH_ALL:
		return "MATCH_ALL"
	case MATCH_ANY:
		return "MATCH_ANY"
	case FIELD_SPECIFIC:
		return "FIELD_SPECIFIC"
	case FUZZY_MATCH:
		return "FUZZY_MATCH"
	case PHRASE_MATCH:
		return "PHRASE_MATCH"
	default:
		return "UNKNOWN_QUERY"
	}
}

// SearchQuery 表示一个内部解析后的复杂搜索查询。
// 它包含了执行搜索所需的所有逻辑和参数。
type SearchQuery struct {
	Keywords  []string          // 经过分词和归一化后的搜索关键词列表
	Fields    []string          // 要搜索的字段列表，如果为空则搜索所有默认字段
	Tables    []string          // 要搜索的表列表，如果为空则搜索所有配置的表
	Page      int               // 当前页码 (从1开始)
	PageSize  int               // 每页结果数量
	QueryType SearchQueryType   // 查询类型，例如 MATCH_ALL, MATCH_ANY, FIELD_SPECIFIC
	Filters   map[string]string // 额外的过滤条件 (例如："category": "tech", "status": "published")
	MinScore  float64           // 最小相关性得分，低于此得分的结果将被过滤
}

// IDFStats 存储一个词元的逆文档频率 (Inverse Document Frequency) 统计。
// 用于 TF-IDF 排名算法。
type IDFStats struct {
	Term                     string  // 词元本身
	DocumentFrequency        int64   // 包含该词元的文档数量 (df)
	TotalDocuments           int64   // 索引中所有文档的总数量 (N)
	InverseDocumentFrequency float64 // 逆文档频率值 (IDF = log(N / df))
}

// CalculateIDF 计算并返回 IDF 值。
func (stats *IDFStats) CalculateIDF() float64 {
	if stats.DocumentFrequency == 0 || stats.TotalDocuments == 0 {
		return 0.0 // 避免除以零或对数错误
	}
	// 经典的 IDF 公式：log(N / df)
	// 实际应用中，通常会加1平滑：log(1 + N / df) 或 log(N / (df + 1))
	// 这里使用简单的 log(N/df)
	return fmt.Log(float64(stats.TotalDocuments) / float64(stats.DocumentFrequency))
}
