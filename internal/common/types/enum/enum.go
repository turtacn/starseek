package enum

import (
	"fmt"
	"strings"
)

// DatabaseType represents different database engine types
type DatabaseType string

const (
	DatabaseTypeStarRocks  DatabaseType = "STARROCKS"
	DatabaseTypeClickHouse DatabaseType = "CLICKHOUSE"
	DatabaseTypeDoris      DatabaseType = "DORIS"
)

// String returns the string representation of DatabaseType
func (dt DatabaseType) String() string {
	return string(dt)
}

// IsValid validates if the database type is supported
func (dt DatabaseType) IsValid() bool {
	switch dt {
	case DatabaseTypeStarRocks, DatabaseTypeClickHouse, DatabaseTypeDoris:
		return true
	default:
		return false
	}
}

// ParseDatabaseType parses string to DatabaseType
func ParseDatabaseType(s string) (DatabaseType, error) {
	dt := DatabaseType(strings.ToUpper(s))
	if !dt.IsValid() {
		return "", fmt.Errorf("invalid database type: %s", s)
	}
	return dt, nil
}

// IndexType represents different index technology types
type IndexType string

const (
	IndexTypeInverted IndexType = "INVERTED"
	IndexTypeNgram    IndexType = "NGRAM"
	IndexTypeFulltext IndexType = "FULLTEXT"
)

// String returns the string representation of IndexType
func (it IndexType) String() string {
	return string(it)
}

// IsValid validates if the index type is supported
func (it IndexType) IsValid() bool {
	switch it {
	case IndexTypeInverted, IndexTypeNgram, IndexTypeFulltext:
		return true
	default:
		return false
	}
}

// ParseIndexType parses string to IndexType
func ParseIndexType(s string) (IndexType, error) {
	it := IndexType(strings.ToUpper(s))
	if !it.IsValid() {
		return "", fmt.Errorf("invalid index type: %s", s)
	}
	return it, nil
}

// TokenizerType represents different text tokenization strategies
type TokenizerType string

const (
	TokenizerTypeChinese      TokenizerType = "CHINESE"
	TokenizerTypeEnglish      TokenizerType = "ENGLISH"
	TokenizerTypeMultilingual TokenizerType = "MULTILINGUAL"
	TokenizerTypeNone         TokenizerType = "NONE"
)

// String returns the string representation of TokenizerType
func (tt TokenizerType) String() string {
	return string(tt)
}

// IsValid validates if the tokenizer type is supported
func (tt TokenizerType) IsValid() bool {
	switch tt {
	case TokenizerTypeChinese, TokenizerTypeEnglish, TokenizerTypeMultilingual, TokenizerTypeNone:
		return true
	default:
		return false
	}
}

// ParseTokenizerType parses string to TokenizerType
func ParseTokenizerType(s string) (TokenizerType, error) {
	tt := TokenizerType(strings.ToUpper(s))
	if !tt.IsValid() {
		return "", fmt.Errorf("invalid tokenizer type: %s", s)
	}
	return tt, nil
}

// QueryType represents different search query types
type QueryType string

const (
	QueryTypeMatch   QueryType = "MATCH"
	QueryTypePhrase  QueryType = "PHRASE"
	QueryTypeFuzzy   QueryType = "FUZZY"
	QueryTypeBoolean QueryType = "BOOLEAN"
)

// String returns the string representation of QueryType
func (qt QueryType) String() string {
	return string(qt)
}

// IsValid validates if the query type is supported
func (qt QueryType) IsValid() bool {
	switch qt {
	case QueryTypeMatch, QueryTypePhrase, QueryTypeFuzzy, QueryTypeBoolean:
		return true
	default:
		return false
	}
}

// ParseQueryType parses string to QueryType
func ParseQueryType(s string) (QueryType, error) {
	qt := QueryType(strings.ToUpper(s))
	if !qt.IsValid() {
		return "", fmt.Errorf("invalid query type: %s", s)
	}
	return qt, nil
}

// RankingAlgorithm represents different relevance scoring algorithms
type RankingAlgorithm string

const (
	RankingAlgorithmTFIDF  RankingAlgorithm = "TFIDF"
	RankingAlgorithmBM25   RankingAlgorithm = "BM25"
	RankingAlgorithmCustom RankingAlgorithm = "CUSTOM"
)

// String returns the string representation of RankingAlgorithm
func (ra RankingAlgorithm) String() string {
	return string(ra)
}

// IsValid validates if the ranking algorithm is supported
func (ra RankingAlgorithm) IsValid() bool {
	switch ra {
	case RankingAlgorithmTFIDF, RankingAlgorithmBM25, RankingAlgorithmCustom:
		return true
	default:
		return false
	}
}

// ParseRankingAlgorithm parses string to RankingAlgorithm
func ParseRankingAlgorithm(s string) (RankingAlgorithm, error) {
	ra := RankingAlgorithm(strings.ToUpper(s))
	if !ra.IsValid() {
		return "", fmt.Errorf("invalid ranking algorithm: %s", s)
	}
	return ra, nil
}

// CacheType represents different cache storage types
type CacheType string

const (
	CacheTypeMemory CacheType = "MEMORY"
	CacheTypeRedis  CacheType = "REDIS"
	CacheTypeHybrid CacheType = "HYBRID"
)

// String returns the string representation of CacheType
func (ct CacheType) String() string {
	return string(ct)
}

// IsValid validates if the cache type is supported
func (ct CacheType) IsValid() bool {
	switch ct {
	case CacheTypeMemory, CacheTypeRedis, CacheTypeHybrid:
		return true
	default:
		return false
	}
}

// ParseCacheType parses string to CacheType
func ParseCacheType(s string) (CacheType, error) {
	ct := CacheType(strings.ToUpper(s))
	if !ct.IsValid() {
		return "", fmt.Errorf("invalid cache type: %s", s)
	}
	return ct, nil
}

// LogLevel represents different log levels
type LogLevel string

const (
	LogLevelDebug LogLevel = "DEBUG"
	LogLevelInfo  LogLevel = "INFO"
	LogLevelWarn  LogLevel = "WARN"
	LogLevelError LogLevel = "ERROR"
	LogLevelFatal LogLevel = "FATAL"
)

// String returns the string representation of LogLevel
func (ll LogLevel) String() string {
	return string(ll)
}

// IsValid validates if the log level is supported
func (ll LogLevel) IsValid() bool {
	switch ll {
	case LogLevelDebug, LogLevelInfo, LogLevelWarn, LogLevelError, LogLevelFatal:
		return true
	default:
		return false
	}
}

// ParseLogLevel parses string to LogLevel
func ParseLogLevel(s string) (LogLevel, error) {
	ll := LogLevel(strings.ToUpper(s))
	if !ll.IsValid() {
		return "", fmt.Errorf("invalid log level: %s", s)
	}
	return ll, nil
}

// ServiceStatus represents different service status types
type ServiceStatus string

const (
	ServiceStatusHealthy   ServiceStatus = "HEALTHY"
	ServiceStatusUnhealthy ServiceStatus = "UNHEALTHY"
	ServiceStatusUnknown   ServiceStatus = "UNKNOWN"
)

// String returns the string representation of ServiceStatus
func (ss ServiceStatus) String() string {
	return string(ss)
}

// IsValid validates if the service status is supported
func (ss ServiceStatus) IsValid() bool {
	switch ss {
	case ServiceStatusHealthy, ServiceStatusUnhealthy, ServiceStatusUnknown:
		return true
	default:
		return false
	}
}

// ParseServiceStatus parses string to ServiceStatus
func ParseServiceStatus(s string) (ServiceStatus, error) {
	ss := ServiceStatus(strings.ToUpper(s))
	if !ss.IsValid() {
		return "", fmt.Errorf("invalid service status: %s", s)
	}
	return ss, nil
}

// TaskStatus represents different task execution status
type TaskStatus string

const (
	TaskStatusPending   TaskStatus = "PENDING"
	TaskStatusRunning   TaskStatus = "RUNNING"
	TaskStatusCompleted TaskStatus = "COMPLETED"
	TaskStatusFailed    TaskStatus = "FAILED"
	TaskStatusCancelled TaskStatus = "CANCELLED"
)

// String returns the string representation of TaskStatus
func (ts TaskStatus) String() string {
	return string(ts)
}

// IsValid validates if the task status is supported
func (ts TaskStatus) IsValid() bool {
	switch ts {
	case TaskStatusPending, TaskStatusRunning, TaskStatusCompleted, TaskStatusFailed, TaskStatusCancelled:
		return true
	default:
		return false
	}
}

// ParseTaskStatus parses string to TaskStatus
func ParseTaskStatus(s string) (TaskStatus, error) {
	ts := TaskStatus(strings.ToUpper(s))
	if !ts.IsValid() {
		return "", fmt.Errorf("invalid task status: %s", s)
	}
	return ts, nil
}

//Personal.AI order the ending
