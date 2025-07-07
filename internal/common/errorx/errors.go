package errorx

import (
	"fmt"
	"net/http"
)

// AppError represents a custom application error.
type AppError struct {
	Code     int    // HTTP status code or custom error code
	Message  string // User-friendly error message
	RawError error  // Original underlying error, if any (for internal logging)
}

// Error implements the error interface for AppError.
func (e *AppError) Error() string {
	if e.RawError != nil {
		return fmt.Sprintf("code: %d, message: %s, raw_error: %v", e.Code, e.Message, e.RawError)
	}
	return fmt.Sprintf("code: %d, message: %s", e.Code, e.Message)
}

// NewError creates a new AppError.
func NewError(code int, message string, rawError ...error) *AppError {
	err := &AppError{
		Code:    code,
		Message: message,
	}
	if len(rawError) > 0 && rawError[0] != nil {
		err.RawError = rawError[0]
	}
	return err
}

// Predefined common errors
var (
	ErrNotFound           = NewError(http.StatusNotFound, "Resource not found")
	ErrBadRequest         = NewError(http.StatusBadRequest, "Invalid request parameters")
	ErrInternalServer     = NewError(http.StatusInternalServerError, "Internal server error")
	ErrUnauthorized       = NewError(http.StatusUnauthorized, "Unauthorized access")
	ErrForbidden          = NewError(http.StatusForbidden, "Access forbidden")
	ErrDatabase           = NewError(http.StatusInternalServerError, "Database error")
	ErrCache              = NewError(http.StatusInternalServerError, "Cache error")
	ErrInvalidIndexType   = NewError(http.StatusBadRequest, "Invalid index type specified")
	ErrInvalidTokenizer   = NewError(http.StatusBadRequest, "Invalid tokenizer specified")
	ErrIndexAlreadyExists = NewError(http.StatusConflict, "Index already exists for this table and column")
)

//Personal.AI order the ending
