package application

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"go.uber.org/zap"

	"github.com/turtacn/starseek/internal/common/enum"
	ferrors "github.com/turtacn/starseek/internal/common/errors"
	"github.com/turtacn/starseek/internal/domain/search"         // 引入 domain.search 包的 SearchService, SearchRequest, SearchResponse
	"github.com/turtacn/starseek/internal/infrastructure/cache"  // 引入 cache 接口
	"github.com/turtacn/starseek/internal/infrastructure/logger" // 引入 logger 接口
	"github.com/turtacn/starseek/pkg/dto"                        // 引入 DTOs
)

// SearchApplicationService 接口定义了搜索的应用层操作。
// 它是 API 层与领域层之间的桥梁，负责请求转换、缓存、以及协调领域服务。
type SearchApplicationService interface {
	// Search 方法接收 DTO 层的 SearchRequest，执行搜索流程，并返回 DTO 层的 SearchResponse。
	Search(ctx context.Context, req *dto.SearchRequest) (*dto.SearchResponse, error)
}

// searchApplicationService 是 SearchApplicationService 接口的实现。
// 它封装了对领域服务 search.SearchService 的调用，并处理应用层的缓存和转换逻辑。
type searchApplicationService struct {
	searchService search.SearchService // 依赖领域层的 SearchService
	cacheClient   cache.CacheClient    // 依赖缓存客户端
	log           logger.Logger
	cacheTTL      time.Duration // 搜索结果缓存时间
}

// NewSearchApplicationService 创建并返回一个新的 SearchApplicationService 实例。
func NewSearchApplicationService(
	searchService search.SearchService,
	cacheClient cache.CacheClient,
	log logger.Logger,
	cacheTTL time.Duration,
) (SearchApplicationService, error) {
	if searchService == nil {
		return nil, ferrors.NewInternalError("search domain service is nil", nil)
	}
	if cacheClient == nil {
		return nil, ferrors.NewInternalError("cache client is nil", nil)
	}
	if log == nil {
		return nil, ferrors.NewInternalError("logger is nil", nil)
	}
	if cacheTTL <= 0 {
		log.Warn("Cache TTL is zero or negative, setting to default (5 minutes)", zap.Duration("original_ttl", cacheTTL))
		cacheTTL = 5 * time.Minute // Default TTL
	}

	log.Info("SearchApplicationService initialized.")
	return &searchApplicationService{
		searchService: searchService,
		cacheClient:   cacheClient,
		log:           log,
		cacheTTL:      cacheTTL,
	}, nil
}

// Search 方法实现了搜索的应用层业务流程。
func (s *searchApplicationService) Search(ctx context.Context, req *dto.SearchRequest) (*dto.SearchResponse, error) {
	s.log.Info("SearchApplicationService received request", zap.Any("request_query", req.Query), zap.String("request_id", req.RequestID))

	// 1. 参数校验 (应用层初步校验)
	if req.Query == "" && len(req.Filters) == 0 {
		return nil, ferrors.NewBadRequestError("search query and filters cannot both be empty", nil)
	}
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 || req.PageSize > 100 { // Limit page size to prevent abuse
		req.PageSize = 20
	}
	if len(req.Tables) == 0 {
		return nil, ferrors.NewBadRequestError("at least one table must be specified for search", nil)
	}
	// 校验 QueryType
	domainQueryType, err := search.ParseQueryType(req.QueryType)
	if err != nil {
		s.log.Warn("Invalid QueryType in request, defaulting to MATCH_ANY", zap.String("received_query_type", req.QueryType))
		domainQueryType = search.MATCH_ANY
	}

	// 2. 缓存查询
	// 使用请求的 JSON 序列化作为缓存键。注意：实际生产中需要更稳定的缓存键生成策略，
	// 例如，对 DTO 进行规范化处理（排序字段、统一大小写等）后再进行哈希。
	cacheKeyBytes, err := json.Marshal(req)
	if err != nil {
		s.log.Error("Failed to marshal search request for cache key generation", zap.Error(err))
		// Continue without caching if key generation fails
	}
	cacheKey := fmt.Sprintf("search:%x", cacheKeyBytes) // Using hex for readability, or use a cryptographic hash

	if cachedRespBytes, found := s.cacheClient.Get(ctx, cacheKey); found {
		var cachedResponse dto.SearchResponse
		if err := json.Unmarshal(cachedRespBytes, &cachedResponse); err == nil {
			s.log.Info("Cache hit for search request", zap.String("cache_key", cacheKey), zap.String("request_id", req.RequestID))
			return &cachedResponse, nil
		}
		s.log.Warn("Failed to unmarshal cached search response, proceeding with actual search", zap.Error(err), zap.String("cache_key", cacheKey))
		// If unmarshal fails, treat as cache miss and proceed
	}
	s.log.Info("Cache miss for search request", zap.String("cache_key", cacheKey), zap.String("request_id", req.RequestID))

	// 3. 将 DTO 转换为领域层请求
	domainReq := &search.SearchRequest{
		Query:        req.Query,
		Tables:       req.Tables,
		Fields:       req.Fields,
		Filters:      req.Filters, // Filters DTO should ideally map to domain.search.Filter
		Page:         req.Page,
		PageSize:     req.PageSize,
		QueryType:    domainQueryType,
		DatabaseType: enum.DatabaseType(strings.ToUpper(req.DatabaseType)), // Ensure consistency with enum
		// MinScore, Language, SortOrder 等其他字段也需要映射
	}

	// 4. 调用领域服务进行实际搜索
	domainResp, err := s.searchService.Search(ctx, domainReq)
	if err != nil {
		s.log.Error("Search domain service failed", zap.Error(err), zap.String("request_id", req.RequestID))
		return nil, ferrors.WrapInternalError(err, "search operation failed") // Wrap domain error
	}

	// 5. 将领域层响应转换为 DTO 响应
	appResp := &dto.SearchResponse{
		TotalHits: domainResp.TotalHits,
		Page:      domainResp.Page,
		PageSize:  domainResp.PageSize,
		Results:   make([]dto.SearchResult, len(domainResp.Results)),
	}

	for i, r := range domainResp.Results {
		// 这里将 domain.search.SearchResult 转换为 dto.SearchResult
		// 注意 Data 和 HighlightedData 是 map[string]interface{} 或 map[string]string，
		// 直接赋值通常是安全的，但如果 DTO 需要更严格的类型，可能需要进一步转换。
		appResp.Results[i] = dto.SearchResult{
			RowID:           r.RowID,
			Score:           r.Score,
			Data:            r.Data,
			HighlightedData: r.HighlightedData,
		}
	}
	// 可以添加 QueryLatency 等统计信息
	appResp.QueryLatency = domainResp.QueryLatency

	// 6. 缓存结果 (仅在成功时缓存)
	if cacheKey != "" { // Only attempt to set if cacheKey generation was successful
		respBytes, err := json.Marshal(appResp)
		if err != nil {
			s.log.Error("Failed to marshal search response for caching", zap.Error(err), zap.String("request_id", req.RequestID))
		} else {
			if err := s.cacheClient.Set(ctx, cacheKey, respBytes, s.cacheTTL); err != nil {
				s.log.Error("Failed to cache search response", zap.Error(err), zap.String("cache_key", cacheKey), zap.String("request_id", req.RequestID))
			} else {
				s.log.Info("Search response cached successfully", zap.String("cache_key", cacheKey), zap.String("request_id", req.RequestID))
			}
		}
	}

	s.log.Info("SearchApplicationService completed successfully",
		zap.Int64("total_hits", appResp.TotalHits),
		zap.Float64("query_latency_ms", appResp.QueryLatency.Milliseconds()),
		zap.String("request_id", req.RequestID))

	return appResp, nil
}
