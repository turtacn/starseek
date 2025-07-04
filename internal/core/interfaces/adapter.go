package interfaces

import (
	"context"
	"time"
)

// DatabaseAdapter 定义数据库适配器接口，抽象不同数据库引擎的操作
type DatabaseAdapter interface {
	// 连接管理
	Connect(ctx context.Context, config *DatabaseConfig) error
	Disconnect(ctx context.Context) error
	Ping(ctx context.Context) error
	IsConnected() bool
	GetConnectionPool() ConnectionPool

	// 查询执行
	Execute(ctx context.Context, query string, args ...interface{}) (*QueryResult, error)
	ExecuteQuery(ctx context.Context, query string, args ...interface{}) (*QueryResultSet, error)
	ExecuteStatement(ctx context.Context, statement string, args ...interface{}) (*ExecutionResult, error)

	// 批量操作
	ExecuteBatch(ctx context.Context, statements []*BatchStatement) (*BatchResult, error)
	BulkInsert(ctx context.Context, table string, columns []string, rows [][]interface{}) error
	BulkUpdate(ctx context.Context, table string, updates []*BulkUpdateItem) error
	BulkDelete(ctx context.Context, table string, conditions []*BulkDeleteCondition) error

	// 元数据获取
	GetDatabaseInfo(ctx context.Context) (*DatabaseInfo, error)
	GetTableInfo(ctx context.Context, tableName string) (*TableInfo, error)
	GetTableList(ctx context.Context, schema string) ([]string, error)
	GetColumnInfo(ctx context.Context, tableName string) ([]*ColumnInfo, error)
	GetIndexInfo(ctx context.Context, tableName string) ([]*IndexInfo, error)

	// DDL操作
	CreateTable(ctx context.Context, definition *TableDefinition) error
	AlterTable(ctx context.Context, tableName string, alterations []*TableAlteration) error
	DropTable(ctx context.Context, tableName string) error
	TruncateTable(ctx context.Context, tableName string) error

	// 索引操作
	CreateIndex(ctx context.Context, definition *IndexDefinition) error
	DropIndex(ctx context.Context, tableName string, indexName string) error

	// 事务管理
	BeginTransaction(ctx context.Context) (DatabaseTransaction, error)
	ExecuteInTransaction(ctx context.Context, fn func(tx DatabaseTransaction) error) error

	// 查询构建器
	GetQueryBuilder() QueryBuilder

	// 数据库特定功能
	GetDialect() DatabaseDialect
	GetCapabilities() *DatabaseCapabilities

	// 统计信息
	GetStatistics(ctx context.Context) (*DatabaseStatistics, error)
	GetTableStatistics(ctx context.Context, tableName string) (*TableStatistics, error)

	// 健康检查
	HealthCheck(ctx context.Context) (*HealthStatus, error)
}

// QueryBuilder 定义SQL查询构建器接口
type QueryBuilder interface {
	// 基础查询构建
	Select(columns ...string) QueryBuilder
	From(table string) QueryBuilder
	Join(table string, condition string) QueryBuilder
	LeftJoin(table string, condition string) QueryBuilder
	RightJoin(table string, condition string) QueryBuilder
	InnerJoin(table string, condition string) QueryBuilder
	Where(condition string, args ...interface{}) QueryBuilder
	GroupBy(columns ...string) QueryBuilder
	Having(condition string, args ...interface{}) QueryBuilder
	OrderBy(column string, direction SortDirection) QueryBuilder
	Limit(limit int) QueryBuilder
	Offset(offset int) QueryBuilder

	// 插入查询构建
	InsertInto(table string) QueryBuilder
	Values(values ...interface{}) QueryBuilder
	OnDuplicateKeyUpdate(updates map[string]interface{}) QueryBuilder

	// 更新查询构建
	Update(table string) QueryBuilder
	Set(column string, value interface{}) QueryBuilder
	SetMap(updates map[string]interface{}) QueryBuilder

	// 删除查询构建
	DeleteFrom(table string) QueryBuilder

	// 聚合函数
	Count(column string) QueryBuilder
	Sum(column string) QueryBuilder
	Avg(column string) QueryBuilder
	Max(column string) QueryBuilder
	Min(column string) QueryBuilder

	// 子查询
	SubQuery(alias string, subQuery QueryBuilder) QueryBuilder
	Exists(subQuery QueryBuilder) QueryBuilder
	NotExists(subQuery QueryBuilder) QueryBuilder

	// 条件构建
	And(condition string, args ...interface{}) QueryBuilder
	Or(condition string, args ...interface{}) QueryBuilder
	In(column string, values ...interface{}) QueryBuilder
	NotIn(column string, values ...interface{}) QueryBuilder
	Between(column string, start, end interface{}) QueryBuilder
	Like(column string, pattern string) QueryBuilder
	IsNull(column string) QueryBuilder
	IsNotNull(column string) QueryBuilder

	// CTE (Common Table Expressions)
	With(name string, query QueryBuilder) QueryBuilder
	WithRecursive(name string, query QueryBuilder) QueryBuilder

	// 窗口函数
	Window(definition string) QueryBuilder
	Over(partition string, order string) QueryBuilder

	// 构建最终查询
	Build() (*Query, error)
	BuildString() (string, error)
	BuildWithArgs() (string, []interface{}, error)

	// 重置构建器
	Reset() QueryBuilder
	Clone() QueryBuilder

	// 数据库方言特定
	SetDialect(dialect DatabaseDialect) QueryBuilder
	GetDialect() DatabaseDialect
}

// ConnectionPool 定义连接池管理接口
type ConnectionPool interface {
	// 连接管理
	GetConnection(ctx context.Context) (Connection, error)
	ReleaseConnection(conn Connection) error
	CloseConnection(conn Connection) error

	// 池状态管理
	GetPoolStats() *PoolStatistics
	SetMaxOpenConns(max int)
	SetMaxIdleConns(max int)
	SetConnMaxLifetime(duration time.Duration)
	SetConnMaxIdleTime(duration time.Duration)

	// 健康检查
	PingAll(ctx context.Context) error
	CleanupIdleConnections(ctx context.Context) error

	// 池生命周期
	Initialize(ctx context.Context, config *PoolConfig) error
	Close(ctx context.Context) error

	// 监控
	GetActiveConnections() int
	GetIdleConnections() int
	GetTotalConnections() int
	GetConnectionWaitCount() int64
	GetConnectionWaitDuration() time.Duration
}

// Connection 定义数据库连接接口
type Connection interface {
	// 基础操作
	Execute(ctx context.Context, query string, args ...interface{}) (*QueryResult, error)
	Query(ctx context.Context, query string, args ...interface{}) (*QueryResultSet, error)
	Prepare(ctx context.Context, query string) (PreparedStatement, error)

	// 事务操作
	Begin(ctx context.Context) (DatabaseTransaction, error)

	// 连接状态
	Ping(ctx context.Context) error
	IsValid() bool
	GetID() string
	GetCreatedAt() time.Time
	GetLastUsedAt() time.Time

	// 关闭连接
	Close() error
}

// PreparedStatement 定义预编译语句接口
type PreparedStatement interface {
	Execute(ctx context.Context, args ...interface{}) (*QueryResult, error)
	Query(ctx context.Context, args ...interface{}) (*QueryResultSet, error)
	Close() error
	GetQuery() string
}

// DatabaseTransaction 定义数据库事务接口
type DatabaseTransaction interface {
	// 事务控制
	Commit(ctx context.Context) error
	Rollback(ctx context.Context) error

	// 查询操作
	Execute(ctx context.Context, query string, args ...interface{}) (*QueryResult, error)
	Query(ctx context.Context, query string, args ...interface{}) (*QueryResultSet, error)
	Prepare(ctx context.Context, query string) (PreparedStatement, error)

	// 批量操作
	ExecuteBatch(ctx context.Context, statements []*BatchStatement) (*BatchResult, error)

	// 事务状态
	IsActive() bool
	GetID() string
	GetStartTime() time.Time

	// 保存点
	Savepoint(ctx context.Context, name string) error
	RollbackToSavepoint(ctx context.Context, name string) error
	ReleaseSavepoint(ctx context.Context, name string) error
}

// BatchProcessor 定义批量处理接口
type BatchProcessor interface {
	// 批量插入
	AddInsert(table string, columns []string, values []interface{}) error
	AddUpdate(table string, updates map[string]interface{}, where string, args ...interface{}) error
	AddDelete(table string, where string, args ...interface{}) error

	// 执行批量操作
	Execute(ctx context.Context) (*BatchResult, error)
	ExecuteInTransaction(ctx context.Context, tx DatabaseTransaction) (*BatchResult, error)

	// 批量配置
	SetBatchSize(size int)
	SetTimeout(timeout time.Duration)
	SetRetryPolicy(policy *RetryPolicy)

	// 状态管理
	GetPendingCount() int
	Clear()
	Reset()
}

// 数据结构定义

// DatabaseConfig 数据库配置
type DatabaseConfig struct {
	Driver       DatabaseDriver `json:"driver"`
	Host         string         `json:"host"`
	Port         int            `json:"port"`
	Database     string         `json:"database"`
	Username     string         `json:"username"`
	Password     string         `json:"password"`
	SSLMode      string         `json:"ssl_mode,omitempty"`
	Charset      string         `json:"charset,omitempty"`
	Timezone     string         `json:"timezone,omitempty"`
	MaxOpenConns int            `json:"max_open_conns"`
	MaxIdleConns int            `json:"max_idle_conns"`
	ConnMaxLifetime time.Duration `json:"conn_max_lifetime"`
	ConnMaxIdleTime time.Duration `json:"conn_max_idle_time"`
	ConnectTimeout  time.Duration `json:"connect_timeout"`
	ReadTimeout     time.Duration `json:"read_timeout"`
	WriteTimeout    time.Duration `json:"write_timeout"`
	ExtraParams     map[string]string `json:"extra_params,omitempty"`
}

// PoolConfig 连接池配置
type PoolConfig struct {
	MaxOpenConns    int           `json:"max_open_conns"`
	MaxIdleConns    int           `json:"max_idle_conns"`
	ConnMaxLifetime time.Duration `json:"conn_max_lifetime"`
	ConnMaxIdleTime time.Duration `json:"conn_max_idle_time"`
	HealthCheckInterval time.Duration `json:"health_check_interval"`
	RetryAttempts   int           `json:"retry_attempts"`
	RetryDelay      time.Duration `json:"retry_delay"`
}

// Query 查询结构
type Query struct {
	SQL  string        `json:"sql"`
	Args []interface{} `json:"args"`
	Type QueryType     `json:"type"`
}

// QueryResult 查询结果
type QueryResult struct {
	RowsAffected int64         `json:"rows_affected"`
	LastInsertID int64         `json:"last_insert_id"`
	Duration     time.Duration `json:"duration"`
	Error        error         `json:"error,omitempty"`
}

// QueryResultSet 查询结果集
type QueryResultSet struct {
	Columns  []string        `json:"columns"`
	Rows     [][]interface{} `json:"rows"`
	RowCount int             `json:"row_count"`
	Duration time.Duration   `json:"duration"`
	Error    error           `json:"error,omitempty"`
}

// ExecutionResult 执行结果
type ExecutionResult struct {
	Success      bool          `json:"success"`
	RowsAffected int64         `json:"rows_affected"`
	Duration     time.Duration `json:"duration"`
	Message      string        `json:"message,omitempty"`
	Error        error         `json:"error,omitempty"`
}

// BatchStatement 批量语句
type BatchStatement struct {
	SQL  string        `json:"sql"`
	Args []interface{} `json:"args"`
	Type QueryType     `json:"type"`
}

// BatchResult 批量操作结果
type BatchResult struct {
	SuccessCount int                 `json:"success_count"`
	FailureCount int                 `json:"failure_count"`
	Results      []*ExecutionResult  `json:"results"`
	Duration     time.Duration       `json:"duration"`
	Errors       []error             `json:"errors,omitempty"`
}

// BulkUpdateItem 批量更新项
type BulkUpdateItem struct {
	Updates   map[string]interface{} `json:"updates"`
	Condition string                 `json:"condition"`
	Args      []interface{}          `json:"args"`
}

// BulkDeleteCondition 批量删除条件
type BulkDeleteCondition struct {
	Condition string        `json:"condition"`
	Args      []interface{} `json:"args"`
}

// DatabaseInfo 数据库信息
type DatabaseInfo struct {
	Name         string            `json:"name"`
	Version      string            `json:"version"`
	Driver       DatabaseDriver    `json:"driver"`
	Charset      string            `json:"charset"`
	Timezone     string            `json:"timezone"`
	MaxConnections int             `json:"max_connections"`
	CurrentConnections int         `json:"current_connections"`
	Uptime       time.Duration     `json:"uptime"`
	Capabilities *DatabaseCapabilities `json:"capabilities"`
}

// TableInfo 表信息
type TableInfo struct {
	Name        string            `json:"name"`
	Schema      string            `json:"schema"`
	Engine      string            `json:"engine"`
	Charset     string            `json:"charset"`
	Collation   string            `json:"collation"`
	RowCount    int64             `json:"row_count"`
	DataSize    int64             `json:"data_size"`
	IndexSize   int64             `json:"index_size"`
	CreatedAt   time.Time         `json:"created_at"`
	UpdatedAt   time.Time         `json:"updated_at"`
	Comment     string            `json:"comment,omitempty"`
	Columns     []*ColumnInfo     `json:"columns,omitempty"`
	Indexes     []*IndexInfo      `json:"indexes,omitempty"`
}

// ColumnInfo 列信息
type ColumnInfo struct {
	Name         string      `json:"name"`
	Type         string      `json:"type"`
	Length       int         `json:"length,omitempty"`
	Precision    int         `json:"precision,omitempty"`
	Scale        int         `json:"scale,omitempty"`
	Nullable     bool        `json:"nullable"`
	PrimaryKey   bool        `json:"primary_key"`
	AutoIncrement bool       `json:"auto_increment"`
	DefaultValue interface{} `json:"default_value,omitempty"`
	Comment      string      `json:"comment,omitempty"`
}

// IndexInfo 索引信息
type IndexInfo struct {
	Name       string   `json:"name"`
	Type       string   `json:"type"`
	Unique     bool     `json:"unique"`
	Primary    bool     `json:"primary"`
	Columns    []string `json:"columns"`
	Size       int64    `json:"size"`
	Cardinality int64   `json:"cardinality"`
}

// TableDefinition 表定义
type TableDefinition struct {
	Name       string              `json:"name"`
	Schema     string              `json:"schema,omitempty"`
	Columns    []*ColumnDefinition `json:"columns"`
	PrimaryKey []string            `json:"primary_key,omitempty"`
	Indexes    []*IndexDefinition  `json:"indexes,omitempty"`
	Engine     string              `json:"engine,omitempty"`
	Charset    string              `json:"charset,omitempty"`
	Comment    string              `json:"comment,omitempty"`
}

// ColumnDefinition 列定义
type ColumnDefinition struct {
	Name         string      `json:"name"`
	Type         string      `json:"type"`
	Length       int         `json:"length,omitempty"`
	Precision    int         `json:"precision,omitempty"`
	Scale        int         `json:"scale,omitempty"`
	Nullable     bool        `json:"nullable"`
	AutoIncrement bool       `json:"auto_increment"`
	DefaultValue interface{} `json:"default_value,omitempty"`
	Comment      string      `json:"comment,omitempty"`
}

// IndexDefinition 索引定义
type IndexDefinition struct {
	Name    string   `json:"name"`
	Table   string   `json:"table"`
	Type    string   `json:"type"`
	Unique  bool     `json:"unique"`
	Columns []string `json:"columns"`
	Comment string   `json:"comment,omitempty"`
}

// TableAlteration 表变更
type TableAlteration struct {
	Type   AlterationType `json:"type"`
	Target string         `json:"target"`
	Definition interface{} `json:"definition,omitempty"`
}

// DatabaseCapabilities 数据库能力
type DatabaseCapabilities struct {
	SupportsTransactions      bool `json:"supports_transactions"`
	SupportsSavepoints        bool `json:"supports_savepoints"`
	SupportsNamedParameters   bool `json:"supports_named_parameters"`
	SupportsMultipleResultSets bool `json:"supports_multiple_result_sets"`
	SupportsBatchUpdates      bool `json:"supports_batch_updates"`
	SupportsStoredProcedures  bool `json:"supports_stored_procedures"`
	SupportsCallableStatements bool `json:"supports_callable_statements"`
	MaxConnections            int  `json:"max_connections"`
	MaxStatementLength        int  `json:"max_statement_length"`
	MaxTableNameLength        int  `json:"max_table_name_length"`
	MaxColumnNameLength       int  `json:"max_column_name_length"`
}

// DatabaseStatistics 数据库统计信息
type DatabaseStatistics struct {
	ConnectionCount    int           `json:"connection_count"`
	ActiveConnections  int           `json:"active_connections"`
	IdleConnections    int           `json:"idle_connections"`
	QueriesPerSecond   float64       `json:"queries_per_second"`
	SlowQueryCount     int64         `json:"slow_query_count"`
	AvgQueryTime       time.Duration `json:"avg_query_time"`
	CacheHitRatio      float64       `json:"cache_hit_ratio"`
	BufferPoolUsage    float64       `json:"buffer_pool_usage"`
	UpdatedAt          time.Time     `json:"updated_at"`
}

// TableStatistics 表统计信息
type TableStatistics struct {
	RowCount      int64     `json:"row_count"`
	DataSize      int64     `json:"data_size"`
	IndexSize     int64     `json:"index_size"`
	ReadCount     int64     `json:"read_count"`
	WriteCount    int64     `json:"write_count"`
	LastRead      time.Time `json:"last_read"`
	LastWrite     time.Time `json:"last_write"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// HealthStatus 健康状态
type HealthStatus struct {
	Status        HealthStatusType `json:"status"`
	Message       string           `json:"message"`
	ResponseTime  time.Duration    `json:"response_time"`
	Connections   *ConnectionHealth `json:"connections"`
	Performance   *PerformanceHealth `json:"performance"`
	CheckedAt     time.Time        `json:"checked_at"`
}

// ConnectionHealth 连接健康状态
type ConnectionHealth struct {
	Total     int `json:"total"`
	Active    int `json:"active"`
	Idle      int `json:"idle"`
	Failed    int `json:"failed"`
}

// PerformanceHealth 性能健康状态
type PerformanceHealth struct {
	AvgResponseTime time.Duration `json:"avg_response_time"`
	SlowQueries     int64         `json:"slow_queries"`
	ErrorRate       float64       `json:"error_rate"`
}

// PoolStatistics 连接池统计信息
type PoolStatistics struct {
	MaxOpenConnections     int           `json:"max_open_connections"`
	OpenConnections        int           `json:"open_connections"`
	InUse                  int           `json:"in_use"`
	Idle                   int           `json:"idle"`
	WaitCount              int64         `json:"wait_count"`
	WaitDuration           time.Duration `json:"wait_duration"`
	MaxIdleClosed          int64         `json:"max_idle_closed"`
	MaxLifetimeClosed      int64         `json:"max_lifetime_closed"`
	MaxIdleTimeClosed      int64         `json:"max_idle_time_closed"`
}

// RetryPolicy 重试策略
type RetryPolicy struct {
	MaxAttempts int           `json:"max_attempts"`
	BaseDelay   time.Duration `json:"base_delay"`
	MaxDelay    time.Duration `json:"max_delay"`
	Multiplier  float64       `json:"multiplier"`
	Jitter      bool          `json:"jitter"`
}

// 枚举类型定义
type (
	DatabaseDriver    string
	DatabaseDialect   string
	SortDirection     string
	QueryType         string
	AlterationType    string
	HealthStatusType  string
)

// DatabaseDriver 常量
const (
	DriverStarRocks  DatabaseDriver = "starrocks"
	DriverClickHouse DatabaseDriver = "clickhouse"
	DriverDoris      DatabaseDriver = "doris"
	DriverMySQL      DatabaseDriver = "mysql"
	DriverPostgreSQL DatabaseDriver = "postgresql"
)

// DatabaseDialect 常量
const (
	DialectStarRocks  DatabaseDialect = "starrocks"
	DialectClickHouse DatabaseDialect = "clickhouse"
	DialectDoris      DatabaseDialect = "doris"
	DialectMySQL      DatabaseDialect = "mysql"
	DialectPostgreSQL DatabaseDialect = "postgresql"
	DialectStandard   DatabaseDialect = "standard"
)

// SortDirection 常量
const (
	SortAscending  SortDirection = "ASC"
	SortDescending SortDirection = "DESC"
)

// QueryType 常量
const (
	QueryTypeSelect QueryType = "SELECT"
	QueryTypeInsert QueryType = "INSERT"
	QueryTypeUpdate QueryType = "UPDATE"
	QueryTypeDelete QueryType = "DELETE"
	QueryTypeDDL    QueryType = "DDL"
	QueryTypeDML    QueryType = "DML"
)

// AlterationType 常量
const (
	AlterationAddColumn    AlterationType = "ADD_COLUMN"
	AlterationDropColumn   AlterationType = "DROP_COLUMN"
	AlterationModifyColumn AlterationType = "MODIFY_COLUMN"
	AlterationAddIndex     AlterationType = "ADD_INDEX"
	AlterationDropIndex    AlterationType = "DROP_INDEX"
	AlterationRename       AlterationType = "RENAME"
)

// HealthStatusType 常量
const (
	HealthStatusHealthy   HealthStatusType = "healthy"
	HealthStatusUnhealthy HealthStatusType = "unhealthy"
	HealthStatusDegraded  HealthStatusType = "degraded"
	HealthStatusUnknown   HealthStatusType = "unknown"
)

//Personal.AI order the ending
