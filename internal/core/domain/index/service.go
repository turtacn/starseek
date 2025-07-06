package index

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/turtacn/starseek/internal/common/errors"
	"github.com/turtacn/starseek/internal/common/logger"
	"github.com/turtacn/starseek/internal/common/types"
)

// =============================================================================
// Index Domain Service Interface
// =============================================================================

// IndexDomainService defines the interface for index domain operations
type IndexDomainService interface {
	// Core index management
	CreateIndex(ctx context.Context, req *CreateIndexRequest) (*Index, error)
	UpdateIndex(ctx context.Context, req *UpdateIndexRequest) (*Index, error)
	DeleteIndex(ctx context.Context, indexID IndexID, actor string) error
	GetIndex(ctx context.Context, indexID IndexID) (*Index, error)
	ListIndexes(ctx context.Context, req *ListIndexesRequest) (*ListIndexesResult, error)

	// Index configuration validation
	ValidateIndexConfig(ctx context.Context, config *IndexConfig) (*ValidationResult, error)
	ValidateIndexColumn(ctx context.Context, column *IndexColumn) (*ValidationResult, error)
	ValidateTokenizerRule(ctx context.Context, rule *TokenizerRule) (*ValidationResult, error)

	// Index metadata collection
	CollectDatabaseMetadata(ctx context.Context, req *MetadataCollectionRequest) (*DatabaseMetadata, error)
	DiscoverTableStructure(ctx context.Context, req *TableDiscoveryRequest) (*TableStructure, error)
	CollectExistingIndexes(ctx context.Context, req *ExistingIndexesRequest) (*ExistingIndexes, error)

	// Index status monitoring
	MonitorIndexHealth(ctx context.Context, indexID IndexID) (*IndexHealthReport, error)
	CheckAllIndexesHealth(ctx context.Context) (*AllIndexesHealthReport, error)
	GetIndexStatistics(ctx context.Context, indexID IndexID) (*IndexStatistics, error)
	UpdateIndexStatistics(ctx context.Context, indexID IndexID, stats *IndexStatistics) error

	// Index optimization recommendations
	AnalyzeQueryPatterns(ctx context.Context, req *QueryPatternAnalysisRequest) (*QueryPatternAnalysis, error)
	GenerateIndexRecommendations(ctx context.Context, req *IndexRecommendationRequest) (*IndexRecommendations, error)
	EvaluateIndexPerformance(ctx context.Context, req *IndexPerformanceRequest) (*IndexPerformanceReport, error)

	// Index change notifications
	NotifyIndexChange(ctx context.Context, event *IndexChangeEvent) error
	RegisterChangeListener(ctx context.Context, listener IndexChangeListener) error
	UnregisterChangeListener(ctx context.Context, listenerID string) error

	// Index lifecycle management
	EnableIndex(ctx context.Context, indexID IndexID, actor string) error
	DisableIndex(ctx context.Context, indexID IndexID, actor string) error
	RebuildIndex(ctx context.Context, indexID IndexID, actor string) error

	// Batch operations
	CreateMultipleIndexes(ctx context.Context, req *CreateMultipleIndexesRequest) (*CreateMultipleIndexesResult, error)
	UpdateMultipleIndexes(ctx context.Context, req *UpdateMultipleIndexesRequest) (*UpdateMultipleIndexesResult, error)
	DeleteMultipleIndexes(ctx context.Context, req *DeleteMultipleIndexesRequest) (*DeleteMultipleIndexesResult, error)
}

// =============================================================================
// Index Domain Service Implementation
// =============================================================================

// indexDomainService implements the IndexDomainService interface
type indexDomainService struct {
	repository        IndexRepository
	logger            logger.Logger
	changeListeners   map[string]IndexChangeListener
	monitoringEnabled bool
}

// NewIndexDomainService creates a new index domain service
func NewIndexDomainService(repository IndexRepository, logger logger.Logger) IndexDomainService {
	return &indexDomainService{
		repository:        repository,
		logger:            logger,
		changeListeners:   make(map[string]IndexChangeListener),
		monitoringEnabled: true,
	}
}

// =============================================================================
// Core Index Management Implementation
// =============================================================================

// CreateIndex creates a new index
func (s *indexDomainService) CreateIndex(ctx context.Context, req *CreateIndexRequest) (*Index, error) {
	// Validate request
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("invalid create index request: %w", err)
	}

	// Check if index already exists
	existingIndex, err := s.repository.GetByName(ctx, req.Name)
	if err != nil && !errors.IsNotFound(err) {
		return nil, fmt.Errorf("failed to check existing index: %w", err)
	}

	if existingIndex != nil {
		return nil, errors.NewBusinessError("index already exists with name: %s", req.Name)
	}

	// Generate index ID
	indexID := s.generateIndexID(req.Name)

	// Create index entity
	index, err := NewIndex(indexID, req.Name, req.Config, req.CreatedBy)
	if err != nil {
		return nil, fmt.Errorf("failed to create index entity: %w", err)
	}

	// Set optional fields
	if req.DisplayName != "" {
		if err := index.UpdateDisplayName(req.DisplayName, req.CreatedBy); err != nil {
			return nil, fmt.Errorf("failed to set display name: %w", err)
		}
	}

	if req.Description != "" {
		if err := index.UpdateDescription(req.Description, req.CreatedBy); err != nil {
			return nil, fmt.Errorf("failed to set description: %w", err)
		}
	}

	// Add columns
	for _, columnReq := range req.Columns {
		column, err := NewIndexColumn(columnReq.Name, columnReq.DataType)
		if err != nil {
			return nil, fmt.Errorf("failed to create column %s: %w", columnReq.Name, err)
		}

		// Configure column
		if err := s.configureColumn(column, columnReq); err != nil {
			return nil, fmt.Errorf("failed to configure column %s: %w", columnReq.Name, err)
		}

		// Add column to index
		if err := index.AddColumn(column, req.CreatedBy); err != nil {
			return nil, fmt.Errorf("failed to add column %s: %w", columnReq.Name, err)
		}
	}

	// Validate the complete index
	if err := index.Validate(); err != nil {
		return nil, fmt.Errorf("index validation failed: %w", err)
	}

	// Save to repository
	if err := s.repository.Create(ctx, index); err != nil {
		return nil, fmt.Errorf("failed to save index: %w", err)
	}

	// Notify change listeners
	s.notifyChangeListeners(ctx, &IndexChangeEvent{
		Type:      IndexChangeTypeCreated,
		IndexID:   indexID,
		Actor:     req.CreatedBy,
		Timestamp: time.Now(),
		Data:      map[string]interface{}{"index_name": req.Name},
	})

	s.logger.Info(ctx, "Index created successfully", map[string]interface{}{
		"index_id": indexID,
		"name":     req.Name,
		"actor":    req.CreatedBy,
	})

	return index, nil
}

// UpdateIndex updates an existing index
func (s *indexDomainService) UpdateIndex(ctx context.Context, req *UpdateIndexRequest) (*Index, error) {
	// Validate request
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("invalid update index request: %w", err)
	}

	// Get existing index
	index, err := s.repository.GetByID(ctx, req.IndexID)
	if err != nil {
		return nil, fmt.Errorf("failed to get index: %w", err)
	}

	// Check if index can be updated
	if !index.CanUpdate() {
		return nil, errors.NewBusinessError("index cannot be updated in current status: %s", index.Status())
	}

	// Update fields
	if req.DisplayName != nil {
		if err := index.UpdateDisplayName(*req.DisplayName, req.UpdatedBy); err != nil {
			return nil, fmt.Errorf("failed to update display name: %w", err)
		}
	}

	if req.Description != nil {
		if err := index.UpdateDescription(*req.Description, req.UpdatedBy); err != nil {
			return nil, fmt.Errorf("failed to update description: %w", err)
		}
	}

	if req.Config != nil {
		if err := index.UpdateConfig(req.Config, req.UpdatedBy); err != nil {
			return nil, fmt.Errorf("failed to update config: %w", err)
		}
	}

	// Handle column updates
	if err := s.processColumnUpdates(ctx, index, req.ColumnUpdates, req.UpdatedBy); err != nil {
		return nil, fmt.Errorf("failed to process column updates: %w", err)
	}

	// Validate the updated index
	if err := index.Validate(); err != nil {
		return nil, fmt.Errorf("index validation failed: %w", err)
	}

	// Save to repository
	if err := s.repository.Update(ctx, index); err != nil {
		return nil, fmt.Errorf("failed to update index: %w", err)
	}

	// Notify change listeners
	s.notifyChangeListeners(ctx, &IndexChangeEvent{
		Type:      IndexChangeTypeUpdated,
		IndexID:   req.IndexID,
		Actor:     req.UpdatedBy,
		Timestamp: time.Now(),
		Data:      map[string]interface{}{"changes": req.GetChanges()},
	})

	s.logger.Info(ctx, "Index updated successfully", map[string]interface{}{
		"index_id": req.IndexID,
		"actor":    req.UpdatedBy,
	})

	return index, nil
}

// DeleteIndex deletes an index
func (s *indexDomainService) DeleteIndex(ctx context.Context, indexID IndexID, actor string) error {
	// Get existing index
	index, err := s.repository.GetByID(ctx, indexID)
	if err != nil {
		return fmt.Errorf("failed to get index: %w", err)
	}

	// Check if index can be deleted
	if !index.CanDelete() {
		return errors.NewBusinessError("index cannot be deleted in current status: %s", index.Status())
	}

	// Mark as deleted
	if err := index.MarkAsDeleted(actor); err != nil {
		return fmt.Errorf("failed to mark index as deleted: %w", err)
	}

	// Update in repository
	if err := s.repository.Update(ctx, index); err != nil {
		return fmt.Errorf("failed to update index: %w", err)
	}

	// Notify change listeners
	s.notifyChangeListeners(ctx, &IndexChangeEvent{
		Type:      IndexChangeTypeDeleted,
		IndexID:   indexID,
		Actor:     actor,
		Timestamp: time.Now(),
		Data:      map[string]interface{}{"index_name": index.Name()},
	})

	s.logger.Info(ctx, "Index deleted successfully", map[string]interface{}{
		"index_id": indexID,
		"actor":    actor,
	})

	return nil
}

// GetIndex retrieves an index by ID
func (s *indexDomainService) GetIndex(ctx context.Context, indexID IndexID) (*Index, error) {
	index, err := s.repository.GetByID(ctx, indexID)
	if err != nil {
		return nil, fmt.Errorf("failed to get index: %w", err)
	}

	return index, nil
}

// ListIndexes lists indexes based on criteria
func (s *indexDomainService) ListIndexes(ctx context.Context, req *ListIndexesRequest) (*ListIndexesResult, error) {
	// Validate request
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("invalid list indexes request: %w", err)
	}

	// Get indexes from repository
	indexes, total, err := s.repository.List(ctx, &IndexListOptions{
		Offset:    req.Offset,
		Limit:     req.Limit,
		Status:    req.Status,
		NameLike:  req.NameLike,
		SortBy:    req.SortBy,
		SortOrder: req.SortOrder,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list indexes: %w", err)
	}

	return &ListIndexesResult{
		Indexes: indexes,
		Total:   total,
		Offset:  req.Offset,
		Limit:   req.Limit,
	}, nil
}

// =============================================================================
// Index Configuration Validation Implementation
// =============================================================================

// ValidateIndexConfig validates index configuration
func (s *indexDomainService) ValidateIndexConfig(ctx context.Context, config *IndexConfig) (*ValidationResult, error) {
	result := &ValidationResult{
		IsValid:  true,
		Errors:   make([]string, 0),
		Warnings: make([]string, 0),
	}

	// Basic validation
	if err := config.Validate(); err != nil {
		result.IsValid = false
		result.Errors = append(result.Errors, err.Error())
		return result, nil
	}

	// Advanced validation rules
	s.validateShardConfiguration(config, result)
	s.validateAnalyzerConfiguration(config, result)
	s.validatePerformanceSettings(config, result)
	s.validateStorageSettings(config, result)

	return result, nil
}

// ValidateIndexColumn validates index column configuration
func (s *indexDomainService) ValidateIndexColumn(ctx context.Context, column *IndexColumn) (*ValidationResult, error) {
	result := &ValidationResult{
		IsValid:  true,
		Errors:   make([]string, 0),
		Warnings: make([]string, 0),
	}

	// Basic validation
	if err := column.Validate(); err != nil {
		result.IsValid = false
		result.Errors = append(result.Errors, err.Error())
		return result, nil
	}

	// Advanced validation rules
	s.validateColumnDataType(column, result)
	s.validateColumnFeatures(column, result)
	s.validateTokenizerConfiguration(column, result)

	return result, nil
}

// ValidateTokenizerRule validates tokenizer rule configuration
func (s *indexDomainService) ValidateTokenizerRule(ctx context.Context, rule *TokenizerRule) (*ValidationResult, error) {
	result := &ValidationResult{
		IsValid:  true,
		Errors:   make([]string, 0),
		Warnings: make([]string, 0),
	}

	// Basic validation
	if err := rule.Validate(); err != nil {
		result.IsValid = false
		result.Errors = append(result.Errors, err.Error())
		return result, nil
	}

	// Advanced validation rules
	s.validateTokenizerType(rule, result)
	s.validateCustomDictionary(rule, result)
	s.validateStopWords(rule, result)
	s.validateSynonyms(rule, result)

	return result, nil
}

// =============================================================================
// Index Metadata Collection Implementation
// =============================================================================

// CollectDatabaseMetadata collects metadata from database
func (s *indexDomainService) CollectDatabaseMetadata(ctx context.Context, req *MetadataCollectionRequest) (*DatabaseMetadata, error) {
	// Validate request
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("invalid metadata collection request: %w", err)
	}

	metadata := &DatabaseMetadata{
		DatabaseName: req.DatabaseName,
		CollectedAt:  time.Now(),
		Tables:       make([]*TableMetadata, 0),
		Indexes:      make([]*IndexMetadata, 0),
	}

	// Collect table metadata
	if req.IncludeTables {
		tables, err := s.repository.GetTableMetadata(ctx, req.DatabaseName)
		if err != nil {
			return nil, fmt.Errorf("failed to get table metadata: %w", err)
		}
		metadata.Tables = tables
	}

	// Collect index metadata
	if req.IncludeIndexes {
		indexes, err := s.repository.GetIndexMetadata(ctx, req.DatabaseName)
		if err != nil {
			return nil, fmt.Errorf("failed to get index metadata: %w", err)
		}
		metadata.Indexes = indexes
	}

	// Collect schema metadata
	if req.IncludeSchemas {
		schemas, err := s.repository.GetSchemaMetadata(ctx, req.DatabaseName)
		if err != nil {
			return nil, fmt.Errorf("failed to get schema metadata: %w", err)
		}
		metadata.Schemas = schemas
	}

	// Collect statistics
	if req.IncludeStatistics {
		stats, err := s.repository.GetDatabaseStatistics(ctx, req.DatabaseName)
		if err != nil {
			return nil, fmt.Errorf("failed to get database statistics: %w", err)
		}
		metadata.Statistics = stats
	}

	s.logger.Info(ctx, "Database metadata collected successfully", map[string]interface{}{
		"database": req.DatabaseName,
		"tables":   len(metadata.Tables),
		"indexes":  len(metadata.Indexes),
		"schemas":  len(metadata.Schemas),
	})

	return metadata, nil
}

// DiscoverTableStructure discovers table structure and columns
func (s *indexDomainService) DiscoverTableStructure(ctx context.Context, req *TableDiscoveryRequest) (*TableStructure, error) {
	// Validate request
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("invalid table discovery request: %w", err)
	}

	// Get table structure from repository
	structure, err := s.repository.DiscoverTableStructure(ctx, req.DatabaseName, req.TableName)
	if err != nil {
		return nil, fmt.Errorf("failed to discover table structure: %w", err)
	}

	// Analyze column types and suggest index configurations
	s.analyzeColumnTypes(structure)
	s.suggestIndexConfigurations(structure)

	s.logger.Info(ctx, "Table structure discovered successfully", map[string]interface{}{
		"database": req.DatabaseName,
		"table":    req.TableName,
		"columns":  len(structure.Columns),
	})

	return structure, nil
}

// CollectExistingIndexes collects existing indexes from database
func (s *indexDomainService) CollectExistingIndexes(ctx context.Context, req *ExistingIndexesRequest) (*ExistingIndexes, error) {
	// Validate request
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("invalid existing indexes request: %w", err)
	}

	// Get existing indexes from repository
	indexes, err := s.repository.GetExistingIndexes(ctx, req.DatabaseName, req.TableName)
	if err != nil {
		return nil, fmt.Errorf("failed to get existing indexes: %w", err)
	}

	// Analyze index usage and performance
	for _, index := range indexes.Indexes {
		s.analyzeIndexUsage(ctx, index)
		s.analyzeIndexPerformance(ctx, index)
	}

	s.logger.Info(ctx, "Existing indexes collected successfully", map[string]interface{}{
		"database": req.DatabaseName,
		"table":    req.TableName,
		"indexes":  len(indexes.Indexes),
	})

	return indexes, nil
}

// =============================================================================
// Index Status Monitoring Implementation
// =============================================================================

// MonitorIndexHealth monitors the health of a specific index
func (s *indexDomainService) MonitorIndexHealth(ctx context.Context, indexID IndexID) (*IndexHealthReport, error) {
	// Get index from repository
	index, err := s.repository.GetByID(ctx, indexID)
	if err != nil {
		return nil, fmt.Errorf("failed to get index: %w", err)
	}

	// Create health report
	report := &IndexHealthReport{
		IndexID:         indexID,
		IndexName:       index.Name(),
		CheckedAt:       time.Now(),
		HealthStatus:    IndexHealthUnknown,
		Metrics:         make(map[string]interface{}),
		Issues:          make([]*HealthIssue, 0),
		Recommendations: make([]*HealthRecommendation, 0),
	}

	// Check index status
	if !index.IsActive() {
		report.Issues = append(report.Issues, &HealthIssue{
			Type:        "status",
			Severity:    "warning",
			Description: fmt.Sprintf("Index is not active, current status: %s", index.Status()),
		})
	}

	// Check index statistics
	if err := s.checkIndexStatistics(ctx, index, report); err != nil {
		s.logger.Warn(ctx, "Failed to check index statistics", map[string]interface{}{
			"index_id": indexID,
			"error":    err.Error(),
		})
	}

	// Check index performance
	if err := s.checkIndexPerformance(ctx, index, report); err != nil {
		s.logger.Warn(ctx, "Failed to check index performance", map[string]interface{}{
			"index_id": indexID,
			"error":    err.Error(),
		})
	}

	// Check index configuration
	if err := s.checkIndexConfiguration(ctx, index, report); err != nil {
		s.logger.Warn(ctx, "Failed to check index configuration", map[string]interface{}{
			"index_id": indexID,
			"error":    err.Error(),
		})
	}

	// Determine overall health status
	report.HealthStatus = s.determineHealthStatus(report)

	// Update index health if changed
	if index.Health() != report.HealthStatus {
		if err := index.UpdateHealth(report.HealthStatus, "system"); err != nil {
			s.logger.Error(ctx, "Failed to update index health", map[string]interface{}{
				"index_id": indexID,
				"error":    err.Error(),
			})
		} else {
			if err := s.repository.Update(ctx, index); err != nil {
				s.logger.Error(ctx, "Failed to save index health update", map[string]interface{}{
					"index_id": indexID,
					"error":    err.Error(),
				})
			}
		}
	}

	return report, nil
}

// CheckAllIndexesHealth checks the health of all indexes
func (s *indexDomainService) CheckAllIndexesHealth(ctx context.Context) (*AllIndexesHealthReport, error) {
	// Get all active indexes
	indexes, _, err := s.repository.List(ctx, &IndexListOptions{
		Status: &IndexStatusEnabled,
		Limit:  1000, // Reasonable limit for health checks
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get indexes: %w", err)
	}

	// Create overall report
	report := &AllIndexesHealthReport{
		CheckedAt:    time.Now(),
		TotalIndexes: len(indexes),
		HealthySummary: &HealthSummary{
			Green:   0,
			Yellow:  0,
			Red:     0,
			Unknown: 0,
		},
		IndexReports: make([]*IndexHealthReport, 0),
	}

	// Check each index
	for _, index := range indexes {
		indexReport, err := s.MonitorIndexHealth(ctx, index.ID())
		if err != nil {
			s.logger.Error(ctx, "Failed to check index health", map[string]interface{}{
				"index_id": index.ID(),
				"error":    err.Error(),
			})
			continue
		}

		report.IndexReports = append(report.IndexReports, indexReport)

		// Update summary
		switch indexReport.HealthStatus {
		case IndexHealthGreen:
			report.HealthySummary.Green++
		case IndexHealthYellow:
			report.HealthySummary.Yellow++
		case IndexHealthRed:
			report.HealthySummary.Red++
		default:
			report.HealthySummary.Unknown++
		}
	}

	s.logger.Info(ctx, "All indexes health check completed", map[string]interface{}{
		"total":   report.TotalIndexes,
		"green":   report.HealthySummary.Green,
		"yellow":  report.HealthySummary.Yellow,
		"red":     report.HealthySummary.Red,
		"unknown": report.HealthySummary.Unknown,
	})

	return report, nil
}

// GetIndexStatistics retrieves index statistics
func (s *indexDomainService) GetIndexStatistics(ctx context.Context, indexID IndexID) (*IndexStatistics, error) {
	// Get index from repository
	index, err := s.repository.GetByID(ctx, indexID)
	if err != nil {
		return nil, fmt.Errorf("failed to get index: %w", err)
	}

	// Get fresh statistics from repository
	stats, err := s.repository.GetIndexStatistics(ctx, indexID)
	if err != nil {
		return nil, fmt.Errorf("failed to get index statistics: %w", err)
	}

	return stats, nil
}

// UpdateIndexStatistics updates index statistics
func (s *indexDomainService) UpdateIndexStatistics(ctx context.Context, indexID IndexID, stats *IndexStatistics) error {
	// Get index from repository
	index, err := s.repository.GetByID(ctx, indexID)
	if err != nil {
		return fmt.Errorf("failed to get index: %w", err)
	}

	// Update statistics
	if err := index.UpdateStatistics(stats, "system"); err != nil {
		return fmt.Errorf("failed to update index statistics: %w", err)
	}

	// Save to repository
	if err := s.repository.Update(ctx, index); err != nil {
		return fmt.Errorf("failed to save index: %w", err)
	}

	return nil
}

// =============================================================================
// Index Optimization Recommendations Implementation
// =============================================================================

// AnalyzeQueryPatterns analyzes query patterns for optimization
func (s *indexDomainService) AnalyzeQueryPatterns(ctx context.Context, req *QueryPatternAnalysisRequest) (*QueryPatternAnalysis, error) {
	// Validate request
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("invalid query pattern analysis request: %w", err)
	}

	// Get query patterns from repository
	patterns, err := s.repository.GetQueryPatterns(ctx, req.IndexID, req.TimeRange)
	if err != nil {
		return nil, fmt.Errorf("failed to get query patterns: %w", err)
	}

	// Analyze patterns
	analysis := &QueryPatternAnalysis{
		IndexID:         req.IndexID,
		AnalyzedAt:      time.Now(),
		TimeRange:       req.TimeRange,
		TotalQueries:    len(patterns),
		PatternGroups:   make([]*PatternGroup, 0),
		Insights:        make([]*QueryInsight, 0),
		Recommendations: make([]*QueryRecommendation, 0),
	}

	// Group similar patterns
	s.groupQueryPatterns(patterns, analysis)

	// Generate insights
	s.generateQueryInsights(analysis)

	// Generate recommendations
	s.generateQueryRecommendations(analysis)

	s.logger.Info(ctx, "Query pattern analysis completed", map[string]interface{}{
		"index_id":        req.IndexID,
		"total_queries":   analysis.TotalQueries,
		"pattern_groups":  len(analysis.PatternGroups),
		"insights":        len(analysis.Insights),
		"recommendations": len(analysis.Recommendations),
	})

	return analysis, nil
}

// GenerateIndexRecommendations generates index optimization recommendations
func (s *indexDomainService) GenerateIndexRecommendations(ctx context.Context, req *IndexRecommendationRequest) (*IndexRecommendations, error) {
	// Validate request
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("invalid index recommendation request: %w", err)
	}

	// Get index and related data
	index, err := s.repository.GetByID(ctx, req.IndexID)
	if err != nil {
		return nil, fmt.Errorf("failed to get index: %w", err)
	}

	// Create recommendations
	recommendations := &IndexRecommendations{
		IndexID:         req.IndexID,
		IndexName:       index.Name(),
		GeneratedAt:     time.Now(),
		Recommendations: make([]*IndexRecommendation, 0),
	}

	// Analyze configuration
	s.analyzeIndexConfiguration(ctx, index, recommendations)

	// Analyze performance
	s.analyzeIndexPerformanceForRecommendations(ctx, index, recommendations)

	// Analyze usage patterns
	s.analyzeIndexUsagePatterns(ctx, index, recommendations)

	// Sort recommendations by priority
	sort.Slice(recommendations.Recommendations, func(i, j int) bool {
		return recommendations.Recommendations[i].Priority > recommendations.Recommendations[j].Priority
	})

	s.logger.Info(ctx, "Index recommendations generated", map[string]interface{}{
		"index_id":        req.IndexID,
		"recommendations": len(recommendations.Recommendations),
	})

	return recommendations, nil
}

// EvaluateIndexPerformance evaluates index performance
func (s *indexDomainService) EvaluateIndexPerformance(ctx context.Context, req *IndexPerformanceRequest) (*IndexPerformanceReport, error) {
	// Validate request
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("invalid index performance request: %w", err)
	}

	// Get index from repository
	index, err := s.repository.GetByID(ctx, req.IndexID)
	if err != nil {
		return nil, fmt.Errorf("failed to get index: %w", err)
	}

	// Get performance metrics
	metrics, err := s.repository.GetIndexPerformanceMetrics(ctx, req.IndexID, req.TimeRange)
	if err != nil {
		return nil, fmt.Errorf("failed to get performance metrics: %w", err)
	}

	// Create performance report
	report := &IndexPerformanceReport{
		IndexID:         req.IndexID,
		IndexName:       index.Name(),
		EvaluatedAt:     time.Now(),
		TimeRange:       req.TimeRange,
		Metrics:         metrics,
		Performance:     &PerformanceAnalysis{},
		Issues:          make([]*PerformanceIssue, 0),
		Recommendations: make([]*PerformanceRecommendation, 0),
	}

	// Analyze performance
	s.analyzePerformanceMetrics(metrics, report)

	// Identify issues
	s.identifyPerformanceIssues(report)

	// Generate recommendations
	s.generatePerformanceRecommendations(report)

	return report, nil
}

// =============================================================================
// Index Change Notifications Implementation
// =============================================================================

// NotifyIndexChange notifies about index changes
func (s *indexDomainService) NotifyIndexChange(ctx context.Context, event *IndexChangeEvent) error {
	// Validate event
	if err := event.Validate(); err != nil {
		return fmt.Errorf("invalid index change event: %w", err)
	}

	// Notify all registered listeners
	s.notifyChangeListeners(ctx, event)

	s.logger.Info(ctx, "Index change notification sent", map[string]interface{}{
		"event_type": event.Type,
		"index_id":   event.IndexID,
		"actor":      event.Actor,
		"listeners":  len(s.changeListeners),
	})

	return nil
}

// RegisterChangeListener registers a change listener
func (s *indexDomainService) RegisterChangeListener(ctx context.Context, listener IndexChangeListener) error {
	listenerID := listener.ID()

	if _, exists := s.changeListeners[listenerID]; exists {
		return errors.NewBusinessError("listener already registered: %s", listenerID)
	}

	s.changeListeners[listenerID] = listener

	s.logger.Info(ctx, "Index change listener registered", map[string]interface{}{
		"listener_id": listenerID,
	})

	return nil
}

// UnregisterChangeListener unregisters a change listener
func (s *indexDomainService) UnregisterChangeListener(ctx context.Context, listenerID string) error {
	if _, exists := s.changeListeners[listenerID]; !exists {
		return errors.NewBusinessError("listener not found: %s", listenerID)
	}

	delete(s.changeListeners, listenerID)

	s.logger.Info(ctx, "Index change listener unregistered", map[string]interface{}{
		"listener_id": listenerID,
	})

	return nil
}

// =============================================================================
// Index Lifecycle Management Implementation
// =============================================================================

// EnableIndex enables an index
func (s *indexDomainService) EnableIndex(ctx context.Context, indexID IndexID, actor string) error {
	// Get index from repository
	index, err := s.repository.GetByID(ctx, indexID)
	if err != nil {
		return fmt.Errorf("failed to get index: %w", err)
	}

	// Enable the index
	if err := index.Enable(actor); err != nil {
		return fmt.Errorf("failed to enable index: %w", err)
	}

	// Update in repository
	if err := s.repository.Update(ctx, index); err != nil {
		return fmt.Errorf("failed to update index: %w", err)
	}

	// Notify change listeners
	s.notifyChangeListeners(ctx, &IndexChangeEvent{
		Type:      IndexChangeTypeEnabled,
		IndexID:   indexID,
		Actor:     actor,
		Timestamp: time.Now(),
		Data:      map[string]interface{}{"index_name": index.Name()},
	})

	s.logger.Info(ctx, "Index enabled successfully", map[string]interface{}{
		"index_id": indexID,
		"actor":    actor,
	})

	return nil
}

// DisableIndex disables an index
func (s *indexDomainService) DisableIndex(ctx context.Context, indexID IndexID, actor string) error {
	// Get index from repository
	index, err := s.repository.GetByID(ctx, indexID)
	if err != nil {
		return fmt.Errorf("failed to get index: %w", err)
	}

	// Disable the index
	if err := index.Disable(actor); err != nil {
		return fmt.Errorf("failed to disable index: %w", err)
	}

	// Update in repository
	if err := s.repository.Update(ctx, index); err != nil {
		return fmt.Errorf("failed to update index: %w", err)
	}

	// Notify change listeners
	s.notifyChangeListeners(ctx, &IndexChangeEvent{
		Type:      IndexChangeTypeDisabled,
		IndexID:   indexID,
		Actor:     actor,
		Timestamp: time.Now(),
		Data:      map[string]interface{}{"index_name": index.Name()},
	})

	s.logger.Info(ctx, "Index disabled successfully", map[string]interface{}{
		"index_id": indexID,
		"actor":    actor,
	})

	return nil
}

// RebuildIndex rebuilds an index
func (s *indexDomainService) RebuildIndex(ctx context.Context, indexID IndexID, actor string) error {
	// Get index from repository
	index, err := s.repository.GetByID(ctx, indexID)
	if err != nil {
		return fmt.Errorf("failed to get index: %w", err)
	}

	// Check if index can be rebuilt
	if !index.CanRebuild() {
		return errors.NewBusinessError("index cannot be rebuilt in current status: %s", index.Status())
	}

	// Start rebuilding
	if err := index.StartRebuilding(actor); err != nil {
		return fmt.Errorf("failed to start rebuilding: %w", err)
	}

	// Update in repository
	if err := s.repository.Update(ctx, index); err != nil {
		return fmt.Errorf("failed to update index: %w", err)
	}

	// Notify change listeners
	s.notifyChangeListeners(ctx, &IndexChangeEvent{
		Type:      IndexChangeTypeRebuildStarted,
		IndexID:   indexID,
		Actor:     actor,
		Timestamp: time.Now(),
		Data:      map[string]interface{}{"index_name": index.Name()},
	})

	s.logger.Info(ctx, "Index rebuild started successfully", map[string]interface{}{
		"index_id": indexID,
		"actor":    actor,
	})

	return nil
}

//Personal.AI order the ending
