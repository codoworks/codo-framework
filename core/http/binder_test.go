package http

import (
	"bytes"
	"mime/multipart"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBinder_BindSingleFileHeader(t *testing.T) {
	e := echo.New()
	e.Binder = NewBinder()

	// Create multipart form with file
	body := new(bytes.Buffer)
	writer := multipart.NewWriter(body)

	fileWriter, err := writer.CreateFormFile("file", "test.txt")
	require.NoError(t, err)
	_, err = fileWriter.Write([]byte("test content"))
	require.NoError(t, err)

	err = writer.WriteField("description", "test description")
	require.NoError(t, err)
	writer.Close()

	req := httptest.NewRequest("POST", "/", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	type UploadForm struct {
		File        *multipart.FileHeader `form:"file"`
		Description string                `form:"description"`
	}

	var form UploadForm
	err = c.Bind(&form)

	assert.NoError(t, err)
	assert.NotNil(t, form.File)
	assert.Equal(t, "test.txt", form.File.Filename)
	assert.Equal(t, "test description", form.Description)
}

func TestBinder_BindMultipleFileHeaders(t *testing.T) {
	e := echo.New()
	e.Binder = NewBinder()

	body := new(bytes.Buffer)
	writer := multipart.NewWriter(body)

	// Add 3 files with same field name
	for i := 0; i < 3; i++ {
		fileWriter, err := writer.CreateFormFile("files", "test.txt")
		require.NoError(t, err)
		_, err = fileWriter.Write([]byte("content"))
		require.NoError(t, err)
	}
	writer.Close()

	req := httptest.NewRequest("POST", "/", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	type UploadForm struct {
		Files []*multipart.FileHeader `form:"files"`
	}

	var form UploadForm
	err := c.Bind(&form)

	assert.NoError(t, err)
	assert.Len(t, form.Files, 3)
}

func TestBinder_BindWithoutFile(t *testing.T) {
	e := echo.New()
	e.Binder = NewBinder()

	// Create multipart form without file
	body := new(bytes.Buffer)
	writer := multipart.NewWriter(body)
	err := writer.WriteField("description", "test description")
	require.NoError(t, err)
	writer.Close()

	req := httptest.NewRequest("POST", "/", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	type UploadForm struct {
		File        *multipart.FileHeader `form:"file"`
		Description string                `form:"description"`
	}

	var form UploadForm
	err = c.Bind(&form)

	assert.NoError(t, err)
	assert.Nil(t, form.File) // File should remain nil
	assert.Equal(t, "test description", form.Description)
}

func TestBinder_ValidationAfterBinding(t *testing.T) {
	e := echo.New()
	e.Binder = NewBinder()
	e.Validator = NewEchoValidator()

	// Empty form - no file
	body := new(bytes.Buffer)
	writer := multipart.NewWriter(body)
	writer.Close()

	req := httptest.NewRequest("POST", "/", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	rec := httptest.NewRecorder()
	c := &Context{Context: e.NewContext(req, rec)}

	type UploadForm struct {
		File *multipart.FileHeader `form:"file" validate:"required"`
	}

	var form UploadForm
	err := c.BindAndValidate(&form)

	assert.Error(t, err)
	// The error should be a validation error indicating file is required
	validationErr, ok := err.(*ValidationErrorList)
	assert.True(t, ok, "Expected ValidationErrorList")
	if ok && len(validationErr.Errors) > 0 {
		assert.Equal(t, "File", validationErr.Errors[0].Field)
		assert.Equal(t, "is required", validationErr.Errors[0].Message)
	}
}

func TestBinder_JSONBindingStillWorks(t *testing.T) {
	e := echo.New()
	e.Binder = NewBinder()

	// Test that regular JSON binding still works
	body := bytes.NewBufferString(`{"name":"test","value":123}`)

	req := httptest.NewRequest("POST", "/", body)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	type JSONForm struct {
		Name  string `json:"name"`
		Value int    `json:"value"`
	}

	var form JSONForm
	err := c.Bind(&form)

	assert.NoError(t, err)
	assert.Equal(t, "test", form.Name)
	assert.Equal(t, 123, form.Value)
}

func TestBinder_FormBindingStillWorks(t *testing.T) {
	e := echo.New()
	e.Binder = NewBinder()

	// Test that regular form binding (non-multipart) still works
	body := bytes.NewBufferString("name=test&value=123")

	req := httptest.NewRequest("POST", "/", body)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	type URLForm struct {
		Name  string `form:"name"`
		Value int    `form:"value"`
	}

	var form URLForm
	err := c.Bind(&form)

	assert.NoError(t, err)
	assert.Equal(t, "test", form.Name)
	assert.Equal(t, 123, form.Value)
}

func TestBinder_FileHeaderWithTagOptions(t *testing.T) {
	e := echo.New()
	e.Binder = NewBinder()

	// Create multipart form with file
	body := new(bytes.Buffer)
	writer := multipart.NewWriter(body)

	fileWriter, err := writer.CreateFormFile("avatar", "profile.png")
	require.NoError(t, err)
	_, err = fileWriter.Write([]byte("fake image content"))
	require.NoError(t, err)
	writer.Close()

	req := httptest.NewRequest("POST", "/", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Test that tag options like "omitempty" don't break parsing
	type ProfileForm struct {
		Avatar *multipart.FileHeader `form:"avatar,omitempty"`
	}

	var form ProfileForm
	err = c.Bind(&form)

	assert.NoError(t, err)
	assert.NotNil(t, form.Avatar)
	assert.Equal(t, "profile.png", form.Avatar.Filename)
}

func TestBinder_MixedFileAndFormFields(t *testing.T) {
	e := echo.New()
	e.Binder = NewBinder()

	body := new(bytes.Buffer)
	writer := multipart.NewWriter(body)

	// Add file
	fileWriter, err := writer.CreateFormFile("document", "report.pdf")
	require.NoError(t, err)
	_, err = fileWriter.Write([]byte("PDF content"))
	require.NoError(t, err)

	// Add various form fields
	err = writer.WriteField("title", "Monthly Report")
	require.NoError(t, err)
	err = writer.WriteField("category_id", "550e8400-e29b-41d4-a716-446655440000")
	require.NoError(t, err)
	err = writer.WriteField("is_public", "true")
	require.NoError(t, err)
	writer.Close()

	req := httptest.NewRequest("POST", "/", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	type DocumentUpload struct {
		Document   *multipart.FileHeader `form:"document"`
		Title      string                `form:"title"`
		CategoryID string                `form:"category_id"`
		IsPublic   bool                  `form:"is_public"`
	}

	var form DocumentUpload
	err = c.Bind(&form)

	assert.NoError(t, err)
	assert.NotNil(t, form.Document)
	assert.Equal(t, "report.pdf", form.Document.Filename)
	assert.Equal(t, "Monthly Report", form.Title)
	assert.Equal(t, "550e8400-e29b-41d4-a716-446655440000", form.CategoryID)
	assert.True(t, form.IsPublic)
}
