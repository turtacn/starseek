package dto

import (
	"github.com/turtacn/starseek/internal/common/types"
	"github.com/turtacn/starseek/internal/domain/model"
)

// SearchReq defines the request payload for a search query.
type SearchReq struct {
	Query  string   `form:"q" binding:"required"`      // The search query string
	Fields []string `form:"fields" binding:"required"` // Comma-separated list of fields to search in
	types.PaginationReq
}

// SearchRes defines the response payload for a search query.
type SearchRes struct {
	types.PaginationRes[SearchItemRes] // Embed PaginationRes for common pagination fields
}

// SearchItemRes represents a single search result item returned to the client.
type SearchItemRes struct {
	ID         string                 `json:"id"`         // Unique identifier for the document (RowID)
	Score      float64                `json:"score"`      // Relevance score
	Data       map[string]interface{} `json:"data"`       // The original document data (selected fields)
	Highlights map[string]string      `json:"highlights"` // Highlighted snippets for matching fields
}

// FromModel converts a model.SearchResult to a SearchRes DTO.
func FromModel(result *model.SearchResult) *SearchRes {
	if result == nil {
		return &SearchRes{
			PaginationRes: types.PaginationRes[SearchItemRes]{
				Total:    0,
				Page:     1,
				PageSize: 0,
				Data:     []SearchItemRes{},
			},
		}
	}

	items := make([]SearchItemRes, len(result.Items))
	for i, item := range result.Items {
		items[i] = SearchItemRes{
			ID:         item.RowID,
			Score:      item.Score,
			Data:       item.OriginalData,
			Highlights: item.Highlights,
		}
	}

	return &SearchRes{
		PaginationRes: types.PaginationRes[SearchItemRes]{
			Total:    result.TotalHits,
			Page:     result.Page,
			PageSize: result.PageSize,
			Data:     items,
		},
	}
}

//Personal.AI order the ending
