package service

import (
	"github.com/turtacn/starseek/internal/domain/model"
	"go.uber.org/zap"
	"strings"
)

// RankingService is a domain service responsible for scoring and re-ranking search results.
type RankingService struct {
	logger *zap.Logger
}

// NewRankingService creates a new RankingService instance.
func NewRankingService(logger *zap.Logger) *RankingService {
	return &RankingService{
		logger: logger.Named("RankingService"),
	}
}

// Rank reorders search results based on relevance.
//
// Initial implementation:
// 1. Prioritize items where all 'MustWords' are present.
// 2. Then, sort by the number of 'Keywords' matched.
// 3. Fallback to a simple alphabetical sort or original order.
//
// Advanced implementation: TF-IDF, BM25, or custom scoring algorithms.
// This would require more context about the actual content of `OriginalData`
// and potentially pre-computed IDF values or a full-text analysis pipeline.
func (rs *RankingService) Rank(results []*model.SearchResultItem, query *model.ParsedQuery) []*model.SearchResultItem {
	if len(results) == 0 {
		return results
	}

	// For simplicity, this example does not implement complex TF-IDF.
	// It performs a basic score based on keyword presence.

	// Step 1: Assign a preliminary score based on keyword matches
	for _, item := range results {
		item.Score = rs.calculateBasicScore(item.OriginalData, query)
	}

	// Step 2: Sort the results
	// Using a bubble sort for clarity in this example; for performance, use sort.Slice
	// Sort by score (descending), then RowID (ascending) for stable sort.
	for i := 0; i < len(results)-1; i++ {
		for j := 0; j < len(results)-i-1; j++ {
			if results[j].Score < results[j+1].Score {
				results[j], results[j+1] = results[j+1], results[j]
			} else if results[j].Score == results[j+1].Score && results[j].RowID > results[j+1].RowID {
				results[j], results[j+1] = results[j+1], results[j]
			}
		}
	}

	rs.logger.Debug("Results ranked", zap.Int("count", len(results)))
	return results
}

// calculateBasicScore calculates a simple score for an item based on query matches.
// This is a placeholder for actual relevance scoring.
func (rs *RankingService) calculateBasicScore(data map[string]interface{}, query *model.ParsedQuery) float64 {
	score := 0.0

	combinedText := ""
	for _, fieldName := range query.TargetFields {
		if val, ok := data[fieldName]; ok {
			if strVal, isStr := val.(string); isStr {
				combinedText += strVal + " "
			}
		}
	}
	combinedText = strings.ToLower(combinedText)

	// Boost for MustWords
	allMustWordsFound := true
	for _, mustWord := range query.MustWords {
		if !strings.Contains(combinedText, strings.ToLower(mustWord)) {
			allMustWordsFound = false
			break
		}
	}
	if allMustWordsFound && len(query.MustWords) > 0 {
		score += 100.0 // Significant boost for all must words
	}

	// Score for Keywords
	matchedKeywords := 0
	for _, keyword := range query.Keywords {
		if strings.Contains(combinedText, strings.ToLower(keyword)) {
			matchedKeywords++
		}
	}
	score += float64(matchedKeywords * 10) // Moderate boost per keyword

	// Score for ShouldWords (less important than Keywords)
	matchedShouldWords := 0
	for _, shouldWord := range query.ShouldWords {
		if strings.Contains(combinedText, strings.ToLower(shouldWord)) {
			matchedShouldWords++
		}
	}
	score += float64(matchedShouldWords * 5) // Small boost per should word

	// Penalize for NotWords (simple negative match)
	for _, notWord := range query.NotWords {
		if strings.Contains(combinedText, strings.ToLower(notWord)) {
			score -= 50.0 // Significant penalty
			break
		}
	}

	// Score for Field-Specific Queries
	for _, fq := range query.FieldQueries {
		if val, ok := data[fq.FieldName]; ok {
			if strVal, isStr := val.(string); isStr {
				fieldContent := strings.ToLower(strVal)
				for _, keyword := range fq.Keywords {
					if strings.Contains(fieldContent, strings.ToLower(keyword)) {
						score += 20.0 // Boost for field-specific matches
					}
				}
				if fq.IsPhrase && strings.Contains(fieldContent, strings.ToLower(strings.Join(fq.Keywords, " "))) {
					score += 30.0 // Extra boost for phrase match
				}
			}
		}
	}

	return score
}

//Personal.AI order the ending
