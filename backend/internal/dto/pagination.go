package dto

// PaginationQuery holds common query parameters for paginated list endpoints.
type PaginationQuery struct {
	Page    int `json:"-"` // parsed from query param "page"
	PerPage int `json:"-"` // parsed from query param "per_page"
}

// Defaults applies sane defaults and clamps values.
func (q *PaginationQuery) Defaults() {
	if q.Page < 1 {
		q.Page = 1
	}
	if q.PerPage < 1 {
		q.PerPage = 20
	}
	if q.PerPage > 100 {
		q.PerPage = 100
	}
}

// Offset calculates the SQL OFFSET for the current page.
func (q *PaginationQuery) Offset() int {
	return (q.Page - 1) * q.PerPage
}

// PaginationMeta is returned with every paginated list response.
type PaginationMeta struct {
	Page       int `json:"page"`
	PerPage    int `json:"per_page"`
	Total      int `json:"total"`
	TotalPages int `json:"total_pages"`
}

// NewMeta builds a PaginationMeta from a query and a total count.
func NewMeta(q PaginationQuery, total int) PaginationMeta {
	totalPages := total / q.PerPage
	if total%q.PerPage != 0 {
		totalPages++
	}
	return PaginationMeta{
		Page:       q.Page,
		PerPage:    q.PerPage,
		Total:      total,
		TotalPages: totalPages,
	}
}

// ListResponse wraps a paginated list of items with metadata.
type ListResponse struct {
	Data interface{}    `json:"data"`
	Meta PaginationMeta `json:"meta"`
}
