package enum

import (
	"fmt"
)

// IndexType 定义索引分词类型
type IndexType int

const (
	_IndexTypeUnknown     IndexType = iota // 0: 默认或未知类型，不应被使用
	IndexTypeChinese                       // 1: 中文分词
	IndexTypeEnglish                       // 2: 英文分词
	IndexTypeMultilingual                  // 3: 多语言分词
	IndexTypeNoTokenizer                   // 4: 不分词
)

// String 实现 fmt.Stringer 接口，返回 IndexType 的字符串表示
func (it IndexType) String() string {
	switch it {
	case IndexTypeChinese:
		return "Chinese"
	case IndexTypeEnglish:
		return "English"
	case IndexTypeMultilingual:
		return "Multilingual"
	case IndexTypeNoTokenizer:
		return "NoTokenizer"
	default:
		return fmt.Sprintf("UnknownIndexType(%d)", it)
	}
}

// IsValid 校验 IndexType 值是否合法
func (it IndexType) IsValid() bool {
	// 检查值是否在定义的有效范围内 (不包括 _IndexTypeUnknown)
	return it > _IndexTypeUnknown && it <= IndexTypeNoTokenizer
}

// DatabaseType 定义数据库类型
type DatabaseType int

const (
	_DatabaseTypeUnknown   DatabaseType = iota // 0: 默认或未知类型，不应被使用
	DatabaseTypeStarRocks                      // 1: StarRocks 数据库
	DatabaseTypeDoris                          // 2: Doris 数据库
	DatabaseTypeClickHouse                     // 3: ClickHouse 数据库
	DatabaseTypeMySQL                          // 4: MySQL 数据库 (通常用于元数据存储)
)

// String 实现 fmt.Stringer 接口，返回 DatabaseType 的字符串表示
func (dt DatabaseType) String() string {
	switch dt {
	case DatabaseTypeStarRocks:
		return "StarRocks"
	case DatabaseTypeDoris:
		return "Doris"
	case DatabaseTypeClickHouse:
		return "ClickHouse"
	case DatabaseTypeMySQL:
		return "MySQL"
	default:
		return fmt.Sprintf("UnknownDatabaseType(%d)", dt)
	}
}

// IsValid 校验 DatabaseType 值是否合法
func (dt DatabaseType) IsValid() bool {
	// 检查值是否在定义的有效范围内 (不包括 _DatabaseTypeUnknown)
	return dt > _DatabaseTypeUnknown && dt <= DatabaseTypeMySQL
}
