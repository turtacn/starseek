package index

import (
	"context"
	"time"

	"github.com/turtacn/starseek/internal/common/types"
)

// =============================================================================
// Index Domain Repository Interface
// =============================================================================

// IndexRepository defines the interface for index domain data operations
type IndexRepository interface {
	// =============================================================================
	// Core CRUD Operations
	// =============================================================================

	// Create creates a new index
	Create(ctx context.Context, index *Index) error

	// Update updates an existing index
	Update(ctx context.Context, index *Index) error

	// Delete permanently deletes an index
	Delete(ctx context.Context, indexID IndexID) error

	// GetByID retrieves an index by ID
	GetByID(ctx context.Context, indexID IndexID) (*Index, error)

	// GetByName retrieves an index by name
	GetByName(ctx context.Context, name string) (*Index, error)

	// Exists checks if an index exists
	Exists(ctx context.Context, indexID IndexID) (bool, error)

	// ExistsByName checks if an index exists by name
	ExistsByName(ctx context.Context, name string) (bool, error)

	// =============================================================================
	// Query and List Operations
	// =============================================================================

	// List lists indexes based on criteria
	List(ctx context.Context, options *IndexListOptions) ([]*Index, int64, error)

	// ListByStatus lists indexes by status
	ListByStatus(ctx context.Context, status IndexStatus) ([]*Index, error)

	// ListByCreator lists indexes by creator
	ListByCreator(ctx context.Context, creator string) ([]*Index, error)

	// ListByDateRange lists indexes created within a date range
	ListByDateRange(ctx context.Context, startDate, endDate time.Time) ([]*Index, error)

	// Search searches indexes by name or description
	Search(ctx context.Context, query string, options *SearchOptions) ([]*Index, int64, error)

	// =============================================================================
	// Index Configuration Operations
	// =============================================================================

	// SaveIndexConfig saves index configuration
	SaveIndexConfig(ctx context.Context, indexID IndexID, config *IndexConfig) error

	// GetIndexConfig retrieves index configuration
	GetIndexConfig(ctx context.Context, indexID IndexID) (*IndexConfig, error)

	// UpdateIndexConfig updates index configuration
	UpdateIndexConfig(ctx context.Context, indexID IndexID, config *IndexConfig) error

	// DeleteIndexConfig deletes index configuration
	DeleteIndexConfig(ctx context.Context, indexID IndexID) error

	// ValidateIndexConfig validates index configuration against database
	ValidateIndexConfig(ctx context.Context, config *IndexConfig) (*ConfigValidationResult, error)

	// =============================================================================
	// Index Column Operations
	// =============================================================================

	// AddColumn adds a column to an index
	AddColumn(ctx context.Context, indexID IndexID, column *IndexColumn) error

	// UpdateColumn updates an index column
	UpdateColumn(ctx context.Context, indexID IndexID, column *IndexColumn) error

	// RemoveColumn removes a column from an index
	RemoveColumn(ctx context.Context, indexID IndexID, columnName string) error

	// GetColumns retrieves all columns for an index
	GetColumns(ctx context.Context, indexID IndexID) ([]*IndexColumn, error)

	// GetColumn retrieves a specific column
	GetColumn(ctx context.Context, indexID IndexID, columnName string) (*IndexColumn, error)

	// ReorderColumns reorders columns in an index
	ReorderColumns(ctx context.Context, indexID IndexID, columnOrder []string) error

	// =============================================================================
	// Index Statistics Operations
	// =============================================================================

	// GetIndexStatistics retrieves index statistics
	GetIndexStatistics(ctx context.Context, indexID IndexID) (*IndexStatistics, error)

	// UpdateIndexStatistics updates index statistics
	UpdateIndexStatistics(ctx context.Context, indexID IndexID, stats *IndexStatistics) error

	// GetStatisticsHistory retrieves statistics history
	GetStatisticsHistory(ctx context.Context, indexID IndexID, timeRange *TimeRange) ([]*IndexStatistics, error)

	// GetAggregatedStatistics retrieves aggregated statistics
	GetAggregatedStatistics(ctx context.Context, indexIDs []IndexID, timeRange *TimeRange) (*AggregatedStatistics, error)

	// GetPerformanceMetrics retrieves performance metrics
	GetPerformanceMetrics(ctx context.Context, indexID IndexID, timeRange *TimeRange) (*PerformanceMetrics, error)

	// RecordPerformanceMetric records a performance metric
	RecordPerformanceMetric(ctx context.Context, indexID IndexID, metric *PerformanceMetric) error

	// =============================================================================
	// Database Metadata Operations
	// =============================================================================

	// GetTableMetadata retrieves table metadata from database
	GetTableMetadata(ctx context.Context, databaseName string) ([]*TableMetadata, error)

	// GetSpecificTableMetadata retrieves specific table metadata
	GetSpecificTableMetadata(ctx context.Context, databaseName, tableName string) (*TableMetadata, error)

	// GetIndexMetadata retrieves index metadata from database
	GetIndexMetadata(ctx context.Context, databaseName string) ([]*IndexMetadata, error)

	// GetSchemaMetadata retrieves schema metadata
	GetSchemaMetadata(ctx context.Context, databaseName string) ([]*SchemaMetadata, error)

	// GetColumnMetadata retrieves column metadata for a table
	GetColumnMetadata(ctx context.Context, databaseName, tableName string) ([]*ColumnMetadata, error)

	// GetDatabaseStatistics retrieves database statistics
	GetDatabaseStatistics(ctx context.Context, databaseName string) (*DatabaseStatistics, error)

	// DiscoverTableStructure discovers table structure and relationships
	DiscoverTableStructure(ctx context.Context, databaseName, tableName string) (*TableStructure, error)

	// GetExistingIndexes retrieves existing indexes for a table
	GetExistingIndexes(ctx context.Context, databaseName, tableName string) (*ExistingIndexes, error)

	// =============================================================================
	// Index Validation Operations
	// =============================================================================

	// ValidateIndexOnDatabase validates index configuration against actual database
	ValidateIndexOnDatabase(ctx context.Context, indexID IndexID) (*DatabaseValidationResult, error)

	// CheckIndexConsistency checks index consistency
	CheckIndexConsistency(ctx context.Context, indexID IndexID) (*ConsistencyCheckResult, error)

	// ValidateColumnMapping validates column mapping
	ValidateColumnMapping(ctx context.Context, indexID IndexID, mapping *ColumnMapping) (*MappingValidationResult, error)

	// CheckSchemaCompatibility checks schema compatibility
	CheckSchemaCompatibility(ctx context.Context, indexID IndexID, schema *Schema) (*CompatibilityCheckResult, error)

	// ValidateTokenizerConfig validates tokenizer configuration
	ValidateTokenizerConfig(ctx context.Context, config *TokenizerRule) (*TokenizerValidationResult, error)

	// =============================================================================
	// Batch Operations
	// =============================================================================

	// CreateMultiple creates multiple indexes
	CreateMultiple(ctx context.Context, indexes []*Index) (*BatchResult, error)

	// UpdateMultiple updates multiple indexes
	UpdateMultiple(ctx context.Context, indexes []*Index) (*BatchResult, error)

	// DeleteMultiple deletes multiple indexes
	DeleteMultiple(ctx context.Context, indexIDs []IndexID) (*BatchResult, error)

	// BatchUpdateStatus updates status for multiple indexes
	BatchUpdateStatus(ctx context.Context, indexIDs []IndexID, status IndexStatus, actor string) (*BatchResult, error)

	// BatchUpdateConfig updates configuration for multiple indexes
	BatchUpdateConfig(ctx context.Context, updates map[IndexID]*IndexConfig) (*BatchResult, error)

	// BatchGetStatistics retrieves statistics for multiple indexes
	BatchGetStatistics(ctx context.Context, indexIDs []IndexID) (map[IndexID]*IndexStatistics, error)

	// =============================================================================
	// Transaction Operations
	// =============================================================================

	// BeginTransaction begins a database transaction
	BeginTransaction(ctx context.Context) (TransactionContext, error)

	// CommitTransaction commits a transaction
	CommitTransaction(ctx context.Context, tx TransactionContext) error

	// RollbackTransaction rolls back a transaction
	RollbackTransaction(ctx context.Context, tx TransactionContext) error

	// WithTransaction executes operations within a transaction
	WithTransaction(ctx context.Context, fn func(ctx context.Context, tx TransactionContext) error) error

	// =============================================================================
	// Index Health and Monitoring
	// =============================================================================

	// GetIndexHealth retrieves index health status
	GetIndexHealth(ctx context.Context, indexID IndexID) (*IndexHealth, error)

	// UpdateIndexHealth updates index health status
	UpdateIndexHealth(ctx context.Context, indexID IndexID, health *IndexHealth) error

	// GetHealthHistory retrieves health history
	GetHealthHistory(ctx context.Context, indexID IndexID, timeRange *TimeRange) ([]*IndexHealth, error)

	// RecordHealthCheck records a health check result
	RecordHealthCheck(ctx context.Context, indexID IndexID, check *HealthCheck) error

	// GetSystemMetrics retrieves system-level metrics
	GetSystemMetrics(ctx context.Context) (*SystemMetrics, error)

	// =============================================================================
	// Query Pattern Analysis
	// =============================================================================

	// GetQueryPatterns retrieves query patterns for an index
	GetQueryPatterns(ctx context.Context, indexID IndexID, timeRange *TimeRange) ([]*QueryPattern, error)

	// RecordQueryPattern records a query pattern
	RecordQueryPattern(ctx context.Context, indexID IndexID, pattern *QueryPattern) error

	// GetQueryStatistics retrieves query statistics
	GetQueryStatistics(ctx context.Context, indexID IndexID, timeRange *TimeRange) (*QueryStatistics, error)

	// AnalyzeQueryPerformance analyzes query performance
	AnalyzeQueryPerformance(ctx context.Context, indexID IndexID, timeRange *TimeRange) (*QueryPerformanceAnalysis, error)

	// =============================================================================
	// Index Optimization
	// =============================================================================

	// GetOptimizationSuggestions retrieves optimization suggestions
	GetOptimizationSuggestions(ctx context.Context, indexID IndexID) ([]*OptimizationSuggestion, error)

	// RecordOptimizationAction records an optimization action
	RecordOptimizationAction(ctx context.Context, indexID IndexID, action *OptimizationAction) error

	// GetOptimizationHistory retrieves optimization history
	GetOptimizationHistory(ctx context.Context, indexID IndexID, timeRange *TimeRange) ([]*OptimizationAction, error)

	// =============================================================================
	// Index Usage Tracking
	// =============================================================================

	// RecordIndexUsage records index usage
	RecordIndexUsage(ctx context.Context, indexID IndexID, usage *IndexUsage) error

	// GetIndexUsageStatistics retrieves usage statistics
	GetIndexUsageStatistics(ctx context.Context, indexID IndexID, timeRange *TimeRange) (*IndexUsageStatistics, error)

	// GetUnusedIndexes retrieves unused indexes
	GetUnusedIndexes(ctx context.Context, threshold time.Duration) ([]*Index, error)

	// GetMostUsedIndexes retrieves most used indexes
	GetMostUsedIndexes(ctx context.Context, limit int, timeRange *TimeRange) ([]*IndexUsageReport, error)

	// =============================================================================
	// Index Backup and Recovery
	// =============================================================================

	// BackupIndexConfiguration backs up index configuration
	BackupIndexConfiguration(ctx context.Context, indexID IndexID) (*IndexBackup, error)

	// RestoreIndexConfiguration restores index configuration
	RestoreIndexConfiguration(ctx context.Context, indexID IndexID, backup *IndexBackup) error

	// GetBackupHistory retrieves backup history
	GetBackupHistory(ctx context.Context, indexID IndexID) ([]*IndexBackup, error)

	// DeleteBackup deletes a backup
	DeleteBackup(ctx context.Context, backupID string) error

	// =============================================================================
	// Index Lifecycle Management
	// =============================================================================

	// GetIndexLifecycleEvents retrieves lifecycle events
	GetIndexLifecycleEvents(ctx context.Context, indexID IndexID) ([]*IndexLifecycleEvent, error)

	// RecordLifecycleEvent records a lifecycle event
	RecordLifecycleEvent(ctx context.Context, indexID IndexID, event *IndexLifecycleEvent) error

	// GetIndexVersionHistory retrieves version history
	GetIndexVersionHistory(ctx context.Context, indexID IndexID) ([]*IndexVersion, error)

	// CreateIndexVersion creates a new version
	CreateIndexVersion(ctx context.Context, indexID IndexID, version *IndexVersion) error

	// =============================================================================
	// Index Relationships
	// =============================================================================

	// GetIndexDependencies retrieves index dependencies
	GetIndexDependencies(ctx context.Context, indexID IndexID) ([]*IndexDependency, error)

	// AddIndexDependency adds a dependency
	AddIndexDependency(ctx context.Context, indexID IndexID, dependency *IndexDependency) error

	// RemoveIndexDependency removes a dependency
	RemoveIndexDependency(ctx context.Context, indexID IndexID, dependencyID string) error

	// GetRelatedIndexes retrieves related indexes
	GetRelatedIndexes(ctx context.Context, indexID IndexID) ([]*Index, error)
}

// =============================================================================
// Repository Options and Parameters
// =============================================================================

// IndexListOptions defines options for listing indexes
type IndexListOptions struct {
	// Pagination
	Offset int
	Limit  int

	// Filtering
	Status   *IndexStatus
	Creator  *string
	NameLike *string

	// Date filtering
	CreatedAfter  *time.Time
	CreatedBefore *time.Time
	UpdatedAfter  *time.Time
	UpdatedBefore *time.Time

	// Sorting
	SortBy    string
	SortOrder types.SortOrder

	// Include options
	IncludeColumns    bool
	IncludeStatistics bool
	IncludeConfig     bool
	IncludeHealth     bool
}

// SearchOptions defines options for searching indexes
type SearchOptions struct {
	// Pagination
	Offset int
	Limit  int

	// Search scope
	SearchInName        bool
	SearchInDescription bool
	SearchInColumns     bool

	// Filtering
	Status *IndexStatus

	// Sorting
	SortBy    string
	SortOrder types.SortOrder

	// Highlight options
	HighlightResults bool

	// Include options
	IncludeColumns bool
	IncludeConfig  bool
}

// TimeRange represents a time range for queries
type TimeRange struct {
	StartTime time.Time
	EndTime   time.Time
}

// TransactionContext represents a database transaction context
type TransactionContext interface {
	// ID returns the transaction ID
	ID() string

	// IsActive returns whether the transaction is active
	IsActive() bool

	// Context returns the context associated with the transaction
	Context() context.Context
}

// =============================================================================
// Repository Result Types
// =============================================================================

// BatchResult represents the result of a batch operation
type BatchResult struct {
	// Total number of operations attempted
	TotalCount int

	// Number of successful operations
	SuccessCount int

	// Number of failed operations
	FailedCount int

	// Individual operation results
	Results []*OperationResult

	// Overall operation duration
	Duration time.Duration

	// Error details for failed operations
	Errors []error
}

// OperationResult represents the result of a single operation
type OperationResult struct {
	// Operation ID or index ID
	ID string

	// Operation type
	Operation string

	// Success status
	Success bool

	// Error message if failed
	Error string

	// Operation duration
	Duration time.Duration

	// Additional data
	Data map[string]interface{}
}

// ConfigValidationResult represents index configuration validation result
type ConfigValidationResult struct {
	// Is valid
	IsValid bool

	// Validation errors
	Errors []string

	// Validation warnings
	Warnings []string

	// Configuration suggestions
	Suggestions []string

	// Validation timestamp
	ValidatedAt time.Time
}

// DatabaseValidationResult represents database validation result
type DatabaseValidationResult struct {
	// Is valid
	IsValid bool

	// Database connection status
	DatabaseConnected bool

	// Table exists
	TableExists bool

	// Columns exist
	ColumnsExist map[string]bool

	// Index exists in database
	IndexExistsInDB bool

	// Validation errors
	Errors []string

	// Validation warnings
	Warnings []string

	// Validation timestamp
	ValidatedAt time.Time
}

// ConsistencyCheckResult represents consistency check result
type ConsistencyCheckResult struct {
	// Is consistent
	IsConsistent bool

	// Configuration consistency
	ConfigConsistent bool

	// Data consistency
	DataConsistent bool

	// Schema consistency
	SchemaConsistent bool

	// Inconsistency details
	Inconsistencies []*InconsistencyDetail

	// Check timestamp
	CheckedAt time.Time
}

// InconsistencyDetail represents an inconsistency detail
type InconsistencyDetail struct {
	// Type of inconsistency
	Type string

	// Description
	Description string

	// Severity
	Severity string

	// Affected component
	Component string

	// Suggested action
	SuggestedAction string
}

// MappingValidationResult represents column mapping validation result
type MappingValidationResult struct {
	// Is valid
	IsValid bool

	// Mapping errors
	Errors []string

	// Mapping warnings
	Warnings []string

	// Source field validation
	SourceFieldValid map[string]bool

	// Target field validation
	TargetFieldValid map[string]bool

	// Data type compatibility
	DataTypeCompatible map[string]bool

	// Validation timestamp
	ValidatedAt time.Time
}

// CompatibilityCheckResult represents schema compatibility check result
type CompatibilityCheckResult struct {
	// Is compatible
	IsCompatible bool

	// Compatibility version
	Version string

	// Breaking changes
	BreakingChanges []string

	// Non-breaking changes
	NonBreakingChanges []string

	// Migration required
	MigrationRequired bool

	// Migration suggestions
	MigrationSuggestions []string

	// Check timestamp
	CheckedAt time.Time
}

// TokenizerValidationResult represents tokenizer configuration validation result
type TokenizerValidationResult struct {
	// Is valid
	IsValid bool

	// Tokenizer type valid
	TokenizerTypeValid bool

	// Language valid
	LanguageValid bool

	// Custom dictionary valid
	CustomDictionaryValid bool

	// Stop words valid
	StopWordsValid bool

	// Synonyms valid
	SynonymsValid bool

	// Validation errors
	Errors []string

	// Validation warnings
	Warnings []string

	// Validation timestamp
	ValidatedAt time.Time
}

// =============================================================================
// Metadata Types
// =============================================================================

// TableMetadata represents database table metadata
type TableMetadata struct {
	// Table name
	Name string

	// Schema name
	Schema string

	// Table type
	Type string

	// Row count
	RowCount int64

	// Table size
	Size int64

	// Columns
	Columns []*ColumnMetadata

	// Indexes
	Indexes []*IndexMetadata

	// Foreign keys
	ForeignKeys []*ForeignKeyMetadata

	// Creation time
	CreatedAt time.Time

	// Last modified time
	ModifiedAt time.Time
}

// ColumnMetadata represents database column metadata
type ColumnMetadata struct {
	// Column name
	Name string

	// Data type
	DataType string

	// Is nullable
	Nullable bool

	// Default value
	DefaultValue interface{}

	// Is primary key
	IsPrimaryKey bool

	// Is unique
	IsUnique bool

	// Is indexed
	IsIndexed bool

	// Column size
	Size int64

	// Precision
	Precision int

	// Scale
	Scale int

	// Character set
	CharacterSet string

	// Collation
	Collation string

	// Comments
	Comments string
}

// IndexMetadata represents database index metadata
type IndexMetadata struct {
	// Index name
	Name string

	// Table name
	TableName string

	// Index type
	Type string

	// Is unique
	IsUnique bool

	// Is primary
	IsPrimary bool

	// Columns
	Columns []string

	// Column order
	ColumnOrder []string

	// Index size
	Size int64

	// Cardinality
	Cardinality int64

	// Creation time
	CreatedAt time.Time

	// Last used time
	LastUsedAt *time.Time
}

// SchemaMetadata represents database schema metadata
type SchemaMetadata struct {
	// Schema name
	Name string

	// Owner
	Owner string

	// Tables
	Tables []*TableMetadata

	// Views
	Views []*ViewMetadata

	// Procedures
	Procedures []*ProcedureMetadata

	// Functions
	Functions []*FunctionMetadata

	// Creation time
	CreatedAt time.Time

	// Last modified time
	ModifiedAt time.Time
}

// ViewMetadata represents database view metadata
type ViewMetadata struct {
	// View name
	Name string

	// Schema name
	Schema string

	// Definition
	Definition string

	// Columns
	Columns []*ColumnMetadata

	// Is materialized
	IsMaterialized bool

	// Creation time
	CreatedAt time.Time

	// Last modified time
	ModifiedAt time.Time
}

// ProcedureMetadata represents database procedure metadata
type ProcedureMetadata struct {
	// Procedure name
	Name string

	// Schema name
	Schema string

	// Definition
	Definition string

	// Parameters
	Parameters []*ParameterMetadata

	// Return type
	ReturnType string

	// Creation time
	CreatedAt time.Time

	// Last modified time
	ModifiedAt time.Time
}

// FunctionMetadata represents database function metadata
type FunctionMetadata struct {
	// Function name
	Name string

	// Schema name
	Schema string

	// Definition
	Definition string

	// Parameters
	Parameters []*ParameterMetadata

	// Return type
	ReturnType string

	// Is deterministic
	IsDeterministic bool

	// Creation time
	CreatedAt time.Time

	// Last modified time
	ModifiedAt time.Time
}

// ParameterMetadata represents parameter metadata
type ParameterMetadata struct {
	// Parameter name
	Name string

	// Data type
	DataType string

	// Parameter mode (IN, OUT, INOUT)
	Mode string

	// Default value
	DefaultValue interface{}

	// Is nullable
	Nullable bool
}

// ForeignKeyMetadata represents foreign key metadata
type ForeignKeyMetadata struct {
	// Foreign key name
	Name string

	// Source columns
	SourceColumns []string

	// Target table
	TargetTable string

	// Target columns
	TargetColumns []string

	// Update rule
	UpdateRule string

	// Delete rule
	DeleteRule string

	// Is deferrable
	IsDeferrable bool

	// Initially deferred
	InitiallyDeferred bool
}

// DatabaseStatistics represents database statistics
type DatabaseStatistics struct {
	// Database name
	DatabaseName string

	// Total tables
	TotalTables int64

	// Total indexes
	TotalIndexes int64

	// Total size
	TotalSize int64

	// Total rows
	TotalRows int64

	// Average table size
	AverageTableSize float64

	// Average index size
	AverageIndexSize float64

	// Collection time
	CollectedAt time.Time
}

// TableStructure represents table structure with relationships
type TableStructure struct {
	// Table metadata
	Table *TableMetadata

	// Column details
	Columns []*ColumnDetail

	// Index suggestions
	IndexSuggestions []*IndexSuggestion

	// Relationships
	Relationships []*TableRelationship

	// Data distribution
	DataDistribution *DataDistribution

	// Discovery time
	DiscoveredAt time.Time
}

// ColumnDetail represents detailed column information
type ColumnDetail struct {
	// Column metadata
	Metadata *ColumnMetadata

	// Data distribution
	Distribution *ColumnDistribution

	// Index suitability
	IndexSuitability *IndexSuitability

	// Suggested configurations
	SuggestedConfigs []*ColumnConfig
}

// IndexSuggestion represents an index suggestion
type IndexSuggestion struct {
	// Suggestion type
	Type string

	// Suggested columns
	Columns []string

	// Index type
	IndexType string

	// Priority
	Priority int

	// Reason
	Reason string

	// Estimated benefit
	EstimatedBenefit float64

	// Configuration
	Config *IndexConfig
}

// TableRelationship represents a table relationship
type TableRelationship struct {
	// Relationship type
	Type string

	// Source table
	SourceTable string

	// Target table
	TargetTable string

	// Join columns
	JoinColumns []string

	// Relationship strength
	Strength float64

	// Cardinality
	Cardinality string
}

// DataDistribution represents data distribution information
type DataDistribution struct {
	// Row count
	RowCount int64

	// Data size
	DataSize int64

	// Average row size
	AverageRowSize float64

	// Column distributions
	ColumnDistributions map[string]*ColumnDistribution

	// Growth rate
	GrowthRate float64

	// Update frequency
	UpdateFrequency float64
}

// ColumnDistribution represents column data distribution
type ColumnDistribution struct {
	// Column name
	ColumnName string

	// Unique values
	UniqueValues int64

	// Null values
	NullValues int64

	// Min value
	MinValue interface{}

	// Max value
	MaxValue interface{}

	// Average value
	AverageValue interface{}

	// Most frequent values
	MostFrequentValues []interface{}

	// Cardinality
	Cardinality float64

	// Distribution type
	DistributionType string
}

// IndexSuitability represents index suitability for a column
type IndexSuitability struct {
	// Column name
	ColumnName string

	// Suitability score
	Score float64

	// Recommended index types
	RecommendedTypes []string

	// Reasons
	Reasons []string

	// Considerations
	Considerations []string
}

// ColumnConfig represents suggested column configuration
type ColumnConfig struct {
	// Configuration type
	Type string

	// Parameters
	Parameters map[string]interface{}

	// Priority
	Priority int

	// Reason
	Reason string

	// Estimated impact
	EstimatedImpact float64
}

// ExistingIndexes represents existing indexes in database
type ExistingIndexes struct {
	// Database name
	DatabaseName string

	// Table name
	TableName string

	// Indexes
	Indexes []*ExistingIndex

	// Collection time
	CollectedAt time.Time
}

// ExistingIndex represents an existing index
type ExistingIndex struct {
	// Index metadata
	Metadata *IndexMetadata

	// Usage statistics
	Usage *IndexUsage

	// Performance metrics
	Performance *IndexPerformance

	// Health status
	Health *IndexHealth

	// Optimization suggestions
	Suggestions []*OptimizationSuggestion
}

// =============================================================================
// Supporting Types
// =============================================================================

// AggregatedStatistics represents aggregated statistics across indexes
type AggregatedStatistics struct {
	// Time range
	TimeRange *TimeRange

	// Index count
	IndexCount int64

	// Total documents
	TotalDocuments int64

	// Total size
	TotalSize int64

	// Average response time
	AverageResponseTime time.Duration

	// Total searches
	TotalSearches int64

	// Total updates
	TotalUpdates int64

	// Statistics per index
	IndexStatistics map[IndexID]*IndexStatistics

	// Collection time
	CollectedAt time.Time
}

// PerformanceMetrics represents performance metrics
type PerformanceMetrics struct {
	// Time range
	TimeRange *TimeRange

	// Metrics
	Metrics map[string]*MetricData

	// Percentiles
	Percentiles map[string]map[int]float64

	// Aggregations
	Aggregations map[string]float64

	// Collection time
	CollectedAt time.Time
}

// MetricData represents metric data points
type MetricData struct {
	// Metric name
	Name string

	// Data points
	DataPoints []*DataPoint

	// Unit
	Unit string

	// Aggregation type
	AggregationType string
}

// DataPoint represents a single data point
type DataPoint struct {
	// Timestamp
	Timestamp time.Time

	// Value
	Value float64

	// Labels
	Labels map[string]string
}

// PerformanceMetric represents a single performance metric
type PerformanceMetric struct {
	// Index ID
	IndexID IndexID

	// Metric name
	Name string

	// Metric value
	Value float64

	// Metric unit
	Unit string

	// Labels
	Labels map[string]string

	// Timestamp
	Timestamp time.Time
}

// =============================================================================
// Health and Monitoring Types
// =============================================================================

// IndexHealth represents index health status
type IndexHealth struct {
	// Index ID
	IndexID IndexID

	// Health status
	Status IndexHealthStatus

	// Health score
	Score float64

	// Last check time
	LastCheckTime time.Time

	// Issues
	Issues []*HealthIssue

	// Metrics
	Metrics map[string]float64

	// Recommendations
	Recommendations []*HealthRecommendation
}

// HealthCheck represents a health check
type HealthCheck struct {
	// Check ID
	ID string

	// Index ID
	IndexID IndexID

	// Check type
	Type string

	// Check result
	Result string

	// Check status
	Status string

	// Check time
	CheckTime time.Time

	// Duration
	Duration time.Duration

	// Details
	Details map[string]interface{}
}

// HealthIssue represents a health issue
type HealthIssue struct {
	// Issue type
	Type string

	// Severity
	Severity string

	// Description
	Description string

	// Impact
	Impact string

	// Suggested action
	SuggestedAction string

	// First detected
	FirstDetected time.Time

	// Last detected
	LastDetected time.Time
}

// HealthRecommendation represents a health recommendation
type HealthRecommendation struct {
	// Recommendation type
	Type string

	// Priority
	Priority int

	// Description
	Description string

	// Action
	Action string

	// Expected benefit
	ExpectedBenefit string

	// Effort required
	EffortRequired string
}

// SystemMetrics represents system-level metrics
type SystemMetrics struct {
	// CPU usage
	CPUUsage float64

	// Memory usage
	MemoryUsage float64

	// Disk usage
	DiskUsage float64

	// Network usage
	NetworkUsage float64

	// Active connections
	ActiveConnections int64

	// Total indexes
	TotalIndexes int64

	// Healthy indexes
	HealthyIndexes int64

	// Collection time
	CollectedAt time.Time
}

// =============================================================================
// Query Pattern Types
// =============================================================================

// QueryPattern represents a query pattern
type QueryPattern struct {
	// Pattern ID
	ID string

	// Index ID
	IndexID IndexID

	// Pattern type
	Type string

	// Query template
	QueryTemplate string

	// Frequency
	Frequency int64

	// Average response time
	AverageResponseTime time.Duration

	// Success rate
	SuccessRate float64

	// First seen
	FirstSeen time.Time

	// Last seen
	LastSeen time.Time

	// Parameters
	Parameters map[string]interface{}
}

// QueryStatistics represents query statistics
type QueryStatistics struct {
	// Index ID
	IndexID IndexID

	// Time range
	TimeRange *TimeRange

	// Total queries
	TotalQueries int64

	// Successful queries
	SuccessfulQueries int64

	// Failed queries
	FailedQueries int64

	// Average response time
	AverageResponseTime time.Duration

	// P95 response time
	P95ResponseTime time.Duration

	// P99 response time
	P99ResponseTime time.Duration

	// Query patterns
	QueryPatterns []*QueryPattern

	// Collection time
	CollectedAt time.Time
}

// QueryPerformanceAnalysis represents query performance analysis
type QueryPerformanceAnalysis struct {
	// Index ID
	IndexID IndexID

	// Time range
	TimeRange *TimeRange

	// Slow queries
	SlowQueries []*SlowQuery

	// Performance trends
	PerformanceTrends []*PerformanceTrend

	// Bottlenecks
	Bottlenecks []*PerformanceBottleneck

	// Recommendations
	Recommendations []*PerformanceRecommendation

	// Analysis time
	AnalysisTime time.Time
}

// SlowQuery represents a slow query
type SlowQuery struct {
	// Query ID
	ID string

	// Query text
	QueryText string

	// Response time
	ResponseTime time.Duration

	// Execution count
	ExecutionCount int64

	// First seen
	FirstSeen time.Time

	// Last seen
	LastSeen time.Time

	// Execution plan
	ExecutionPlan string

	// Optimization suggestions
	OptimizationSuggestions []string
}

// PerformanceTrend represents a performance trend
type PerformanceTrend struct {
	// Metric name
	MetricName string

	// Trend direction
	Direction string

	// Trend strength
	Strength float64

	// Data points
	DataPoints []*TrendDataPoint

	// Prediction
	Prediction *TrendPrediction
}

// TrendDataPoint represents a trend data point
type TrendDataPoint struct {
	// Timestamp
	Timestamp time.Time

	// Value
	Value float64

	// Moving average
	MovingAverage float64
}

// TrendPrediction represents a trend prediction
type TrendPrediction struct {
	// Predicted value
	PredictedValue float64

	// Confidence
	Confidence float64

	// Prediction time
	PredictionTime time.Time

	// Prediction horizon
	PredictionHorizon time.Duration
}

// PerformanceBottleneck represents a performance bottleneck
type PerformanceBottleneck struct {
	// Bottleneck type
	Type string

	// Component
	Component string

	// Severity
	Severity string

	// Impact
	Impact string

	// Root cause
	RootCause string

	// Recommendations
	Recommendations []string
}

// PerformanceRecommendation represents a performance recommendation
type PerformanceRecommendation struct {
	// Recommendation type
	Type string

	// Priority
	Priority int

	// Description
	Description string

	// Action
	Action string

	// Expected improvement
	ExpectedImprovement string

	// Implementation effort
	ImplementationEffort string
}

// =============================================================================
// Optimization Types
// =============================================================================

// OptimizationSuggestion represents an optimization suggestion
type OptimizationSuggestion struct {
	// Suggestion ID
	ID string

	// Index ID
	IndexID IndexID

	// Suggestion type
	Type string

	// Priority
	Priority int

	// Description
	Description string

	// Action
	Action string

	// Expected benefit
	ExpectedBenefit string

	// Effort required
	EffortRequired string

	// Generated time
	GeneratedAt time.Time

	// Status
	Status string
}

// OptimizationAction represents an optimization action
type OptimizationAction struct {
	// Action ID
	ID string

	// Index ID
	IndexID IndexID

	// Action type
	Type string

	// Action description
	Description string

	// Executed by
	ExecutedBy string

	// Executed at
	ExecutedAt time.Time

	// Result
	Result string

	// Before metrics
	BeforeMetrics map[string]float64

	// After metrics
	AfterMetrics map[string]float64

	// Improvement
	Improvement map[string]float64
}

// =============================================================================
// Usage Tracking Types
// =============================================================================

// IndexUsage represents index usage information
type IndexUsage struct {
	// Index ID
	IndexID IndexID

	// Usage type
	Type string

	// Usage count
	Count int64

	// Usage time
	UsageTime time.Time

	// User
	User string

	// Application
	Application string

	// Query type
	QueryType string

	// Response time
	ResponseTime time.Duration

	// Result count
	ResultCount int64
}

// IndexUsageStatistics represents index usage statistics
type IndexUsageStatistics struct {
	// Index ID
	IndexID IndexID

	// Time range
	TimeRange *TimeRange

	// Total usage
	TotalUsage int64

	// Unique users
	UniqueUsers int64

	// Usage by type
	UsageByType map[string]int64

	// Usage by hour
	UsageByHour map[int]int64

	// Usage by day
	UsageByDay map[string]int64

	// Top users
	TopUsers []*UserUsage

	// Top applications
	TopApplications []*ApplicationUsage

	// Collection time
	CollectedAt time.Time
}

// UserUsage represents user usage statistics
type UserUsage struct {
	// User ID
	UserID string

	// Usage count
	UsageCount int64

	// Average response time
	AverageResponseTime time.Duration

	// Most frequent query types
	MostFrequentQueryTypes []string
}

// ApplicationUsage represents application usage statistics
type ApplicationUsage struct {
	// Application name
	ApplicationName string

	// Usage count
	UsageCount int64

	// Average response time
	AverageResponseTime time.Duration

	// Most frequent query types
	MostFrequentQueryTypes []string
}

// IndexUsageReport represents index usage report
type IndexUsageReport struct {
	// Index information
	Index *Index

	// Usage statistics
	Statistics *IndexUsageStatistics

	// Usage ranking
	Ranking int

	// Usage score
	UsageScore float64

	// Recommendations
	Recommendations []string
}

// =============================================================================
// Backup and Recovery Types
// =============================================================================

// IndexBackup represents an index backup
type IndexBackup struct {
	// Backup ID
	ID string

	// Index ID
	IndexID IndexID

	// Backup type
	Type string

	// Backup data
	Data []byte

	// Backup size
	Size int64

	// Backup time
	BackupTime time.Time

	// Backup version
	Version string

	// Compression
	Compression string

	// Checksum
	Checksum string

	// Metadata
	Metadata map[string]interface{}
}

// =============================================================================
// Lifecycle Management Types
// =============================================================================

// IndexLifecycleEvent represents an index lifecycle event
type IndexLifecycleEvent struct {
	// Event ID
	ID string

	// Index ID
	IndexID IndexID

	// Event type
	Type string

	// Event description
	Description string

	// Actor
	Actor string

	// Event time
	EventTime time.Time

	// Before state
	BeforeState map[string]interface{}

	// After state
	AfterState map[string]interface{}

	// Metadata
	Metadata map[string]interface{}
}

// IndexVersion represents an index version
type IndexVersion struct {
	// Version ID
	ID string

	// Index ID
	IndexID IndexID

	// Version number
	Version int64

	// Version description
	Description string

	// Configuration snapshot
	ConfigSnapshot []byte

	// Created by
	CreatedBy string

	// Created at
	CreatedAt time.Time

	// Tags
	Tags []string

	// Metadata
	Metadata map[string]interface{}
}

// =============================================================================
// Relationship Types
// =============================================================================

// IndexDependency represents an index dependency
type IndexDependency struct {
	// Dependency ID
	ID string

	// Source index ID
	SourceIndexID IndexID

	// Target index ID
	TargetIndexID IndexID

	// Dependency type
	Type string

	// Description
	Description string

	// Strength
	Strength float64

	// Created at
	CreatedAt time.Time

	// Metadata
	Metadata map[string]interface{}
}

// ColumnMapping represents column mapping between source and target
type ColumnMapping struct {
	// Source field
	SourceField string

	// Target field
	TargetField string

	// Mapping type
	Type string

	// Transformation
	Transformation string

	// Validation rules
	ValidationRules []string

	// Metadata
	Metadata map[string]interface{}
}

// Schema represents a data schema
type Schema struct {
	// Schema name
	Name string

	// Version
	Version string

	// Fields
	Fields []*SchemaField

	// Metadata
	Metadata map[string]interface{}
}

// SchemaField represents a schema field
type SchemaField struct {
	// Field name
	Name string

	// Field type
	Type string

	// Required
	Required bool

	// Description
	Description string

	// Constraints
	Constraints []string

	// Metadata
	Metadata map[string]interface{}
}

//Personal.AI order the ending
