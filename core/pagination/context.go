package pagination

import (
	"context"
	"fmt"
	"math"
	"strconv"

	"github.com/labstack/echo/v4"

	"github.com/codoworks/codo-framework/core/forms"
)

// Context keys
type contextKey string

const (
	paramsKey contextKey = "pagination"
	metaKey   contextKey = "pagination_meta"
)

// ErrNoPagination is returned when no pagination params are in context
var ErrNoPagination = fmt.Errorf("no pagination in context")

// ContextWithParams adds pagination params to the context
func ContextWithParams(ctx context.Context, params *Params) context.Context {
	return context.WithValue(ctx, paramsKey, params)
}

// ParamsFromContext retrieves pagination params from context
func ParamsFromContext(ctx context.Context) (*Params, bool) {
	params, ok := ctx.Value(paramsKey).(*Params)
	return params, ok
}

// Set stores pagination params in the Echo context
func Set(c echo.Context, params *Params) {
	ctx := ContextWithParams(c.Request().Context(), params)
	c.SetRequest(c.Request().WithContext(ctx))
}

// Get retrieves pagination params from the Echo context.
// Returns nil if pagination middleware is not enabled or not applied to this request.
func Get(c echo.Context) *Params {
	params, _ := ParamsFromContext(c.Request().Context())
	return params
}

// MustGet retrieves pagination params or panics if not present.
// Use this only when you're certain the pagination middleware is enabled.
func MustGet(c echo.Context) *Params {
	params := Get(c)
	if params == nil {
		panic(ErrNoPagination)
	}
	return params
}

// GetOrDefault retrieves pagination params, or extracts them from query params
// if not set by middleware. This is useful when the middleware is disabled
// but the handler still wants pagination support.
func GetOrDefault(c echo.Context, defaultPage, defaultPerPage int) *Params {
	if params := Get(c); params != nil {
		return params
	}

	// Build from query params manually with default config
	return parseFromQuery(c, defaultPage, defaultPerPage, 100, "page", "per_page")
}

// GetOrDefaultWithMax is like GetOrDefault but allows specifying a custom max per page
func GetOrDefaultWithMax(c echo.Context, defaultPage, defaultPerPage, maxPerPage int) *Params {
	if params := Get(c); params != nil {
		return params
	}

	return parseFromQuery(c, defaultPage, defaultPerPage, maxPerPage, "page", "per_page")
}

// parseFromQuery extracts pagination from query params (internal helper)
func parseFromQuery(c echo.Context, defaultPage, defaultPerPage, maxPerPage int, pageParam, perPageParam string) *Params {
	page := queryInt(c, pageParam, defaultPage)
	perPage := queryInt(c, perPageParam, defaultPerPage)

	// Normalize
	if page < 1 {
		page = 1
	}
	if perPage < 1 {
		perPage = defaultPerPage
	}
	if perPage > maxPerPage {
		perPage = maxPerPage
	}

	return &Params{
		Type:       TypeOffset,
		Page:       page,
		PerPage:    perPage,
		Offset:     (page - 1) * perPage,
		MaxPerPage: maxPerPage,
		RawPage:    c.QueryParam(pageParam),
		RawPerPage: c.QueryParam(perPageParam),
	}
}

// queryInt extracts an integer from query params with a default value
func queryInt(c echo.Context, name string, defaultVal int) int {
	val := c.QueryParam(name)
	if val == "" {
		return defaultVal
	}
	result, err := strconv.Atoi(val)
	if err != nil {
		return defaultVal
	}
	return result
}

// SetMeta stores computed pagination metadata in context for response inclusion.
// Call this from handlers after querying total count to populate the Page field
// in the standard response.
//
// Example:
//
//	func (h *Handler) List(c *http.Context) error {
//	    pg := pagination.Get(c)
//	    items, _ := h.service.FindAll(ctx, pg.QueryOptions()...)
//	    total, _ := h.service.Count(ctx)
//	    pagination.SetMeta(c, total)  // Sets Page metadata
//	    return c.Success(items)       // Response includes Page field
//	}
func SetMeta(c echo.Context, total int64) {
	params := Get(c)
	if params == nil {
		return
	}

	perPage := params.PerPage
	if perPage <= 0 {
		perPage = 20
	}

	pages := int(math.Ceil(float64(total) / float64(perPage)))
	if pages == 0 {
		pages = 1
	}

	meta := &forms.ListMeta{
		Total:   total,
		Page:    params.Page,
		PerPage: perPage,
		Pages:   pages,
		HasNext: params.Page < pages,
		HasPrev: params.Page > 1,
	}

	// Store meta in context for response methods to pick up
	ctx := context.WithValue(c.Request().Context(), metaKey, meta)
	c.SetRequest(c.Request().WithContext(ctx))
}

// GetMeta retrieves computed pagination metadata from context.
// Returns nil if SetMeta was not called or if pagination middleware is not enabled.
func GetMeta(c echo.Context) *forms.ListMeta {
	meta, _ := c.Request().Context().Value(metaKey).(*forms.ListMeta)
	return meta
}
