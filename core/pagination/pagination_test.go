package pagination_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/codoworks/codo-framework/core/pagination"
)

func TestParams_IsOffset(t *testing.T) {
	params := &pagination.Params{Type: pagination.TypeOffset}
	assert.True(t, params.IsOffset())
	assert.False(t, params.IsCursor())
}

func TestParams_IsCursor(t *testing.T) {
	params := &pagination.Params{Type: pagination.TypeCursor}
	assert.True(t, params.IsCursor())
	assert.False(t, params.IsOffset())
}

func TestParams_HasCursor(t *testing.T) {
	params := &pagination.Params{Type: pagination.TypeCursor, Cursor: "abc123"}
	assert.True(t, params.HasCursor())

	params.Cursor = ""
	assert.False(t, params.HasCursor())
}

func TestParams_Limit(t *testing.T) {
	params := &pagination.Params{PerPage: 25}
	assert.Equal(t, 25, params.Limit())
}

func TestParams_WithMaxPerPage(t *testing.T) {
	params := &pagination.Params{
		Type:       pagination.TypeOffset,
		Page:       2,
		PerPage:    100,
		Offset:     100,
		MaxPerPage: 100,
	}

	// Increase max - should not change PerPage
	newParams := params.WithMaxPerPage(200)
	assert.Equal(t, 100, newParams.PerPage)
	assert.Equal(t, 200, newParams.MaxPerPage)
	assert.Equal(t, 100, newParams.Offset) // Offset unchanged

	// Decrease max - should cap PerPage and recalculate offset
	newParams = params.WithMaxPerPage(50)
	assert.Equal(t, 50, newParams.PerPage)
	assert.Equal(t, 50, newParams.MaxPerPage)
	assert.Equal(t, 50, newParams.Offset) // (Page 2 - 1) * 50 = 50
}

func TestParams_WithMaxPerPage_Nil(t *testing.T) {
	var params *pagination.Params
	result := params.WithMaxPerPage(100)
	assert.Nil(t, result)
}

func TestParams_WithPerPage(t *testing.T) {
	params := &pagination.Params{
		Type:       pagination.TypeOffset,
		Page:       3,
		PerPage:    20,
		Offset:     40,
		MaxPerPage: 100,
	}

	// Change per_page
	newParams := params.WithPerPage(50)
	assert.Equal(t, 50, newParams.PerPage)
	assert.Equal(t, 100, newParams.Offset) // (Page 3 - 1) * 50 = 100

	// Exceeds max - should cap
	newParams = params.WithPerPage(200)
	assert.Equal(t, 100, newParams.PerPage)
	assert.Equal(t, 200, newParams.Offset) // (Page 3 - 1) * 100 = 200

	// Below minimum - should be 1
	newParams = params.WithPerPage(0)
	assert.Equal(t, 1, newParams.PerPage)

	newParams = params.WithPerPage(-5)
	assert.Equal(t, 1, newParams.PerPage)
}

func TestParams_WithPerPage_Nil(t *testing.T) {
	var params *pagination.Params
	result := params.WithPerPage(50)
	assert.Nil(t, result)
}

func TestParams_WithPage(t *testing.T) {
	params := &pagination.Params{
		Type:       pagination.TypeOffset,
		Page:       1,
		PerPage:    20,
		Offset:     0,
		MaxPerPage: 100,
	}

	// Change page
	newParams := params.WithPage(5)
	assert.Equal(t, 5, newParams.Page)
	assert.Equal(t, 80, newParams.Offset) // (Page 5 - 1) * 20 = 80

	// Invalid page - should be 1
	newParams = params.WithPage(0)
	assert.Equal(t, 1, newParams.Page)
	assert.Equal(t, 0, newParams.Offset)

	newParams = params.WithPage(-3)
	assert.Equal(t, 1, newParams.Page)
	assert.Equal(t, 0, newParams.Offset)
}

func TestParams_WithPage_Nil(t *testing.T) {
	var params *pagination.Params
	result := params.WithPage(5)
	assert.Nil(t, result)
}

func TestParams_Immutability(t *testing.T) {
	original := &pagination.Params{
		Type:       pagination.TypeOffset,
		Page:       2,
		PerPage:    25,
		Offset:     25,
		MaxPerPage: 100,
	}

	// Ensure mutations return new copies
	newParams := original.WithPage(5)
	assert.Equal(t, 2, original.Page)   // Original unchanged
	assert.Equal(t, 5, newParams.Page)  // New has change
	assert.Equal(t, 25, original.Offset) // Original unchanged

	newParams = original.WithPerPage(50)
	assert.Equal(t, 25, original.PerPage) // Original unchanged
	assert.Equal(t, 50, newParams.PerPage)

	newParams = original.WithMaxPerPage(200)
	assert.Equal(t, 100, original.MaxPerPage) // Original unchanged
	assert.Equal(t, 200, newParams.MaxPerPage)
}

func TestTypeConstants(t *testing.T) {
	assert.Equal(t, pagination.Type("offset"), pagination.TypeOffset)
	assert.Equal(t, pagination.Type("cursor"), pagination.TypeCursor)
}

func TestDirectionConstants(t *testing.T) {
	assert.Equal(t, pagination.Direction("next"), pagination.DirectionNext)
	assert.Equal(t, pagination.Direction("prev"), pagination.DirectionPrev)
}
