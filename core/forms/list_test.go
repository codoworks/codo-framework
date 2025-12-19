package forms_test

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/codoworks/codo-framework/core/forms"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewListMeta(t *testing.T) {
	meta := forms.NewListMeta(100, 2, 10)

	assert.Equal(t, int64(100), meta.Total)
	assert.Equal(t, 2, meta.Page)
	assert.Equal(t, 10, meta.PerPage)
	assert.Equal(t, 10, meta.Pages)
	assert.True(t, meta.HasNext)
	assert.True(t, meta.HasPrev)
}

func TestNewListMeta_DefaultPerPage(t *testing.T) {
	meta := forms.NewListMeta(100, 1, 0)

	assert.Equal(t, 20, meta.PerPage)
	assert.Equal(t, 5, meta.Pages)
}

func TestNewListMeta_NegativePerPage(t *testing.T) {
	meta := forms.NewListMeta(100, 1, -5)

	assert.Equal(t, 20, meta.PerPage)
}

func TestNewListMeta_DefaultPage(t *testing.T) {
	meta := forms.NewListMeta(100, 0, 10)

	assert.Equal(t, 1, meta.Page)
}

func TestNewListMeta_NegativePage(t *testing.T) {
	meta := forms.NewListMeta(100, -5, 10)

	assert.Equal(t, 1, meta.Page)
}

func TestNewListMeta_Pages(t *testing.T) {
	tests := []struct {
		total       int64
		perPage     int
		expectedPages int
	}{
		{100, 10, 10},
		{101, 10, 11},
		{99, 10, 10},
		{10, 10, 1},
		{1, 10, 1},
		{0, 10, 1},
		{50, 20, 3},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("total=%d,perPage=%d", tt.total, tt.perPage), func(t *testing.T) {
			meta := forms.NewListMeta(tt.total, 1, tt.perPage)
			assert.Equal(t, tt.expectedPages, meta.Pages)
		})
	}
}

func TestNewListMeta_HasNext(t *testing.T) {
	tests := []struct {
		total    int64
		page     int
		perPage  int
		expected bool
	}{
		{100, 1, 10, true},
		{100, 5, 10, true},
		{100, 10, 10, false},
		{100, 11, 10, false},
		{10, 1, 10, false},
		{0, 1, 10, false},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("total=%d,page=%d", tt.total, tt.page), func(t *testing.T) {
			meta := forms.NewListMeta(tt.total, tt.page, tt.perPage)
			assert.Equal(t, tt.expected, meta.HasNext)
		})
	}
}

func TestNewListMeta_HasPrev(t *testing.T) {
	tests := []struct {
		page     int
		expected bool
	}{
		{1, false},
		{2, true},
		{10, true},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("page=%d", tt.page), func(t *testing.T) {
			meta := forms.NewListMeta(100, tt.page, 10)
			assert.Equal(t, tt.expected, meta.HasPrev)
		})
	}
}

func TestNewListMeta_SinglePage(t *testing.T) {
	meta := forms.NewListMeta(5, 1, 10)

	assert.Equal(t, 1, meta.Pages)
	assert.False(t, meta.HasNext)
	assert.False(t, meta.HasPrev)
}

func TestNewListMeta_Empty(t *testing.T) {
	meta := forms.NewListMeta(0, 1, 10)

	assert.Equal(t, int64(0), meta.Total)
	assert.Equal(t, 1, meta.Pages)
	assert.False(t, meta.HasNext)
	assert.False(t, meta.HasPrev)
}

func TestNewListMeta_JSON(t *testing.T) {
	meta := forms.NewListMeta(100, 2, 10)

	data, err := json.Marshal(meta)
	require.NoError(t, err)

	var decoded forms.ListMeta
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)

	assert.Equal(t, meta.Total, decoded.Total)
	assert.Equal(t, meta.Page, decoded.Page)
	assert.Equal(t, meta.PerPage, decoded.PerPage)
	assert.Equal(t, meta.Pages, decoded.Pages)
	assert.Equal(t, meta.HasNext, decoded.HasNext)
	assert.Equal(t, meta.HasPrev, decoded.HasPrev)
}

func TestNewListMeta_JSON_FieldNames(t *testing.T) {
	meta := forms.NewListMeta(100, 2, 10)

	data, err := json.Marshal(meta)
	require.NoError(t, err)

	var raw map[string]interface{}
	err = json.Unmarshal(data, &raw)
	require.NoError(t, err)

	assert.Contains(t, raw, "total")
	assert.Contains(t, raw, "page")
	assert.Contains(t, raw, "per_page")
	assert.Contains(t, raw, "pages")
	assert.Contains(t, raw, "has_next")
	assert.Contains(t, raw, "has_prev")
}

func TestNewListResponse(t *testing.T) {
	items := []string{"a", "b", "c"}

	resp := forms.NewListResponse(items, 100, 2, 10)

	assert.Equal(t, items, resp.Items)
	assert.Equal(t, int64(100), resp.Meta.Total)
	assert.Equal(t, 2, resp.Meta.Page)
	assert.Equal(t, 10, resp.Meta.PerPage)
	assert.Equal(t, 10, resp.Meta.Pages)
	assert.True(t, resp.Meta.HasNext)
	assert.True(t, resp.Meta.HasPrev)
}

func TestNewListResponse_NilItems(t *testing.T) {
	resp := forms.NewListResponse[string](nil, 0, 1, 10)

	require.NotNil(t, resp.Items)
	assert.Len(t, resp.Items, 0)
}

func TestNewListResponse_EmptySlice(t *testing.T) {
	items := []string{}

	resp := forms.NewListResponse(items, 0, 1, 10)

	require.NotNil(t, resp.Items)
	assert.Len(t, resp.Items, 0)
}

func TestNewListResponse_JSON(t *testing.T) {
	items := []string{"a", "b", "c"}
	resp := forms.NewListResponse(items, 100, 2, 10)

	data, err := json.Marshal(resp)
	require.NoError(t, err)

	var decoded forms.ListResponse[string]
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)

	assert.Equal(t, items, decoded.Items)
	assert.Equal(t, resp.Meta.Total, decoded.Meta.Total)
}

func TestListResponse_Empty(t *testing.T) {
	resp := forms.NewListResponse[string](nil, 0, 1, 10)
	assert.True(t, resp.Empty())
}

func TestListResponse_Empty_WithItems(t *testing.T) {
	items := []string{"a", "b"}
	resp := forms.NewListResponse(items, 2, 1, 10)
	assert.False(t, resp.Empty())
}

func TestListResponse_Count(t *testing.T) {
	items := []string{"a", "b", "c"}
	resp := forms.NewListResponse(items, 100, 1, 10)

	assert.Equal(t, 3, resp.Count())
}

func TestListResponse_Count_Empty(t *testing.T) {
	resp := forms.NewListResponse[string](nil, 0, 1, 10)
	assert.Equal(t, 0, resp.Count())
}

func TestListResponse_Offset(t *testing.T) {
	tests := []struct {
		page     int
		perPage  int
		expected int
	}{
		{1, 10, 0},
		{2, 10, 10},
		{3, 10, 20},
		{1, 20, 0},
		{2, 20, 20},
		{5, 25, 100},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("page=%d,perPage=%d", tt.page, tt.perPage), func(t *testing.T) {
			resp := forms.NewListResponse([]string{"a"}, 100, tt.page, tt.perPage)
			assert.Equal(t, tt.expected, resp.Offset())
		})
	}
}

func TestMapItems(t *testing.T) {
	numbers := []int{1, 2, 3}

	strings := forms.MapItems(numbers, func(n int) string {
		return fmt.Sprintf("item-%d", n)
	})

	assert.Equal(t, []string{"item-1", "item-2", "item-3"}, strings)
}

func TestMapItems_Empty(t *testing.T) {
	numbers := []int{}

	strings := forms.MapItems(numbers, func(n int) string {
		return fmt.Sprintf("item-%d", n)
	})

	require.NotNil(t, strings)
	assert.Len(t, strings, 0)
}

func TestMapItems_TypeConversion(t *testing.T) {
	type User struct {
		ID   int
		Name string
	}

	type UserDTO struct {
		ID   string
		Name string
	}

	users := []User{
		{ID: 1, Name: "John"},
		{ID: 2, Name: "Jane"},
	}

	dtos := forms.MapItems(users, func(u User) UserDTO {
		return UserDTO{
			ID:   fmt.Sprintf("user-%d", u.ID),
			Name: u.Name,
		}
	})

	require.Len(t, dtos, 2)
	assert.Equal(t, "user-1", dtos[0].ID)
	assert.Equal(t, "John", dtos[0].Name)
	assert.Equal(t, "user-2", dtos[1].ID)
	assert.Equal(t, "Jane", dtos[1].Name)
}

func TestMapListResponse(t *testing.T) {
	numbers := []int{1, 2, 3}
	resp := forms.NewListResponse(numbers, 100, 2, 10)

	mapped := forms.MapListResponse(resp, func(n int) string {
		return fmt.Sprintf("num-%d", n)
	})

	assert.Equal(t, []string{"num-1", "num-2", "num-3"}, mapped.Items)
	assert.Equal(t, resp.Meta, mapped.Meta)
}

func TestMapListResponse_PreservesMeta(t *testing.T) {
	items := []int{1, 2}
	resp := forms.NewListResponse(items, 50, 3, 15)

	mapped := forms.MapListResponse(resp, func(n int) string {
		return ""
	})

	assert.Equal(t, int64(50), mapped.Meta.Total)
	assert.Equal(t, 3, mapped.Meta.Page)
	assert.Equal(t, 15, mapped.Meta.PerPage)
	assert.Equal(t, 4, mapped.Meta.Pages)
}

func TestPaginationParams_Normalize(t *testing.T) {
	params := forms.PaginationParams{Page: 0, PerPage: 0}
	params.Normalize()

	assert.Equal(t, 1, params.Page)
	assert.Equal(t, 20, params.PerPage)
}

func TestPaginationParams_Normalize_NegativeValues(t *testing.T) {
	params := forms.PaginationParams{Page: -5, PerPage: -10}
	params.Normalize()

	assert.Equal(t, 1, params.Page)
	assert.Equal(t, 20, params.PerPage)
}

func TestPaginationParams_Normalize_MaxPerPage(t *testing.T) {
	params := forms.PaginationParams{Page: 1, PerPage: 200}
	params.Normalize()

	assert.Equal(t, 100, params.PerPage)
}

func TestPaginationParams_Normalize_ValidValues(t *testing.T) {
	params := forms.PaginationParams{Page: 5, PerPage: 50}
	params.Normalize()

	assert.Equal(t, 5, params.Page)
	assert.Equal(t, 50, params.PerPage)
}

func TestPaginationParams_Normalize_EdgeCase(t *testing.T) {
	params := forms.PaginationParams{Page: 1, PerPage: 100}
	params.Normalize()

	assert.Equal(t, 1, params.Page)
	assert.Equal(t, 100, params.PerPage)
}

func TestPaginationParams_ToLimitOffset(t *testing.T) {
	tests := []struct {
		page           int
		perPage        int
		expectedLimit  int
		expectedOffset int
	}{
		{1, 10, 10, 0},
		{2, 10, 10, 10},
		{3, 10, 10, 20},
		{1, 20, 20, 0},
		{5, 25, 25, 100},
		{0, 0, 20, 0},     // Defaults
		{-1, -5, 20, 0},   // Negative values normalized
		{1, 200, 100, 0},  // Max per_page
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("page=%d,perPage=%d", tt.page, tt.perPage), func(t *testing.T) {
			params := forms.PaginationParams{Page: tt.page, PerPage: tt.perPage}
			limit, offset := params.ToLimitOffset()

			assert.Equal(t, tt.expectedLimit, limit)
			assert.Equal(t, tt.expectedOffset, offset)
		})
	}
}

func TestPaginationParams_ToLimitOffset_NormalizesValues(t *testing.T) {
	params := forms.PaginationParams{Page: 0, PerPage: 0}
	limit, offset := params.ToLimitOffset()

	assert.Equal(t, 20, limit)
	assert.Equal(t, 0, offset)
	// Also verify params are normalized
	assert.Equal(t, 1, params.Page)
	assert.Equal(t, 20, params.PerPage)
}

func TestNewCursorListResponse(t *testing.T) {
	items := []string{"a", "b", "c"}

	resp := forms.NewCursorListResponse(items, "next123", "prev456", true, 10)

	assert.Equal(t, items, resp.Items)
	assert.Equal(t, "next123", resp.Meta.NextCursor)
	assert.Equal(t, "prev456", resp.Meta.PrevCursor)
	assert.True(t, resp.Meta.HasMore)
	assert.Equal(t, 10, resp.Meta.PerPage)
}

func TestNewCursorListResponse_NilItems(t *testing.T) {
	resp := forms.NewCursorListResponse[string](nil, "", "", false, 10)

	require.NotNil(t, resp.Items)
	assert.Len(t, resp.Items, 0)
}

func TestNewCursorListResponse_EmptySlice(t *testing.T) {
	items := []string{}

	resp := forms.NewCursorListResponse(items, "", "", false, 10)

	require.NotNil(t, resp.Items)
	assert.Len(t, resp.Items, 0)
}

func TestNewCursorListResponse_NoCursors(t *testing.T) {
	items := []string{"a"}

	resp := forms.NewCursorListResponse(items, "", "", false, 10)

	assert.Empty(t, resp.Meta.NextCursor)
	assert.Empty(t, resp.Meta.PrevCursor)
	assert.False(t, resp.Meta.HasMore)
}

func TestCursorListResponse_JSON(t *testing.T) {
	items := []string{"a", "b"}
	resp := forms.NewCursorListResponse(items, "next123", "prev456", true, 10)

	data, err := json.Marshal(resp)
	require.NoError(t, err)

	var decoded forms.CursorListResponse[string]
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)

	assert.Equal(t, items, decoded.Items)
	assert.Equal(t, "next123", decoded.Meta.NextCursor)
	assert.Equal(t, "prev456", decoded.Meta.PrevCursor)
	assert.True(t, decoded.Meta.HasMore)
	assert.Equal(t, 10, decoded.Meta.PerPage)
}

func TestCursorMeta_JSON_OmitsEmptyCursors(t *testing.T) {
	resp := forms.NewCursorListResponse([]string{"a"}, "", "", false, 10)

	data, err := json.Marshal(resp)
	require.NoError(t, err)

	var raw map[string]interface{}
	err = json.Unmarshal(data, &raw)
	require.NoError(t, err)

	meta, ok := raw["meta"].(map[string]interface{})
	require.True(t, ok)

	assert.NotContains(t, meta, "next_cursor")
	assert.NotContains(t, meta, "prev_cursor")
	assert.Contains(t, meta, "has_more")
	assert.Contains(t, meta, "per_page")
}

func TestCursorMeta_JSON_FieldNames(t *testing.T) {
	resp := forms.NewCursorListResponse([]string{"a"}, "next", "prev", true, 10)

	data, err := json.Marshal(resp)
	require.NoError(t, err)

	var raw map[string]interface{}
	err = json.Unmarshal(data, &raw)
	require.NoError(t, err)

	meta, ok := raw["meta"].(map[string]interface{})
	require.True(t, ok)

	assert.Contains(t, meta, "next_cursor")
	assert.Contains(t, meta, "prev_cursor")
	assert.Contains(t, meta, "has_more")
	assert.Contains(t, meta, "per_page")
}

func TestListResponse_WithStructItems(t *testing.T) {
	type Item struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	}

	items := []Item{
		{ID: 1, Name: "First"},
		{ID: 2, Name: "Second"},
	}

	resp := forms.NewListResponse(items, 100, 1, 10)

	assert.Len(t, resp.Items, 2)
	assert.Equal(t, 1, resp.Items[0].ID)
	assert.Equal(t, "First", resp.Items[0].Name)
}

func TestListResponse_WithPointerItems(t *testing.T) {
	type Item struct {
		ID int
	}

	items := []*Item{
		{ID: 1},
		{ID: 2},
	}

	resp := forms.NewListResponse(items, 100, 1, 10)

	assert.Len(t, resp.Items, 2)
	assert.Equal(t, 1, resp.Items[0].ID)
}
