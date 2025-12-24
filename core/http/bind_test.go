package http

import (
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/assert"
)

type testForm struct {
	Email    string `json:"email" validate:"required,email"`
	Name     string `json:"name" validate:"required,min=2,max=50"`
	Age      int    `json:"age" validate:"gte=0,lte=150"`
	Password string `json:"password" validate:"required,min=8"`
	Website  string `json:"website" validate:"omitempty,url"`
	UserID   string `json:"user_id" validate:"omitempty,uuid"`
	Role     string `json:"role" validate:"omitempty,oneof=admin user guest"`
	Code     string `json:"code" validate:"omitempty,numeric"`
	Label    string `json:"label" validate:"omitempty,alpha"`
	Tag      string `json:"tag" validate:"omitempty,alphanum"`
}

func TestValidate_Success(t *testing.T) {
	form := &testForm{
		Email:    "test@example.com",
		Name:     "John",
		Age:      30,
		Password: "password123",
	}

	err := Validate(form)
	assert.NoError(t, err)
}

func TestValidate_Required(t *testing.T) {
	form := &testForm{}

	err := Validate(form)
	assert.Error(t, err)
	assert.IsType(t, &ValidationErrorList{}, err)

	verr := err.(*ValidationErrorList)
	assert.NotEmpty(t, verr.Errors)
	assert.Contains(t, verr.Errors[0].Message, "is required")
}

func TestValidate_Email(t *testing.T) {
	form := &testForm{
		Email:    "invalid-email",
		Name:     "John",
		Password: "password123",
	}

	err := Validate(form)
	assert.Error(t, err)

	verr := err.(*ValidationErrorList)
	foundEmailError := false
	for _, e := range verr.Errors {
		if e.Field == "email" && e.Message == "must be a valid email" {
			foundEmailError = true
			break
		}
	}
	assert.True(t, foundEmailError)
}

func TestValidate_Min(t *testing.T) {
	form := &testForm{
		Email:    "test@example.com",
		Name:     "J", // Too short
		Password: "password123",
	}

	err := Validate(form)
	assert.Error(t, err)

	verr := err.(*ValidationErrorList)
	foundMinError := false
	for _, e := range verr.Errors {
		if e.Field == "name" && e.Message == "must be at least 2 characters" {
			foundMinError = true
			break
		}
	}
	assert.True(t, foundMinError)
}

func TestValidate_Max(t *testing.T) {
	form := &testForm{
		Email:    "test@example.com",
		Name:     "This is a very long name that exceeds the maximum allowed length for the name field",
		Password: "password123",
	}

	err := Validate(form)
	assert.Error(t, err)

	verr := err.(*ValidationErrorList)
	foundMaxError := false
	for _, e := range verr.Errors {
		if e.Field == "name" && e.Message == "must be at most 50 characters" {
			foundMaxError = true
			break
		}
	}
	assert.True(t, foundMaxError)
}

func TestValidate_Gte(t *testing.T) {
	form := &testForm{
		Email:    "test@example.com",
		Name:     "John",
		Password: "password123",
		Age:      -1, // Below 0
	}

	err := Validate(form)
	assert.Error(t, err)

	verr := err.(*ValidationErrorList)
	foundGteError := false
	for _, e := range verr.Errors {
		if e.Field == "age" && e.Message == "must be greater than or equal to 0" {
			foundGteError = true
			break
		}
	}
	assert.True(t, foundGteError)
}

func TestValidate_Lte(t *testing.T) {
	form := &testForm{
		Email:    "test@example.com",
		Name:     "John",
		Password: "password123",
		Age:      200, // Above 150
	}

	err := Validate(form)
	assert.Error(t, err)

	verr := err.(*ValidationErrorList)
	foundLteError := false
	for _, e := range verr.Errors {
		if e.Field == "age" && e.Message == "must be less than or equal to 150" {
			foundLteError = true
			break
		}
	}
	assert.True(t, foundLteError)
}

func TestValidate_URL(t *testing.T) {
	form := &testForm{
		Email:    "test@example.com",
		Name:     "John",
		Password: "password123",
		Website:  "not-a-url",
	}

	err := Validate(form)
	assert.Error(t, err)

	verr := err.(*ValidationErrorList)
	foundURLError := false
	for _, e := range verr.Errors {
		if e.Field == "website" && e.Message == "must be a valid URL" {
			foundURLError = true
			break
		}
	}
	assert.True(t, foundURLError)
}

func TestValidate_UUID(t *testing.T) {
	form := &testForm{
		Email:    "test@example.com",
		Name:     "John",
		Password: "password123",
		UserID:   "not-a-uuid",
	}

	err := Validate(form)
	assert.Error(t, err)

	verr := err.(*ValidationErrorList)
	foundUUIDError := false
	for _, e := range verr.Errors {
		if e.Field == "user_id" && e.Message == "must be a valid UUID" {
			foundUUIDError = true
			break
		}
	}
	assert.True(t, foundUUIDError)
}

func TestValidate_OneOf(t *testing.T) {
	form := &testForm{
		Email:    "test@example.com",
		Name:     "John",
		Password: "password123",
		Role:     "invalid",
	}

	err := Validate(form)
	assert.Error(t, err)

	verr := err.(*ValidationErrorList)
	foundOneOfError := false
	for _, e := range verr.Errors {
		if e.Field == "role" && e.Message == "must be one of: admin user guest" {
			foundOneOfError = true
			break
		}
	}
	assert.True(t, foundOneOfError)
}

func TestValidate_Numeric(t *testing.T) {
	form := &testForm{
		Email:    "test@example.com",
		Name:     "John",
		Password: "password123",
		Code:     "abc",
	}

	err := Validate(form)
	assert.Error(t, err)

	verr := err.(*ValidationErrorList)
	foundNumericError := false
	for _, e := range verr.Errors {
		if e.Field == "code" && e.Message == "must be numeric" {
			foundNumericError = true
			break
		}
	}
	assert.True(t, foundNumericError)
}

func TestValidate_Alpha(t *testing.T) {
	form := &testForm{
		Email:    "test@example.com",
		Name:     "John",
		Password: "password123",
		Label:    "abc123",
	}

	err := Validate(form)
	assert.Error(t, err)

	verr := err.(*ValidationErrorList)
	foundAlphaError := false
	for _, e := range verr.Errors {
		if e.Field == "label" && e.Message == "must contain only letters" {
			foundAlphaError = true
			break
		}
	}
	assert.True(t, foundAlphaError)
}

func TestValidate_AlphaNum(t *testing.T) {
	form := &testForm{
		Email:    "test@example.com",
		Name:     "John",
		Password: "password123",
		Tag:      "abc-123",
	}

	err := Validate(form)
	assert.Error(t, err)

	verr := err.(*ValidationErrorList)
	foundAlphaNumError := false
	for _, e := range verr.Errors {
		if e.Field == "tag" && e.Message == "must contain only letters and numbers" {
			foundAlphaNumError = true
			break
		}
	}
	assert.True(t, foundAlphaNumError)
}

func TestValidationErrorList_Error(t *testing.T) {
	t.Run("with errors", func(t *testing.T) {
		v := &ValidationErrorList{Errors: []ValidationError{
			{Field: "field1", Message: "error 1"},
			{Field: "field2", Message: "error 2"},
		}}
		assert.Equal(t, "field1 error 1", v.Error())
	})

	t.Run("empty errors", func(t *testing.T) {
		v := &ValidationErrorList{Errors: []ValidationError{}}
		assert.Equal(t, "validation failed", v.Error())
	})
}

func TestSetValidator(t *testing.T) {
	original := GetValidator()
	defer SetValidator(original)

	newValidator := GetValidator()
	SetValidator(newValidator)

	assert.Equal(t, newValidator, GetValidator())
}

func TestGetValidator(t *testing.T) {
	v := GetValidator()
	assert.NotNil(t, v)
}

func TestRegisterValidation(t *testing.T) {
	// Register a custom validation
	err := RegisterValidation("customtest", func(fl validator.FieldLevel) bool {
		return fl.Field().String() == "valid"
	})
	assert.NoError(t, err)

	// Test using the custom validation
	type form struct {
		Field string `json:"field" validate:"customtest"`
	}

	t.Run("custom validation passes", func(t *testing.T) {
		f := &form{Field: "valid"}
		err := Validate(f)
		assert.NoError(t, err)
	})

	t.Run("custom validation fails", func(t *testing.T) {
		f := &form{Field: "invalid"}
		err := Validate(f)
		assert.Error(t, err)
	})
}

func TestValidate_JSONTagFieldName(t *testing.T) {
	// Test that validation errors use JSON field names, not Go field names
	type form struct {
		FirstName string `json:"first_name" validate:"required"`
	}

	f := &form{}
	err := Validate(f)
	assert.Error(t, err)

	verr, ok := err.(*ValidationErrorList)
	assert.True(t, ok)
	// Should use json tag name "first_name" not "FirstName"
	assert.Equal(t, "first_name", verr.Errors[0].Field)
}

func TestValidate_NoJSONTag(t *testing.T) {
	// Test that validation falls back to field name when no json tag
	type form struct {
		FieldWithoutTag string `validate:"required"`
	}

	f := &form{}
	err := Validate(f)
	assert.Error(t, err)

	verr, ok := err.(*ValidationErrorList)
	assert.True(t, ok)
	assert.Equal(t, "FieldWithoutTag", verr.Errors[0].Field)
}

func TestValidate_JSONTagDash(t *testing.T) {
	// Test that "-" json tag falls back to field name
	type form struct {
		HiddenField string `json:"-" validate:"required"`
	}

	f := &form{}
	err := Validate(f)
	assert.Error(t, err)

	verr, ok := err.(*ValidationErrorList)
	assert.True(t, ok)
	assert.Equal(t, "HiddenField", verr.Errors[0].Field)
}

func TestValidate_GtLtValidation(t *testing.T) {
	type form struct {
		Count int `json:"count" validate:"gt=5,lt=10"`
	}

	t.Run("gt fails", func(t *testing.T) {
		f := &form{Count: 3}
		err := Validate(f)
		assert.Error(t, err)
		verr := err.(*ValidationErrorList)
		assert.Contains(t, verr.Errors[0].Message, "greater than")
	})

	t.Run("lt fails", func(t *testing.T) {
		f := &form{Count: 15}
		err := Validate(f)
		assert.Error(t, err)
		verr := err.(*ValidationErrorList)
		assert.Contains(t, verr.Errors[0].Message, "less than")
	})
}

func TestValidate_NonValidationError(t *testing.T) {
	// Validate with nil should return error
	err := Validate(nil)
	assert.Error(t, err)
	// Should not be ValidationErrorList
	_, ok := err.(*ValidationErrorList)
	assert.False(t, ok)
}

func TestValidate_DefaultValidation(t *testing.T) {
	// Test an unknown/default validation error format
	type form struct {
		Field string `json:"field" validate:"len=5"`
	}

	f := &form{Field: "abc"} // Wrong length
	err := Validate(f)
	assert.Error(t, err)
	verr, ok := err.(*ValidationErrorList)
	assert.True(t, ok)
	assert.Contains(t, verr.Errors[0].Message, "failed")
}

func TestNewEchoValidator(t *testing.T) {
	v := NewEchoValidator()
	assert.NotNil(t, v)

	type form struct {
		Name string `json:"name" validate:"required"`
	}

	t.Run("validates successfully", func(t *testing.T) {
		f := &form{Name: "test"}
		err := v.Validate(f)
		assert.NoError(t, err)
	})

	t.Run("returns validation error", func(t *testing.T) {
		f := &form{Name: ""}
		err := v.Validate(f)
		assert.Error(t, err)
		assert.IsType(t, &ValidationErrorList{}, err)
	})
}
