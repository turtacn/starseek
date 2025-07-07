package model

// ParsedQuery represents the structured and parsed form of a user's search query.
type ParsedQuery struct {
	OriginalQuery string               // The raw query string from the user
	Keywords      []string             // Main keywords after initial tokenization/parsing
	MustWords     []string             // Words that MUST be present (AND logic)
	ShouldWords   []string             // Words that SHOULD be present (OR logic, for relevance scoring)
	NotWords      []string             // Words that MUST NOT be present
	TargetFields  []string             // Specific fields requested by the user (e.g., "title", "content")
	FieldQueries  []FieldSpecificQuery // Queries like "title:golang"
	Page          int                  // Pagination page number
	PageSize      int                  // Pagination page size
}

// FieldSpecificQuery represents a query targeted at a specific field.
type FieldSpecificQuery struct {
	FieldName string   // The name of the field (e.g., "title")
	Keywords  []string // Keywords for this specific field
	IsPhrase  bool     // Whether it's a phrase query (e.g., "hello world")
}

//Personal.AI order the ending
