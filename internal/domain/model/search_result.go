package model

import "encoding/json"

// SearchResultItem represents a single item found in the search results.
type SearchResultItem struct {
	RowID        string                 `json:"row_id"`        // Unique identifier for the row in the underlying database
	Score        float64                `json:"score"`         // Relevance score of the item
	OriginalData map[string]interface{} `json:"original_data"` // The full original data for the row, typically retrieved later
	Highlights   map[string]string      `json:"highlights"`    // Highlighted snippets for matching fields
}

// SearchResult represents the complete search result, including items and metadata.
type SearchResult struct {
	Items     []*SearchResultItem `json:"items"`      // List of search result items
	TotalHits int64               `json:"total_hits"` // Total number of matching documents
	Page      int                 `json:"page"`
	PageSize  int                 `json:"page_size"`
}

// ConvertToMap converts the SearchResultItem to a map, handling JSON marshal for OriginalData.
func (s *SearchResultItem) ConvertToMap() (map[string]interface{}, error) {
	data, err := json.Marshal(s)
	if err != nil {
		return nil, err
	}
	var m map[string]interface{}
	err = json.Unmarshal(data, &m)
	return m, err
}

//Personal.AI order the ending
