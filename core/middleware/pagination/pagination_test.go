package pagination_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/codoworks/codo-framework/core/config"
	"github.com/codoworks/codo-framework/core/middleware"
	paginationmw "github.com/codoworks/codo-framework/core/middleware/pagination"
	"github.com/codoworks/codo-framework/core/pagination"
)

// Ensure the middleware is registered
var _ = paginationmw.PaginationMiddleware{}

func createMiddleware(t *testing.T, cfg *config.PaginationMiddlewareConfig) middleware.Middleware {
	mw := &paginationmw.PaginationMiddleware{
		BaseMiddleware: middleware.NewBaseMiddleware(
			"pagination",
			"middleware.pagination",
			middleware.PriorityPagination,
			middleware.RouterAll,
		),
	}

	err := mw.Configure(cfg)
	require.NoError(t, err)

	return mw
}

func TestPaginationMiddleware_DisabledByDefault(t *testing.T) {
	mw := &paginationmw.PaginationMiddleware{
		BaseMiddleware: middleware.NewBaseMiddleware(
			"pagination",
			"middleware.pagination",
			middleware.PriorityPagination,
			middleware.RouterAll,
		),
	}

	// Nil config - disabled by default
	assert.False(t, mw.Enabled(nil))

	// Empty config - disabled by default
	cfg := &config.PaginationMiddlewareConfig{}
	assert.False(t, mw.Enabled(cfg))

	// Explicitly enabled
	cfg.Enabled = true
	assert.True(t, mw.Enabled(cfg))
}

func TestPaginationMiddleware_SkipsNonGet(t *testing.T) {
	cfg := &config.PaginationMiddlewareConfig{
		BaseMiddlewareConfig: config.BaseMiddlewareConfig{Enabled: true},
		DefaultPageSize:      20,
		MaxPageSize:          100,
	}

	mw := createMiddleware(t, cfg)
	handler := mw.Handler()

	methods := []string{http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete}

	for _, method := range methods {
		t.Run(method, func(t *testing.T) {
			e := echo.New()
			req := httptest.NewRequest(method, "/test?page=2&per_page=25", nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			called := false
			err := handler(func(c echo.Context) error {
				called = true
				// Pagination should NOT be set for non-GET
				params := pagination.Get(c)
				assert.Nil(t, params)
				return nil
			})(c)

			assert.NoError(t, err)
			assert.True(t, called)
		})
	}
}

func TestPaginationMiddleware_ExtractsOffsetParams(t *testing.T) {
	cfg := &config.PaginationMiddlewareConfig{
		BaseMiddlewareConfig: config.BaseMiddlewareConfig{Enabled: true},
		DefaultPageSize:      20,
		MaxPageSize:          100,
		DefaultType:          "offset",
		ParamNames: config.PaginationParamNames{
			Page:    "page",
			PerPage: "per_page",
		},
	}

	mw := createMiddleware(t, cfg)
	handler := mw.Handler()

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/test?page=3&per_page=25", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	var params *pagination.Params
	err := handler(func(c echo.Context) error {
		params = pagination.Get(c)
		return nil
	})(c)

	assert.NoError(t, err)
	require.NotNil(t, params)
	assert.True(t, params.IsOffset())
	assert.Equal(t, 3, params.Page)
	assert.Equal(t, 25, params.PerPage)
	assert.Equal(t, 50, params.Offset) // (3-1) * 25
	assert.Equal(t, 100, params.MaxPerPage)
}

func TestPaginationMiddleware_ExtractsCursorParams(t *testing.T) {
	cfg := &config.PaginationMiddlewareConfig{
		BaseMiddlewareConfig: config.BaseMiddlewareConfig{Enabled: true},
		DefaultPageSize:      20,
		MaxPageSize:          100,
		DefaultType:          "offset",
		ParamNames: config.PaginationParamNames{
			Page:      "page",
			PerPage:   "per_page",
			Cursor:    "cursor",
			Direction: "direction",
		},
	}

	mw := createMiddleware(t, cfg)
	handler := mw.Handler()

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/test?cursor=abc123&direction=prev&per_page=30", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	var params *pagination.Params
	err := handler(func(c echo.Context) error {
		params = pagination.Get(c)
		return nil
	})(c)

	assert.NoError(t, err)
	require.NotNil(t, params)
	assert.True(t, params.IsCursor())
	assert.Equal(t, "abc123", params.Cursor)
	assert.Equal(t, pagination.DirectionPrev, params.Direction)
	assert.Equal(t, 30, params.PerPage)
}

func TestPaginationMiddleware_UsesDefaults(t *testing.T) {
	cfg := &config.PaginationMiddlewareConfig{
		BaseMiddlewareConfig: config.BaseMiddlewareConfig{Enabled: true},
		DefaultPageSize:      15,
		MaxPageSize:          50,
	}

	mw := createMiddleware(t, cfg)
	handler := mw.Handler()

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	var params *pagination.Params
	err := handler(func(c echo.Context) error {
		params = pagination.Get(c)
		return nil
	})(c)

	assert.NoError(t, err)
	require.NotNil(t, params)
	assert.Equal(t, 1, params.Page)
	assert.Equal(t, 15, params.PerPage)
	assert.Equal(t, 0, params.Offset)
	assert.Equal(t, 50, params.MaxPerPage)
}

func TestPaginationMiddleware_CapsPerPage(t *testing.T) {
	cfg := &config.PaginationMiddlewareConfig{
		BaseMiddlewareConfig: config.BaseMiddlewareConfig{Enabled: true},
		DefaultPageSize:      20,
		MaxPageSize:          50,
	}

	mw := createMiddleware(t, cfg)
	handler := mw.Handler()

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/test?per_page=200", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	var params *pagination.Params
	err := handler(func(c echo.Context) error {
		params = pagination.Get(c)
		return nil
	})(c)

	assert.NoError(t, err)
	require.NotNil(t, params)
	assert.Equal(t, 50, params.PerPage) // Capped at max
}

func TestPaginationMiddleware_InvalidParams(t *testing.T) {
	cfg := &config.PaginationMiddlewareConfig{
		BaseMiddlewareConfig: config.BaseMiddlewareConfig{Enabled: true},
		DefaultPageSize:      20,
		MaxPageSize:          100,
	}

	mw := createMiddleware(t, cfg)
	handler := mw.Handler()

	tests := []struct {
		name            string
		query           string
		expectedPage    int
		expectedPerPage int
	}{
		{
			name:            "negative page",
			query:           "?page=-1",
			expectedPage:    1,
			expectedPerPage: 20,
		},
		{
			name:            "zero page",
			query:           "?page=0",
			expectedPage:    1,
			expectedPerPage: 20,
		},
		{
			name:            "non-numeric page",
			query:           "?page=abc",
			expectedPage:    1,
			expectedPerPage: 20,
		},
		{
			name:            "negative per_page",
			query:           "?per_page=-10",
			expectedPage:    1,
			expectedPerPage: 20,
		},
		{
			name:            "zero per_page",
			query:           "?per_page=0",
			expectedPage:    1,
			expectedPerPage: 20,
		},
		{
			name:            "non-numeric per_page",
			query:           "?per_page=xyz",
			expectedPage:    1,
			expectedPerPage: 20,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := echo.New()
			req := httptest.NewRequest(http.MethodGet, "/test"+tt.query, nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			var params *pagination.Params
			err := handler(func(c echo.Context) error {
				params = pagination.Get(c)
				return nil
			})(c)

			assert.NoError(t, err)
			require.NotNil(t, params)
			assert.Equal(t, tt.expectedPage, params.Page)
			assert.Equal(t, tt.expectedPerPage, params.PerPage)
		})
	}
}

func TestPaginationMiddleware_CustomParamNames(t *testing.T) {
	cfg := &config.PaginationMiddlewareConfig{
		BaseMiddlewareConfig: config.BaseMiddlewareConfig{Enabled: true},
		DefaultPageSize:      20,
		MaxPageSize:          100,
		ParamNames: config.PaginationParamNames{
			Page:      "p",
			PerPage:   "limit",
			Cursor:    "after",
			Direction: "dir",
		},
	}

	mw := createMiddleware(t, cfg)
	handler := mw.Handler()

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/test?p=5&limit=30", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	var params *pagination.Params
	err := handler(func(c echo.Context) error {
		params = pagination.Get(c)
		return nil
	})(c)

	assert.NoError(t, err)
	require.NotNil(t, params)
	assert.Equal(t, 5, params.Page)
	assert.Equal(t, 30, params.PerPage)
}

func TestPaginationMiddleware_CursorDirectionDefault(t *testing.T) {
	cfg := &config.PaginationMiddlewareConfig{
		BaseMiddlewareConfig: config.BaseMiddlewareConfig{Enabled: true},
		DefaultPageSize:      20,
		MaxPageSize:          100,
		ParamNames: config.PaginationParamNames{
			Cursor:    "cursor",
			Direction: "direction",
		},
	}

	mw := createMiddleware(t, cfg)
	handler := mw.Handler()

	e := echo.New()
	// No direction specified - should default to "next"
	req := httptest.NewRequest(http.MethodGet, "/test?cursor=abc", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	var params *pagination.Params
	err := handler(func(c echo.Context) error {
		params = pagination.Get(c)
		return nil
	})(c)

	assert.NoError(t, err)
	require.NotNil(t, params)
	assert.Equal(t, pagination.DirectionNext, params.Direction)
}

func TestPaginationMiddleware_Priority(t *testing.T) {
	mw := &paginationmw.PaginationMiddleware{
		BaseMiddleware: middleware.NewBaseMiddleware(
			"pagination",
			"middleware.pagination",
			middleware.PriorityPagination,
			middleware.RouterAll,
		),
	}

	assert.Equal(t, middleware.PriorityPagination, mw.Priority())
	assert.Equal(t, 102, mw.Priority())
}

func TestPaginationMiddleware_Routers(t *testing.T) {
	mw := &paginationmw.PaginationMiddleware{
		BaseMiddleware: middleware.NewBaseMiddleware(
			"pagination",
			"middleware.pagination",
			middleware.PriorityPagination,
			middleware.RouterAll,
		),
	}

	assert.Equal(t, middleware.RouterAll, mw.Routers())
}

func TestPaginationMiddleware_RawValues(t *testing.T) {
	cfg := &config.PaginationMiddlewareConfig{
		BaseMiddlewareConfig: config.BaseMiddlewareConfig{Enabled: true},
		DefaultPageSize:      20,
		MaxPageSize:          100,
	}

	mw := createMiddleware(t, cfg)
	handler := mw.Handler()

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/test?page=3&per_page=25", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	var params *pagination.Params
	err := handler(func(c echo.Context) error {
		params = pagination.Get(c)
		return nil
	})(c)

	assert.NoError(t, err)
	require.NotNil(t, params)
	assert.Equal(t, "3", params.RawPage)
	assert.Equal(t, "25", params.RawPerPage)
}
