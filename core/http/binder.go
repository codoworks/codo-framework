package http

import (
	"mime/multipart"
	"reflect"
	"strings"

	"github.com/labstack/echo/v4"
)

// Binder extends Echo's DefaultBinder with multipart file header support.
// It automatically binds *multipart.FileHeader and []*multipart.FileHeader
// struct fields from multipart form data.
type Binder struct {
	defaultBinder echo.DefaultBinder
}

// NewBinder creates a new Binder instance.
func NewBinder() *Binder {
	return &Binder{}
}

// Bind implements echo.Binder interface.
// It first delegates to Echo's DefaultBinder for standard fields,
// then handles multipart file header binding.
func (b *Binder) Bind(i interface{}, c echo.Context) error {
	// First, use Echo's default binder for standard fields (JSON, form values, query params)
	if err := b.defaultBinder.Bind(i, c); err != nil {
		return err
	}

	// Then, handle multipart file headers if this is a multipart request
	contentType := c.Request().Header.Get("Content-Type")
	if strings.HasPrefix(contentType, "multipart/form-data") {
		if err := b.bindFileHeaders(i, c); err != nil {
			return err
		}
	}

	return nil
}

// bindFileHeaders binds *multipart.FileHeader and []*multipart.FileHeader
// struct fields from the multipart form.
func (b *Binder) bindFileHeaders(i interface{}, c echo.Context) error {
	// Get the multipart form (parses if not already parsed)
	form, err := c.MultipartForm()
	if err != nil {
		// No multipart form available, nothing to bind
		return nil
	}

	val := reflect.ValueOf(i)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}
	if val.Kind() != reflect.Struct {
		return nil
	}

	typ := val.Type()
	for i := 0; i < val.NumField(); i++ {
		field := val.Field(i)
		fieldType := typ.Field(i)

		// Skip unexported fields
		if !field.CanSet() {
			continue
		}

		// Get form field name from tag or use lowercase field name
		formTag := fieldType.Tag.Get("form")
		if formTag == "" || formTag == "-" {
			formTag = strings.ToLower(fieldType.Name)
		}
		// Handle tag options like `form:"file,omitempty"`
		if idx := strings.Index(formTag, ","); idx != -1 {
			formTag = formTag[:idx]
		}

		// Handle *multipart.FileHeader (single file)
		if field.Type() == reflect.TypeOf((*multipart.FileHeader)(nil)) {
			if files, ok := form.File[formTag]; ok && len(files) > 0 {
				field.Set(reflect.ValueOf(files[0]))
			}
			continue
		}

		// Handle []*multipart.FileHeader (multiple files)
		if field.Type() == reflect.TypeOf([]*multipart.FileHeader{}) {
			if files, ok := form.File[formTag]; ok {
				field.Set(reflect.ValueOf(files))
			}
			continue
		}
	}

	return nil
}
