package pagination

import (
	"fmt"
	"math"
	"strconv"
)

// PageRequest represents pagination parameters
type PageRequest struct {
	Page     int    `json:"page"`
	PageSize int    `json:"page_size"`
	SortBy   string `json:"sort_by"`
	SortDir  string `json:"sort_dir"` // "asc" or "desc"
}

// PageResponse represents paginated response
type PageResponse[T any] struct {
	Data       []T   `json:"data"`
	Page       int   `json:"page"`
	PageSize   int   `json:"page_size"`
	Total      int64 `json:"total"`
	TotalPages int   `json:"total_pages"`
	HasNext    bool  `json:"has_next"`
	HasPrev    bool  `json:"has_prev"`
}

// DefaultPageSize is the default number of items per page
const DefaultPageSize = 20

// MaxPageSize is the maximum allowed page size
const MaxPageSize = 100

// NewPageRequest creates a new page request with defaults
func NewPageRequest(page, pageSize int, sortBy, sortDir string) PageRequest {
	// Set defaults
	if page < 1 {
		page = 1
	}

	if pageSize < 1 {
		pageSize = DefaultPageSize
	}

	if pageSize > MaxPageSize {
		pageSize = MaxPageSize
	}

	if sortDir != "asc" && sortDir != "desc" {
		sortDir = "desc"
	}

	return PageRequest{
		Page:     page,
		PageSize: pageSize,
		SortBy:   sortBy,
		SortDir:  sortDir,
	}
}

// ParsePageRequest parses page request from query parameters
func ParsePageRequest(pageStr, pageSizeStr, sortBy, sortDir string) PageRequest {
	page, _ := strconv.Atoi(pageStr)
	pageSize, _ := strconv.Atoi(pageSizeStr)

	return NewPageRequest(page, pageSize, sortBy, sortDir)
}

// Offset calculates the offset for SQL queries
func (pr PageRequest) Offset() int {
	return (pr.Page - 1) * pr.PageSize
}

// Limit returns the page size (alias for clarity)
func (pr PageRequest) Limit() int {
	return pr.PageSize
}

// NewPageResponse creates a new paginated response
func NewPageResponse[T any](data []T, page, pageSize int, total int64) PageResponse[T] {
	totalPages := int(math.Ceil(float64(total) / float64(pageSize)))

	return PageResponse[T]{
		Data:       data,
		Page:       page,
		PageSize:   pageSize,
		Total:      total,
		TotalPages: totalPages,
		HasNext:    page < totalPages,
		HasPrev:    page > 1,
	}
}

// GetSortClause generates SQL ORDER BY clause
func (pr PageRequest) GetSortClause() string {
	if pr.SortBy == "" {
		return ""
	}

	return fmt.Sprintf("ORDER BY %s %s", pr.SortBy, pr.SortDir)
}

// GetLimitOffset generates SQL LIMIT and OFFSET clause
func (pr PageRequest) GetLimitOffset() string {
	return fmt.Sprintf("LIMIT %d OFFSET %d", pr.Limit(), pr.Offset())
}
