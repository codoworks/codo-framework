package pagination_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/codoworks/codo-framework/core/pagination"
)

type testItem struct {
	ID   string
	Name string
}

func TestListResponse(t *testing.T) {
	items := []testItem{
		{ID: "1", Name: "First"},
		{ID: "2", Name: "Second"},
	}

	params := &pagination.Params{
		Type:    pagination.TypeOffset,
		Page:    2,
		PerPage: 25,
	}

	response := pagination.ListResponse(items, 100, params)
	require.NotNil(t, response)
	assert.Len(t, response.Items, 2)
	assert.Equal(t, int64(100), response.Meta.Total)
	assert.Equal(t, 2, response.Meta.Page)
	assert.Equal(t, 25, response.Meta.PerPage)
	assert.Equal(t, 4, response.Meta.Pages) // 100 / 25 = 4
	assert.True(t, response.Meta.HasNext)
	assert.True(t, response.Meta.HasPrev)
}

func TestListResponse_NilParams(t *testing.T) {
	items := []testItem{{ID: "1", Name: "Test"}}

	response := pagination.ListResponse(items, 50, nil)
	require.NotNil(t, response)
	assert.Equal(t, 1, response.Meta.Page)
	assert.Equal(t, 20, response.Meta.PerPage) // Default
}

func TestListResponse_EmptyItems(t *testing.T) {
	var items []testItem

	params := &pagination.Params{Page: 1, PerPage: 20}
	response := pagination.ListResponse(items, 0, params)
	require.NotNil(t, response)
	assert.Empty(t, response.Items)
	assert.Equal(t, int64(0), response.Meta.Total)
}

func TestListResponseFromContext(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	params := &pagination.Params{
		Type:    pagination.TypeOffset,
		Page:    3,
		PerPage: 10,
	}
	pagination.Set(c, params)

	items := []testItem{{ID: "1", Name: "Test"}}
	response := pagination.ListResponseFromContext(c, items, 50)

	require.NotNil(t, response)
	assert.Equal(t, 3, response.Meta.Page)
	assert.Equal(t, 10, response.Meta.PerPage)
}

func TestCursorListResponse(t *testing.T) {
	items := []testItem{
		{ID: "1", Name: "First"},
		{ID: "2", Name: "Second"},
	}

	params := &pagination.Params{
		Type:    pagination.TypeCursor,
		PerPage: 25,
	}

	response := pagination.CursorListResponse(items, "cursor_next", "cursor_prev", true, params)
	require.NotNil(t, response)
	assert.Len(t, response.Items, 2)
	assert.Equal(t, "cursor_next", response.Meta.NextCursor)
	assert.Equal(t, "cursor_prev", response.Meta.PrevCursor)
	assert.True(t, response.Meta.HasMore)
	assert.Equal(t, 25, response.Meta.PerPage)
}

func TestCursorListResponse_NilParams(t *testing.T) {
	items := []testItem{{ID: "1", Name: "Test"}}

	response := pagination.CursorListResponse(items, "next", "", false, nil)
	require.NotNil(t, response)
	assert.Equal(t, 20, response.Meta.PerPage) // Default
}

func TestCursorListResponseFromContext(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	params := &pagination.Params{
		Type:    pagination.TypeCursor,
		PerPage: 30,
	}
	pagination.Set(c, params)

	items := []testItem{{ID: "1", Name: "Test"}}
	response := pagination.CursorListResponseFromContext(c, items, "next_cursor", "", true)

	require.NotNil(t, response)
	assert.Equal(t, 30, response.Meta.PerPage)
	assert.Equal(t, "next_cursor", response.Meta.NextCursor)
}

func TestNewCursorResult_HasMore(t *testing.T) {
	// Simulate N+1 fetch strategy: request 10, got 11
	items := make([]testItem, 11)
	for i := 0; i < 11; i++ {
		items[i] = testItem{ID: string(rune('a' + i)), Name: "Item"}
	}

	result := pagination.NewCursorResult(items, 10, func(item testItem) string {
		return "cursor_" + item.ID
	})

	assert.True(t, result.HasMore)
	assert.Len(t, result.Items, 10) // Trimmed to perPage
	assert.Equal(t, "cursor_j", result.NextCursor) // Last item in trimmed list
}

func TestNewCursorResult_NoMore(t *testing.T) {
	// Exact count: request 10, got 10
	items := make([]testItem, 10)
	for i := 0; i < 10; i++ {
		items[i] = testItem{ID: string(rune('a' + i)), Name: "Item"}
	}

	result := pagination.NewCursorResult(items, 10, func(item testItem) string {
		return "cursor_" + item.ID
	})

	assert.False(t, result.HasMore)
	assert.Len(t, result.Items, 10)
	assert.Equal(t, "cursor_j", result.NextCursor)
}

func TestNewCursorResult_LessThanPage(t *testing.T) {
	// Less than full page: request 10, got 5
	items := make([]testItem, 5)
	for i := 0; i < 5; i++ {
		items[i] = testItem{ID: string(rune('a' + i)), Name: "Item"}
	}

	result := pagination.NewCursorResult(items, 10, func(item testItem) string {
		return "cursor_" + item.ID
	})

	assert.False(t, result.HasMore)
	assert.Len(t, result.Items, 5)
	assert.Equal(t, "cursor_e", result.NextCursor)
}

func TestNewCursorResult_Empty(t *testing.T) {
	var items []testItem

	result := pagination.NewCursorResult(items, 10, func(item testItem) string {
		return "cursor_" + item.ID
	})

	assert.False(t, result.HasMore)
	assert.Empty(t, result.Items)
	assert.Empty(t, result.NextCursor)
}

func TestNewCursorResult_NilEncoder(t *testing.T) {
	items := []testItem{{ID: "1", Name: "Test"}}

	result := pagination.NewCursorResult(items, 10, nil)

	assert.False(t, result.HasMore)
	assert.Len(t, result.Items, 1)
	assert.Empty(t, result.NextCursor) // No encoder, no cursor
}

func TestCursorResult_ToCursorListResponse(t *testing.T) {
	items := []testItem{
		{ID: "1", Name: "First"},
		{ID: "2", Name: "Second"},
	}

	result := pagination.NewCursorResult(items, 10, func(item testItem) string {
		return "cursor_" + item.ID
	})

	response := result.ToCursorListResponse(10)
	require.NotNil(t, response)
	assert.Len(t, response.Items, 2)
	assert.Equal(t, 10, response.Meta.PerPage)
	assert.Equal(t, "cursor_2", response.Meta.NextCursor)
	assert.False(t, response.Meta.HasMore)
}
