package forms_test

import (
	"testing"
	"time"

	"github.com/codoworks/codo-framework/core/forms"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestModel is a sample model for testing
type TestModel struct {
	ID        string
	Name      string
	Email     string
	CreatedAt time.Time
	UpdatedAt time.Time
}

// TestCreateForm implements CreateForm[TestModel]
type TestCreateForm struct {
	Name  string `json:"name" validate:"required,min=2"`
	Email string `json:"email" validate:"required,email"`
}

func (f *TestCreateForm) ToModel() *TestModel {
	return &TestModel{
		Name:  f.Name,
		Email: f.Email,
	}
}

func (f *TestCreateForm) ApplyTo(m *TestModel) {
	m.Name = f.Name
	m.Email = f.Email
}

func (f *TestCreateForm) Validate() error {
	return forms.Validate(f)
}

// Compile-time interface check
var _ forms.CreateForm[TestModel] = (*TestCreateForm)(nil)

// TestUpdateForm implements UpdateForm[TestModel]
type TestUpdateForm struct {
	Name  *string `json:"name,omitempty" validate:"omitempty,min=2"`
	Email *string `json:"email,omitempty" validate:"omitempty,email"`
}

func (f *TestUpdateForm) ApplyTo(m *TestModel) {
	if f.Name != nil {
		m.Name = *f.Name
	}
	if f.Email != nil {
		m.Email = *f.Email
	}
}

func (f *TestUpdateForm) Validate() error {
	return forms.Validate(f)
}

// Compile-time interface check
var _ forms.UpdateForm[TestModel] = (*TestUpdateForm)(nil)

// TestResponseForm implements ResponseForm[TestModel]
type TestResponseForm struct {
	forms.FormBase
	Name  string `json:"name"`
	Email string `json:"email"`
}

func (f *TestResponseForm) FromModel(m *TestModel) forms.ResponseForm[TestModel] {
	f.ID = m.ID
	f.Name = m.Name
	f.Email = m.Email
	f.CreatedAt = m.CreatedAt
	f.UpdatedAt = m.UpdatedAt
	return f
}

// Compile-time interface check
var _ forms.ResponseForm[TestModel] = (*TestResponseForm)(nil)

// TestMapper implements Mapper[TestCreateForm, TestModel]
type TestMapper struct{}

func (m *TestMapper) ToModel(form *TestCreateForm) *TestModel {
	return form.ToModel()
}

func (m *TestMapper) ToForm(model *TestModel) *TestCreateForm {
	return &TestCreateForm{
		Name:  model.Name,
		Email: model.Email,
	}
}

func (m *TestMapper) ToModels(formList []*TestCreateForm) []*TestModel {
	models := make([]*TestModel, len(formList))
	for i, f := range formList {
		models[i] = m.ToModel(f)
	}
	return models
}

func (m *TestMapper) ToForms(models []*TestModel) []*TestCreateForm {
	formList := make([]*TestCreateForm, len(models))
	for i, model := range models {
		formList[i] = m.ToForm(model)
	}
	return formList
}

// Compile-time interface check
var _ forms.Mapper[TestCreateForm, TestModel] = (*TestMapper)(nil)

func TestRequestForm_ToModel(t *testing.T) {
	form := &TestCreateForm{
		Name:  "John Doe",
		Email: "john@example.com",
	}

	model := form.ToModel()

	require.NotNil(t, model)
	assert.Equal(t, "John Doe", model.Name)
	assert.Equal(t, "john@example.com", model.Email)
	assert.Empty(t, model.ID)
}

func TestRequestForm_ApplyTo(t *testing.T) {
	model := &TestModel{
		ID:    "123",
		Name:  "Old Name",
		Email: "old@example.com",
	}

	form := &TestCreateForm{
		Name:  "New Name",
		Email: "new@example.com",
	}

	form.ApplyTo(model)

	assert.Equal(t, "123", model.ID)
	assert.Equal(t, "New Name", model.Name)
	assert.Equal(t, "new@example.com", model.Email)
}

func TestResponseForm_FromModel(t *testing.T) {
	now := time.Now()
	model := &TestModel{
		ID:        "123",
		Name:      "John Doe",
		Email:     "john@example.com",
		CreatedAt: now,
		UpdatedAt: now,
	}

	form := &TestResponseForm{}
	result := form.FromModel(model)

	require.NotNil(t, result)
	assert.Equal(t, "123", form.ID)
	assert.Equal(t, "John Doe", form.Name)
	assert.Equal(t, "john@example.com", form.Email)
	assert.Equal(t, now, form.CreatedAt)
	assert.Equal(t, now, form.UpdatedAt)
}

func TestCreateForm_Validate_Valid(t *testing.T) {
	form := &TestCreateForm{
		Name:  "John Doe",
		Email: "john@example.com",
	}

	err := form.Validate()
	assert.NoError(t, err)
}

func TestCreateForm_Validate_Invalid(t *testing.T) {
	form := &TestCreateForm{
		Name:  "",
		Email: "invalid-email",
	}

	err := form.Validate()
	require.Error(t, err)

	validationErr, ok := err.(*forms.ValidationErrors)
	require.True(t, ok)
	assert.True(t, validationErr.HasErrors())
	assert.GreaterOrEqual(t, len(validationErr.Errors), 2)
}

func TestUpdateForm_ApplyTo_PartialUpdate(t *testing.T) {
	model := &TestModel{
		ID:    "123",
		Name:  "Old Name",
		Email: "old@example.com",
	}

	newName := "New Name"
	form := &TestUpdateForm{
		Name:  &newName,
		Email: nil,
	}

	form.ApplyTo(model)

	assert.Equal(t, "123", model.ID)
	assert.Equal(t, "New Name", model.Name)
	assert.Equal(t, "old@example.com", model.Email)
}

func TestUpdateForm_ApplyTo_FullUpdate(t *testing.T) {
	model := &TestModel{
		ID:    "123",
		Name:  "Old Name",
		Email: "old@example.com",
	}

	newName := "New Name"
	newEmail := "new@example.com"
	form := &TestUpdateForm{
		Name:  &newName,
		Email: &newEmail,
	}

	form.ApplyTo(model)

	assert.Equal(t, "New Name", model.Name)
	assert.Equal(t, "new@example.com", model.Email)
}

func TestUpdateForm_ApplyTo_NoUpdate(t *testing.T) {
	model := &TestModel{
		ID:    "123",
		Name:  "Old Name",
		Email: "old@example.com",
	}

	form := &TestUpdateForm{
		Name:  nil,
		Email: nil,
	}

	form.ApplyTo(model)

	assert.Equal(t, "Old Name", model.Name)
	assert.Equal(t, "old@example.com", model.Email)
}

func TestUpdateForm_Validate_Valid(t *testing.T) {
	name := "John"
	email := "john@example.com"
	form := &TestUpdateForm{
		Name:  &name,
		Email: &email,
	}

	err := form.Validate()
	assert.NoError(t, err)
}

func TestUpdateForm_Validate_Empty(t *testing.T) {
	form := &TestUpdateForm{
		Name:  nil,
		Email: nil,
	}

	err := form.Validate()
	assert.NoError(t, err)
}

func TestMapper_ToModel(t *testing.T) {
	mapper := &TestMapper{}
	form := &TestCreateForm{
		Name:  "John",
		Email: "john@example.com",
	}

	model := mapper.ToModel(form)

	require.NotNil(t, model)
	assert.Equal(t, "John", model.Name)
	assert.Equal(t, "john@example.com", model.Email)
}

func TestMapper_ToForm(t *testing.T) {
	mapper := &TestMapper{}
	model := &TestModel{
		ID:    "123",
		Name:  "John",
		Email: "john@example.com",
	}

	form := mapper.ToForm(model)

	require.NotNil(t, form)
	assert.Equal(t, "John", form.Name)
	assert.Equal(t, "john@example.com", form.Email)
}

func TestMapper_ToModels(t *testing.T) {
	mapper := &TestMapper{}
	formList := []*TestCreateForm{
		{Name: "John", Email: "john@example.com"},
		{Name: "Jane", Email: "jane@example.com"},
	}

	models := mapper.ToModels(formList)

	require.Len(t, models, 2)
	assert.Equal(t, "John", models[0].Name)
	assert.Equal(t, "Jane", models[1].Name)
}

func TestMapper_ToForms(t *testing.T) {
	mapper := &TestMapper{}
	models := []*TestModel{
		{ID: "1", Name: "John", Email: "john@example.com"},
		{ID: "2", Name: "Jane", Email: "jane@example.com"},
	}

	formList := mapper.ToForms(models)

	require.Len(t, formList, 2)
	assert.Equal(t, "John", formList[0].Name)
	assert.Equal(t, "Jane", formList[1].Name)
}

func TestMapper_ToModels_Empty(t *testing.T) {
	mapper := &TestMapper{}
	formList := []*TestCreateForm{}

	models := mapper.ToModels(formList)

	require.NotNil(t, models)
	assert.Len(t, models, 0)
}

func TestMapper_ToForms_Empty(t *testing.T) {
	mapper := &TestMapper{}
	models := []*TestModel{}

	formList := mapper.ToForms(models)

	require.NotNil(t, formList)
	assert.Len(t, formList, 0)
}

// TestValidatable ensures the interface can be satisfied
func TestValidatable_Interface(t *testing.T) {
	var v forms.Validatable = &TestCreateForm{Name: "test", Email: "test@example.com"}
	err := v.Validate()
	assert.NoError(t, err)
}
