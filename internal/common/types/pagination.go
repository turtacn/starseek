package types

// PaginationReq defines the common structure for pagination requests.
type PaginationReq struct {
	Page     int `json:"page" form:"page" binding:"gte=1"`           // Current page number, starts from 1
	PageSize int `json:"page_size" form:"page_size" binding:"gte=1"` // Number of items per page
}

// PaginationRes defines the common structure for pagination responses.
// T is a generic type for the data slice.
type PaginationRes[T any] struct {
	Total    int64 `json:"total"`     // Total number of items
	Page     int   `json:"page"`      // Current page number
	PageSize int   `json:"page_size"` // Number of items per page
	Data     []T   `json:"data"`      // Slice of data items for the current page
}

//Personal.AI order the ending
