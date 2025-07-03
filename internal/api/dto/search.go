package dto

import (
	"time"

	"github.com/go-playground/validator/v10"
)

// 定义全局的 validator 实例
var validate *validator.Validate

func init() {
	validate = validator.New()
}

// =============================================================================
// Search DTOs
// =============================================================================

// SearchRequest represents the request payload for a search operation.
type SearchRequest struct {
	// Q is the main query string.
	// Either Q or Filters must be provided.
	Q string `json:"q" validate:"omitempty"` // Custom validation in application layer will handle "Q or Filters" rule

	// Fields is a comma-separated list of fields to search within.
	// If empty, default fields will be searched.
	Fields string `json:"fields,omitempty"`

	// Tables is a comma-separated list of table names to search in.
	// At least one table must be specified.
	Tables string `json:"tables" validate:"required"` // Must not be empty, application layer splits by comma

	// Filters is a map of key-value pairs for filtering results.
	// Example: {"category": "electronics", "price_range": "[100 TO 500]"}
	// Either Q or Filters must be provided.
	Filters map[string]string `json:"filters,omitempty"`

	// Page is the requested page number, 1-indexed.
	Page int `json:"page" validate:"min=1"`

	// PageSize is the number of results per page. Max 100.
	PageSize int `json:"pageSize" validate:"min=1,max=100"`

	// QueryType specifies the type of query (e.g., "MATCH_ANY", "MATCH_ALL").
	QueryType string `json:"queryType" validate:"oneof=MATCH_ANY MATCH_ALL"` // Example values

	// DatabaseType specifies the type of database/search engine to target (e.g., "CLICKHOUSE", "ELASTICSEARCH").
	DatabaseType string `json:"databaseType" validate:"required,oneof=CLICKHOUSE ELASTICSEARCH"` // Example values
}

// SearchResponse represents the response payload for a search operation.
type SearchResponse struct {
	TotalHits    int64          `json:"totalHits"`
	Results      []SearchResult `json:"results"`
	CurrentPage  int            `json:"currentPage"` // Renamed from Page for clarity in response
	PageSize     int            `json:"pageSize"`
	QueryLatency float64        `json:"queryLatencyMs"` // Latency in milliseconds
}

// SearchResult represents a single search result document.
type SearchResult struct {
	RowID           string                 `json:"rowId"`
	Score           float64                `json:"score"`
	Data            map[string]interface{} `json:"data"`            // Original document data (e.g., from DB row)
	HighlightedData map[string]string      `json:"highlightedData"` // Highlighted snippets for display
}

// =============================================================================
// Index Management DTOs
// =============================================================================

// RegisterIndexRequest represents the request payload for registering/updating an index.
type RegisterIndexRequest struct {
	TableName     string `json:"tableName" validate:"required"`
	ColumnName    string `json:"columnName" validate:"required"`
	IndexType     string `json:"indexType" validate:"required,oneof=FULL_TEXT_INDEX KEYWORD_INDEX NUMERIC_INDEX GEOSPATIAL_INDEX"` // Example types
	TokenizerName string `json:"tokenizerName,omitempty"`                                                                          // Required for FULL_TEXT_INDEX
	DatabaseType  string `json:"databaseType" validate:"required,oneof=CLICKHOUSE ELASTICSEARCH"`                                  // e.g., "CLICKHOUSE", "ELASTICSEARCH"
	// Other index-specific parameters could be added, e.g., Analyzer, ShardCount, Replicas, etc.
}

// RegisterIndexResponse represents the response payload after an index registration.
// Often, for successful registration, an empty success response or just a status message is enough.
type RegisterIndexResponse struct {
	Message string `json:"message"`
	Status  string `json:"status"`            // e.g., "success", "failure"
	IndexID string `json:"indexId,omitempty"` // Potentially a unique ID generated for the index
}

// DeleteIndexRequest represents the request payload for deleting an index.
type DeleteIndexRequest struct {
	TableName  string `json:"tableName" validate:"required"`
	ColumnName string `json:"columnName" validate:"required"`
	// Additional fields for more specific deletion (e.g., only delete for a specific IndexType)
	// IndexType string `json:"indexType,omitempty"`
	DatabaseType string `json:"databaseType" validate:"required,oneof=CLICKHOUSE ELASTICSEARCH"`
}

// DeleteIndexResponse represents the response payload after an index deletion.
type DeleteIndexResponse struct {
	Message string `json:"message"`
	Status  string `json:"status"` // e.g., "success", "failure"
}

// Validate method for SearchRequest
func (r *SearchRequest) Validate() error {
	// Custom validation for Q or Filters must be present, and Tables not empty
	if r.Q == "" && (r.Filters == nil || len(r.Filters) == 0) {
		return validator.ValidationErrorsTranslations{
			"Q": "Q or Filters must be provided",
		}
	}
	if r.Tables == "" {
		return validator.ValidationErrorsTranslations{
			"Tables": "at least one table must be specified for search",
		}
	}
	return validate.Struct(r)
}

// Validate method for RegisterIndexRequest
func (r *RegisterIndexRequest) Validate() error {
	err := validate.Struct(r)
	if err != nil {
		return err
	}
	// Additional conditional validation for Full-text index
	if r.IndexType == "FULL_TEXT_INDEX" && r.TokenizerName == "" {
		return validator.ValidationErrorsTranslations{
			"TokenizerName": "tokenizerName is required for FULL_TEXT_INDEX",
		}
	}
	return nil
}

// Validate method for DeleteIndexRequest
func (r *DeleteIndexRequest) Validate() error {
	return validate.Struct(r)
}
