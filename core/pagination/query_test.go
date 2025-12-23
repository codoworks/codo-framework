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

func TestQueryOptions_FromContext(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	params := &pagination.Params{
		Type:    pagination.TypeOffset,
		Page:    2,
		PerPage: 25,
		Offset:  25,
	}
	pagination.Set(c, params)

	opts := pagination.QueryOptions(c)
	require.NotNil(t, opts)
	assert.Len(t, opts, 2) // Limit and Offset
}

func TestQueryOptions_NotSet(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	opts := pagination.QueryOptions(c)
	assert.Nil(t, opts)
}

func TestParams_QueryOptions_Offset(t *testing.T) {
	params := &pagination.Params{
		Type:    pagination.TypeOffset,
		Page:    3,
		PerPage: 50,
		Offset:  100,
	}

	opts := params.QueryOptions()
	require.NotNil(t, opts)
	assert.Len(t, opts, 2) // Limit and Offset
}

func TestParams_QueryOptions_Cursor(t *testing.T) {
	params := &pagination.Params{
		Type:      pagination.TypeCursor,
		PerPage:   25,
		Cursor:    "abc123",
		Direction: pagination.DirectionNext,
	}

	opts := params.QueryOptions()
	require.NotNil(t, opts)
	assert.Len(t, opts, 1) // Only Limit (PerPage + 1 for HasMore detection)
}

func TestParams_QueryOptions_Nil(t *testing.T) {
	var params *pagination.Params
	opts := params.QueryOptions()
	assert.Nil(t, opts)
}

func TestParams_QueryOptionsWithOrder(t *testing.T) {
	params := &pagination.Params{
		Type:    pagination.TypeOffset,
		Page:    1,
		PerPage: 20,
		Offset:  0,
	}

	opts := params.QueryOptionsWithOrder("created_at", "DESC")
	require.NotNil(t, opts)
	assert.Len(t, opts, 3) // Limit, Offset, and OrderBy
}

func TestParams_QueryOptionsWithOrderAsc(t *testing.T) {
	params := &pagination.Params{
		Type:    pagination.TypeOffset,
		Page:    1,
		PerPage: 20,
		Offset:  0,
	}

	opts := params.QueryOptionsWithOrderAsc("name")
	require.NotNil(t, opts)
	assert.Len(t, opts, 3)
}

func TestParams_QueryOptionsWithOrderDesc(t *testing.T) {
	params := &pagination.Params{
		Type:    pagination.TypeOffset,
		Page:    1,
		PerPage: 20,
		Offset:  0,
	}

	opts := params.QueryOptionsWithOrderDesc("updated_at")
	require.NotNil(t, opts)
	assert.Len(t, opts, 3)
}

func TestParams_QueryOptionsWithOrder_Nil(t *testing.T) {
	var params *pagination.Params
	opts := params.QueryOptionsWithOrder("id", "ASC")
	assert.Nil(t, opts)
}
