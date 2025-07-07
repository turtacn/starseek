package constants

import (
	"encoding/json"
	"fmt"
	"time"
)

// Error codes organized by layers
const (
	// Parameter validation errors (1000-1999)
	ErrCodeInvalidParameter    = 1001
	ErrCodeMissingParameter    = 1002
	ErrCodeParameterOutOfRange = 1003
	ErrCodeInvalidQuerySyntax  = 1004
	ErrCodeInvalidPageSize     = 1005
	ErrCodeInvalidSortField    = 1006
	ErrCodeInvalidTimeRange    = 1007
	ErrCodeInvalidTokenizer    = 1008
	ErrCodeInvalidDatabaseType = 1009
	ErrCodeInvalidIndexType    = 1010

	// Business logic errors (2000-2999)
	ErrCodeIndexNotFound        = 2001
	ErrCodeIndexAlreadyExists   = 2002
	ErrCodeTableNotFound        = 2003
	ErrCodeColumnNotFound       = 2004
	ErrCodeQueryTimeout         = 2005
	ErrCodeSearchResultEmpty    = 2006
	ErrCodeIndexConfigInvalid   = 2007
	ErrCodeTokenizationFailed   = 2008
	ErrCodeRankingFailed        = 2009
	ErrCodeCacheOperationFailed = 2010
	ErrCodePermissionDenied     = 2011
	ErrCodeQuotaExceeded        = 2012

	// Infrastructure errors (3000-3999)
	ErrCodeDatabaseConnection   = 3001
	ErrCodeDatabaseQuery        = 3002
	ErrCodeRedisConnection      = 3003
	ErrCodeRedisOperation       = 3004
	ErrCodeConfigurationError   = 3005
	ErrCodeNetworkError         = 3006
	ErrCodeServiceUnavailable   = 3007
	ErrCodeInternalServerError  = 3008
	ErrCodeExternalServiceError = 3009
	ErrCodeResourceExhausted    = 3010

	// Authentication and authorization errors (4000-4999)
	ErrCodeUnauthorized           = 4001
	ErrCodeInvalidToken           = 4002
	ErrCodeTokenExpired           = 4003
	ErrCodeInvalidAPIKey          = 4004
	ErrCodeInsufficientPermission = 4005
	ErrCodeRateLimitExceeded      = 4006
)

// StarSeekError represents a custom error with code, message and details
type StarSeekError struct {
	Code      int                    `json:"code"`
	Message   string                 `json:"message"`
	Details   map[string]interface{} `json:"details,omitempty"`
	Timestamp time.Time              `json:"timestamp"`
	TraceID   string                 `json:"trace_id,omitempty"`
}

// Error implements the error interface
func (e *StarSeekError) Error() string {
	return fmt.Sprintf("StarSeek Error [%d]: %s", e.Code, e.Message)
}

// JSON returns the JSON representation of the error
func (e *StarSeekError) JSON() string {
	jsonBytes, _ := json.Marshal(e)
	return string(jsonBytes)
}

// WithDetail adds additional detail to the error
func (e *StarSeekError) WithDetail(key string, value interface{}) *StarSeekError {
	if e.Details == nil {
		e.Details = make(map[string]interface{})
	}
	e.Details[key] = value
	return e
}

// WithTraceID adds trace ID to the error
func (e *StarSeekError) WithTraceID(traceID string) *StarSeekError {
	e.TraceID = traceID
	return e
}

// NewError creates a new StarSeekError
func NewError(code int, message string) *StarSeekError {
	return &StarSeekError{
		Code:      code,
		Message:   message,
		Timestamp: time.Now(),
		Details:   make(map[string]interface{}),
	}
}

// Error constructors for common errors

// NewInvalidParameterError creates an invalid parameter error
func NewInvalidParameterError(parameter string, value interface{}) *StarSeekError {
	return NewError(ErrCodeInvalidParameter, fmt.Sprintf("Invalid parameter: %s", parameter)).
		WithDetail("parameter", parameter).
		WithDetail("value", value)
}

// NewMissingParameterError creates a missing parameter error
func NewMissingParameterError(parameter string) *StarSeekError {
	return NewError(ErrCodeMissingParameter, fmt.Sprintf("Missing required parameter: %s", parameter)).
		WithDetail("parameter", parameter)
}

// NewIndexNotFoundError creates an index not found error
func NewIndexNotFoundError(indexName string) *StarSeekError {
	return NewError(ErrCodeIndexNotFound, fmt.Sprintf("Index not found: %s", indexName)).
		WithDetail("index_name", indexName)
}

// NewTableNotFoundError creates a table not found error
func NewTableNotFoundError(tableName string) *StarSeekError {
	return NewError(ErrCodeTableNotFound, fmt.Sprintf("Table not found: %s", tableName)).
		WithDetail("table_name", tableName)
}

// NewDatabaseConnectionError creates a database connection error
func NewDatabaseConnectionError(database string, err error) *StarSeekError {
	return NewError(ErrCodeDatabaseConnection, fmt.Sprintf("Failed to connect to database: %s", database)).
		WithDetail("database", database).
		WithDetail("original_error", err.Error())
}

// NewQueryTimeoutError creates a query timeout error
func NewQueryTimeoutError(timeout time.Duration) *StarSeekError {
	return NewError(ErrCodeQueryTimeout, fmt.Sprintf("Query timeout after %v", timeout)).
		WithDetail("timeout", timeout.String())
}

// NewPermissionDeniedError creates a permission denied error
func NewPermissionDeniedError(resource string, action string) *StarSeekError {
	return NewError(ErrCodePermissionDenied, fmt.Sprintf("Permission denied for %s on %s", action, resource)).
		WithDetail("resource", resource).
		WithDetail("action", action)
}

// NewRateLimitExceededError creates a rate limit exceeded error
func NewRateLimitExceededError(limit int, window time.Duration) *StarSeekError {
	return NewError(ErrCodeRateLimitExceeded, fmt.Sprintf("Rate limit exceeded: %d requests per %v", limit, window)).
		WithDetail("limit", limit).
		WithDetail("window", window.String())
}

// Error categorization functions

// IsParameterError checks if the error is a parameter validation error
func IsParameterError(err error) bool {
	if starseekErr, ok := err.(*StarSeekError); ok {
		return starseekErr.Code >= 1000 && starseekErr.Code < 2000
	}
	return false
}

// IsBusinessError checks if the error is a business logic error
func IsBusinessError(err error) bool {
	if starseekErr, ok := err.(*StarSeekError); ok {
		return starseekErr.Code >= 2000 && starseekErr.Code < 3000
	}
	return false
}

// IsInfrastructureError checks if the error is an infrastructure error
func IsInfrastructureError(err error) bool {
	if starseekErr, ok := err.(*StarSeekError); ok {
		return starseekErr.Code >= 3000 && starseekErr.Code < 4000
	}
	return false
}

// IsAuthenticationError checks if the error is an authentication error
func IsAuthenticationError(err error) bool {
	if starseekErr, ok := err.(*StarSeekError); ok {
		return starseekErr.Code >= 4000 && starseekErr.Code < 5000
	}
	return false
}

// IsRetryableError checks if the error is retryable
func IsRetryableError(err error) bool {
	if starseekErr, ok := err.(*StarSeekError); ok {
		switch starseekErr.Code {
		case ErrCodeQueryTimeout, ErrCodeDatabaseConnection, ErrCodeNetworkError,
			ErrCodeServiceUnavailable, ErrCodeResourceExhausted:
			return true
		default:
			return false
		}
	}
	return false
}

// GetErrorCategory returns the category of the error
func GetErrorCategory(err error) string {
	if starseekErr, ok := err.(*StarSeekError); ok {
		switch {
		case starseekErr.Code >= 1000 && starseekErr.Code < 2000:
			return "parameter"
		case starseekErr.Code >= 2000 && starseekErr.Code < 3000:
			return "business"
		case starseekErr.Code >= 3000 && starseekErr.Code < 4000:
			return "infrastructure"
		case starseekErr.Code >= 4000 && starseekErr.Code < 5000:
			return "authentication"
		default:
			return "unknown"
		}
	}
	return "system"
}

// Common error messages
var (
	ErrInvalidQuerySyntax  = NewError(ErrCodeInvalidQuerySyntax, "Invalid query syntax")
	ErrSearchResultEmpty   = NewError(ErrCodeSearchResultEmpty, "No search results found")
	ErrInternalServerError = NewError(ErrCodeInternalServerError, "Internal server error")
	ErrServiceUnavailable  = NewError(ErrCodeServiceUnavailable, "Service temporarily unavailable")
	ErrUnauthorized        = NewError(ErrCodeUnauthorized, "Authentication required")
	ErrPermissionDenied    = NewError(ErrCodePermissionDenied, "Permission denied")
)

//Personal.AI order the ending
