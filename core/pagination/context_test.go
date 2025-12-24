package pagination_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/codoworks/codo-framework/core/pagination"
)

func TestContextWithParams(t *testing.T) {
	params := &pagination.Params{
		Type:    pagination.TypeOffset,
		Page:    2,
		PerPage: 25,
		Offset:  25,
	}

	ctx := pagination.ContextWithParams(context.Background(), params)

	retrieved, ok := pagination.ParamsFromContext(ctx)
	require.True(t, ok)
	assert.Equal(t, params, retrieved)
}

func TestParamsFromContext_NotSet(t *testing.T) {
	ctx := context.Background()

	retrieved, ok := pagination.ParamsFromContext(ctx)
	assert.False(t, ok)
	assert.Nil(t, retrieved)
}

func TestSetAndGet(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	params := &pagination.Params{
		Type:    pagination.TypeOffset,
		Page:    3,
		PerPage: 50,
		Offset:  100,
	}

	pagination.Set(c, params)

	retrieved := pagination.Get(c)
	require.NotNil(t, retrieved)
	assert.Equal(t, params.Page, retrieved.Page)
	assert.Equal(t, params.PerPage, retrieved.PerPage)
	assert.Equal(t, params.Offset, retrieved.Offset)
}

func TestGet_NotSet(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	retrieved := pagination.Get(c)
	assert.Nil(t, retrieved)
}

func TestMustGet_Panics(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	assert.PanicsWithError(t, pagination.ErrNoPagination.Error(), func() {
		pagination.MustGet(c)
	})
}

func TestMustGet_Success(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	params := &pagination.Params{
		Type:    pagination.TypeOffset,
		Page:    1,
		PerPage: 20,
	}
	pagination.Set(c, params)

	assert.NotPanics(t, func() {
		retrieved := pagination.MustGet(c)
		assert.Equal(t, params, retrieved)
	})
}

func TestGetOrDefault_WithMiddleware(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	params := &pagination.Params{
		Type:    pagination.TypeOffset,
		Page:    5,
		PerPage: 30,
	}
	pagination.Set(c, params)

	retrieved := pagination.GetOrDefault(c, 1, 20)
	assert.Equal(t, 5, retrieved.Page)
	assert.Equal(t, 30, retrieved.PerPage)
}

func TestGetOrDefault_WithoutMiddleware(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/test?page=3&per_page=50", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	retrieved := pagination.GetOrDefault(c, 1, 20)
	assert.Equal(t, 3, retrieved.Page)
	assert.Equal(t, 50, retrieved.PerPage)
}

func TestGetOrDefault_UsesDefaults(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	retrieved := pagination.GetOrDefault(c, 1, 25)
	assert.Equal(t, 1, retrieved.Page)
	assert.Equal(t, 25, retrieved.PerPage)
}

func TestGetOrDefault_CapsAtMaxPerPage(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/test?per_page=500", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	retrieved := pagination.GetOrDefault(c, 1, 20)
	assert.Equal(t, 100, retrieved.PerPage) // Default max is 100
}

func TestGetOrDefaultWithMax(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/test?per_page=500", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	retrieved := pagination.GetOrDefaultWithMax(c, 1, 20, 200)
	assert.Equal(t, 200, retrieved.PerPage) // Custom max is 200
	assert.Equal(t, 200, retrieved.MaxPerPage)
}

func TestGetOrDefault_InvalidParams(t *testing.T) {
	tests := []struct {
		name          string
		query         string
		expectedPage  int
		expectedLimit int
	}{
		{
			name:          "negative page",
			query:         "?page=-1&per_page=20",
			expectedPage:  1,
			expectedLimit: 20,
		},
		{
			name:          "zero page",
			query:         "?page=0&per_page=20",
			expectedPage:  1,
			expectedLimit: 20,
		},
		{
			name:          "negative per_page",
			query:         "?page=1&per_page=-10",
			expectedPage:  1,
			expectedLimit: 20,
		},
		{
			name:          "non-numeric page",
			query:         "?page=abc&per_page=20",
			expectedPage:  1,
			expectedLimit: 20,
		},
		{
			name:          "non-numeric per_page",
			query:         "?page=1&per_page=xyz",
			expectedPage:  1,
			expectedLimit: 20,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := echo.New()
			req := httptest.NewRequest(http.MethodGet, "/test"+tt.query, nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			retrieved := pagination.GetOrDefault(c, 1, 20)
			assert.Equal(t, tt.expectedPage, retrieved.Page)
			assert.Equal(t, tt.expectedLimit, retrieved.PerPage)
		})
	}
}

func TestSetMeta(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Set pagination params first
	params := &pagination.Params{
		Type:    pagination.TypeOffset,
		Page:    2,
		PerPage: 25,
		Offset:  25,
	}
	pagination.Set(c, params)

	// Now set meta with total
	pagination.SetMeta(c, 100)

	// Retrieve and verify
	meta := pagination.GetMeta(c)
	require.NotNil(t, meta)
	assert.Equal(t, int64(100), meta.Total)
	assert.Equal(t, 2, meta.Page)
	assert.Equal(t, 25, meta.PerPage)
	assert.Equal(t, 4, meta.Pages) // 100 / 25 = 4
	assert.True(t, meta.HasNext)   // Page 2 < 4
	assert.True(t, meta.HasPrev)   // Page 2 > 1
}

func TestSetMeta_WithoutParams(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// SetMeta without params set should be a no-op
	pagination.SetMeta(c, 100)

	meta := pagination.GetMeta(c)
	assert.Nil(t, meta)
}

func TestSetMeta_FirstPage(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	params := &pagination.Params{
		Type:    pagination.TypeOffset,
		Page:    1,
		PerPage: 20,
	}
	pagination.Set(c, params)
	pagination.SetMeta(c, 50)

	meta := pagination.GetMeta(c)
	require.NotNil(t, meta)
	assert.Equal(t, 1, meta.Page)
	assert.Equal(t, 3, meta.Pages) // 50 / 20 = 2.5 → 3
	assert.True(t, meta.HasNext)
	assert.False(t, meta.HasPrev) // First page
}

func TestSetMeta_LastPage(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	params := &pagination.Params{
		Type:    pagination.TypeOffset,
		Page:    5,
		PerPage: 20,
	}
	pagination.Set(c, params)
	pagination.SetMeta(c, 100)

	meta := pagination.GetMeta(c)
	require.NotNil(t, meta)
	assert.Equal(t, 5, meta.Page)
	assert.Equal(t, 5, meta.Pages) // 100 / 20 = 5
	assert.False(t, meta.HasNext) // Last page
	assert.True(t, meta.HasPrev)
}

func TestSetMeta_EmptyResult(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	params := &pagination.Params{
		Type:    pagination.TypeOffset,
		Page:    1,
		PerPage: 20,
	}
	pagination.Set(c, params)
	pagination.SetMeta(c, 0)

	meta := pagination.GetMeta(c)
	require.NotNil(t, meta)
	assert.Equal(t, int64(0), meta.Total)
	assert.Equal(t, 1, meta.Pages)  // At least 1 page
	assert.False(t, meta.HasNext)
	assert.False(t, meta.HasPrev)
}

func TestGetMeta_NotSet(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	meta := pagination.GetMeta(c)
	assert.Nil(t, meta)
}
