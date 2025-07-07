package handler

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/turtacn/starseek/internal/api/dto"
	"github.com/turtacn/starseek/internal/application/service"
	"github.com/turtacn/starseek/internal/common/constant"
	"github.com/turtacn/starseek/internal/common/errorx"
	"go.uber.org/zap"
)

// SearchHandler handles HTTP requests related to search operations.
type SearchHandler struct {
	searchAppService *service.SearchAppService
	logger           *zap.Logger
}

// NewSearchHandler creates a new SearchHandler.
func NewSearchHandler(searchAppService *service.SearchAppService, logger *zap.Logger) *SearchHandler {
	return &SearchHandler{
		searchAppService: searchAppService,
		logger:           logger.Named("SearchHandler"),
	}
}

// Search handles the GET /search API request.
func (h *SearchHandler) Search(c *gin.Context) {
	var req dto.SearchReq

	// Bind query parameters to SearchReq struct
	if err := c.ShouldBindQuery(&req); err != nil {
		h.logger.Warn("Bad request - failed to bind query parameters", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"code": errorx.ErrBadRequest.Code, "message": fmt.Sprintf("Invalid request parameters: %v", err.Error())})
		return
	}

	// Validate fields (comma-separated string)
	if len(req.Fields) == 0 {
		req.Fields = []string{} // Initialize if empty
	}

	// Set default pagination values if not provided or invalid
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = constant.DefaultPageSize
	}

	h.logger.Debug("Received search request",
		zap.String("query", req.Query),
		zap.Strings("fields", req.Fields),
		zap.Int("page", req.Page),
		zap.Int("pageSize", req.PageSize))

	// Call the application service
	searchResult, appErr := h.searchAppService.Search(c.Request.Context(), req.Query, req.Fields, req.Page, req.PageSize)
	if appErr != nil {
		h.logger.Error("Search application service failed", zap.Error(appErr),
			zap.String("query", req.Query), zap.Strings("fields", req.Fields))
		c.JSON(appErr.Code, gin.H{"code": appErr.Code, "message": appErr.Message})
		return
	}

	// Convert domain model result to DTO response
	res := dto.FromModel(searchResult)

	c.JSON(http.StatusOK, gin.H{
		"code":    http.StatusOK,
		"message": "Success",
		"data":    res,
	})
}

//Personal.AI order the ending
