package forms

import (
	"math"
)

// ListMeta contains pagination metadata
type ListMeta struct {
	Total   int64 `json:"total"`
	Page    int   `json:"page"`
	PerPage int   `json:"per_page"`
	Pages   int   `json:"pages"`
	HasNext bool  `json:"has_next"`
	HasPrev bool  `json:"has_prev"`
}

// NewListMeta creates a ListMeta and calculates derived fields
func NewListMeta(total int64, page, perPage int) ListMeta {
	if perPage <= 0 {
		perPage = 20
	}
	if page <= 0 {
		page = 1
	}

	pages := int(math.Ceil(float64(total) / float64(perPage)))
	if pages == 0 {
		pages = 1
	}

	return ListMeta{
		Total:   total,
		Page:    page,
		PerPage: perPage,
		Pages:   pages,
		HasNext: page < pages,
		HasPrev: page > 1,
	}
}

// ListResponse wraps a slice of items with pagination metadata
type ListResponse[T any] struct {
	Items []T      `json:"items"`
	Meta  ListMeta `json:"meta"`
}

// NewListResponse creates a ListResponse with calculated metadata
func NewListResponse[T any](items []T, total int64, page, perPage int) *ListResponse[T] {
	if items == nil {
		items = make([]T, 0)
	}

	return &ListResponse[T]{
		Items: items,
		Meta:  NewListMeta(total, page, perPage),
	}
}

// Empty returns true if the list has no items
func (r *ListResponse[T]) Empty() bool {
	return len(r.Items) == 0
}

// Count returns the number of items in the current page
func (r *ListResponse[T]) Count() int {
	return len(r.Items)
}

// Offset returns the offset for the current page
func (r *ListResponse[T]) Offset() int {
	return (r.Meta.Page - 1) * r.Meta.PerPage
}

// MapItems transforms items using a mapper function
func MapItems[T any, U any](items []T, mapper func(T) U) []U {
	result := make([]U, len(items))
	for i, item := range items {
		result[i] = mapper(item)
	}
	return result
}

// MapListResponse transforms a ListResponse to a different type
func MapListResponse[T any, U any](list *ListResponse[T], mapper func(T) U) *ListResponse[U] {
	return &ListResponse[U]{
		Items: MapItems(list.Items, mapper),
		Meta:  list.Meta,
	}
}

// PaginationParams holds pagination query parameters
type PaginationParams struct {
	Page    int `query:"page"`
	PerPage int `query:"per_page"`
}

// Normalize ensures pagination params have valid values
func (p *PaginationParams) Normalize() {
	if p.Page <= 0 {
		p.Page = 1
	}
	if p.PerPage <= 0 {
		p.PerPage = 20
	}
	if p.PerPage > 100 {
		p.PerPage = 100
	}
}

// ToLimitOffset converts to LIMIT/OFFSET values
func (p *PaginationParams) ToLimitOffset() (limit, offset int) {
	p.Normalize()
	return p.PerPage, (p.Page - 1) * p.PerPage
}

// CursorMeta contains cursor-based pagination metadata
type CursorMeta struct {
	NextCursor string `json:"next_cursor,omitempty"`
	PrevCursor string `json:"prev_cursor,omitempty"`
	HasMore    bool   `json:"has_more"`
	PerPage    int    `json:"per_page"`
}

// CursorListResponse wraps items with cursor-based pagination
type CursorListResponse[T any] struct {
	Items []T        `json:"items"`
	Meta  CursorMeta `json:"meta"`
}

// NewCursorListResponse creates a cursor-based list response
func NewCursorListResponse[T any](items []T, nextCursor, prevCursor string, hasMore bool, perPage int) *CursorListResponse[T] {
	if items == nil {
		items = make([]T, 0)
	}

	return &CursorListResponse[T]{
		Items: items,
		Meta: CursorMeta{
			NextCursor: nextCursor,
			PrevCursor: prevCursor,
			HasMore:    hasMore,
			PerPage:    perPage,
		},
	}
}
