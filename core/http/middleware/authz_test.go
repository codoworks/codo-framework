package middleware

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"

	"github.com/codoworks/codo-framework/core/auth"
	"github.com/codoworks/codo-framework/core/clients/keto"
)

func setupAuthzTest(t *testing.T, mockKeto *keto.MockClient, objectFunc func(c echo.Context) string) (*echo.Echo, echo.MiddlewareFunc) {
	cfg := &AuthzConfig{
		Keto:       mockKeto,
		Namespace:  "documents",
		Relation:   "viewer",
		ObjectFunc: objectFunc,
	}

	return echo.New(), Authz(cfg)
}

func TestAuthzMiddleware(t *testing.T) {
	mockKeto := keto.NewMockClient()
	mockKeto.CheckFunc = func(ctx context.Context, subject, relation, namespace, object string) (bool, error) {
		assert.Equal(t, "user-123", subject)
		assert.Equal(t, "viewer", relation)
		assert.Equal(t, "documents", namespace)
		assert.Equal(t, "doc-456", object)
		return true, nil
	}

	e, authzMiddleware := setupAuthzTest(t, mockKeto, func(c echo.Context) string {
		return "doc-456"
	})

	e.Use(authzMiddleware)
	e.GET("/test", func(c echo.Context) error {
		return c.String(http.StatusOK, "OK")
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()

	// Set identity in context
	ctx := auth.ContextWithIdentity(req.Context(), &auth.Identity{ID: "user-123"})
	req = req.WithContext(ctx)

	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestAuthzMiddleware_Denied(t *testing.T) {
	mockKeto := keto.NewMockClient()
	mockKeto.CheckFunc = func(ctx context.Context, subject, relation, namespace, object string) (bool, error) {
		return false, nil
	}

	e, authzMiddleware := setupAuthzTest(t, mockKeto, func(c echo.Context) string {
		return "doc-456"
	})

	e.Use(authzMiddleware)
	e.GET("/test", func(c echo.Context) error {
		return c.String(http.StatusOK, "OK")
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()

	ctx := auth.ContextWithIdentity(req.Context(), &auth.Identity{ID: "user-123"})
	req = req.WithContext(ctx)

	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusForbidden, rec.Code)
	assert.Contains(t, rec.Body.String(), "Permission denied")
}

func TestAuthzMiddleware_NoIdentity(t *testing.T) {
	mockKeto := keto.NewMockClient()

	e, authzMiddleware := setupAuthzTest(t, mockKeto, func(c echo.Context) string {
		return "doc-456"
	})

	e.Use(authzMiddleware)
	e.GET("/test", func(c echo.Context) error {
		return c.String(http.StatusOK, "OK")
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()
	// No identity set

	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
	assert.Contains(t, rec.Body.String(), "No identity in context")
}

func TestAuthzMiddleware_NoObject(t *testing.T) {
	mockKeto := keto.NewMockClient()

	e, authzMiddleware := setupAuthzTest(t, mockKeto, func(c echo.Context) string {
		return "" // empty object
	})

	e.Use(authzMiddleware)
	e.GET("/test", func(c echo.Context) error {
		return c.String(http.StatusOK, "OK")
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()

	ctx := auth.ContextWithIdentity(req.Context(), &auth.Identity{ID: "user-123"})
	req = req.WithContext(ctx)

	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
	assert.Contains(t, rec.Body.String(), "No object specified")
}

func TestAuthzMiddleware_Error(t *testing.T) {
	mockKeto := keto.NewMockClient()
	mockKeto.CheckFunc = func(ctx context.Context, subject, relation, namespace, object string) (bool, error) {
		return false, fmt.Errorf("connection error")
	}

	e, authzMiddleware := setupAuthzTest(t, mockKeto, func(c echo.Context) string {
		return "doc-456"
	})

	e.Use(authzMiddleware)
	e.GET("/test", func(c echo.Context) error {
		return c.String(http.StatusOK, "OK")
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()

	ctx := auth.ContextWithIdentity(req.Context(), &auth.Identity{ID: "user-123"})
	req = req.WithContext(ctx)

	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusInternalServerError, rec.Code)
	assert.Contains(t, rec.Body.String(), "Permission check failed")
}

func TestRequirePermission(t *testing.T) {
	mockKeto := keto.NewMockClient()
	mockKeto.CheckFunc = func(ctx context.Context, subject, relation, namespace, object string) (bool, error) {
		assert.Equal(t, "user-123", subject)
		assert.Equal(t, "editor", relation)
		assert.Equal(t, "projects", namespace)
		assert.Equal(t, "proj-789", object)
		return true, nil
	}

	middleware := RequirePermission(mockKeto, "projects", "editor", func(c echo.Context) string {
		return "proj-789"
	})

	e := echo.New()
	e.Use(middleware)
	e.GET("/test", func(c echo.Context) error {
		return c.String(http.StatusOK, "OK")
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()

	ctx := auth.ContextWithIdentity(req.Context(), &auth.Identity{ID: "user-123"})
	req = req.WithContext(ctx)

	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestObjectFromParam(t *testing.T) {
	e := echo.New()
	e.GET("/documents/:id", func(c echo.Context) error {
		objectFunc := ObjectFromParam("id")
		result := objectFunc(c)
		return c.String(http.StatusOK, result)
	})

	req := httptest.NewRequest(http.MethodGet, "/documents/doc-123", nil)
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "doc-123", rec.Body.String())
}

func TestObjectFromParam_Missing(t *testing.T) {
	e := echo.New()
	e.GET("/documents", func(c echo.Context) error {
		objectFunc := ObjectFromParam("id")
		result := objectFunc(c)
		return c.String(http.StatusOK, result)
	})

	req := httptest.NewRequest(http.MethodGet, "/documents", nil)
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "", rec.Body.String())
}

func TestObjectFromQuery(t *testing.T) {
	e := echo.New()
	e.GET("/documents", func(c echo.Context) error {
		objectFunc := ObjectFromQuery("doc_id")
		result := objectFunc(c)
		return c.String(http.StatusOK, result)
	})

	req := httptest.NewRequest(http.MethodGet, "/documents?doc_id=doc-456", nil)
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "doc-456", rec.Body.String())
}

func TestObjectFromQuery_Missing(t *testing.T) {
	e := echo.New()
	e.GET("/documents", func(c echo.Context) error {
		objectFunc := ObjectFromQuery("doc_id")
		result := objectFunc(c)
		return c.String(http.StatusOK, result)
	})

	req := httptest.NewRequest(http.MethodGet, "/documents", nil)
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "", rec.Body.String())
}

func TestAuthzConfig_Fields(t *testing.T) {
	mockKeto := keto.NewMockClient()
	objectFunc := func(c echo.Context) string { return "obj" }

	cfg := &AuthzConfig{
		Keto:       mockKeto,
		Namespace:  "test-namespace",
		Relation:   "test-relation",
		ObjectFunc: objectFunc,
	}

	assert.Equal(t, mockKeto, cfg.Keto)
	assert.Equal(t, "test-namespace", cfg.Namespace)
	assert.Equal(t, "test-relation", cfg.Relation)
	assert.NotNil(t, cfg.ObjectFunc)
}
