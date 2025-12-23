package pagination

import (
	"github.com/labstack/echo/v4"

	"github.com/codoworks/codo-framework/core/forms"
)

// ListResponse creates a forms.ListResponse with pagination metadata.
// If params is nil, uses defaults (page=1, per_page=20).
func ListResponse[T any](items []T, total int64, params *Params) *forms.ListResponse[T] {
	if params == nil {
		return forms.NewListResponse(items, total, 1, 20)
	}
	return forms.NewListResponse(items, total, params.Page, params.PerPage)
}

// ListResponseFromContext creates a ListResponse using pagination from context.
// Convenience function that combines Get() and ListResponse().
func ListResponseFromContext[T any](c echo.Context, items []T, total int64) *forms.ListResponse[T] {
	return ListResponse(items, total, Get(c))
}

// CursorListResponse creates a forms.CursorListResponse.
// If params is nil, uses default per_page of 20.
func CursorListResponse[T any](items []T, nextCursor, prevCursor string, hasMore bool, params *Params) *forms.CursorListResponse[T] {
	perPage := 20
	if params != nil {
		perPage = params.PerPage
	}
	return forms.NewCursorListResponse(items, nextCursor, prevCursor, hasMore, perPage)
}

// CursorListResponseFromContext creates a CursorListResponse using pagination from context.
// Convenience function that combines Get() and CursorListResponse().
func CursorListResponseFromContext[T any](c echo.Context, items []T, nextCursor, prevCursor string, hasMore bool) *forms.CursorListResponse[T] {
	return CursorListResponse(items, nextCursor, prevCursor, hasMore, Get(c))
}

// CursorResult helps build cursor-based responses from query results.
// Use this when you've fetched N+1 items to detect HasMore.
type CursorResult[T any] struct {
	// Items contains the trimmed items (at most perPage)
	Items []T

	// NextCursor is the cursor for the next page (if any)
	NextCursor string

	// PrevCursor is the cursor for the previous page (if any)
	PrevCursor string

	// HasMore indicates if there are more items after this page
	HasMore bool
}

// NewCursorResult processes items fetched with the N+1 strategy.
// Pass the items fetched with Limit(perPage + 1).
// Returns the trimmed items and sets HasMore if there were extra items.
//
// The cursorEncoder function should convert an item to its cursor string.
// If nil, no NextCursor will be generated.
//
// Example:
//
//	items, _ := repo.FindAll(ctx, pagination.QueryOptions(c)...)
//	result := pagination.NewCursorResult(items, pg.PerPage, func(item *Model) string {
//	    return encodeCursor(item.ID, item.CreatedAt)
//	})
//	return c.Success(pagination.CursorListResponse(result.Items, result.NextCursor, "", result.HasMore, pg))
func NewCursorResult[T any](items []T, perPage int, cursorEncoder func(T) string) *CursorResult[T] {
	result := &CursorResult[T]{
		Items: items,
	}

	// Check if we have more items than requested
	if len(items) > perPage {
		result.HasMore = true
		result.Items = items[:perPage]
	}

	// Generate next cursor from last item
	if len(result.Items) > 0 && cursorEncoder != nil {
		result.NextCursor = cursorEncoder(result.Items[len(result.Items)-1])
	}

	return result
}

// ToCursorListResponse converts CursorResult to a forms.CursorListResponse.
func (r *CursorResult[T]) ToCursorListResponse(perPage int) *forms.CursorListResponse[T] {
	return forms.NewCursorListResponse(r.Items, r.NextCursor, r.PrevCursor, r.HasMore, perPage)
}
