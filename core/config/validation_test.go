package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidationError_Error(t *testing.T) {
	err := &ValidationError{
		Field:   "database.host",
		Message: "is required",
	}

	result := err.Error()

	assert.Equal(t, "database.host: is required", result)
}

func TestValidationError_Error_EmptyField(t *testing.T) {
	err := &ValidationError{
		Field:   "",
		Message: "general error",
	}

	result := err.Error()

	assert.Equal(t, ": general error", result)
}

func TestValidationErrors_Error(t *testing.T) {
	errs := ValidationErrors{
		{Field: "server.port", Message: "must be between 1 and 65535"},
		{Field: "database.name", Message: "is required"},
	}

	result := errs.Error()

	assert.Equal(t, "configuration validation failed: server.port: must be between 1 and 65535; database.name: is required", result)
}

func TestValidationErrors_Error_Empty(t *testing.T) {
	errs := ValidationErrors{}

	result := errs.Error()

	assert.Equal(t, "", result)
}

func TestValidationErrors_Error_Single(t *testing.T) {
	errs := ValidationErrors{
		{Field: "service.name", Message: "is required"},
	}

	result := errs.Error()

	assert.Equal(t, "configuration validation failed: service.name: is required", result)
}

func TestValidationErrors_Add(t *testing.T) {
	var errs ValidationErrors

	errs.Add("field1", "message1")
	errs.Add("field2", "message2")

	assert.Len(t, errs, 2)
	assert.Equal(t, "field1", errs[0].Field)
	assert.Equal(t, "message1", errs[0].Message)
	assert.Equal(t, "field2", errs[1].Field)
	assert.Equal(t, "message2", errs[1].Message)
}

func TestValidationErrors_Add_ToEmpty(t *testing.T) {
	var errs ValidationErrors

	errs.Add("test.field", "test message")

	assert.Len(t, errs, 1)
	assert.Equal(t, "test.field", errs[0].Field)
}

func TestValidationErrors_HasErrors(t *testing.T) {
	errs := ValidationErrors{
		{Field: "test", Message: "error"},
	}

	assert.True(t, errs.HasErrors())
}

func TestValidationErrors_HasErrors_Empty(t *testing.T) {
	errs := ValidationErrors{}

	assert.False(t, errs.HasErrors())
}

func TestValidationErrors_HasErrors_Nil(t *testing.T) {
	var errs ValidationErrors

	assert.False(t, errs.HasErrors())
}

func TestValidationErrors_ToError(t *testing.T) {
	errs := ValidationErrors{
		{Field: "test", Message: "error"},
	}

	result := errs.ToError()

	assert.NotNil(t, result)
	assert.Equal(t, errs, result)
}

func TestValidationErrors_ToError_Empty(t *testing.T) {
	errs := ValidationErrors{}

	result := errs.ToError()

	assert.Nil(t, result)
}

func TestValidationErrors_ToError_Nil(t *testing.T) {
	var errs ValidationErrors

	result := errs.ToError()

	assert.Nil(t, result)
}

func TestValidationErrors_ErrorInterface(t *testing.T) {
	errs := ValidationErrors{
		{Field: "test", Message: "error"},
	}

	// Verify it implements error interface
	var err error = errs
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "test: error")
}

func TestValidationError_ErrorInterface(t *testing.T) {
	verr := &ValidationError{Field: "field", Message: "msg"}

	// Verify it implements error interface
	var err error = verr
	assert.NotNil(t, err)
	assert.Equal(t, "field: msg", err.Error())
}
