package index

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/turtacn/starseek/internal/common/errors"
)

// =============================================================================
// Index Entity - 索引实体
// =============================================================================

// Index represents an index entity in the domain
type Index struct {
	// Basic identification
	id          IndexID
	name        string
	displayName string
	description string

	// Configuration
	config      *IndexConfig
	columns     []*IndexColumn

	// Status and lifecycle
	status      IndexStatus
	health      IndexHealth

	// Statistics
	statistics  *IndexStatistics

	// Metadata
	createdAt   time.Time
	updatedAt   time.Time
	createdBy   string
	updatedBy   string
	version     int64

	// Business rules
	rules       []*IndexRule

	// Events
	events      []IndexEvent
}

// NewIndex creates a new index instance
func NewIndex(id IndexID, name string, config *IndexConfig, createdBy string) (*Index, error) {
	// Validate inputs
	if err := validateIndexID(id); err != nil {
		return nil, fmt.Errorf("invalid index ID: %w", err)
	}

	if err := validateIndexName(name); err != nil {
		return nil, fmt.Errorf("invalid index name: %w", err)
	}

	if config == nil {
		return nil, errors.NewValidationError("index config is required")
	}

	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid index config: %w", err)
	}

	now := time.Now()

	index := &Index{
		id:          id,
		name:        name,
		displayName: name,
		config:      config,
		columns:     make([]*IndexColumn, 0),
		status:      IndexStatusCreating,
		health:      IndexHealthUnknown,
		statistics:  NewIndexStatistics(),
		createdAt:   now,
		updatedAt:   now,
		createdBy:   createdBy,
		updatedBy:   createdBy,
		version:     1,
		rules:       make([]*IndexRule, 0),
		events:      make([]IndexEvent, 0),
	}

	// Add creation event
	index.addEvent(IndexEventCreated, "Index created", createdBy)

	return index, nil
}

// ID returns the index ID
func (i *Index) ID() IndexID {
	return i.id
}

// Name returns the index name
func (i *Index) Name() string {
	return i.name
}

// DisplayName returns the display name
func (i *Index) DisplayName() string {
	return i.displayName
}

// Description returns the description
func (i *Index) Description() string {
	return i.description
}

// Config returns the index configuration
func (i *Index) Config() *IndexConfig {
	return i.config
}

// Status returns the current status
func (i *Index) Status() IndexStatus {
	return i.status
}

// Health returns the current health status
func (i *Index) Health() IndexHealth {
	return i.health
}

// Statistics returns the index statistics
func (i *Index) Statistics() *IndexStatistics {
	return i.statistics
}

// Columns returns the index columns
func (i *Index) Columns() []*IndexColumn {
	return i.columns
}

// Version returns the entity version
func (i *Index) Version() int64 {
	return i.version
}

// CreatedAt returns the creation time
func (i *Index) CreatedAt() time.Time {
	return i.createdAt
}

// UpdatedAt returns the last update time
func (i *Index) UpdatedAt() time.Time {
	return i.updatedAt
}

// CreatedBy returns the creator
func (i *Index) CreatedBy() string {
	return i.createdBy
}

// UpdatedBy returns the last updater
func (i *Index) UpdatedBy() string {
	return i.updatedBy
}

// Events returns the domain events
func (i *Index) Events() []IndexEvent {
	return i.events
}

// ClearEvents clears all pending events
func (i *Index) ClearEvents() {
	i.events = make([]IndexEvent, 0)
}

// =============================================================================
// Index Business Methods
// =============================================================================

// UpdateDisplayName updates the display name
func (i *Index) UpdateDisplayName(displayName string, updatedBy string) error {
	if strings.TrimSpace(displayName) == "" {
		return errors.NewValidationError("display name cannot be empty")
	}

	if len(displayName) > 100 {
		return errors.NewValidationError("display name cannot exceed 100 characters")
	}

	i.displayName = displayName
	i.updatedAt = time.Now()
	i.updatedBy = updatedBy
	i.version++

	i.addEvent(IndexEventUpdated, "Display name updated", updatedBy)

	return nil
}

// UpdateDescription updates the description
func (i *Index) UpdateDescription(description string, updatedBy string) error {
	if len(description) > 1000 {
		return errors.NewValidationError("description cannot exceed 1000 characters")
	}

	i.description = description
	i.updatedAt = time.Now()
	i.updatedBy = updatedBy
	i.version++

	i.addEvent(IndexEventUpdated, "Description updated", updatedBy)

	return nil
}

// UpdateConfig updates the index configuration
func (i *Index) UpdateConfig(config *IndexConfig, updatedBy string) error {
	if config == nil {
		return errors.NewValidationError("index config is required")
	}

	if err := config.Validate(); err != nil {
		return fmt.Errorf("invalid index config: %w", err)
	}

	// Check if index can be updated
	if !i.CanUpdate() {
		return errors.NewBusinessError("index cannot be updated in current status: %s", i.status)
	}

	i.config = config
	i.updatedAt = time.Now()
	i.updatedBy = updatedBy
	i.version++

	i.addEvent(IndexEventConfigUpdated, "Configuration updated", updatedBy)

	return nil
}

// AddColumn adds a new column to the index
func (i *Index) AddColumn(column *IndexColumn, updatedBy string) error {
	if column == nil {
		return errors.NewValidationError("column is required")
	}

	if err := column.Validate(); err != nil {
		return fmt.Errorf("invalid column: %w", err)
	}

	// Check if column already exists
	if i.HasColumn(column.Name()) {
		return errors.NewBusinessError("column already exists: %s", column.Name())
	}

	// Check if index can be updated
	if !i.CanUpdate() {
		return errors.NewBusinessError("index cannot be updated in current status: %s", i.status)
	}

	i.columns = append(i.columns, column)
	i.updatedAt = time.Now()
	i.updatedBy = updatedBy
	i.version++

	i.addEvent(IndexEventColumnAdded, fmt.Sprintf("Column added: %s", column.Name()), updatedBy)

	return nil
}

// RemoveColumn removes a column from the index
func (i *Index) RemoveColumn(columnName string, updatedBy string) error {
	if strings.TrimSpace(columnName) == "" {
		return errors.NewValidationError("column name is required")
	}

	// Check if index can be updated
	if !i.CanUpdate() {
		return errors.NewBusinessError("index cannot be updated in current status: %s", i.status)
	}

	// Find and remove the column
	for j, column := range i.columns {
		if column.Name() == columnName {
			i.columns = append(i.columns[:j], i.columns[j+1:]...)
			i.updatedAt = time.Now()
			i.updatedBy = updatedBy
			i.version++

			i.addEvent(IndexEventColumnRemoved, fmt.Sprintf("Column removed: %s", columnName), updatedBy)

			return nil
		}
	}

	return errors.NewBusinessError("column not found: %s", columnName)
}

// HasColumn checks if the index has a specific column
func (i *Index) HasColumn(columnName string) bool {
	for _, column := range i.columns {
		if column.Name() == columnName {
			return true
		}
	}
	return false
}

// GetColumn returns a specific column by name
func (i *Index) GetColumn(columnName string) *IndexColumn {
	for _, column := range i.columns {
		if column.Name() == columnName {
			return column
		}
	}
	return nil
}

// =============================================================================
// Index Status Management
// =============================================================================

// Enable enables the index
func (i *Index) Enable(updatedBy string) error {
	if i.status == IndexStatusEnabled {
		return errors.NewBusinessError("index is already enabled")
	}

	if i.status == IndexStatusDeleted {
		return errors.NewBusinessError("cannot enable deleted index")
	}

	i.status = IndexStatusEnabled
	i.updatedAt = time.Now()
	i.updatedBy = updatedBy
	i.version++

	i.addEvent(IndexEventEnabled, "Index enabled", updatedBy)

	return nil
}

// Disable disables the index
func (i *Index) Disable(updatedBy string) error {
	if i.status == IndexStatusDisabled {
		return errors.NewBusinessError("index is already disabled")
	}

	if i.status == IndexStatusDeleted {
		return errors.NewBusinessError("cannot disable deleted index")
	}

	i.status = IndexStatusDisabled
	i.updatedAt = time.Now()
	i.updatedBy = updatedBy
	i.version++

	i.addEvent(IndexEventDisabled, "Index disabled", updatedBy)

	return nil
}

// StartRebuilding starts the index rebuilding process
func (i *Index) StartRebuilding(updatedBy string) error {
	if i.status == IndexStatusRebuilding {
		return errors.NewBusinessError("index is already rebuilding")
	}

	if i.status == IndexStatusDeleted {
		return errors.NewBusinessError("cannot rebuild deleted index")
	}

	i.status = IndexStatusRebuilding
	i.updatedAt = time.Now()
	i.updatedBy = updatedBy
	i.version++

	i.addEvent(IndexEventRebuildStarted, "Index rebuild started", updatedBy)

	return nil
}

// CompleteRebuilding completes the index rebuilding process
func (i *Index) CompleteRebuilding(updatedBy string) error {
	if i.status != IndexStatusRebuilding {
		return errors.NewBusinessError("index is not in rebuilding status")
	}

	i.status = IndexStatusEnabled
	i.updatedAt = time.Now()
	i.updatedBy = updatedBy
	i.version++

	i.addEvent(IndexEventRebuildCompleted, "Index rebuild completed", updatedBy)

	return nil
}

// MarkAsDeleted marks the index as deleted
func (i *Index) MarkAsDeleted(updatedBy string) error {
	if i.status == IndexStatusDeleted {
		return errors.NewBusinessError("index is already deleted")
	}

	i.status = IndexStatusDeleted
	i.updatedAt = time.Now()
	i.updatedBy = updatedBy
	i.version++

	i.addEvent(IndexEventDeleted, "Index deleted", updatedBy)

	return nil
}

// UpdateHealth updates the index health status
func (i *Index) UpdateHealth(health IndexHealth, updatedBy string) error {
	if i.health == health {
		return nil // No change needed
	}

	i.health = health
	i.updatedAt = time.Now()
	i.updatedBy = updatedBy
	i.version++

	i.addEvent(IndexEventHealthUpdated, fmt.Sprintf("Health updated to: %s", health), updatedBy)

	return nil
}

// =============================================================================
// Index Business Rules
// =============================================================================

// CanUpdate checks if the index can be updated
func (i *Index) CanUpdate() bool {
	return i.status == IndexStatusEnabled || i.status == IndexStatusDisabled
}

// CanDelete checks if the index can be deleted
func (i *Index) CanDelete() bool {
	return i.status != IndexStatusDeleted
}

// CanRebuild checks if the index can be rebuilt
func (i *Index) CanRebuild() bool {
	return i.status == IndexStatusEnabled || i.status == IndexStatusDisabled
}

// IsActive checks if the index is active
func (i *Index) IsActive() bool {
	return i.status == IndexStatusEnabled
}

// IsHealthy checks if the index is healthy
func (i *Index) IsHealthy() bool {
	return i.health == IndexHealthGreen
}

// =============================================================================
// Index Statistics Update
// =============================================================================

// UpdateStatistics updates the index statistics
func (i *Index) UpdateStatistics(stats *IndexStatistics, updatedBy string) error {
	if stats == nil {
		return errors.NewValidationError("statistics is required")
	}

	if err := stats.Validate(); err != nil {
		return fmt.Errorf("invalid statistics: %w", err)
	}

	i.statistics = stats
	i.updatedAt = time.Now()
	i.updatedBy = updatedBy
	i.version++

	i.addEvent(IndexEventStatisticsUpdated, "Statistics updated", updatedBy)

	return nil
}

// =============================================================================
// Index Validation
// =============================================================================

// Validate validates the index entity
func (i *Index) Validate() error {
	// Validate ID
	if err := validateIndexID(i.id); err != nil {
		return fmt.Errorf("invalid index ID: %w", err)
	}

	// Validate name
	if err := validateIndexName(i.name); err != nil {
		return fmt.Errorf("invalid index name: %w", err)
	}

	// Validate config
	if i.config == nil {
		return errors.NewValidationError("index config is required")
	}

	if err := i.config.Validate(); err != nil {
		return fmt.Errorf("invalid index config: %w", err)
	}

	// Validate columns
	for _, column := range i.columns {
		if err := column.Validate(); err != nil {
			return fmt.Errorf("invalid column %s: %w", column.Name(), err)
		}
	}

	// Validate statistics
	if i.statistics != nil {
		if err := i.statistics.Validate(); err != nil {
			return fmt.Errorf("invalid statistics: %w", err)
		}
	}

	return nil
}

// =============================================================================
// Private Helper Methods
// =============================================================================

// addEvent adds a domain event
func (i *Index) addEvent(eventType IndexEventType, message string, actor string) {
	event := IndexEvent{
		Type:      eventType,
		IndexID:   i.id,
		Message:   message,
		Actor:     actor,
		Timestamp: time.Now(),
		Data:      make(map[string]interface{}),
	}

	i.events = append(i.events, event)
}

// validateIndexID validates the index ID
func validateIndexID(id IndexID) error {
	if string(id) == "" {
		return errors.NewValidationError("index ID cannot be empty")
	}

	if len(string(id)) > 50 {
		return errors.NewValidationError("index ID cannot exceed 50 characters")
	}

	// Check format (alphanumeric, dash, underscore)
	matched, err := regexp.MatchString("^[a-zA-Z0-9_-]+$", string(id))
	if err != nil {
		return errors.NewValidationError("invalid index ID format")
	}

	if !matched {
		return errors.NewValidationError("index ID can only contain alphanumeric characters, dashes, and underscores")
	}

	return nil
}

// validateIndexName validates the index name
func validateIndexName(name string) error {
	if strings.TrimSpace(name) == "" {
		return errors.NewValidationError("index name cannot be empty")
	}

	if len(name) > 100 {
		return errors.NewValidationError("index name cannot exceed 100 characters")
	}

	// Check for invalid characters
	matched, err := regexp.MatchString("^[a-zA-Z0-9_-]+$", name)
	if err != nil {
		return errors.NewValidationError("invalid index name format")
	}

	if !matched {
		return errors.NewValidationError("index name can only contain alphanumeric characters, dashes, and underscores")
	}

	return nil
}

// =============================================================================
// Index Value Objects and Types
// =============================================================================

// IndexID represents the index identifier
type IndexID string

// String returns the string representation
func (id IndexID) String() string {
	return string(id)
}

// IndexStatus represents the index status
type IndexStatus string

const (
	IndexStatusCreating    IndexStatus = "creating"
	IndexStatusEnabled     IndexStatus = "enabled"
	IndexStatusDisabled    IndexStatus = "disabled"
	IndexStatusRebuilding  IndexStatus = "rebuilding"
	IndexStatusDeleted     IndexStatus = "deleted"
	IndexStatusError       IndexStatus = "error"
)

// String returns the string representation
func (s IndexStatus) String() string {
	return string(s)
}

// IsValid checks if the status is valid
func (s IndexStatus) IsValid() bool {
	switch s {
	case IndexStatusCreating, IndexStatusEnabled, IndexStatusDisabled,
		IndexStatusRebuilding, IndexStatusDeleted, IndexStatusError:
		return true
	default:
		return false
	}
}

// IndexHealth represents the index health status
type IndexHealth string

const (
	IndexHealthGreen   IndexHealth = "green"
	IndexHealthYellow  IndexHealth = "yellow"
	IndexHealthRed     IndexHealth = "red"
	IndexHealthUnknown IndexHealth = "unknown"
)

// String returns the string representation
func (h IndexHealth) String() string {
	return string(h)
}

// IsValid checks if the health status is valid
func (h IndexHealth) IsValid() bool {
	switch h {
	case IndexHealthGreen, IndexHealthYellow, IndexHealthRed, IndexHealthUnknown:
		return true
	default:
		return false
	}
}

// IndexEventType represents the type of index event
type IndexEventType string

const (
	IndexEventCreated            IndexEventType = "created"
	IndexEventUpdated           IndexEventType = "updated"
	IndexEventConfigUpdated     IndexEventType = "config_updated"
	IndexEventColumnAdded       IndexEventType = "column_added"
	IndexEventColumnRemoved     IndexEventType = "column_removed"
	IndexEventColumnUpdated     IndexEventType = "column_updated"
	IndexEventEnabled           IndexEventType = "enabled"
	IndexEventDisabled          IndexEventType = "disabled"
	IndexEventRebuildStarted    IndexEventType = "rebuild_started"
	IndexEventRebuildCompleted  IndexEventType = "rebuild_completed"
	IndexEventDeleted           IndexEventType = "deleted"
	IndexEventHealthUpdated     IndexEventType = "health_updated"
	IndexEventStatisticsUpdated IndexEventType = "statistics_updated"
)

// String returns the string representation
func (t IndexEventType) String() string {
	return string(t)
}

// IndexEvent represents a domain event for index operations
type IndexEvent struct {
	Type      IndexEventType         `json:"type"`
	IndexID   IndexID                `json:"index_id"`
	Message   string                 `json:"message"`
	Actor     string                 `json:"actor"`
	Timestamp time.Time              `json:"timestamp"`
	Data      map[string]interface{} `json:"data"`
}

// =============================================================================
// Index Column Entity - 索引列实体
// =============================================================================

// IndexColumn represents a column in an index
type IndexColumn struct {
	// Basic information
	name        string
	displayName string
	description string

	// Data type and configuration
	dataType    ColumnDataType
	isSearchable bool
	isSortable  bool
	isFilterable bool
	isAggregatable bool

	// Tokenizer configuration
	tokenizerRule *TokenizerRule

	// Field mapping
	sourceField string
	fieldPath   string

	// Boost and scoring
	boost       float64

	// Validation rules
	required    bool
	maxLength   int
	minLength   int
	pattern     string

	// Metadata
	createdAt   time.Time
	updatedAt   time.Time
	version     int64
}

// NewIndexColumn creates a new index column
func NewIndexColumn(name string, dataType ColumnDataType) (*IndexColumn, error) {
	if err := validateColumnName(name); err != nil {
		return nil, fmt.Errorf("invalid column name: %w", err)
	}

	if !dataType.IsValid() {
		return nil, errors.NewValidationError("invalid data type: %s", dataType)
	}

	now := time.Now()

	return &IndexColumn{
		name:        name,
		displayName: name,
		dataType:    dataType,
		isSearchable: true,
		isSortable:  true,
		isFilterable: true,
		isAggregatable: false,
		boost:       1.0,
		createdAt:   now,
		updatedAt:   now,
		version:     1,
	}, nil
}

// Name returns the column name
func (c *IndexColumn) Name() string {
	return c.name
}

// DisplayName returns the display name
func (c *IndexColumn) DisplayName() string {
	return c.displayName
}

// Description returns the description
func (c *IndexColumn) Description() string {
	return c.description
}

// DataType returns the data type
func (c *IndexColumn) DataType() ColumnDataType {
	return c.dataType
}

// IsSearchable returns if the column is searchable
func (c *IndexColumn) IsSearchable() bool {
	return c.isSearchable
}

// IsSortable returns if the column is sortable
func (c *IndexColumn) IsSortable() bool {
	return c.isSortable
}

// IsFilterable returns if the column is filterable
func (c *IndexColumn) IsFilterable() bool {
	return c.isFilterable
}

// IsAggregatable returns if the column is aggregatable
func (c *IndexColumn) IsAggregatable() bool {
	return c.isAggregatable
}

// TokenizerRule returns the tokenizer rule
func (c *IndexColumn) TokenizerRule() *TokenizerRule {
	return c.tokenizerRule
}

// SourceField returns the source field
func (c *IndexColumn) SourceField() string {
	return c.sourceField
}

// Boost returns the boost value
func (c *IndexColumn) Boost() float64 {
	return c.boost
}

// UpdateDisplayName updates the display name
func (c *IndexColumn) UpdateDisplayName(displayName string) error {
	if strings.TrimSpace(displayName) == "" {
		return errors.NewValidationError("display name cannot be empty")
	}

	if len(displayName) > 100 {
		return errors.NewValidationError("display name cannot exceed 100 characters")
	}

	c.displayName = displayName
	c.updatedAt = time.Now()
	c.version++

	return nil
}

// UpdateDescription updates the description
func (c *IndexColumn) UpdateDescription(description string) error {
	if len(description) > 500 {
		return errors.NewValidationError("description cannot exceed 500 characters")
	}

	c.description = description
	c.updatedAt = time.Now()
	c.version++

	return nil
}

// SetSearchable sets the searchable flag
func (c *IndexColumn) SetSearchable(searchable bool) {
	c.isSearchable = searchable
	c.updatedAt = time.Now()
	c.version++
}

// SetSortable sets the sortable flag
func (c *IndexColumn) SetSortable(sortable bool) {
	c.isSortable = sortable
	c.updatedAt = time.Now()
	c.version++
}

// SetFilterable sets the filterable flag
func (c *IndexColumn) SetFilterable(filterable bool) {
	c.isFilterable = filterable
	c.updatedAt = time.Now()
	c.version++
}

// SetAggregatable sets the aggregatable flag
func (c *IndexColumn) SetAggregatable(aggregatable bool) {
	c.isAggregatable = aggregatable
	c.updatedAt = time.Now()
	c.version++
}

// SetTokenizerRule sets the tokenizer rule
func (c *IndexColumn) SetTokenizerRule(rule *TokenizerRule) error {
	if rule != nil {
		if err := rule.Validate(); err != nil {
			return fmt.Errorf("invalid tokenizer rule: %w", err)
		}
	}

	c.tokenizerRule = rule
	c.updatedAt = time.Now()
	c.version++

	return nil
}

// SetSourceField sets the source field
func (c *IndexColumn) SetSourceField(sourceField string) error {
	if strings.TrimSpace(sourceField) == "" {
		return errors.NewValidationError("source field cannot be empty")
	}

	c.sourceField = sourceField
	c.updatedAt = time.Now()
	c.version++

	return nil
}

// SetBoost sets the boost value
func (c *IndexColumn) SetBoost(boost float64) error {
	if boost < 0 {
		return errors.NewValidationError("boost cannot be negative")
	}

	if boost > 100 {
		return errors.NewValidationError("boost cannot exceed 100")
	}

	c.boost = boost
	c.updatedAt = time.Now()
	c.version++

	return nil
}

// Validate validates the column
func (c *IndexColumn) Validate() error {
	if err := validateColumnName(c.name); err != nil {
		return fmt.Errorf("invalid column name: %w", err)
	}

	if !c.dataType.IsValid() {
		return errors.NewValidationError("invalid data type: %s", c.dataType)
	}

	if c.tokenizerRule != nil {
		if err := c.tokenizerRule.Validate(); err != nil {
			return fmt.Errorf("invalid tokenizer rule: %w", err)
		}
	}

	if c.boost < 0 || c.boost > 100 {
		return errors.NewValidationError("boost must be between 0 and 100")
	}

	return nil
}

// =============================================================================
// Index Statistics Entity - 索引统计信息实体
// =============================================================================

// IndexStatistics represents index statistics
type IndexStatistics struct {
	// Document statistics
	documentCount     int64
	totalSize         int64
	averageDocSize    float64

	// Index statistics
	termCount         int64
	dictionarySize    int64
	segmentCount      int64

	// Performance statistics
	searchCount       int64
	indexCount        int64
	updateCount       int64
	deleteCount       int64

	// Time-based statistics
	avgSearchTime     time.Duration
	avgIndexTime      time.Duration
	avgUpdateTime     time.Duration

	// Memory usage
	memoryUsage       int64
	diskUsage         int64
	cacheUsage        int64

	// Health metrics
	errorCount        int64
	warningCount      int64

	// Metadata
	lastUpdated       time.Time
	collectionTime    time.Time
	version           int64
}

// NewIndexStatistics creates a new index statistics
func NewIndexStatistics() *IndexStatistics {
	now := time.Now()
	return &IndexStatistics{
		lastUpdated:    now,
		collectionTime: now,
		version:        1,
	}
}

// DocumentCount returns the document count
func (s *IndexStatistics) DocumentCount() int64 {
	return s.documentCount
}

// TotalSize returns the total size
func (s *IndexStatistics) TotalSize() int64 {
	return s.totalSize
}

// AverageDocSize returns the average document size
func (s *IndexStatistics) AverageDocSize() float64 {
	return s.averageDocSize
}

// TermCount returns the term count
func (s *IndexStatistics) TermCount() int64 {
	return s.termCount
}

// DictionarySize returns the dictionary size
func (s *IndexStatistics) DictionarySize() int64 {
	return s.dictionarySize
}

// SegmentCount returns the segment count
func (s *IndexStatistics) SegmentCount() int64 {
	return s.segmentCount
}

// SearchCount returns the search count
func (s *IndexStatistics) SearchCount() int64 {
	return s.searchCount
}

// MemoryUsage returns the memory usage
func (s *IndexStatistics) MemoryUsage() int64 {
	return s.memoryUsage
}

// DiskUsage returns the disk usage
func (s *IndexStatistics) DiskUsage() int64 {
	return s.diskUsage
}

// LastUpdated returns the last updated time
func (s *IndexStatistics) LastUpdated() time.Time {
	return s.lastUpdated
}

// UpdateDocumentStatistics updates document statistics
func (s *IndexStatistics) UpdateDocumentStatistics(docCount int64, totalSize int64) {
	s.documentCount = docCount
	s.totalSize = totalSize
	if docCount > 0 {
		s.averageDocSize = float64(totalSize) / float64(docCount)
	}
	s.lastUpdated = time.Now()
	s.version++
}

// UpdateIndexStatistics updates index statistics
func (s *IndexStatistics) UpdateIndexStatistics(termCount, dictSize, segmentCount int64) {
	s.termCount = termCount
	s.dictionarySize = dictSize
	s.segmentCount = segmentCount
	s.lastUpdated = time.Now()
	s.version++
}

// UpdatePerformanceStatistics updates performance statistics
func (s *IndexStatistics) UpdatePerformanceStatistics(searchCount, indexCount, updateCount, deleteCount int64) {
	s.searchCount = searchCount
	s.indexCount = indexCount
	s.updateCount = updateCount
	s.deleteCount = deleteCount
	s.lastUpdated = time.Now()
	s.version++
}

// UpdateTimeStatistics updates time-based statistics
func (s *IndexStatistics) UpdateTimeStatistics(avgSearchTime, avgIndexTime, avgUpdateTime time.Duration) {
	s.avgSearchTime = avgSearchTime
	s.avgIndexTime = avgIndexTime
	s.avgUpdateTime = avgUpdateTime
	s.lastUpdated = time.Now()
	s.version++
}

// UpdateMemoryStatistics updates memory usage statistics
func (s *IndexStatistics) UpdateMemoryStatistics(memoryUsage, diskUsage, cacheUsage int64) {
	s.memoryUsage = memoryUsage
	s.diskUsage = diskUsage
	s.cacheUsage = cacheUsage
	s.lastUpdated = time.Now()
	s.version++
}

// UpdateHealthMetrics updates health metrics
func (s *IndexStatistics) UpdateHealthMetrics(errorCount, warningCount int64) {
	s.errorCount = errorCount
	s.warningCount = warningCount
	s.lastUpdated = time.Now()
	s.version++
}

// Validate validates the statistics
func (s *IndexStatistics) Validate() error {
	if s.documentCount < 0 {
		return errors.NewValidationError("document count cannot be negative")
	}

	if s.totalSize < 0 {
		return errors.NewValidationError("total size cannot be negative")
	}

	if s.termCount < 0 {
		return errors.NewValidationError("term count cannot be negative")
	}

	if s.dictionarySize < 0 {
		return errors.NewValidationError("dictionary size cannot be negative")
	}

	return nil
}

// =============================================================================
// Tokenizer Rule Entity - 分词规则实体
// =============================================================================

// TokenizerRule represents a tokenizer rule
type TokenizerRule struct {
	// Basic information
	name        string
	description string

	// Tokenizer configuration
	tokenizerType   TokenizerType
	language        string
	caseSensitive   bool

	// Custom dictionary
	customDictionary []string
	stopWords       []string
	synonyms        map[string][]string

	// Character filters
	charFilters     []*CharFilter

	// Token filters
	tokenFilters    []*TokenFilter

	// Advanced options
	minTokenLength  int
	maxTokenLength  int
	enableStemming  bool
	enableLemmatization bool

	// Metadata
	createdAt   time.Time
	updatedAt   time.Time
	version     int64
}

// NewTokenizerRule creates a new tokenizer rule
func NewTokenizerRule(name string, tokenizerType TokenizerType) (*TokenizerRule, error) {
	if err := validateTokenizerName(name); err != nil {
		return nil, fmt.Errorf("invalid tokenizer name: %w", err)
	}

	if !tokenizerType.IsValid() {
		return nil, errors.NewValidationError("invalid tokenizer type: %s", tokenizerType)
	}

	now := time.Now()

	return &TokenizerRule{
		name:            name,
		tokenizerType:   tokenizerType,
		language:        "en",
		caseSensitive:   false,
		customDictionary: make([]string, 0),
		stopWords:       make([]string, 0),
		synonyms:        make(map[string][]string),
		charFilters:     make([]*CharFilter, 0),
		tokenFilters:    make([]*TokenFilter, 0),
		minTokenLength:  1,
		maxTokenLength:  255,
		enableStemming:  false,
		enableLemmatization: false,
		createdAt:       now,
		updatedAt:       now,
		version:         1,
	}, nil
}

// Name returns the tokenizer name
func (r *TokenizerRule) Name() string {
	return r.name
}

// Description returns the description
func (r *TokenizerRule) Description() string {
	return r.description
}

// TokenizerType returns the tokenizer type
func (r *TokenizerRule) TokenizerType() TokenizerType {
	return r.tokenizerType
}

// Language returns the language
func (r *TokenizerRule) Language() string {
	return r.language
}

// IsCaseSensitive returns if the tokenizer is case sensitive
func (r *TokenizerRule) IsCaseSensitive() bool {
	return r.caseSensitive
}

// CustomDictionary returns the custom dictionary
func (r *TokenizerRule) CustomDictionary() []string {
	return r.customDictionary
}

// StopWords returns the stop words
func (r *TokenizerRule) StopWords() []string {
	return r.stopWords
}

// Synonyms returns the synonyms
func (r *TokenizerRule) Synonyms() map[string][]string {
	return r.synonyms
}

// UpdateDescription updates the description
func (r *TokenizerRule) UpdateDescription(description string) error {
	if len(description) > 500 {
		return errors.NewValidationError("description cannot exceed 500 characters")
	}

	r.description = description
	r.updatedAt = time.Now()
	r.version++

	return nil
}

// SetLanguage sets the language
func (r *TokenizerRule) SetLanguage(language string) error {
	if !isValidLanguage(language) {
		return errors.NewValidationError("invalid language: %s", language)
	}

	r.language = language
	r.updatedAt = time.Now()
	r.version++

	return nil
}

// SetCaseSensitive sets the case sensitivity
func (r *TokenizerRule) SetCaseSensitive(caseSensitive bool) {
	r.caseSensitive = caseSensitive
	r.updatedAt = time.Now()
	r.version++
}

// AddCustomDictionaryWord adds a word to the custom dictionary
func (r *TokenizerRule) AddCustomDictionaryWord(word string) error {
	if strings.TrimSpace(word) == "" {
		return errors.NewValidationError("word cannot be empty")
	}

	// Check if word already exists
	for _, existingWord := range r.customDictionary {
		if existingWord == word {
			return errors.NewBusinessError("word already exists in dictionary: %s", word)
		}
	}

	r.customDictionary = append(r.customDictionary, word)
	r.updatedAt = time.Now()
	r.version++

	return nil
}

// RemoveCustomDictionaryWord removes a word from the custom dictionary
func (r *TokenizerRule) RemoveCustomDictionaryWord(word string) error {
	for i, existingWord := range r.customDictionary {
		if existingWord == word {
			r.customDictionary = append(r.customDictionary[:i], r.customDictionary[i+1:]...)
			r.updatedAt = time.Now()
			r.version++
			return nil
		}
	}

	return errors.NewBusinessError("word not found in dictionary: %s", word)
}

// AddStopWord adds a stop word
func (r *TokenizerRule) AddStopWord(word string) error {
	if strings.TrimSpace(word) == "" {
		return errors.NewValidationError("stop word cannot be empty")
	}

	// Check if word already exists
	for _, existingWord := range r.stopWords {
		if existingWord == word {
			return errors.NewBusinessError("stop word already exists: %s", word)
		}
	}

	r.stopWords = append(r.stopWords, word)
	r.updatedAt = time.Now()
	r.version++

	return nil
}

// RemoveStopWord removes a stop word
func (r *TokenizerRule) RemoveStopWord(word string) error {
	for i, existingWord := range r.stopWords {
		if existingWord == word {
			r.stopWords = append(r.stopWords[:i], r.stopWords[i+1:]...)
			r.updatedAt = time.Now()
			r.version++
			return nil
		}
	}

	return errors.NewBusinessError("stop word not found: %s", word)
}

// AddSynonym adds a synonym mapping
func (r *TokenizerRule) AddSynonym(word string, synonyms []string) error {
	if strings.TrimSpace(word) == "" {
		return errors.NewValidationError("word cannot be empty")
	}

	if len(synonyms) == 0 {
		return errors.NewValidationError("synonyms cannot be empty")
	}

	r.synonyms[word] = synonyms
	r.updatedAt = time.Now()
	r.version++

	return nil
}

// RemoveSynonym removes a synonym mapping
func (r *TokenizerRule) RemoveSynonym(word string) error {
	if _, exists := r.synonyms[word]; !exists {
		return errors.NewBusinessError("synonym not found: %s", word)
	}

	delete(r.synonyms, word)
	r.updatedAt = time.Now()
	r.version++

	return nil
}

// SetTokenLength sets the token length constraints
func (r *TokenizerRule) SetTokenLength(minLength, maxLength int) error {
	if minLength < 1 {
		return errors.NewValidationError("minimum token length must be at least 1")
	}

	if maxLength < minLength {
		return errors.NewValidationError("maximum token length must be greater than minimum")
	}

	if maxLength > 1000 {
		return errors.NewValidationError("maximum token length cannot exceed 1000")
	}

	r.minTokenLength = minLength
	r.maxTokenLength = maxLength
	r.updatedAt = time.Now()
	r.version++

	return nil
}

// SetStemming enables or disables stemming
func (r *TokenizerRule) SetStemming(enabled bool) {
	r.enableStemming = enabled
	r.updatedAt = time.Now()
	r.version++
}

// SetLemmatization enables or disables lemmatization
func (r *TokenizerRule) SetLemmatization(enabled bool) {
	r.enableLemmatization = enabled
	r.updatedAt = time.Now()
	r.version++
}

// Validate validates the tokenizer rule
func (r *TokenizerRule) Validate() error {
	if err := validateTokenizerName(r.name); err != nil {
		return fmt.Errorf("invalid tokenizer name: %w", err)
	}

	if !r.tokenizerType.IsValid() {
		return errors.NewValidationError("invalid tokenizer type: %s", r.tokenizerType)
	}

	if !isValidLanguage(r.language) {
		return errors.NewValidationError("invalid language: %s", r.language)
	}

	if r.minTokenLength < 1 {
		return errors.NewValidationError("minimum token length must be at least 1")
	}

	if r.maxTokenLength < r.minTokenLength {
		return errors.NewValidationError("maximum token length must be greater than minimum")
	}

	return nil
}

// =============================================================================
// Index Configuration Entity - 索引配置实体
// =============================================================================

// IndexConfig represents index configuration
type IndexConfig struct {
	// Basic settings
	numberOfShards   int
	numberOfReplicas int

	// Analysis settings
	defaultAnalyzer  string
	searchAnalyzer   string

	// Performance settings
	refreshInterval  time.Duration
	maxResultWindow  int

	// Storage settings
	compressionType  CompressionType

	// Advanced settings
	settings map[string]interface{}

	// Metadata
	createdAt time.Time
	updatedAt time.Time
	version   int64
}

// NewIndexConfig creates a new index configuration
func NewIndexConfig() *IndexConfig {
	now := time.Now()
	return &IndexConfig{
		numberOfShards:   1,
		numberOfReplicas: 0,
		defaultAnalyzer:  "standard",
		searchAnalyzer:   "standard",
		refreshInterval:  time.Second,
		maxResultWindow:  10000,
		compressionType:  CompressionTypeDefault,
		settings:         make(map[string]interface{}),
		createdAt:        now,
		updatedAt:        now,
		version:          1,
	}
}

// NumberOfShards returns the number of shards
func (c *IndexConfig) NumberOfShards() int {
	return c.numberOfShards
}

// NumberOfReplicas returns the number of replicas
func (c *IndexConfig) NumberOfReplicas() int {
	return c.numberOfReplicas
}

// DefaultAnalyzer returns the default analyzer
func (c *IndexConfig) DefaultAnalyzer() string {
	return c.defaultAnalyzer
}

// SearchAnalyzer returns the search analyzer
func (c *IndexConfig) SearchAnalyzer() string {
	return c.searchAnalyzer
}

// RefreshInterval returns the refresh interval
func (c *IndexConfig) RefreshInterval() time.Duration {
	return c.refreshInterval
}

// MaxResultWindow returns the max result window
func (c *IndexConfig) MaxResultWindow() int {
	return c.maxResultWindow
}

// CompressionType returns the compression type
func (c *IndexConfig) CompressionType() CompressionType {
	return c.compressionType
}

// Settings returns the additional settings
func (c *IndexConfig) Settings() map[string]interface{} {
	return c.settings
}

// SetShardConfiguration sets the shard configuration
func (c *IndexConfig) SetShardConfiguration(shards, replicas int) error {
	if shards < 1 {
		return errors.NewValidationError("number of shards must be at least 1")
	}

	if replicas < 0 {
		return errors.NewValidationError("number of replicas cannot be negative")
	}

	c.numberOfShards = shards
	c.numberOfReplicas = replicas
	c.updatedAt = time.Now()
	c.version++

	return nil
}

// SetAnalyzers sets the analyzers
func (c *IndexConfig) SetAnalyzers(defaultAnalyzer, searchAnalyzer string) error {
	if strings.TrimSpace(defaultAnalyzer) == "" {
		return errors.NewValidationError("default analyzer cannot be empty")
	}

	if strings.TrimSpace(searchAnalyzer) == "" {
		return errors.NewValidationError("search analyzer cannot be empty")
	}

	c.defaultAnalyzer = defaultAnalyzer
	c.searchAnalyzer = searchAnalyzer
	c.updatedAt = time.Now()
	c.version++

	return nil
}

// SetRefreshInterval sets the refresh interval
func (c *IndexConfig) SetRefreshInterval(interval time.Duration) error {
	if interval < 0 {
		return errors.NewValidationError("refresh interval cannot be negative")
	}

	c.refreshInterval = interval
	c.updatedAt = time.Now()
	c.version++

	return nil
}

// SetMaxResultWindow sets the max result window
func (c *IndexConfig) SetMaxResultWindow(window int) error {
	if window < 1 {
		return errors.NewValidationError("max result window must be at least 1")
	}

	if window > 1000000 {
		return errors.NewValidationError("max result window cannot exceed 1,000,000")
	}

	c.maxResultWindow = window
	c.updatedAt = time.Now()
	c.version++

	return nil
}

// SetCompressionType sets the compression type
func (c *IndexConfig) SetCompressionType(compressionType CompressionType) error {
	if !compressionType.IsValid() {
		return errors.NewValidationError("invalid compression type: %s", compressionType)
	}

	c.compressionType = compressionType
	c.updatedAt = time.Now()
	c.version++

	return nil
}

// SetSetting sets a custom setting
func (c *IndexConfig) SetSetting(key string, value interface{}) error {
	if strings.TrimSpace(key) == "" {
		return errors.NewValidationError("setting key cannot be empty")
	}

	c.settings[key] = value
	c.updatedAt = time.Now()
	c.version++

	return nil
}

// GetSetting gets a custom setting
func (c *IndexConfig) GetSetting(key string) (interface{}, bool) {
	value, exists := c.settings[key]
	return value, exists
}

// RemoveSetting removes a custom setting
func (c *IndexConfig) RemoveSetting(key string) {
	delete(c.settings, key)
	c.updatedAt = time.Now()
	c.version++
}

// Validate validates the index configuration
func (c *IndexConfig) Validate() error {
	if c.numberOfShards < 1 {
		return errors.NewValidationError("number of shards must be at least 1")
	}

	if c.numberOfReplicas < 0 {
		return errors.NewValidationError("number of replicas cannot be negative")
	}

	if strings.TrimSpace(c.defaultAnalyzer) == "" {
		return errors.NewValidationError("default analyzer cannot be empty")
	}

	if strings.TrimSpace(c.searchAnalyzer) == "" {
		return errors.NewValidationError("search analyzer cannot be empty")
	}

	if c.refreshInterval < 0 {
		return errors.NewValidationError("refresh interval cannot be negative")
	}

	if c.maxResultWindow < 1 {
		return errors.NewValidationError("max result window must be at least 1")
	}

	if !c.compressionType.IsValid() {
		return errors.NewValidationError("invalid compression type: %s", c.compressionType)
	}

	return nil
}

// =============================================================================
// Supporting Types and Enums
// =============================================================================

// ColumnDataType represents the data type of a column
type ColumnDataType string

const (
	ColumnDataTypeText     ColumnDataType = "text"
	ColumnDataTypeKeyword  ColumnDataType = "keyword"
	ColumnDataTypeLong     ColumnDataType = "long"
	ColumnDataTypeInteger  ColumnDataType = "integer"
	ColumnDataTypeShort    ColumnDataType = "short"
	ColumnDataTypeByte     ColumnDataType = "byte"
	ColumnDataTypeDouble   ColumnDataType = "double"
	ColumnDataTypeFloat    ColumnDataType = "float"
	ColumnDataTypeHalfFloat ColumnDataType = "half_float"
	ColumnDataTypeScaledFloat ColumnDataType = "scaled_float"
	ColumnDataTypeDate     ColumnDataType = "date"
	ColumnDataTypeBoolean  ColumnDataType = "boolean"
	ColumnDataTypeBinary   ColumnDataType = "binary"
	ColumnDataTypeGeoPoint ColumnDataType = "geo_point"
	ColumnDataTypeGeoShape ColumnDataType = "geo_shape"
	ColumnDataTypeIP       ColumnDataType = "ip"
	ColumnDataTypeCompletion ColumnDataType = "completion"
	ColumnDataTypeTokenCount ColumnDataType = "token_count"
	ColumnDataTypeObject   ColumnDataType = "object"
	ColumnDataTypeNested   ColumnDataType = "nested"
)

// IsValid checks if the data type is valid
func (t ColumnDataType) IsValid() bool {
	switch t {
	case ColumnDataTypeText, ColumnDataTypeKeyword, ColumnDataTypeLong, ColumnDataTypeInteger,
		 ColumnDataTypeShort, ColumnDataTypeByte, ColumnDataTypeDouble, ColumnDataTypeFloat,
		 ColumnDataTypeHalfFloat, ColumnDataTypeScaledFloat, ColumnDataTypeDate, ColumnDataTypeBoolean,
		 ColumnDataTypeBinary, ColumnDataTypeGeoPoint, ColumnDataTypeGeoShape, ColumnDataTypeIP,
		 ColumnDataTypeCompletion, ColumnDataTypeTokenCount, ColumnDataTypeObject, ColumnDataTypeNested:
		return true
	default:
		return false
	}
}

// String returns the string representation
func (t ColumnDataType) String() string {
	return string(t)
}

// TokenizerType represents the type of tokenizer
type TokenizerType string

const (
	TokenizerTypeStandard    TokenizerType = "standard"
	TokenizerTypeKeyword     TokenizerType = "keyword"
	TokenizerTypeWhitespace  TokenizerType = "whitespace"
	TokenizerTypePattern     TokenizerType = "pattern"
	TokenizerTypeSimplePattern TokenizerType = "simple_pattern"
	TokenizerTypeCharGroup   TokenizerType = "char_group"
	TokenizerTypeClassic     TokenizerType = "classic"
	TokenizerTypeLetter      TokenizerType = "letter"
	TokenizerTypeLowercase   TokenizerType = "lowercase"
	TokenizerTypeNgram       TokenizerType = "ngram"
	TokenizerTypeEdgeNgram   TokenizerType = "edge_ngram"
	TokenizerTypePathHierarchy TokenizerType = "path_hierarchy"
	TokenizerTypeUaxUrlEmail TokenizerType = "uax_url_email"
	TokenizerTypeCJK         TokenizerType = "cjk"
	TokenizerTypeIK          TokenizerType = "ik"
	TokenizerTypeCustom      TokenizerType = "custom"
)

// IsValid checks if the tokenizer type is valid
func (t TokenizerType) IsValid() bool {
	switch t {
	case TokenizerTypeStandard, TokenizerTypeKeyword, TokenizerTypeWhitespace, TokenizerTypePattern,
		 TokenizerTypeSimplePattern, TokenizerTypeCharGroup, TokenizerTypeClassic, TokenizerTypeLetter,
		 TokenizerTypeLowercase, TokenizerTypeNgram, TokenizerTypeEdgeNgram, TokenizerTypePathHierarchy,
		 TokenizerTypeUaxUrlEmail, TokenizerTypeCJK, TokenizerTypeIK, TokenizerTypeCustom:
		return true
	default:
		return false
	}
}

// String returns the string representation
func (t TokenizerType) String() string {
	return string(t)
}

// CompressionType represents the compression type
type CompressionType string

const (
	CompressionTypeDefault CompressionType = "default"
	CompressionTypeLZ4     CompressionType = "lz4"
	CompressionTypeGzip    CompressionType = "gzip"
	CompressionTypeSnappy  CompressionType = "snappy"
	CompressionTypeZstd    CompressionType = "zstd"
)

// IsValid checks if the compression type is valid
func (t CompressionType) IsValid() bool {
	switch t {
	case CompressionTypeDefault, CompressionTypeLZ4, CompressionTypeGzip, CompressionTypeSnappy, CompressionTypeZstd:
		return true
	default:
		return false
	}
}

// String returns the string representation
func (t CompressionType) String() string {
	return string(t)
}

// CharFilter represents a character filter
type CharFilter struct {
	Type     string                 `json:"type"`
	Settings map[string]interface{} `json:"settings"`
}

// TokenFilter represents a token filter
type TokenFilter struct {
	Type     string                 `json:"type"`
	Settings map[string]interface{} `json:"settings"`
}

// IndexRule represents a business rule for index
type IndexRule struct {
	Name        string                 `json:"name"`
	Type        string                 `json:"type"`
	Condition   string                 `json:"condition"`
	Action      string                 `json:"action"`
	Parameters  map[string]interface{} `json:"parameters"`
	Enabled     bool                   `json:"enabled"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
}

// =============================================================================
// Validation Helper Functions
// =============================================================================

// validateColumnName validates the column name
func validateColumnName(name string) error {
	if strings.TrimSpace(name) == "" {
		return errors.NewValidationError("column name cannot be empty")
	}

	if len(name) > 100 {
		return errors.NewValidationError("column name cannot exceed 100 characters")
	}

	// Check for invalid characters
	matched, err := regexp.MatchString("^[a-zA-Z0-9_.-]+$", name)
	if err != nil {
		return errors.NewValidationError("invalid column name format")
	}

	if !matched {
		return errors.NewValidationError("column name can only contain alphanumeric characters, dots, dashes, and underscores")
	}

	return nil
}

// validateTokenizerName validates the tokenizer name
func validateTokenizerName(name string) error {
	if strings.TrimSpace(name) == "" {
		return errors.NewValidationError("tokenizer name cannot be empty")
	}

	if len(name) > 50 {
		return errors.NewValidationError("tokenizer name cannot exceed 50 characters")
	}

	// Check for invalid characters
	matched, err := regexp.MatchString("^[a-zA-Z0-9_-]+$", name)
	if err != nil {
		return errors.NewValidationError("invalid tokenizer name format")
	}

	if !matched {
		return errors.NewValidationError("tokenizer name can only contain alphanumeric characters, dashes, and underscores")
	}

	return nil
}

// isValidLanguage checks if the language code is valid
func isValidLanguage(language string) bool {
	validLanguages := []string{
		"en", "zh", "ja", "ko", "es", "fr", "de", "it", "pt", "ru",
		"ar", "hi", "tr", "nl", "sv", "da", "no", "fi", "pl", "cs",
		"hu", "ro", "bg", "hr", "sk", "sl", "et", "lv", "lt", "mt",
		"ga", "eu", "ca", "gl", "ast", "br", "cy", "gd", "gv", "kw",
	}

	for _, validLang := range validLanguages {
		if language == validLang {
			return true
		}
	}

	return false
}

//Personal.AI order the ending
```

