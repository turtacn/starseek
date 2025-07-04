package constants

import (
	"fmt"
)

// Error code ranges
const (
	// Parameter errors: 1000-1999
	ErrorCodeParameterBase = 1000

	// Business logic errors: 2000-2999
	ErrorCodeBusinessBase = 2000

	// Infrastructure errors: 3000-3999
	ErrorCodeInfrastructureBase = 3000

	// System errors: 4000-4999
	ErrorCodeSystemBase = 4000
)

// Parameter error codes (1000-1999)
const (
	ErrorCodeInvalidParameter     = 1001
	ErrorCodeMissingParameter     = 1002
	ErrorCodeParameterOutOfRange  = 1003
	ErrorCodeInvalidFormat        = 1004
	ErrorCodeInvalidQuerySyntax   = 1005
	ErrorCodeInvalidPagination    = 1006
	ErrorCodeInvalidSortField     = 1007
	ErrorCodeInvalidFilterFormat  = 1008
	ErrorCodeQueryTooLong         = 1009
	ErrorCodeResultSetTooLarge    = 1010
)

// Business logic error codes (2000-2999)
const (
	ErrorCodeIndexNotFound        = 2001
	ErrorCodeIndexAlreadyExists   = 2002
	ErrorCodeIndexConfigInvalid   = 2003
	ErrorCodeSearchFailed         = 2004
	ErrorCodeNoSearchResults      = 2005
	ErrorCodeRankingFailed        = 2006
	ErrorCodeTokenizationFailed   = 2007
	ErrorCodeCacheKeyNotFound     = 2008
	ErrorCodeCacheExpired         = 2009
	ErrorCodeOperationNotAllowed  = 2010
	ErrorCodeResourceNotFound     = 2011
	ErrorCodeResourceConflict     = 2012
	ErrorCodePermissionDenied     = 2013
	ErrorCodeQuotaExceeded        = 2014
	ErrorCodeRateLimitExceeded    = 2015
)

// Infrastructure error codes (3000-3999)
const (
	ErrorCodeDatabaseConnection   = 3001
	ErrorCodeDatabaseQuery        = 3002
	ErrorCodeDatabaseTimeout      = 3003
	ErrorCodeCacheConnection      = 3004
	ErrorCodeCacheOperation       = 3005
	ErrorCodeNetworkTimeout       = 3006
	ErrorCodeServiceUnavailable   = 3007
	ErrorCodeConfigurationError   = 3008
	ErrorCodeExternalServiceError = 3009
	ErrorCodeStorageError         = 3010
	ErrorCodeSerializationError   = 3011
	ErrorCodeDeserializationError = 3012
)

// System error codes (4000-4999)
const (
	ErrorCodeInternalError        = 4001
	ErrorCodeUnknownError         = 4002
	ErrorCodeNotImplemented       = 4003
	ErrorCodeServiceStartupFailed = 4004
	ErrorCodeMemoryExhausted      = 4005
	ErrorCodeResourceExhausted    = 4006
	ErrorCodeDeadlockDetected     = 4007
	ErrorCodeConcurrencyError     = 4008
	ErrorCodeContextCanceled      = 4009
	ErrorCodeContextTimeout       = 4010
)

// StarSeekError represents a custom error type for the StarSeek system
type StarSeekError struct {
	Code      int                    `json:"code"`
	Message   string                 `json:"message"`
	Details   string                 `json:"details,omitempty"`
	Cause     error                  `json:"cause,omitempty"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
	Timestamp int64                  `json:"timestamp"`
}

// Error implements the error interface
func (e *StarSeekError) Error() string {
	if e.Details != "" {
		return fmt.Sprintf("[%d] %s: %s", e.Code, e.Message, e.Details)
	}
	return fmt.Sprintf("[%d] %s", e.Code, e.Message)
}

// Unwrap returns the underlying cause error
func (e *StarSeekError) Unwrap() error {
	return e.Cause
}

// WithMetadata adds metadata to the error
func (e *StarSeekError) WithMetadata(key string, value interface{}) *StarSeekError {
	if e.Metadata == nil {
		e.Metadata = make(map[string]interface{})
	}
	e.Metadata[key] = value
	return e
}

// WithCause adds a cause error
func (e *StarSeekError) WithCause(cause error) *StarSeekError {
	e.Cause = cause
	return e
}

// NewError creates a new StarSeekError
func NewError(code int, message string) *StarSeekError {
	return &StarSeekError{
		Code:      code,
		Message:   message,
		Timestamp: getCurrentTimestamp(),
	}
}

// NewErrorWithDetails creates a new StarSeekError with details
func NewErrorWithDetails(code int, message, details string) *StarSeekError {
	return &StarSeekError{
		Code:      code,
		Message:   message,
		Details:   details,
		Timestamp: getCurrentTimestamp(),
	}
}

// NewErrorWithCause creates a new StarSeekError with a cause
func NewErrorWithCause(code int, message string, cause error) *StarSeekError {
	return &StarSeekError{
		Code:      code,
		Message:   message,
		Cause:     cause,
		Timestamp: getCurrentTimestamp(),
	}
}

// Predefined common errors

// Parameter errors
var (
	ErrInvalidParameter     = NewError(ErrorCodeInvalidParameter, "Invalid parameter")
	ErrMissingParameter     = NewError(ErrorCodeMissingParameter, "Missing required parameter")
	ErrParameterOutOfRange  = NewError(ErrorCodeParameterOutOfRange, "Parameter value out of range")
	ErrInvalidFormat        = NewError(ErrorCodeInvalidFormat, "Invalid format")
	ErrInvalidQuerySyntax   = NewError(ErrorCodeInvalidQuerySyntax, "Invalid query syntax")
	ErrInvalidPagination    = NewError(ErrorCodeInvalidPagination, "Invalid pagination parameters")
	ErrInvalidSortField     = NewError(ErrorCodeInvalidSortField, "Invalid sort field")
	ErrInvalidFilterFormat  = NewError(ErrorCodeInvalidFilterFormat, "Invalid filter format")
	ErrQueryTooLong         = NewError(ErrorCodeQueryTooLong, "Query string too long")
	ErrResultSetTooLarge    = NewError(ErrorCodeResultSetTooLarge, "Result set too large")
)

// Business logic errors
var (
	ErrIndexNotFound        = NewError(ErrorCodeIndexNotFound, "Index not found")
	ErrIndexAlreadyExists   = NewError(ErrorCodeIndexAlreadyExists, "Index already exists")
	ErrIndexConfigInvalid   = NewError(ErrorCodeIndexConfigInvalid, "Invalid index configuration")
	ErrSearchFailed         = NewError(ErrorCodeSearchFailed, "Search operation failed")
	ErrNoSearchResults      = NewError(ErrorCodeNoSearchResults, "No search results found")
	ErrRankingFailed        = NewError(ErrorCodeRankingFailed, "Ranking calculation failed")
	ErrTokenizationFailed   = NewError(ErrorCodeTokenizationFailed, "Text tokenization failed")
	ErrCacheKeyNotFound     = NewError(ErrorCodeCacheKeyNotFound, "Cache key not found")
	ErrCacheExpired         = NewError(ErrorCodeCacheExpired, "Cache entry expired")
	ErrOperationNotAllowed  = NewError(ErrorCodeOperationNotAllowed, "Operation not allowed")
	ErrResourceNotFound     = NewError(ErrorCodeResourceNotFound, "Resource not found")
	ErrResourceConflict     = NewError(ErrorCodeResourceConflict, "Resource conflict")
	ErrPermissionDenied     = NewError(ErrorCodePermissionDenied, "Permission denied")
	ErrQuotaExceeded        = NewError(ErrorCodeQuotaExceeded, "Quota exceeded")
	ErrRateLimitExceeded    = NewError(ErrorCodeRateLimitExceeded, "Rate limit exceeded")
)

// Infrastructure errors
var (
	ErrDatabaseConnection   = NewError(ErrorCodeDatabaseConnection, "Database connection failed")
	ErrDatabaseQuery        = NewError(ErrorCodeDatabaseQuery, "Database query failed")
	ErrDatabaseTimeout      = NewError(ErrorCodeDatabaseTimeout, "Database operation timeout")
	ErrCacheConnection      = NewError(ErrorCodeCacheConnection, "Cache connection failed")
	ErrCacheOperation       = NewError(ErrorCodeCacheOperation, "Cache operation failed")
	ErrNetworkTimeout       = NewError(ErrorCodeNetworkTimeout, "Network operation timeout")
	ErrServiceUnavailable   = NewError(ErrorCodeServiceUnavailable, "Service unavailable")
	ErrConfigurationError   = NewError(ErrorCodeConfigurationError, "Configuration error")
	ErrExternalServiceError = NewError(ErrorCodeExternalServiceError, "External service error")
	ErrStorageError         = NewError(ErrorCodeStorageError, "Storage operation failed")
	ErrSerializationError   = NewError(ErrorCodeSerializationError, "Serialization failed")
	ErrDeserializationError = NewError(ErrorCodeDeserializationError, "Deserialization failed")
)

// System errors
var (
	ErrInternalError        = NewError(ErrorCodeInternalError, "Internal server error")
	ErrUnknownError         = NewError(ErrorCodeUnknownError, "Unknown error")
	ErrNotImplemented       = NewError(ErrorCodeNotImplemented, "Feature not implemented")
	ErrServiceStartupFailed = NewError(ErrorCodeServiceStartupFailed, "Service startup failed")
	ErrMemoryExhausted      = NewError(ErrorCodeMemoryExhausted, "Memory exhausted")
	ErrResourceExhausted    = NewError(ErrorCodeResourceExhausted, "Resource exhausted")
	ErrDeadlockDetected     = NewError(ErrorCodeDeadlockDetected, "Deadlock detected")
	ErrConcurrencyError     = NewError(ErrorCodeConcurrencyError, "Concurrency error")
	ErrContextCanceled      = NewError(ErrorCodeContextCanceled, "Context canceled")
	ErrContextTimeout       = NewError(ErrorCodeContextTimeout, "Context timeout")
)

// Error classification functions

// IsParameterError checks if the error is a parameter error
func IsParameterError(err error) bool {
	if starSeekErr, ok := err.(*StarSeekError); ok {
		return starSeekErr.Code >= ErrorCodeParameterBase && starSeekErr.Code < ErrorCodeBusinessBase
	}
	return false
}

// IsBusinessError checks if the error is a business logic error
func IsBusinessError(err error) bool {
	if starSeekErr, ok := err.(*StarSeekError); ok {
		return starSeekErr.Code >= ErrorCodeBusinessBase && starSeekErr.Code < ErrorCodeInfrastructureBase
	}
	return false
}

// IsInfrastructureError checks if the error is an infrastructure error
func IsInfrastructureError(err error) bool {
	if starSeekErr, ok := err.(*StarSeekError); ok {
		return starSeekErr.Code >= ErrorCodeInfrastructureBase && starSeekErr.Code < ErrorCodeSystemBase
	}
	return false
}

// IsSystemError checks if the error is a system error
func IsSystemError(err error) bool {
	if starSeekErr, ok := err.(*StarSeekError); ok {
		return starSeekErr.Code >= ErrorCodeSystemBase
	}
	return false
}

// IsRetryable checks if the error is retryable
func IsRetryable(err error) bool {
	if starSeekErr, ok := err.(*StarSeekError); ok {
		switch starSeekErr.Code {
		case ErrorCodeDatabaseTimeout,
			ErrorCodeNetworkTimeout,
			ErrorCodeServiceUnavailable,
			ErrorCodeCacheConnection,
			ErrorCodeExternalServiceError,
			ErrorCodeContextTimeout:
			return true
		}
	}
	return false
}

// IsTemporary checks if the error is temporary
func IsTemporary(err error) bool {
	if starSeekErr, ok := err.(*StarSeekError); ok {
		switch starSeekErr.Code {
		case ErrorCodeRateLimitExceeded,
			ErrorCodeQuotaExceeded,
			ErrorCodeMemoryExhausted,
			ErrorCodeResourceExhausted,
			ErrorCodeConcurrencyError:
			return true
		}
	}
	return IsRetryable(err)
}

// GetErrorCategory returns the error category as string
func GetErrorCategory(err error) string {
	if IsParameterError(err) {
		return "PARAMETER"
	}
	if IsBusinessError(err) {
		return "BUSINESS"
	}
	if IsInfrastructureError(err) {
		return "INFRASTRUCTURE"
	}
	if IsSystemError(err) {
		return "SYSTEM"
	}
	return "UNKNOWN"
}

// GetErrorCode extracts error code from error
func GetErrorCode(err error) int {
	if starSeekErr, ok := err.(*StarSeekError); ok {
		return starSeekErr.Code
	}
	return ErrorCodeUnknownError
}

// WrapError wraps an existing error with StarSeekError
func WrapError(code int, message string, cause error) *StarSeekError {
	return &StarSeekError{
		Code:      code,
		Message:   message,
		Cause:     cause,
		Timestamp: getCurrentTimestamp(),
	}
}

// getCurrentTimestamp returns current unix timestamp
func getCurrentTimestamp() int64 {
	// This would typically use time.Now().Unix()
	// For now, return 0 to avoid importing time package
	return 0
}

// Error message templates for consistent error messaging
const (
	TemplateInvalidParameter     = "Invalid parameter '%s': %s"
	TemplateMissingParameter     = "Missing required parameter: %s"
	TemplateParameterOutOfRange  = "Parameter '%s' value %v is out of range [%v, %v]"
	TemplateResourceNotFound     = "Resource '%s' with ID '%s' not found"
	TemplateResourceConflict     = "Resource '%s' with ID '%s' already exists"
	TemplateOperationFailed      = "Operation '%s' failed: %s"
	TemplateServiceUnavailable   = "Service '%s' is currently unavailable"
	TemplateConnectionFailed     = "Failed to connect to %s: %s"
	TemplateTimeoutError         = "Operation '%s' timed out after %v"
)

// Helper functions for creating formatted errors

// NewInvalidParameterError creates a formatted invalid parameter error
func NewInvalidParameterError(paramName, reason string) *StarSeekError {
	return NewErrorWithDetails(ErrorCodeInvalidParameter,
		"Invalid parameter",
		fmt.Sprintf(TemplateInvalidParameter, paramName, reason))
}

// NewMissingParameterError creates a formatted missing parameter error
func NewMissingParameterError(paramName string) *StarSeekError {
	return NewErrorWithDetails(ErrorCodeMissingParameter,
		"Missing parameter",
		fmt.Sprintf(TemplateMissingParameter, paramName))
}

// NewResourceNotFoundError creates a formatted resource not found error
func NewResourceNotFoundError(resourceType, resourceID string) *StarSeekError {
	return NewErrorWithDetails(ErrorCodeResourceNotFound,
		"Resource not found",
		fmt.Sprintf(TemplateResourceNotFound, resourceType, resourceID))
}

// NewResourceConflictError creates a formatted resource conflict error
func NewResourceConflictError(resourceType, resourceID string) *StarSeekError {
	return NewErrorWithDetails(ErrorCodeResourceConflict,
		"Resource conflict",
		fmt.Sprintf(TemplateResourceConflict, resourceType, resourceID))
}

// NewOperationFailedError creates a formatted operation failed error
func NewOperationFailedError(operation, reason string) *StarSeekError {
	return NewErrorWithDetails(ErrorCodeInternalError,
		"Operation failed",
		fmt.Sprintf(TemplateOperationFailed, operation, reason))
}

//Personal.AI order the ending
