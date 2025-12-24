package http

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
)

func newTestContext(method, path string, body string) (*Context, *httptest.ResponseRecorder) {
	e := echo.New()
	var req *http.Request
	if body != "" {
		req = httptest.NewRequest(method, path, strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
	} else {
		req = httptest.NewRequest(method, path, nil)
	}
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	return &Context{Context: c}, rec
}

func TestContext_ParamUUID(t *testing.T) {
	t.Run("valid UUID", func(t *testing.T) {
		c, _ := newTestContext(http.MethodGet, "/users/123e4567-e89b-12d3-a456-426614174000", "")
		c.SetParamNames("id")
		c.SetParamValues("123e4567-e89b-12d3-a456-426614174000")

		id, err := c.ParamUUID("id")
		assert.NoError(t, err)
		assert.Equal(t, "123e4567-e89b-12d3-a456-426614174000", id)
	})

	t.Run("missing param", func(t *testing.T) {
		c, _ := newTestContext(http.MethodGet, "/users/", "")

		_, err := c.ParamUUID("id")
		assert.Error(t, err)
		assert.IsType(t, &ParamError{}, err)
	})

	t.Run("invalid UUID", func(t *testing.T) {
		c, _ := newTestContext(http.MethodGet, "/users/invalid", "")
		c.SetParamNames("id")
		c.SetParamValues("invalid")

		_, err := c.ParamUUID("id")
		assert.Error(t, err)
		assert.IsType(t, &ParamError{}, err)
	})
}

func TestContext_QueryInt(t *testing.T) {
	t.Run("valid int", func(t *testing.T) {
		c, _ := newTestContext(http.MethodGet, "/?page=5", "")
		val := c.QueryInt("page", 1)
		assert.Equal(t, 5, val)
	})

	t.Run("missing param uses default", func(t *testing.T) {
		c, _ := newTestContext(http.MethodGet, "/", "")
		val := c.QueryInt("page", 1)
		assert.Equal(t, 1, val)
	})

	t.Run("invalid int uses default", func(t *testing.T) {
		c, _ := newTestContext(http.MethodGet, "/?page=abc", "")
		val := c.QueryInt("page", 1)
		assert.Equal(t, 1, val)
	})
}

func TestContext_QueryInt64(t *testing.T) {
	t.Run("valid int64", func(t *testing.T) {
		c, _ := newTestContext(http.MethodGet, "/?offset=9223372036854775807", "")
		val := c.QueryInt64("offset", 0)
		assert.Equal(t, int64(9223372036854775807), val)
	})

	t.Run("missing param uses default", func(t *testing.T) {
		c, _ := newTestContext(http.MethodGet, "/", "")
		val := c.QueryInt64("offset", 100)
		assert.Equal(t, int64(100), val)
	})

	t.Run("invalid int64 uses default", func(t *testing.T) {
		c, _ := newTestContext(http.MethodGet, "/?offset=abc", "")
		val := c.QueryInt64("offset", 100)
		assert.Equal(t, int64(100), val)
	})
}

func TestContext_QueryBool(t *testing.T) {
	tests := []struct {
		query       string
		defaultVal  bool
		expected    bool
	}{
		{"?active=true", false, true},
		{"?active=false", true, false},
		{"?active=1", false, true},
		{"?active=0", true, false},
		{"", false, false},
		{"?active=invalid", true, true},
	}

	for _, tt := range tests {
		t.Run(tt.query, func(t *testing.T) {
			c, _ := newTestContext(http.MethodGet, "/"+tt.query, "")
			val := c.QueryBool("active", tt.defaultVal)
			assert.Equal(t, tt.expected, val)
		})
	}
}

func TestContext_Success(t *testing.T) {
	c, rec := newTestContext(http.MethodGet, "/", "")

	payload := map[string]string{"message": "hello"}
	err := c.Success(payload)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), `"code":"OK"`)
}

func TestContext_Created(t *testing.T) {
	c, rec := newTestContext(http.MethodPost, "/", "")

	payload := map[string]string{"id": "123"}
	err := c.Created(payload)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusCreated, rec.Code)
	assert.Contains(t, rec.Body.String(), `"code":"CREATED"`)
}

func TestContext_NoContent(t *testing.T) {
	c, rec := newTestContext(http.MethodDelete, "/", "")

	err := c.NoContent()

	assert.NoError(t, err)
	assert.Equal(t, http.StatusNoContent, rec.Code)
}

func TestContext_SendError(t *testing.T) {
	c, rec := newTestContext(http.MethodGet, "/", "")

	err := c.SendError(&ParamError{Param: "id", Message: "invalid"})

	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestContext_GetRequestID(t *testing.T) {
	t.Run("from request header", func(t *testing.T) {
		e := echo.New()
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("X-Request-ID", "test-id-123")
		rec := httptest.NewRecorder()
		c := &Context{Context: e.NewContext(req, rec)}

		assert.Equal(t, "test-id-123", c.GetRequestID())
	})

	t.Run("from response header", func(t *testing.T) {
		e := echo.New()
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rec := httptest.NewRecorder()
		c := &Context{Context: e.NewContext(req, rec)}
		c.Response().Header().Set("X-Request-ID", "resp-id-456")

		assert.Equal(t, "resp-id-456", c.GetRequestID())
	})

	t.Run("no request ID", func(t *testing.T) {
		c, _ := newTestContext(http.MethodGet, "/", "")
		assert.Equal(t, "", c.GetRequestID())
	})
}

func TestContext_RealIP(t *testing.T) {
	c, _ := newTestContext(http.MethodGet, "/", "")
	ip := c.RealIP()
	assert.NotEmpty(t, ip)
}

func TestBindError_Error(t *testing.T) {
	err := &BindError{Cause: echo.ErrBadRequest}
	assert.Contains(t, err.Error(), "binding error")
}

func TestBindError_Unwrap(t *testing.T) {
	cause := echo.ErrBadRequest
	err := &BindError{Cause: cause}
	assert.Equal(t, cause, err.Unwrap())
}

func TestParamError_Error(t *testing.T) {
	err := &ParamError{Param: "id", Message: "required"}
	assert.Equal(t, "parameter id: required", err.Error())
}

func TestContext_BindAndValidate(t *testing.T) {
	type testForm struct {
		Name string `json:"name" validate:"required"`
	}

	t.Run("success", func(t *testing.T) {
		e := echo.New()
		e.Validator = &testContextValidator{}
		req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{"name":"test"}`))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		c := &Context{Context: e.NewContext(req, rec)}

		var form testForm
		err := c.BindAndValidate(&form)
		assert.NoError(t, err)
		assert.Equal(t, "test", form.Name)
	})

	t.Run("bind error", func(t *testing.T) {
		e := echo.New()
		req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`invalid json`))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		c := &Context{Context: e.NewContext(req, rec)}

		var form testForm
		err := c.BindAndValidate(&form)
		assert.Error(t, err)
		assert.IsType(t, &BindError{}, err)
	})

	t.Run("validation error", func(t *testing.T) {
		e := echo.New()
		e.Validator = &testContextValidator{shouldFail: true}
		req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{"name":""}`))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		c := &Context{Context: e.NewContext(req, rec)}

		var form testForm
		err := c.BindAndValidate(&form)
		assert.Error(t, err)
	})
}

type testContextValidator struct {
	shouldFail bool
}

func (v *testContextValidator) Validate(i interface{}) error {
	if v.shouldFail {
		return echo.NewHTTPError(http.StatusBadRequest, "validation failed")
	}
	return nil
}

func TestContext_SendError_StrictMode(t *testing.T) {
	t.Run("strict mode disabled - errors can be null", func(t *testing.T) {
		// Save and restore original config
		originalCfg := GetHandlerConfig()
		defer SetHandlerConfig(originalCfg)

		SetHandlerConfig(HandlerConfig{StrictResponse: false})

		c, rec := newTestContext(http.MethodGet, "/", "")
		err := c.SendError(&ParamError{Param: "id", Message: "required"})

		assert.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, rec.Code)
		// In non-strict mode, empty arrays may be null/omitted
		body := rec.Body.String()
		assert.Contains(t, body, `"code":"BAD_REQUEST"`)
	})

	t.Run("strict mode enabled - errors is empty array", func(t *testing.T) {
		// Save and restore original config
		originalCfg := GetHandlerConfig()
		defer SetHandlerConfig(originalCfg)

		SetHandlerConfig(HandlerConfig{StrictResponse: true})

		c, rec := newTestContext(http.MethodGet, "/", "")
		err := c.SendError(&ParamError{Param: "id", Message: "required"})

		assert.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, rec.Code)
		// In strict mode, errors must be an empty array not null
		body := rec.Body.String()
		assert.Contains(t, body, `"code":"BAD_REQUEST"`)
		assert.Contains(t, body, `"errors":[]`)
		assert.Contains(t, body, `"warnings":[]`)
	})
}
