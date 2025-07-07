package handler

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/turtacn/starseek/internal/api/dto"
	"github.com/turtacn/starseek/internal/application/service"
	"github.com/turtacn/starseek/internal/common/errorx"
	"go.uber.org/zap"
)

// IndexHandler handles HTTP requests related to index management.
type IndexHandler struct {
	indexAppService *service.IndexAppService
	logger          *zap.Logger
}

// NewIndexHandler creates a new IndexHandler.
func NewIndexHandler(indexAppService *service.IndexAppService, logger *zap.Logger) *IndexHandler {
	return &IndexHandler{
		indexAppService: indexAppService,
		logger:          logger.Named("IndexHandler"),
	}
}

// Register handles the POST /indexes API request to register a new index metadata.
func (h *IndexHandler) Register(c *gin.Context) {
	var req dto.RegisterIndexReq
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warn("Bad request - failed to bind JSON for index registration", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"code": errorx.ErrBadRequest.Code, "message": fmt.Sprintf("Invalid request body: %v", err.Error())})
		return
	}

	h.logger.Debug("Received index registration request", zap.Any("request", req))

	// Convert DTO to domain model
	metadata := req.ToModel()

	// Call application service to register the index
	registeredMetadata, appErr := h.indexAppService.RegisterIndex(c.Request.Context(), metadata)
	if appErr != nil {
		h.logger.Error("Index registration application service failed", zap.Error(appErr), zap.Any("request", req))
		c.JSON(appErr.Code, gin.H{"code": appErr.Code, "message": appErr.Message})
		return
	}

	// Convert domain model result back to DTO response
	res := dto.IndexResFromModel(registeredMetadata)

	c.JSON(http.StatusCreated, gin.H{
		"code":    http.StatusCreated,
		"message": "Index registered successfully",
		"data":    res,
	})
}

// List handles the GET /indexes API request to list all registered indexes.
func (h *IndexHandler) List(c *gin.Context) {
	h.logger.Debug("Received list indexes request")

	// Call application service to list all indexes
	indexes, appErr := h.indexAppService.ListIndexes(c.Request.Context())
	if appErr != nil {
		h.logger.Error("List indexes application service failed", zap.Error(appErr))
		c.JSON(appErr.Code, gin.H{"code": appErr.Code, "message": appErr.Message})
		return
	}

	// Convert domain models to DTO response
	res := dto.ListIndexesResFromModels(indexes)

	c.JSON(http.StatusOK, gin.H{
		"code":    http.StatusOK,
		"message": "Success",
		"data":    res,
	})
}

// Delete handles the DELETE /indexes/:id API request to delete an index metadata.
func (h *IndexHandler) Delete(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		h.logger.Warn("Bad request - invalid index ID format", zap.String("id", idStr), zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"code": errorx.ErrBadRequest.Code, "message": "Invalid index ID format"})
		return
	}

	h.logger.Debug("Received delete index request", zap.Uint("id", uint(id)))

	// Call application service to delete the index
	appErr := h.indexAppService.DeleteIndex(c.Request.Context(), uint(id))
	if appErr != nil {
		h.logger.Error("Delete index application service failed", zap.Error(appErr), zap.Uint("id", uint(id)))
		c.JSON(appErr.Code, gin.H{"code": appErr.Code, "message": appErr.Message})
		return
	}

	c.JSON(http.StatusNoContent, nil) // 204 No Content for successful deletion
}

//Personal.AI order the ending
