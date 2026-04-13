package model

const (
	// DefaultPage is used when the page query parameter is omitted.
	DefaultPage = 1
	// DefaultLimit is used when the limit query parameter is omitted.
	DefaultLimit = 20
	// MaxLimit caps page size to protect the API from very large queries.
	MaxLimit = 100
)

// PaginationParams represents pagination input from query parameters.
type PaginationParams struct {
	Page  int
	Limit int
}

// Offset returns the SQL offset for the current page and limit.
func (p PaginationParams) Offset() int {
	return (p.Page - 1) * p.Limit
}

// PaginationMeta is returned in list responses.
type PaginationMeta struct {
	Page       int `json:"page"`
	Limit      int `json:"limit"`
	Total      int `json:"total"`
	TotalPages int `json:"total_pages"`
}

// NewPaginationMeta builds response metadata from paging parameters and total rows.
func NewPaginationMeta(params PaginationParams, total int) PaginationMeta {
	totalPages := 0
	if total > 0 {
		totalPages = (total + params.Limit - 1) / params.Limit
	}

	return PaginationMeta{
		Page:       params.Page,
		Limit:      params.Limit,
		Total:      total,
		TotalPages: totalPages,
	}
}
