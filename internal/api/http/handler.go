// internal/api/http/handler.go
package http

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"

	"github.com/turtacn/starseek/internal/application"   // Adjust module path as per your project
	"github.com/turtacn/starseek/internal/common/errors" // Adjust module path as per your project
)

// Handler represents the HTTP request handler.
type Handler struct {
	IndexService  application.IndexService
	SearchService application.SearchService
}

// NewHandler creates a new Handler instance.
func NewHandler(indexSvc application.IndexService, searchSvc application.SearchService) *Handler {
	return &Handler{
		IndexService:  indexSvc,
		SearchService: searchSvc,
	}
}

// apiError represents the standardized error response format.
type apiError struct {
	Code    errors.ErrorCode `json:"code"`
	Message string           `json:"message"`
}

// handleAppError converts an internal AppError to an appropriate HTTP response.
func handleAppError(c *gin.Context, appErr *errors.AppError, log *logrus.Entry) {
	var httpStatus int
	switch appErr.Code {
	case errors.ErrCodeNotFound:
		httpStatus = http.StatusNotFound
	case errors.ErrCodeAlreadyExists:
		httpStatus = http.StatusConflict // 409 Conflict for resource already existing
	case errors.ErrCodeInvalidInput:
		httpStatus = http.StatusBadRequest
	case errors.ErrCodeUnauthorized:
		httpStatus = http.StatusUnauthorized
	case errors.ErrCodeForbidden:
		httpStatus = http.StatusForbidden
	case errors.ErrCodeServiceUnavailable:
		httpStatus = http.StatusServiceUnavailable
	default:
		// Default to Internal Server Error for unhandled or generic internal errors
		httpStatus = http.StatusInternalServerError
	}

	log.WithError(appErr).Warnf("Handling application error: %s", appErr.Message)

	c.JSON(httpStatus, apiError{
		Code:    appErr.Code,
		Message: appErr.Message,
	})
}

// RegisterIndexRequest defines the request body for registering an index.
type RegisterIndexRequest struct {
	Name     string                 `json:"name" binding:"required"`
	Schema   map[string]interface{} `json:"schema" binding:"required"`
	Settings map[string]interface{} `json:"settings"`
}

// RegisterIndexResponse defines the response body for registering an index.
type RegisterIndexResponse struct {
	IndexID string `json:"index_id"`
	Message string `json:"message"`
}

// RegisterIndexHandler handles the HTTP request to register a new index.
// @Summary Register a new index
// @Description Registers a new search index with a specified name, schema, and optional settings.
// @Tags Index
// @Accept json
// @Produce json
// @Param index body RegisterIndexRequest true "Index registration details"
// @Success 201 {object} RegisterIndexResponse "Index successfully registered"
// @Failure 400 {object} apiError "Invalid request payload or parameters"
// @Failure 409 {object} apiError "Index with this name already exists"
// @Failure 500 {object} apiError "Internal server error"
// @Router /indices [post]
func (h *Handler) RegisterIndexHandler(c *gin.Context) {
	log := logrus.WithContext(c.Request.Context()).WithField("handler", "RegisterIndexHandler")

	var req RegisterIndexRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.WithError(err).Warn("Failed to bind register index request")
		c.JSON(http.StatusBadRequest, apiError{
			Code:    errors.ErrCodeInvalidInput,
			Message: "Invalid request payload: " + err.Error(),
		})
		return
	}

	log = log.WithField("index_name", req.Name)
	log.Infof("Received request to register index '%s'", req.Name)

	indexID, err := h.IndexService.RegisterIndex(c.Request.Context(), req.Name, req.Schema, req.Settings)
	if err != nil {
		if appErr, ok := err.(*errors.AppError); ok {
			handleAppError(c, appErr, log)
		} else {
			// Catch any unexpected non-AppError from service layer
			log.WithError(err).Error("Unexpected error from IndexService.RegisterIndex")
			c.JSON(http.StatusInternalServerError, apiError{
				Code:    errors.ErrCodeInternal,
				Message: "An unexpected internal error occurred.",
			})
		}
		return
	}

	log.Infof("Successfully registered index '%s' with ID '%s'", req.Name, indexID)
	c.JSON(http.StatusCreated, RegisterIndexResponse{
		IndexID: indexID,
		Message: "Index registered successfully",
	})
}

// SearchRequest defines the request body for a search query.
type SearchRequest struct {
	Query map[string]interface{} `json:"query" binding:"required"`
}

// SearchResponse defines the response body for a search query.
type SearchResponse struct {
	Results []map[string]interface{} `json:"results"`
	Total   int                      `json:"total"`
	Message string                   `json:"message"`
}

// SearchHandler handles the HTTP request to perform a search.
// @Summary Perform a search query
// @Description Executes a search query against a specified index.
// @Tags Search
// @Accept json
// @Produce json
// @Param index_name path string true "Name of the index to search against"
// @Param query body SearchRequest true "Search query details"
// @Success 200 {object} SearchResponse "Search results"
// @Failure 400 {object} apiError "Invalid request payload or parameters"
// @Failure 404 {object} apiError "Index not found"
// @Failure 500 {object} apiError "Internal server error"
// @Router /indices/{index_name}/_search [post]
func (h *Handler) SearchHandler(c *gin.Context) {
	log := logrus.WithContext(c.Request.Context()).WithField("handler", "SearchHandler")

	indexName := c.Param("index_name")
	if indexName == "" {
		log.Warn("Missing index_name path parameter")
		c.JSON(http.StatusBadRequest, apiError{
			Code:    errors.ErrCodeInvalidInput,
			Message: "Missing index name in path parameter",
		})
		return
	}

	var req SearchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.WithError(err).Warnf("Failed to bind search request for index '%s'", indexName)
		c.JSON(http.StatusBadRequest, apiError{
			Code:    errors.ErrCodeInvalidInput,
			Message: "Invalid request payload: " + err.Error(),
		})
		return
	}

	log = log.WithFields(logrus.Fields{
		"index_name": indexName,
		"query":      req.Query,
	})
	log.Infof("Received search request for index '%s'", indexName)

	results, err := h.SearchService.Search(c.Request.Context(), indexName, req.Query)
	if err != nil {
		if appErr, ok := err.(*errors.AppError); ok {
			handleAppError(c, appErr, log)
		} else {
			log.WithError(err).Error("Unexpected error from SearchService.Search")
			c.JSON(http.StatusInternalServerError, apiError{
				Code:    errors.ErrCodeInternal,
				Message: "An unexpected internal error occurred.",
			})
		}
		return
	}

	log.Infof("Successfully performed search for index '%s', found %d results", indexName, len(results))
	c.JSON(http.StatusOK, SearchResponse{
		Results: results,
		Total:   len(results),
		Message: "Search completed successfully",
	})
}
