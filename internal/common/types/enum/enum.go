package enum

import (
	"fmt"
	"strings"
)

// DatabaseType represents different column-store database engine types
type DatabaseType int

const (
	DatabaseTypeStarRocks DatabaseType = iota
	DatabaseTypeClickHouse
	DatabaseTypeDoris
)

// String returns the string representation of DatabaseType
func (dt DatabaseType) String() string {
	switch dt {
	case DatabaseTypeStarRocks:
		return "STARROCKS"
	case DatabaseTypeClickHouse:
		return "CLICKHOUSE"
	case DatabaseTypeDoris:
		return "DORIS"
	default:
		return "UNKNOWN"
	}
}

// IsValid checks if the DatabaseType is valid
func (dt DatabaseType) IsValid() bool {
	return dt >= DatabaseTypeStarRocks && dt <= DatabaseTypeDoris
}

// ParseDatabaseType parses a string to DatabaseType
func ParseDatabaseType(s string) (DatabaseType, error) {
	switch strings.ToUpper(strings.TrimSpace(s)) {
	case "STARROCKS":
		return DatabaseTypeStarRocks, nil
	case "CLICKHOUSE":
		return DatabaseTypeClickHouse, nil
	case "DORIS":
		return DatabaseTypeDoris, nil
	default:
		return DatabaseType(-1), fmt.Errorf("invalid database type: %s", s)
	}
}

// IndexType represents different index technology types
type IndexType int

const (
	IndexTypeInverted IndexType = iota
	IndexTypeNGram
	IndexTypeFullText
)

// String returns the string representation of IndexType
func (it IndexType) String() string {
	switch it {
	case IndexTypeInverted:
		return "INVERTED"
	case IndexTypeNGram:
		return "NGRAM"
	case IndexTypeFullText:
		return "FULLTEXT"
	default:
		return "UNKNOWN"
	}
}

// IsValid checks if the IndexType is valid
func (it IndexType) IsValid() bool {
	return it >= IndexTypeInverted && it <= IndexTypeFullText
}

// ParseIndexType parses a string to IndexType
func ParseIndexType(s string) (IndexType, error) {
	switch strings.ToUpper(strings.TrimSpace(s)) {
	case "INVERTED":
		return IndexTypeInverted, nil
	case "NGRAM":
		return IndexTypeNGram, nil
	case "FULLTEXT":
		return IndexTypeFullText, nil
	default:
		return IndexType(-1), fmt.Errorf("invalid index type: %s", s)
	}
}

// TokenizerType represents different text tokenization strategies
type TokenizerType int

const (
	TokenizerTypeChinese TokenizerType = iota
	TokenizerTypeEnglish
	TokenizerTypeMultilingual
	TokenizerTypeNone
)

// String returns the string representation of TokenizerType
func (tt TokenizerType) String() string {
	switch tt {
	case TokenizerTypeChinese:
		return "CHINESE"
	case TokenizerTypeEnglish:
		return "ENGLISH"
	case TokenizerTypeMultilingual:
		return "MULTILINGUAL"
	case TokenizerTypeNone:
		return "NONE"
	default:
		return "UNKNOWN"
	}
}

// IsValid checks if the TokenizerType is valid
func (tt TokenizerType) IsValid() bool {
	return tt >= TokenizerTypeChinese && tt <= TokenizerTypeNone
}

// ParseTokenizerType parses a string to TokenizerType
func ParseTokenizerType(s string) (TokenizerType, error) {
	switch strings.ToUpper(strings.TrimSpace(s)) {
	case "CHINESE":
		return TokenizerTypeChinese, nil
	case "ENGLISH":
		return TokenizerTypeEnglish, nil
	case "MULTILINGUAL":
		return TokenizerTypeMultilingual, nil
	case "NONE":
		return TokenizerTypeNone, nil
	default:
		return TokenizerType(-1), fmt.Errorf("invalid tokenizer type: %s", s)
	}
}

// QueryType represents different search query types
type QueryType int

const (
	QueryTypeMatch QueryType = iota
	QueryTypePhrase
	QueryTypeFuzzy
	QueryTypeBoolean
)

// String returns the string representation of QueryType
func (qt QueryType) String() string {
	switch qt {
	case QueryTypeMatch:
		return "MATCH"
	case QueryTypePhrase:
		return "PHRASE"
	case QueryTypeFuzzy:
		return "FUZZY"
	case QueryTypeBoolean:
		return "BOOLEAN"
	default:
		return "UNKNOWN"
	}
}

// IsValid checks if the QueryType is valid
func (qt QueryType) IsValid() bool {
	return qt >= QueryTypeMatch && qt <= QueryTypeBoolean
}

// ParseQueryType parses a string to QueryType
func ParseQueryType(s string) (QueryType, error) {
	switch strings.ToUpper(strings.TrimSpace(s)) {
	case "MATCH":
		return QueryTypeMatch, nil
	case "PHRASE":
		return QueryTypePhrase, nil
	case "FUZZY":
		return QueryTypeFuzzy, nil
	case "BOOLEAN":
		return QueryTypeBoolean, nil
	default:
		return QueryType(-1), fmt.Errorf("invalid query type: %s", s)
	}
}

// RankingAlgorithm represents different relevance scoring algorithms
type RankingAlgorithm int

const (
	RankingAlgorithmTFIDF RankingAlgorithm = iota
	RankingAlgorithmBM25
	RankingAlgorithmCustom
)

// String returns the string representation of RankingAlgorithm
func (ra RankingAlgorithm) String() string {
	switch ra {
	case RankingAlgorithmTFIDF:
		return "TFIDF"
	case RankingAlgorithmBM25:
		return "BM25"
	case RankingAlgorithmCustom:
		return "CUSTOM"
	default:
		return "UNKNOWN"
	}
}

// IsValid checks if the RankingAlgorithm is valid
func (ra RankingAlgorithm) IsValid() bool {
	return ra >= RankingAlgorithmTFIDF && ra <= RankingAlgorithmCustom
}

// ParseRankingAlgorithm parses a string to RankingAlgorithm
func ParseRankingAlgorithm(s string) (RankingAlgorithm, error) {
	switch strings.ToUpper(strings.TrimSpace(s)) {
	case "TFIDF":
		return RankingAlgorithmTFIDF, nil
	case "BM25":
		return RankingAlgorithmBM25, nil
	case "CUSTOM":
		return RankingAlgorithmCustom, nil
	default:
		return RankingAlgorithm(-1), fmt.Errorf("invalid ranking algorithm: %s", s)
	}
}

// CacheType represents different cache storage types
type CacheType int

const (
	CacheTypeMemory CacheType = iota
	CacheTypeRedis
	CacheTypeHybrid
)

// String returns the string representation of CacheType
func (ct CacheType) String() string {
	switch ct {
	case CacheTypeMemory:
		return "MEMORY"
	case CacheTypeRedis:
		return "REDIS"
	case CacheTypeHybrid:
		return "HYBRID"
	default:
		return "UNKNOWN"
	}
}

// IsValid checks if the CacheType is valid
func (ct CacheType) IsValid() bool {
	return ct >= CacheTypeMemory && ct <= CacheTypeHybrid
}

// ParseCacheType parses a string to CacheType
func ParseCacheType(s string) (CacheType, error) {
	switch strings.ToUpper(strings.TrimSpace(s)) {
	case "MEMORY":
		return CacheTypeMemory, nil
	case "REDIS":
		return CacheTypeRedis, nil
	case "HYBRID":
		return CacheTypeHybrid, nil
	default:
		return CacheType(-1), fmt.Errorf("invalid cache type: %s", s)
	}
}

// Status represents general status values
type Status int

const (
	StatusUnknown Status = iota
	StatusActive
	StatusInactive
	StatusPending
	StatusFailed
	StatusDeleted
)

// String returns the string representation of Status
func (s Status) String() string {
	switch s {
	case StatusUnknown:
		return "UNKNOWN"
	case StatusActive:
		return "ACTIVE"
	case StatusInactive:
		return "INACTIVE"
	case StatusPending:
		return "PENDING"
	case StatusFailed:
		return "FAILED"
	case StatusDeleted:
		return "DELETED"
	default:
		return "INVALID"
	}
}

// IsValid checks if the Status is valid
func (s Status) IsValid() bool {
	return s >= StatusUnknown && s <= StatusDeleted
}

// ParseStatus parses a string to Status
func ParseStatus(str string) (Status, error) {
	switch strings.ToUpper(strings.TrimSpace(str)) {
	case "UNKNOWN":
		return StatusUnknown, nil
	case "ACTIVE":
		return StatusActive, nil
	case "INACTIVE":
		return StatusInactive, nil
	case "PENDING":
		return StatusPending, nil
	case "FAILED":
		return StatusFailed, nil
	case "DELETED":
		return StatusDeleted, nil
	default:
		return Status(-1), fmt.Errorf("invalid status: %s", str)
	}
}

// LogLevel represents different logging levels
type LogLevel int

const (
	LogLevelDebug LogLevel = iota
	LogLevelInfo
	LogLevelWarn
	LogLevelError
	LogLevelFatal
)

// String returns the string representation of LogLevel
func (ll LogLevel) String() string {
	switch ll {
	case LogLevelDebug:
		return "DEBUG"
	case LogLevelInfo:
		return "INFO"
	case LogLevelWarn:
		return "WARN"
	case LogLevelError:
		return "ERROR"
	case LogLevelFatal:
		return "FATAL"
	default:
		return "UNKNOWN"
	}
}

// IsValid checks if the LogLevel is valid
func (ll LogLevel) IsValid() bool {
	return ll >= LogLevelDebug && ll <= LogLevelFatal
}

// ParseLogLevel parses a string to LogLevel
func ParseLogLevel(s string) (LogLevel, error) {
	switch strings.ToUpper(strings.TrimSpace(s)) {
	case "DEBUG":
		return LogLevelDebug, nil
	case "INFO":
		return LogLevelInfo, nil
	case "WARN":
		return LogLevelWarn, nil
	case "ERROR":
		return LogLevelError, nil
	case "FATAL":
		return LogLevelFatal, nil
	default:
		return LogLevel(-1), fmt.Errorf("invalid log level: %s", s)
	}
}

// GetAllDatabaseTypes returns all valid database types
func GetAllDatabaseTypes() []DatabaseType {
	return []DatabaseType{
		DatabaseTypeStarRocks,
		DatabaseTypeClickHouse,
		DatabaseTypeDoris,
	}
}

// GetAllIndexTypes returns all valid index types
func GetAllIndexTypes() []IndexType {
	return []IndexType{
		IndexTypeInverted,
		IndexTypeNGram,
		IndexTypeFullText,
	}
}

// GetAllTokenizerTypes returns all valid tokenizer types
func GetAllTokenizerTypes() []TokenizerType {
	return []TokenizerType{
		TokenizerTypeChinese,
		TokenizerTypeEnglish,
		TokenizerTypeMultilingual,
		TokenizerTypeNone,
	}
}

// GetAllQueryTypes returns all valid query types
func GetAllQueryTypes() []QueryType {
	return []QueryType{
		QueryTypeMatch,
		QueryTypePhrase,
		QueryTypeFuzzy,
		QueryTypeBoolean,
	}
}

// GetAllRankingAlgorithms returns all valid ranking algorithms
func GetAllRankingAlgorithms() []RankingAlgorithm {
	return []RankingAlgorithm{
		RankingAlgorithmTFIDF,
		RankingAlgorithmBM25,
		RankingAlgorithmCustom,
	}
}

// GetAllCacheTypes returns all valid cache types
func GetAllCacheTypes() []CacheType {
	return []CacheType{
		CacheTypeMemory,
		CacheTypeRedis,
		CacheTypeHybrid,
	}
}

// GetAllStatuses returns all valid status values
func GetAllStatuses() []Status {
	return []Status{
		StatusUnknown,
		StatusActive,
		StatusInactive,
		StatusPending,
		StatusFailed,
		StatusDeleted,
	}
}

// GetAllLogLevels returns all valid log levels
func GetAllLogLevels() []LogLevel {
	return []LogLevel{
		LogLevelDebug,
		LogLevelInfo,
		LogLevelWarn,
		LogLevelError,
		LogLevelFatal,
	}
}

//Personal.AI order the ending
