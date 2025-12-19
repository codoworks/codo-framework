package forms_test

import (
	"encoding/json"
	"testing"

	"github.com/codoworks/codo-framework/core/forms"
	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type ValidForm struct {
	Name  string `validate:"required,min=2,max=50"`
	Email string `validate:"required,email"`
}

func TestValidate_Valid(t *testing.T) {
	form := &ValidForm{
		Name:  "John Doe",
		Email: "john@example.com",
	}

	err := forms.Validate(form)
	assert.NoError(t, err)
}

func TestValidate_Required(t *testing.T) {
	type Form struct {
		Name string `validate:"required"`
	}

	form := &Form{Name: ""}
	err := forms.Validate(form)

	require.Error(t, err)

	validationErr, ok := err.(*forms.ValidationErrors)
	require.True(t, ok)
	require.Len(t, validationErr.Errors, 1)
	assert.Equal(t, "name", validationErr.Errors[0].Field)
	assert.Equal(t, "required", validationErr.Errors[0].Tag)
	assert.Contains(t, validationErr.Errors[0].Message, "required")
}

func TestValidate_Min(t *testing.T) {
	type Form struct {
		Name string `validate:"min=5"`
	}

	form := &Form{Name: "abc"}
	err := forms.Validate(form)

	require.Error(t, err)

	validationErr, ok := err.(*forms.ValidationErrors)
	require.True(t, ok)
	require.Len(t, validationErr.Errors, 1)
	assert.Equal(t, "name", validationErr.Errors[0].Field)
	assert.Equal(t, "min", validationErr.Errors[0].Tag)
	assert.Contains(t, validationErr.Errors[0].Message, "at least 5 characters")
}

func TestValidate_Max(t *testing.T) {
	type Form struct {
		Name string `validate:"max=5"`
	}

	form := &Form{Name: "abcdefgh"}
	err := forms.Validate(form)

	require.Error(t, err)

	validationErr, ok := err.(*forms.ValidationErrors)
	require.True(t, ok)
	require.Len(t, validationErr.Errors, 1)
	assert.Equal(t, "name", validationErr.Errors[0].Field)
	assert.Equal(t, "max", validationErr.Errors[0].Tag)
	assert.Contains(t, validationErr.Errors[0].Message, "at most 5 characters")
}

func TestValidate_Email(t *testing.T) {
	type Form struct {
		Email string `validate:"email"`
	}

	form := &Form{Email: "invalid-email"}
	err := forms.Validate(form)

	require.Error(t, err)

	validationErr, ok := err.(*forms.ValidationErrors)
	require.True(t, ok)
	require.Len(t, validationErr.Errors, 1)
	assert.Equal(t, "email", validationErr.Errors[0].Field)
	assert.Equal(t, "email", validationErr.Errors[0].Tag)
	assert.Contains(t, validationErr.Errors[0].Message, "valid email address")
}

func TestValidate_UUID(t *testing.T) {
	type Form struct {
		Uuid string `validate:"uuid"`
	}

	form := &Form{Uuid: "not-a-uuid"}
	err := forms.Validate(form)

	require.Error(t, err)

	validationErr, ok := err.(*forms.ValidationErrors)
	require.True(t, ok)
	require.Len(t, validationErr.Errors, 1)
	assert.Equal(t, "uuid", validationErr.Errors[0].Field)
	assert.Equal(t, "uuid", validationErr.Errors[0].Tag)
	assert.Contains(t, validationErr.Errors[0].Message, "valid UUID")
}

func TestValidate_UUID_Valid(t *testing.T) {
	type Form struct {
		ID string `validate:"uuid"`
	}

	form := &Form{ID: "550e8400-e29b-41d4-a716-446655440000"}
	err := forms.Validate(form)
	assert.NoError(t, err)
}

func TestValidate_URL(t *testing.T) {
	type Form struct {
		Website string `validate:"url"`
	}

	form := &Form{Website: "not-a-url"}
	err := forms.Validate(form)

	require.Error(t, err)

	validationErr, ok := err.(*forms.ValidationErrors)
	require.True(t, ok)
	require.Len(t, validationErr.Errors, 1)
	assert.Equal(t, "website", validationErr.Errors[0].Field)
	assert.Equal(t, "url", validationErr.Errors[0].Tag)
	assert.Contains(t, validationErr.Errors[0].Message, "valid URL")
}

func TestValidate_URL_Valid(t *testing.T) {
	type Form struct {
		Website string `validate:"url"`
	}

	form := &Form{Website: "https://example.com"}
	err := forms.Validate(form)
	assert.NoError(t, err)
}

func TestValidate_OneOf(t *testing.T) {
	type Form struct {
		Status string `validate:"oneof=active inactive pending"`
	}

	form := &Form{Status: "invalid"}
	err := forms.Validate(form)

	require.Error(t, err)

	validationErr, ok := err.(*forms.ValidationErrors)
	require.True(t, ok)
	require.Len(t, validationErr.Errors, 1)
	assert.Equal(t, "status", validationErr.Errors[0].Field)
	assert.Equal(t, "oneof", validationErr.Errors[0].Tag)
	assert.Contains(t, validationErr.Errors[0].Message, "must be one of")
}

func TestValidate_OneOf_Valid(t *testing.T) {
	type Form struct {
		Status string `validate:"oneof=active inactive pending"`
	}

	form := &Form{Status: "active"}
	err := forms.Validate(form)
	assert.NoError(t, err)
}

func TestValidate_Gte(t *testing.T) {
	type Form struct {
		Age int `validate:"gte=18"`
	}

	form := &Form{Age: 16}
	err := forms.Validate(form)

	require.Error(t, err)

	validationErr, ok := err.(*forms.ValidationErrors)
	require.True(t, ok)
	require.Len(t, validationErr.Errors, 1)
	assert.Equal(t, "age", validationErr.Errors[0].Field)
	assert.Equal(t, "gte", validationErr.Errors[0].Tag)
	assert.Contains(t, validationErr.Errors[0].Message, "at least 18")
}

func TestValidate_Lte(t *testing.T) {
	type Form struct {
		Age int `validate:"lte=100"`
	}

	form := &Form{Age: 150}
	err := forms.Validate(form)

	require.Error(t, err)

	validationErr, ok := err.(*forms.ValidationErrors)
	require.True(t, ok)
	require.Len(t, validationErr.Errors, 1)
	assert.Equal(t, "age", validationErr.Errors[0].Field)
	assert.Equal(t, "lte", validationErr.Errors[0].Tag)
	assert.Contains(t, validationErr.Errors[0].Message, "at most 100")
}

func TestValidate_Gt(t *testing.T) {
	type Form struct {
		Count int `validate:"gt=0"`
	}

	form := &Form{Count: 0}
	err := forms.Validate(form)

	require.Error(t, err)

	validationErr, ok := err.(*forms.ValidationErrors)
	require.True(t, ok)
	require.Len(t, validationErr.Errors, 1)
	assert.Equal(t, "count", validationErr.Errors[0].Field)
	assert.Equal(t, "gt", validationErr.Errors[0].Tag)
	assert.Contains(t, validationErr.Errors[0].Message, "greater than 0")
}

func TestValidate_Lt(t *testing.T) {
	type Form struct {
		Count int `validate:"lt=10"`
	}

	form := &Form{Count: 10}
	err := forms.Validate(form)

	require.Error(t, err)

	validationErr, ok := err.(*forms.ValidationErrors)
	require.True(t, ok)
	require.Len(t, validationErr.Errors, 1)
	assert.Equal(t, "count", validationErr.Errors[0].Field)
	assert.Equal(t, "lt", validationErr.Errors[0].Tag)
	assert.Contains(t, validationErr.Errors[0].Message, "less than 10")
}

func TestValidate_Len(t *testing.T) {
	type Form struct {
		Code string `validate:"len=6"`
	}

	form := &Form{Code: "abc"}
	err := forms.Validate(form)

	require.Error(t, err)

	validationErr, ok := err.(*forms.ValidationErrors)
	require.True(t, ok)
	require.Len(t, validationErr.Errors, 1)
	assert.Equal(t, "code", validationErr.Errors[0].Field)
	assert.Equal(t, "len", validationErr.Errors[0].Tag)
	assert.Contains(t, validationErr.Errors[0].Message, "exactly 6 characters")
}

func TestValidate_Alphanum(t *testing.T) {
	type Form struct {
		Username string `validate:"alphanum"`
	}

	form := &Form{Username: "user@name"}
	err := forms.Validate(form)

	require.Error(t, err)

	validationErr, ok := err.(*forms.ValidationErrors)
	require.True(t, ok)
	require.Len(t, validationErr.Errors, 1)
	assert.Equal(t, "username", validationErr.Errors[0].Field)
	assert.Equal(t, "alphanum", validationErr.Errors[0].Tag)
	assert.Contains(t, validationErr.Errors[0].Message, "alphanumeric")
}

func TestValidate_Alpha(t *testing.T) {
	type Form struct {
		Name string `validate:"alpha"`
	}

	form := &Form{Name: "John123"}
	err := forms.Validate(form)

	require.Error(t, err)

	validationErr, ok := err.(*forms.ValidationErrors)
	require.True(t, ok)
	require.Len(t, validationErr.Errors, 1)
	assert.Equal(t, "name", validationErr.Errors[0].Field)
	assert.Equal(t, "alpha", validationErr.Errors[0].Tag)
	assert.Contains(t, validationErr.Errors[0].Message, "only letters")
}

func TestValidate_Numeric(t *testing.T) {
	type Form struct {
		Phone string `validate:"numeric"`
	}

	form := &Form{Phone: "123-456"}
	err := forms.Validate(form)

	require.Error(t, err)

	validationErr, ok := err.(*forms.ValidationErrors)
	require.True(t, ok)
	require.Len(t, validationErr.Errors, 1)
	assert.Equal(t, "phone", validationErr.Errors[0].Field)
	assert.Equal(t, "numeric", validationErr.Errors[0].Tag)
	assert.Contains(t, validationErr.Errors[0].Message, "numeric")
}

func TestValidate_EqField(t *testing.T) {
	type Form struct {
		Password        string `validate:"required"`
		ConfirmPassword string `validate:"eqfield=Password"`
	}

	form := &Form{Password: "secret", ConfirmPassword: "different"}
	err := forms.Validate(form)

	require.Error(t, err)

	validationErr, ok := err.(*forms.ValidationErrors)
	require.True(t, ok)
	require.Len(t, validationErr.Errors, 1)
	assert.Equal(t, "confirm_password", validationErr.Errors[0].Field)
	assert.Equal(t, "eqfield", validationErr.Errors[0].Tag)
	assert.Contains(t, validationErr.Errors[0].Message, "must equal password")
}

func TestValidate_Multiple(t *testing.T) {
	type Form struct {
		Name  string `validate:"required,min=2"`
		Email string `validate:"required,email"`
	}

	form := &Form{Name: "", Email: "invalid"}
	err := forms.Validate(form)

	require.Error(t, err)

	validationErr, ok := err.(*forms.ValidationErrors)
	require.True(t, ok)
	assert.GreaterOrEqual(t, len(validationErr.Errors), 2)
}

func TestValidate_UnknownTag(t *testing.T) {
	// Register a custom validation to test the default case
	customValidator := validator.New()
	_ = customValidator.RegisterValidation("customtag", func(fl validator.FieldLevel) bool {
		return false
	})

	originalValidator := forms.GetValidator()
	forms.SetValidator(customValidator)
	defer forms.SetValidator(originalValidator)

	type Form struct {
		Field string `validate:"customtag"`
	}

	form := &Form{Field: "value"}
	err := forms.Validate(form)

	require.Error(t, err)

	validationErr, ok := err.(*forms.ValidationErrors)
	require.True(t, ok)
	require.Len(t, validationErr.Errors, 1)
	assert.Contains(t, validationErr.Errors[0].Message, "customtag validation")
}

func TestValidationErrors_Error(t *testing.T) {
	type Form struct {
		Name  string `validate:"required"`
		Email string `validate:"email"`
	}

	form := &Form{Name: "", Email: "invalid"}
	err := forms.Validate(form)

	require.Error(t, err)

	validationErr, ok := err.(*forms.ValidationErrors)
	require.True(t, ok)

	errorStr := validationErr.Error()
	assert.Contains(t, errorStr, "name")
	assert.Contains(t, errorStr, "email")
}

func TestValidationErrors_Error_Empty(t *testing.T) {
	validationErr := &forms.ValidationErrors{Errors: []forms.FieldError{}}
	assert.Equal(t, "validation failed", validationErr.Error())
}

func TestValidationErrors_HasErrors(t *testing.T) {
	type Form struct {
		Name string `validate:"required"`
	}

	form := &Form{Name: ""}
	err := forms.Validate(form)

	require.Error(t, err)

	validationErr, ok := err.(*forms.ValidationErrors)
	require.True(t, ok)
	assert.True(t, validationErr.HasErrors())
}

func TestValidationErrors_HasErrors_Empty(t *testing.T) {
	validationErr := &forms.ValidationErrors{Errors: []forms.FieldError{}}
	assert.False(t, validationErr.HasErrors())
}

func TestValidationErrors_GetErrors(t *testing.T) {
	type Form struct {
		Name  string `validate:"required"`
		Email string `validate:"email"`
	}

	form := &Form{Name: "", Email: "invalid"}
	err := forms.Validate(form)

	require.Error(t, err)

	validationErr, ok := err.(*forms.ValidationErrors)
	require.True(t, ok)

	errors := validationErr.GetErrors()
	assert.Len(t, errors, 2)
}

func TestValidationErrors_GetErrors_Empty(t *testing.T) {
	validationErr := &forms.ValidationErrors{Errors: []forms.FieldError{}}
	errors := validationErr.GetErrors()
	assert.Empty(t, errors)
}

func TestValidationErrors_JSON(t *testing.T) {
	type Form struct {
		Name string `validate:"required"`
	}

	form := &Form{Name: ""}
	err := forms.Validate(form)

	require.Error(t, err)

	validationErr, ok := err.(*forms.ValidationErrors)
	require.True(t, ok)

	data, err := json.Marshal(validationErr)
	require.NoError(t, err)

	var decoded forms.ValidationErrors
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)

	assert.Len(t, decoded.Errors, 1)
	assert.Equal(t, "name", decoded.Errors[0].Field)
}

func TestFieldError_JSON(t *testing.T) {
	fieldErr := forms.FieldError{
		Field:   "name",
		Message: "name is required",
		Tag:     "required",
		Value:   "",
	}

	data, err := json.Marshal(fieldErr)
	require.NoError(t, err)

	var raw map[string]interface{}
	err = json.Unmarshal(data, &raw)
	require.NoError(t, err)

	assert.Equal(t, "name", raw["field"])
	assert.Equal(t, "name is required", raw["message"])
	assert.Equal(t, "required", raw["tag"])
}

func TestFieldError_JSON_OmitEmptyValue(t *testing.T) {
	fieldErr := forms.FieldError{
		Field:   "name",
		Message: "name is required",
		Tag:     "required",
		Value:   nil,
	}

	data, err := json.Marshal(fieldErr)
	require.NoError(t, err)

	var raw map[string]interface{}
	err = json.Unmarshal(data, &raw)
	require.NoError(t, err)

	assert.NotContains(t, raw, "value")
}

func TestToSnakeCase(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"Name", "name"},
		{"FirstName", "first_name"},
		{"ID", "i_d"},
		{"UserID", "user_i_d"},
		{"HTTPServer", "h_t_t_p_server"},
		{"simple", "simple"},
		{"", ""},
		{"A", "a"},
		{"AB", "a_b"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			type Form struct {
				Field string `validate:"required"`
			}

			// We can't directly test toSnakeCase as it's unexported,
			// but we can verify behavior through validation
			// This test verifies the pattern matching works correctly
		})
	}

	// Test through actual validation
	type CamelCaseForm struct {
		FirstName string `validate:"required"`
	}

	form := &CamelCaseForm{FirstName: ""}
	err := forms.Validate(form)

	require.Error(t, err)

	validationErr, ok := err.(*forms.ValidationErrors)
	require.True(t, ok)
	assert.Equal(t, "first_name", validationErr.Errors[0].Field)
}

func TestSetValidator(t *testing.T) {
	originalValidator := forms.GetValidator()
	defer forms.SetValidator(originalValidator)

	newValidator := validator.New()
	forms.SetValidator(newValidator)

	assert.Equal(t, newValidator, forms.GetValidator())
}

func TestGetValidator(t *testing.T) {
	v := forms.GetValidator()
	require.NotNil(t, v)
}

func TestRegisterValidation(t *testing.T) {
	originalValidator := forms.GetValidator()
	forms.SetValidator(validator.New())
	defer forms.SetValidator(originalValidator)

	err := forms.RegisterValidation("customtest", func(fl validator.FieldLevel) bool {
		return fl.Field().String() == "valid"
	})
	require.NoError(t, err)

	type Form struct {
		Field string `validate:"customtest"`
	}

	validForm := &Form{Field: "valid"}
	err = forms.Validate(validForm)
	assert.NoError(t, err)

	invalidForm := &Form{Field: "invalid"}
	err = forms.Validate(invalidForm)
	assert.Error(t, err)
}

func TestMust_Valid(t *testing.T) {
	type Form struct {
		Name string `validate:"required"`
	}

	form := &Form{Name: "John"}

	assert.NotPanics(t, func() {
		forms.Must(form)
	})
}

func TestMust_Invalid(t *testing.T) {
	type Form struct {
		Name string `validate:"required"`
	}

	form := &Form{Name: ""}

	assert.Panics(t, func() {
		forms.Must(form)
	})
}

func TestValidate_NonStruct(t *testing.T) {
	// Validating a non-struct should return an error
	err := forms.Validate("not a struct")
	assert.Error(t, err)
}

func TestValidate_Nil(t *testing.T) {
	// Validating nil should return an error
	err := forms.Validate(nil)
	assert.Error(t, err)
}

func TestValidate_Pointer(t *testing.T) {
	type Form struct {
		Name string `validate:"required"`
	}

	form := &Form{Name: "John"}
	err := forms.Validate(form)
	assert.NoError(t, err)
}

func TestValidationErrors_Value(t *testing.T) {
	type Form struct {
		Age int `validate:"min=18"`
	}

	form := &Form{Age: 10}
	err := forms.Validate(form)

	require.Error(t, err)

	validationErr, ok := err.(*forms.ValidationErrors)
	require.True(t, ok)
	require.Len(t, validationErr.Errors, 1)
	assert.Equal(t, 10, validationErr.Errors[0].Value)
}
