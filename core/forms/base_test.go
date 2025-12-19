package forms_test

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/codoworks/codo-framework/core/forms"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFormBase_Fields(t *testing.T) {
	now := time.Now()
	form := forms.FormBase{
		ID:        "test-id",
		CreatedAt: now,
		UpdatedAt: now,
	}

	assert.Equal(t, "test-id", form.ID)
	assert.Equal(t, now, form.CreatedAt)
	assert.Equal(t, now, form.UpdatedAt)
}

func TestFormBase_JSON(t *testing.T) {
	now := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)
	form := forms.FormBase{
		ID:        "test-id",
		CreatedAt: now,
		UpdatedAt: now,
	}

	data, err := json.Marshal(form)
	require.NoError(t, err)

	var decoded forms.FormBase
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)

	assert.Equal(t, form.ID, decoded.ID)
	assert.True(t, form.CreatedAt.Equal(decoded.CreatedAt))
	assert.True(t, form.UpdatedAt.Equal(decoded.UpdatedAt))
}

func TestFormBase_JSON_FieldNames(t *testing.T) {
	now := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)
	form := forms.FormBase{
		ID:        "test-id",
		CreatedAt: now,
		UpdatedAt: now,
	}

	data, err := json.Marshal(form)
	require.NoError(t, err)

	var raw map[string]interface{}
	err = json.Unmarshal(data, &raw)
	require.NoError(t, err)

	assert.Contains(t, raw, "id")
	assert.Contains(t, raw, "created_at")
	assert.Contains(t, raw, "updated_at")
}

func TestNewFormBase(t *testing.T) {
	createdAt := time.Date(2024, 1, 10, 0, 0, 0, 0, time.UTC)
	updatedAt := time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)

	form := forms.NewFormBase("my-id", createdAt, updatedAt)

	assert.Equal(t, "my-id", form.ID)
	assert.Equal(t, createdAt, form.CreatedAt)
	assert.Equal(t, updatedAt, form.UpdatedAt)
}

func TestNewFormBase_EmptyID(t *testing.T) {
	now := time.Now()
	form := forms.NewFormBase("", now, now)

	assert.Empty(t, form.ID)
	assert.Equal(t, now, form.CreatedAt)
}

func TestNewFormBase_ZeroTime(t *testing.T) {
	var zeroTime time.Time
	form := forms.NewFormBase("id", zeroTime, zeroTime)

	assert.Equal(t, "id", form.ID)
	assert.True(t, form.CreatedAt.IsZero())
	assert.True(t, form.UpdatedAt.IsZero())
}

func TestFormWithID(t *testing.T) {
	form := forms.FormWithID{ID: "test-id"}
	assert.Equal(t, "test-id", form.ID)
}

func TestFormWithID_JSON(t *testing.T) {
	form := forms.FormWithID{ID: "test-id"}

	data, err := json.Marshal(form)
	require.NoError(t, err)

	var decoded forms.FormWithID
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)

	assert.Equal(t, "test-id", decoded.ID)
}

func TestFormWithID_JSON_FieldNames(t *testing.T) {
	form := forms.FormWithID{ID: "test-id"}

	data, err := json.Marshal(form)
	require.NoError(t, err)

	var raw map[string]interface{}
	err = json.Unmarshal(data, &raw)
	require.NoError(t, err)

	assert.Contains(t, raw, "id")
	assert.Len(t, raw, 1)
}

func TestNewFormWithID(t *testing.T) {
	form := forms.NewFormWithID("my-id")
	assert.Equal(t, "my-id", form.ID)
}

func TestNewFormWithID_Empty(t *testing.T) {
	form := forms.NewFormWithID("")
	assert.Empty(t, form.ID)
}

func TestTimestampFields(t *testing.T) {
	createdAt := time.Date(2024, 1, 10, 0, 0, 0, 0, time.UTC)
	updatedAt := time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)

	form := forms.TimestampFields{
		CreatedAt: createdAt,
		UpdatedAt: updatedAt,
	}

	assert.Equal(t, createdAt, form.CreatedAt)
	assert.Equal(t, updatedAt, form.UpdatedAt)
}

func TestTimestampFields_JSON(t *testing.T) {
	now := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)
	form := forms.TimestampFields{
		CreatedAt: now,
		UpdatedAt: now,
	}

	data, err := json.Marshal(form)
	require.NoError(t, err)

	var decoded forms.TimestampFields
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)

	assert.True(t, form.CreatedAt.Equal(decoded.CreatedAt))
	assert.True(t, form.UpdatedAt.Equal(decoded.UpdatedAt))
}

func TestTimestampFields_JSON_FieldNames(t *testing.T) {
	now := time.Now()
	form := forms.TimestampFields{
		CreatedAt: now,
		UpdatedAt: now,
	}

	data, err := json.Marshal(form)
	require.NoError(t, err)

	var raw map[string]interface{}
	err = json.Unmarshal(data, &raw)
	require.NoError(t, err)

	assert.Contains(t, raw, "created_at")
	assert.Contains(t, raw, "updated_at")
	assert.Len(t, raw, 2)
}

func TestNewTimestampFields(t *testing.T) {
	createdAt := time.Date(2024, 1, 10, 0, 0, 0, 0, time.UTC)
	updatedAt := time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)

	form := forms.NewTimestampFields(createdAt, updatedAt)

	assert.Equal(t, createdAt, form.CreatedAt)
	assert.Equal(t, updatedAt, form.UpdatedAt)
}

func TestNewTimestampFields_ZeroTime(t *testing.T) {
	var zeroTime time.Time
	form := forms.NewTimestampFields(zeroTime, zeroTime)

	assert.True(t, form.CreatedAt.IsZero())
	assert.True(t, form.UpdatedAt.IsZero())
}

func TestDeletedFormBase(t *testing.T) {
	now := time.Now()
	deletedAt := now.Add(time.Hour)

	form := forms.DeletedFormBase{
		FormBase: forms.FormBase{
			ID:        "test-id",
			CreatedAt: now,
			UpdatedAt: now,
		},
		DeletedAt: &deletedAt,
	}

	assert.Equal(t, "test-id", form.ID)
	assert.Equal(t, now, form.CreatedAt)
	assert.Equal(t, now, form.UpdatedAt)
	require.NotNil(t, form.DeletedAt)
	assert.Equal(t, deletedAt, *form.DeletedAt)
}

func TestDeletedFormBase_NilDeletedAt(t *testing.T) {
	now := time.Now()

	form := forms.DeletedFormBase{
		FormBase: forms.FormBase{
			ID:        "test-id",
			CreatedAt: now,
			UpdatedAt: now,
		},
		DeletedAt: nil,
	}

	assert.Equal(t, "test-id", form.ID)
	assert.Nil(t, form.DeletedAt)
}

func TestNewDeletedFormBase(t *testing.T) {
	createdAt := time.Date(2024, 1, 10, 0, 0, 0, 0, time.UTC)
	updatedAt := time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)
	deletedAt := time.Date(2024, 1, 20, 0, 0, 0, 0, time.UTC)

	form := forms.NewDeletedFormBase("my-id", createdAt, updatedAt, &deletedAt)

	assert.Equal(t, "my-id", form.ID)
	assert.Equal(t, createdAt, form.CreatedAt)
	assert.Equal(t, updatedAt, form.UpdatedAt)
	require.NotNil(t, form.DeletedAt)
	assert.Equal(t, deletedAt, *form.DeletedAt)
}

func TestNewDeletedFormBase_NilDeletedAt(t *testing.T) {
	createdAt := time.Date(2024, 1, 10, 0, 0, 0, 0, time.UTC)
	updatedAt := time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)

	form := forms.NewDeletedFormBase("my-id", createdAt, updatedAt, nil)

	assert.Equal(t, "my-id", form.ID)
	assert.Nil(t, form.DeletedAt)
}

func TestDeletedFormBase_JSON(t *testing.T) {
	now := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)
	deletedAt := time.Date(2024, 1, 20, 10, 30, 0, 0, time.UTC)

	form := forms.DeletedFormBase{
		FormBase: forms.FormBase{
			ID:        "test-id",
			CreatedAt: now,
			UpdatedAt: now,
		},
		DeletedAt: &deletedAt,
	}

	data, err := json.Marshal(form)
	require.NoError(t, err)

	var decoded forms.DeletedFormBase
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)

	assert.Equal(t, form.ID, decoded.ID)
	require.NotNil(t, decoded.DeletedAt)
	assert.True(t, deletedAt.Equal(*decoded.DeletedAt))
}

func TestDeletedFormBase_JSON_NilDeleted(t *testing.T) {
	now := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)

	form := forms.DeletedFormBase{
		FormBase: forms.FormBase{
			ID:        "test-id",
			CreatedAt: now,
			UpdatedAt: now,
		},
		DeletedAt: nil,
	}

	data, err := json.Marshal(form)
	require.NoError(t, err)

	// deleted_at should be omitted when nil
	var raw map[string]interface{}
	err = json.Unmarshal(data, &raw)
	require.NoError(t, err)

	assert.NotContains(t, raw, "deleted_at")
	assert.Contains(t, raw, "id")
	assert.Contains(t, raw, "created_at")
	assert.Contains(t, raw, "updated_at")
}

func TestDeletedFormBase_JSON_FieldNames(t *testing.T) {
	now := time.Now()
	deletedAt := now.Add(time.Hour)

	form := forms.DeletedFormBase{
		FormBase: forms.FormBase{
			ID:        "test-id",
			CreatedAt: now,
			UpdatedAt: now,
		},
		DeletedAt: &deletedAt,
	}

	data, err := json.Marshal(form)
	require.NoError(t, err)

	var raw map[string]interface{}
	err = json.Unmarshal(data, &raw)
	require.NoError(t, err)

	assert.Contains(t, raw, "id")
	assert.Contains(t, raw, "created_at")
	assert.Contains(t, raw, "updated_at")
	assert.Contains(t, raw, "deleted_at")
}
