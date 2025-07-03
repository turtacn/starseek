package search

import (
	"context"
	"fmt"
	"math"
	"sort"
	"strconv"
	"strings"

	"go.uber.org/zap"

	ferrors "github.com/turtacn/starseek/internal/common/errors"
	"github.com/turtacn/starseek/internal/domain/tokenizer"      // 引入 Tokenizer 接口
	"github.com/turtacn/starseek/internal/infrastructure/cache"  // 引入 Cache 接口 (例如 Redis 客户端)
	"github.com/turtacn/starseek/internal/infrastructure/logger" // 引入日志接口
)

// Ranker 接口定义了对搜索结果进行相关度打分和重新排序的能力。
type Ranker interface {
	// Rank 方法接收搜索结果列表和原始查询，并根据相关性算法对结果进行打分和排序。
	// 结果的 Score 字段会被更新，并且切片会被就地排序。
	Rank(ctx context.Context, results []*SearchResult, query *SearchQuery) error
}

// IDFProvider 接口定义了获取词项逆文档频率 (IDF) 统计的能力。
type IDFProvider interface {
	// GetIDF 方法根据词项获取其 IDF 统计数据。
	// 如果词项不存在，可能返回 nil 或带有 DocumentFrequency=0 的 IDFStats。
	GetIDF(ctx context.Context, term string) (*IDFStats, error)
}

// ==============================================================================
// BM25 Ranker 实现
// ==============================================================================

// BM25 算法的默认参数
const (
	bm25k1Default = 1.2
	bm25bDefault  = 0.75
)

// bm25Ranker 是 Ranker 接口的实现，使用 BM25 算法。
type bm25Ranker struct {
	idfProvider IDFProvider
	tokenizer   tokenizer.Tokenizer // 用于在计算TF时对文档内容进行分词
	log         logger.Logger
	k1          float64 // BM25参数 k1 (词频饱和度)
	b           float64 // BM25参数 b (文档长度归一化)
	// avgDocumentLength 实际项目中会从一个持久化存储中获取，而不是每次计算。
	// 为了简化，这里在 Rank 方法内部计算。
	// documentLengths 也可以预先计算并存储在索引中。
}

// NewBM25Ranker 创建并返回一个新的 BM25 Ranker 实例。
// k1和b参数可以用于调优BM25算法。
func NewBM25Ranker(idfProvider IDFProvider, tokenizer tokenizer.Tokenizer, log logger.Logger, k1, b float64) (Ranker, error) {
	if idfProvider == nil {
		return nil, ferrors.NewInternalError("IDF provider is nil", nil)
	}
	if tokenizer == nil {
		return nil, ferrors.NewInternalError("tokenizer is nil", nil)
	}
	if log == nil {
		return nil, ferrors.NewInternalError("logger is nil", nil)
	}

	// 使用默认值如果传入参数无效
	if k1 <= 0 {
		k1 = bm25k1Default
	}
	if b < 0 || b > 1 {
		b = bm25bDefault
	}

	log.Info("BM25 Ranker initialized", zap.Float64("k1", k1), zap.Float64("b", b))
	return &bm25Ranker{
		idfProvider: idfProvider,
		tokenizer:   tokenizer,
		log:         log,
		k1:          k1,
		b:           b,
	}, nil
}

// Rank 方法根据 BM25 算法对搜索结果进行打分和排序。
func (r *bm25Ranker) Rank(ctx context.Context, results []*SearchResult, query *SearchQuery) error {
	if len(results) == 0 || len(query.Keywords) == 0 {
		r.log.Debug("No results or keywords to rank, returning.",
			zap.Int("results_count", len(results)), zap.Int("keywords_count", len(query.Keywords)))
		return nil
	}

	// 1. 获取所有查询关键词的 IDF 值
	idfMap := make(map[string]*IDFStats)
	for _, term := range query.Keywords {
		idfStats, err := r.idfProvider.GetIDF(ctx, term)
		if err != nil {
			r.log.Warn("Failed to get IDF for term, using 0.0", zap.String("term", term), zap.Error(err))
			idfMap[term] = &IDFStats{Term: term, InverseDocumentFrequency: 0.0} // Fallback to 0.0 IDF
			continue
		}
		if idfStats == nil || idfStats.DocumentFrequency == 0 || idfStats.TotalDocuments == 0 {
			r.log.Debug("IDF stats not found or zero frequency for term, using 0.0", zap.String("term", term))
			idfMap[term] = &IDFStats{Term: term, InverseDocumentFrequency: 0.0} // Fallback to 0.0 IDF
			continue
		}
		idfMap[term] = idfStats
	}

	// 2. 计算平均文档长度 (AvgDL)
	// 在实际系统中，AvgDL 和每个文档的长度 DL 应该预先计算并存储在索引中，
	// 而不是在每次查询时动态计算，因为这会非常耗时。
	// 这里为了示例的完整性，进行简化计算。
	totalDocumentLength := 0.0
	documentLengths := make(map[string]float64) // Map of RowID -> DocumentLength

	// Identify relevant text fields from `query.Fields` or all string fields in `Data`
	// For simplicity, let's assume all string values in `Data` map are relevant.
	// In a real scenario, you'd have metadata about which fields are text-searchable.
	for _, res := range results {
		currentDocLen := 0.0
		for field, val := range res.Data {
			strVal, ok := val.(string)
			if !ok {
				continue // Only consider string fields for document length
			}
			// If specific fields are specified in query, only consider those
			if len(query.Fields) > 0 {
				found := false
				for _, f := range query.Fields {
					if f == field {
						found = true
						break
					}
				}
				if !found {
					continue
				}
			}

			// Tokenize field content to get a more accurate length based on words
			fieldTokens, err := r.tokenizer.Tokenize(strVal)
			if err != nil {
				r.log.Warn("Failed to tokenize field for document length calculation",
					zap.String("row_id", res.RowID), zap.String("field", field), zap.Error(err))
				continue
			}
			currentDocLen += float64(len(fieldTokens))
		}
		documentLengths[res.RowID] = currentDocLen
		totalDocumentLength += currentDocLen
	}

	avgDocumentLength := 0.0
	if len(results) > 0 {
		avgDocumentLength = totalDocumentLength / float64(len(results))
	}
	r.log.Debug("Calculated average document length", zap.Float64("avg_dl", avgDocumentLength))

	// 3. 计算每个文档的 BM25 分数
	for _, res := range results {
		docScore := 0.0
		dl := documentLengths[res.RowID] // Current document length

		for _, qTerm := range query.Keywords {
			idfStats, ok := idfMap[qTerm]
			if !ok || idfStats.InverseDocumentFrequency == 0.0 {
				continue // Skip if IDF not found or zero (term not important)
			}
			idf := idfStats.InverseDocumentFrequency

			// Calculate Term Frequency (TF) for qTerm in the current document (res)
			tf := 0.0
			for field, val := range res.Data {
				strVal, ok := val.(string)
				if !ok {
					continue
				}
				if len(query.Fields) > 0 {
					found := false
					for _, f := range query.Fields {
						if f == field {
							found = true
							break
						}
					}
					if !found {
						continue
					}
				}

				fieldTokens, err := r.tokenizer.Tokenize(strVal)
				if err != nil {
					r.log.Warn("Failed to tokenize field for TF calculation",
						zap.String("row_id", res.RowID), zap.String("field", field), zap.Error(err))
					continue
				}
				// Count occurrences of qTerm (case-insensitive for simplicity, or pre-normalize)
				count := 0
				for _, token := range fieldTokens {
					// Assuming keywords are already normalized (e.g., lowercased).
					// If not, perform normalization here for `token`.
					if strings.EqualFold(token, qTerm) { // Use EqualFold for case-insensitive matching
						count++
					}
				}
				tf += float64(count)
			}

			// Apply BM25 formula for this term and document
			// BM25 = IDF * ( (TF * (k1 + 1)) / (TF + k1 * (1 - b + b * (DL / AvgDL))) )
			numerator := tf * (r.k1 + 1)
			denominator := tf + r.k1*(1-r.b+r.b*(dl/avgDocumentLength))

			if denominator == 0 { // Avoid division by zero
				r.log.Warn("BM25 denominator is zero for term, skipping score contribution",
					zap.String("term", qTerm), zap.String("row_id", res.RowID))
				continue
			}

			termScore := idf * (numerator / denominator)
			docScore += termScore
		}
		res.Score = docScore
		r.log.Debug("Document scored", zap.String("row_id", res.RowID), zap.Float64("score", res.Score))
	}

	// 4. 对结果进行排序 (降序)
	sort.Slice(results, func(i, j int) bool {
		return results[i].Score > results[j].Score
	})

	r.log.Info("Results ranked successfully", zap.Int("ranked_count", len(results)))
	return nil
}

// ==============================================================================
// Redis IDF Provider 实现
// ==============================================================================

// redisIDFProvider 是 IDFProvider 接口的实现，从 Redis 获取 IDF 值。
// 假设 Redis 中 IDF 值存储的键格式为 `idf:term`，值为 `doc_freq:total_docs`。
type redisIDFProvider struct {
	cacheClient cache.CacheClient // Redis 客户端
	log         logger.Logger
}

// NewRedisIDFProvider 创建并返回一个新的 Redis IDF Provider 实例。
func NewRedisIDFProvider(cacheClient cache.CacheClient, log logger.Logger) (IDFProvider, error) {
	if cacheClient == nil {
		return nil, ferrors.NewInternalError("cache client is nil", nil)
	}
	if log == nil {
		return nil, ferrors.NewInternalError("logger is nil", nil)
	}
	log.Info("Redis IDF Provider initialized.")
	return &redisIDFProvider{
		cacheClient: cacheClient,
		log:         log,
	}, nil
}

// GetIDF 方法从 Redis 获取指定词项的 IDF 统计。
func (r *redisIDFProvider) GetIDF(ctx context.Context, term string) (*IDFStats, error) {
	key := fmt.Sprintf("idf:%s", term)

	// 从 Redis 获取字符串值
	val, err := r.cacheClient.Get(ctx, key)
	if err != nil {
		// 如果是 Key Not Found 错误，返回 nil IDFStats 和 nil 错误
		if ferrors.IsNotFoundError(err) {
			r.log.Debug("IDF not found for term in Redis", zap.String("term", term), zap.String("key", key))
			return &IDFStats{Term: term, DocumentFrequency: 0, TotalDocuments: 0}, nil // Return zero-stats for not found
		}
		r.log.Error("Failed to get IDF from Redis", zap.String("key", key), zap.Error(err))
		return nil, ferrors.NewExternalServiceError(fmt.Sprintf("failed to get IDF for term '%s' from cache", term), err)
	}

	// 解析存储的格式: "documentFrequency:totalDocuments"
	parts := strings.Split(val, ":")
	if len(parts) != 2 {
		r.log.Warn("Invalid IDF data format in Redis", zap.String("key", key), zap.String("value", val))
		return nil, ferrors.NewInternalError(fmt.Sprintf("invalid IDF data format for term '%s'", term), nil)
	}

	df, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		r.log.Warn("Failed to parse document frequency from Redis", zap.String("key", key), zap.String("value", val), zap.Error(err))
		return nil, ferrors.NewInternalError(fmt.Sprintf("failed to parse document frequency for term '%s'", term), err)
	}

	totalDocs, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil {
		r.log.Warn("Failed to parse total documents from Redis", zap.String("key", key), zap.String("value", val), zap.Error(err))
		return nil, ferrors.NewInternalError(fmt.Sprintf("failed to parse total documents for term '%s'", term), err)
	}

	stats := &IDFStats{
		Term:              term,
		DocumentFrequency: df,
		TotalDocuments:    totalDocs,
	}
	stats.InverseDocumentFrequency = stats.CalculateIDF() // Calculate IDF using the helper method

	r.log.Debug("Retrieved IDF for term",
		zap.String("term", term),
		zap.Int64("df", stats.DocumentFrequency),
		zap.Int64("total_docs", stats.TotalDocuments),
		zap.Float64("idf", stats.InverseDocumentFrequency))

	return stats, nil
}
